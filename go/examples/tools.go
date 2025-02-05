package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/dotprompt/go/src/dotprompt"
	"github.com/google/dotprompt/go/src/dotprompt/stores"
)

// Calculator represents a simple tool for arithmetic operations
type Calculator struct{}

func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

func (c *Calculator) Multiply(a, b float64) float64 {
	return a * b
}

func main() {
	// Create prompts directory
	promptsDir := filepath.Join(".", "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Create example prompt with tool usage
	promptContent := `{{role "system"}}
You are a helpful math assistant that can perform calculations.

{{role "user"}}
I need help with some math. Can you add {{a}} and {{b}}, then multiply the result by {{c}}?

{{role "assistant"}}
Let me help you with that calculation.

First, let's add {{a}} and {{b}}:
{{#with (call "calculator.Add" a b)}}
{{a}} + {{b}} = {{this}}
{{/with}}

Now, let's multiply that result by {{c}}:
{{#with (call "calculator.Multiply" (call "calculator.Add" a b) c)}}
The final result is: {{this}}
{{/with}}`

	if err := os.WriteFile(filepath.Join(promptsDir, "calculator.prompt"), []byte(promptContent), 0644); err != nil {
		log.Fatal(err)
	}

	// Initialize dotprompt with calculator tool
	store := stores.NewDirStore(stores.DirStoreOptions{
		Directory: promptsDir,
	})

	calc := &Calculator{}

	dp := dotprompt.New(&dotprompt.DotpromptOptions{
		DefaultModel: "gpt-4",
		Store:       store,
		Tools: map[string]dotprompt.ToolDefinition{
			"calculator.Add": {
				Name:        "calculator.Add",
				Description: "Adds two numbers",
				Schema: &dotprompt.JSONSchema{
					Type: "object",
					Properties: map[string]*dotprompt.JSONSchema{
						"a": {Type: "number"},
						"b": {Type: "number"},
					},
					Required: []string{"a", "b"},
				},
			},
			"calculator.Multiply": {
				Name:        "calculator.Multiply",
				Description: "Multiplies two numbers",
				Schema: &dotprompt.JSONSchema{
					Type: "object",
					Properties: map[string]*dotprompt.JSONSchema{
						"a": {Type: "number"},
						"b": {Type: "number"},
					},
					Required: []string{"a", "b"},
				},
			},
		},
	})

	// Register tool methods
	dp.DefineHelper("call", func(name string, args ...interface{}) interface{} {
		switch name {
		case "calculator.Add":
			if len(args) >= 2 {
				if a, ok := args[0].(float64); ok {
					if b, ok := args[1].(float64); ok {
						return calc.Add(a, b)
					}
				}
			}
		case "calculator.Multiply":
			if len(args) >= 2 {
				if a, ok := args[0].(float64); ok {
					if b, ok := args[1].(float64); ok {
						return calc.Multiply(a, b)
					}
				}
			}
		}
		return nil
	})

	// Load and render the prompt
	fn, err := dp.Get("calculator", nil)
	if err != nil {
		log.Fatal(err)
	}

	result, err := fn(map[string]any{
		"a": 5.0,
		"b": 3.0,
		"c": 2.0,
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Print the messages
	fmt.Println("Messages:")
	for _, msg := range result.Messages {
		fmt.Printf("%s: %s\n", msg.Role, msg.Content)
	}
}
