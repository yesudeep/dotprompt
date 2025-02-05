package dotprompt

import (
	"testing"

	"github.com/aymerick/raymond"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates instance with default options", func(t *testing.T) {
		d := New(nil)
		assert.NotNil(t, d)
		assert.NotNil(t, d.handlebars)
		assert.NotNil(t, d.knownHelpers)
		assert.NotNil(t, d.modelConfigs)
		assert.NotNil(t, d.tools)
		assert.NotNil(t, d.schemas)
	})

	t.Run("creates instance with custom options", func(t *testing.T) {
		options := &DotpromptOptions{
			DefaultModel: "gpt-4",
			ModelConfigs: map[string]any{
				"gpt-4": map[string]any{"temperature": 0.7},
			},
			Helpers: map[string]raymond.Helper{
				"custom": func(options *raymond.Options) string {
					return "custom"
				},
			},
			Partials: map[string]string{
				"header": "Header content",
			},
		}

		d := New(options)
		assert.NotNil(t, d)
		assert.Equal(t, "gpt-4", d.defaultModel)
		assert.Equal(t, options.ModelConfigs, d.modelConfigs)
	})
}

func TestDefineHelper(t *testing.T) {
	d := New(nil)
	helperCalled := false

	d.DefineHelper("test", func(options *raymond.Options) string {
		helperCalled = true
		return "test"
	})

	// Compile and render a template using the helper
	tmpl, err := d.handlebars.Parse("{{test}}")
	require.NoError(t, err)

	result, err := tmpl.Exec(nil)
	require.NoError(t, err)
	assert.Equal(t, "test", result)
	assert.True(t, helperCalled)
}

func TestDefinePartial(t *testing.T) {
	d := New(nil)
	d.DefinePartial("header", "Header: {{title}}")

	// Compile and render a template using the partial
	tmpl, err := d.handlebars.Parse("{{> header}}")
	require.NoError(t, err)

	result, err := tmpl.Exec(map[string]any{"title": "Test"})
	require.NoError(t, err)
	assert.Equal(t, "Header: Test", result)
}

func TestDefineTool(t *testing.T) {
	d := New(nil)
	tool := ToolDefinition{
		Name:        "test",
		Description: "Test tool",
	}

	d.DefineTool(tool)
	assert.Equal(t, tool, d.tools["test"])
}

func TestRender(t *testing.T) {
	d := New(nil)

	t.Run("renders simple template", func(t *testing.T) {
		source := "{{role 'system'}}Hello {{name}}"
		data := map[string]any{"name": "World"}

		result, err := d.Render(source, data, nil)
		require.NoError(t, err)
		assert.Len(t, result.Messages, 1)
		assert.Equal(t, "system", result.Messages[0].Role)
		assert.Equal(t, "Hello World", result.Messages[0].Content)
	})

	t.Run("renders template with history", func(t *testing.T) {
		source := "{{history}}{{role 'user'}}Hello"
		data := map[string]any{
			"history": []Message{
				{Role: "system", Content: "Previous message"},
			},
		}

		result, err := d.Render(source, data, nil)
		require.NoError(t, err)
		assert.Len(t, result.Messages, 2)
		assert.Equal(t, "system", result.Messages[0].Role)
		assert.Equal(t, "Previous message", result.Messages[0].Content)
		assert.Equal(t, "user", result.Messages[1].Role)
		assert.Equal(t, "Hello", result.Messages[1].Content)
	})

	t.Run("renders template with media", func(t *testing.T) {
		source := "{{media url='test.jpg' contentType='image/jpeg'}}{{role 'user'}}See image"
		data := map[string]any{}

		result, err := d.Render(source, data, nil)
		require.NoError(t, err)
		assert.Len(t, result.Document.Content, 2)

		mediaPart, ok := result.Document.Content[0].(MediaPart)
		require.True(t, ok)
		assert.Equal(t, "test.jpg", mediaPart.Media.URL)
		assert.Equal(t, "image/jpeg", mediaPart.Media.ContentType)

		assert.Len(t, result.Messages, 1)
		assert.Equal(t, "user", result.Messages[0].Role)
		assert.Equal(t, "See image", result.Messages[0].Content)
	})
}

func TestCompile(t *testing.T) {
	d := New(nil)

	t.Run("compiles valid template", func(t *testing.T) {
		source := "{{role 'system'}}Template"
		fn, err := d.Compile(source, nil)
		require.NoError(t, err)
		assert.NotNil(t, fn)

		result, err := fn(nil, nil)
		require.NoError(t, err)
		assert.Len(t, result.Messages, 1)
		assert.Equal(t, "system", result.Messages[0].Role)
		assert.Equal(t, "Template", result.Messages[0].Content)
	})

	t.Run("fails on invalid template", func(t *testing.T) {
		source := "{{role 'system'}}{{invalid}}"
		_, err := d.Compile(source, nil)
		assert.Error(t, err)
	})
}

func TestGet(t *testing.T) {
	t.Run("fails without store", func(t *testing.T) {
		d := New(nil)
		_, err := d.Get("test", nil)
		assert.Error(t, err)
	})

	// Add more tests with a mock store
}
