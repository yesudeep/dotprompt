// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"testing"

	"github.com/aymerick/raymond"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperNames is a list of all helper names that we register
var helperNames = []string{
	"section",
	"media",
	"role",
	"history",
	"eq",
	"not",
	"and",
	"or",
	"json",
}

func resetHelpers() {
	// Unregister all our helpers
	for _, name := range helperNames {
		raymond.RegisterHelper(name, nil)
	}
}

func TestRegisterDefaultHelpers(t *testing.T) {
	resetHelpers()
	RegisterDefaultHelpers()

	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "section helper",
			template: `{{section "test"}}`,
			data:     nil,
			expected: "<<<dotprompt:section test>>>",
		},
		{
			name:     "media helper with url only",
			template: `{{media url="test.jpg"}}`,
			data:     nil,
			expected: "<<<dotprompt:media:url test.jpg>>>",
		},
		{
			name:     "media helper with content type",
			template: `{{media url="test.jpg" contentType="image/jpeg"}}`,
			data:     nil,
			expected: "<<<dotprompt:media:url test.jpg image/jpeg>>>",
		},
		{
			name:     "role helper",
			template: `{{role "system"}}`,
			data:     nil,
			expected: "<<<dotprompt:role system>>>",
		},
		{
			name:     "history helper",
			template: `{{history}}`,
			data:     nil,
			expected: "<<<dotprompt:history>>>",
		},
		{
			name:     "eq helper true",
			template: `{{#if (eq value 42)}}true{{else}}false{{/if}}`,
			data:     map[string]interface{}{"value": 42},
			expected: "true",
		},
		{
			name:     "eq helper false",
			template: `{{#if (eq value 42)}}true{{else}}false{{/if}}`,
			data:     map[string]interface{}{"value": 41},
			expected: "false",
		},
		{
			name:     "not helper",
			template: `{{#if (not value)}}true{{else}}false{{/if}}`,
			data:     map[string]interface{}{"value": false},
			expected: "true",
		},
		{
			name:     "and helper true",
			template: `{{#if (and true true)}}true{{else}}false{{/if}}`,
			data:     nil,
			expected: "true",
		},
		{
			name:     "and helper false",
			template: `{{#if (and true false)}}true{{else}}false{{/if}}`,
			data:     nil,
			expected: "false",
		},
		{
			name:     "or helper true",
			template: `{{#if (or false true)}}true{{else}}false{{/if}}`,
			data:     nil,
			expected: "true",
		},
		{
			name:     "or helper false",
			template: `{{#if (or false false)}}true{{else}}false{{/if}}`,
			data:     nil,
			expected: "false",
		},
		{
			name:     "json helper with string",
			template: `{{json value}}`,
			data:     map[string]interface{}{"value": "test"},
			expected: `"test"`,
		},
		{
			name:     "json helper with number",
			template: `{{json value}}`,
			data:     map[string]interface{}{"value": 42},
			expected: "42",
		},
		{
			name:     "json helper with null",
			template: `{{json value}}`,
			data:     map[string]interface{}{"value": nil},
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := raymond.Parse(tt.template)
			require.NoError(t, err)

			result, err := tmpl.Exec(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHelperCombinations(t *testing.T) {
	resetHelpers()
	RegisterDefaultHelpers()

	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name: "combine role and section",
			template: `{{role "system"}}
{{section "instructions"}}
Hello`,
			data: nil,
			expected: `<<<dotprompt:role system>>>
<<<dotprompt:section instructions>>>
Hello`,
		},
		{
			name: "combine media and role",
			template: `{{role "user"}}
{{media url="test.jpg" contentType="image/jpeg"}}
Check this image`,
			data: nil,
			expected: `<<<dotprompt:role user>>>
<<<dotprompt:media:url test.jpg image/jpeg>>>
Check this image`,
		},
		{
			name: "logical operators",
			template: `{{#if (and (eq value 42) (not isDisabled))}}
{{role "assistant"}}Yes
{{else}}
{{role "assistant"}}No
{{/if}}`,
			data: map[string]interface{}{
				"value":      42,
				"isDisabled": false,
			},
			expected: `<<<dotprompt:role assistant>>>Yes
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := raymond.Parse(tt.template)
			require.NoError(t, err)

			result, err := tmpl.Exec(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
