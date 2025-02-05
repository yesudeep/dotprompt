// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dotprompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// SpecTest represents a single test case within a spec suite
type SpecTest struct {
	Desc    string                 `yaml:"desc"`
	Data    map[string]any         `yaml:"data"`
	Options map[string]interface{} `yaml:"options"`
	Expect  map[string]interface{} `yaml:"expect"`
}

// SpecSuite represents a complete test suite from a YAML file
type SpecSuite struct {
	Name             string                    `yaml:"name"`
	Template         string                    `yaml:"template"`
	Data             map[string]any            `yaml:"data"`
	Schemas          map[string]*JSONSchema    `yaml:"schemas"`
	Tools            map[string]ToolDefinition `yaml:"tools"`
	Partials         map[string]string         `yaml:"partials"`
	ResolverPartials map[string]string         `yaml:"resolverPartials"`
	Tests            []SpecTest                `yaml:"tests"`
}

func TestSpecs(t *testing.T) {
	// Find all YAML files in the spec directory
	specDir := filepath.Join("..", "..", "spec")
	entries, err := os.ReadDir(specDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		// Read and parse the YAML file
		filePath := filepath.Join(specDir, entry.Name())
		yamlData, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var suites []SpecSuite
		err = yaml.Unmarshal(yamlData, &suites)
		require.NoError(t, err)

		// Run each suite
		for _, suite := range suites {
			t.Run(filepath.Join(entry.Name(), suite.Name), func(t *testing.T) {
				env := New(&DotpromptOptions{
					DefaultModel: "gpt-4",
					Schemas:      suite.Schemas,
					Tools:        suite.Tools,
				})

				// Register partials if any
				if suite.Partials != nil {
					for name, template := range suite.Partials {
						env.DefinePartial(name, template)
					}
				}

				// Run each test in the suite
				for _, test := range suite.Tests {
					testName := test.Desc
					if testName == "" {
						testName = "unnamed test"
					}

					t.Run(testName, func(t *testing.T) {
						// Create metadata from test options
						var metadata *PromptMetadata
						if test.Options != nil {
							metadata = &PromptMetadata{}
							if input, ok := test.Options["input"].(map[string]interface{}); ok {
								metadata.Input = &InputConfig{
									Default: input["default"].(map[string]any),
								}
							}
						}

						// Compile and execute the template
						fn, err := env.Compile(suite.Template, metadata)
						require.NoError(t, err)

						result, err := fn(test.Data, metadata)
						require.NoError(t, err)

						// Compare with expected output
						if messages, ok := test.Expect["messages"].([]interface{}); ok {
							assert.Len(t, result.Messages, len(messages))
							for i, expected := range messages {
								expectedMsg := expected.(map[string]interface{})
								assert.Equal(t, expectedMsg["role"], result.Messages[i].Role)

								// Compare content
								if content, ok := expectedMsg["content"].([]interface{}); ok {
									for _, part := range content {
										partMap := part.(map[string]interface{})
										if text, ok := partMap["text"].(string); ok {
											assert.Equal(t, text, result.Messages[i].Content)
										}
									}
								}
							}
						}

						// Compare input defaults if specified
						if input, ok := test.Expect["input"].(map[string]interface{}); ok {
							if defaults, ok := input["default"].(map[string]interface{}); ok {
								assert.Equal(t, defaults, result.Metadata.Input.Default)
							}
						}
					})
				}
			})
		}
	}
}
