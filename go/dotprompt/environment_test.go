package dotprompt

import (
	"testing"

	"github.com/aymerick/raymond"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDotprompt(t *testing.T) {
	dp := New(&DotpromptOptions{
		DefaultModel: "gpt-4",
	})

	t.Run("define and use helper", func(t *testing.T) {
		dp.DefineHelper("greet", func(name string, options *raymond.Options) string {
			return "Hello, " + name + "!"
		})

		fn, err := dp.Compile("{{greet name}}", nil)
		require.NoError(t, err)

		result, err := fn(map[string]any{
			"name": "Alice",
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, "Hello, Alice!", result.Messages[0].Content)
	})

	t.Run("define and use partial", func(t *testing.T) {
		dp.DefinePartial("footer", "Best regards, {{name}}")

		fn, err := dp.Compile("Dear {{recipient}},\n\n{{> footer}}", nil)
		require.NoError(t, err)

		result, err := fn(map[string]any{
			"recipient": "Alice",
			"name":     "Bob",
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, "Dear Alice,\n\nBest regards, Bob", result.Messages[0].Content)
	})

	t.Run("use built-in helpers", func(t *testing.T) {
		fn, err := dp.Compile(`{{role "system"}}
I am an AI assistant.

{{role "user"}}
Hello!

{{role "assistant"}}
Hi there! How can I help you today?`, nil)
		require.NoError(t, err)

		result, err := fn(nil, nil)
		require.NoError(t, err)
		assert.Len(t, result.Messages, 3)
		assert.Equal(t, "system", result.Messages[0].Role)
		assert.Equal(t, "user", result.Messages[1].Role)
		assert.Equal(t, "assistant", result.Messages[2].Role)
	})
}

func TestDotpromptStore(t *testing.T) {
	store := &mockStore{
		prompts: map[string]string{
			"greeting": `{{role "system"}}
I am a friendly AI assistant.

{{role "user"}}
Hello {{name}}!

{{role "assistant"}}
Hi {{name}}! How can I help you today?`,
		},
	}

	dp := New(&DotpromptOptions{
		DefaultModel: "gpt-4",
		Store:       store,
	})

	t.Run("load and render prompt", func(t *testing.T) {
		fn, err := dp.Get("greeting", nil)
		require.NoError(t, err)

		result, err := fn(map[string]any{
			"name": "Alice",
		}, nil)
		require.NoError(t, err)
		assert.Len(t, result.Messages, 3)
		assert.Equal(t, "Hello Alice!", result.Messages[1].Content)
		assert.Equal(t, "Hi Alice! How can I help you today?", result.Messages[2].Content)
	})
}

// mockStore is a simple in-memory store for testing
type mockStore struct {
	prompts map[string]string
}

func (s *mockStore) Load(name string, options map[string]string) (*PromptData, error) {
	if source, ok := s.prompts[name]; ok {
		return &PromptData{
			Name:   name,
			Source: source,
		}, nil
	}
	return nil, nil
}

func (s *mockStore) List(options map[string]interface{}) ([]PromptRef, string, error) {
	var refs []PromptRef
	for name := range s.prompts {
		refs = append(refs, PromptRef{
			Name: name,
		})
	}
	return refs, "", nil
}

func (s *mockStore) LoadPartial(name string, options map[string]string) (*PromptData, error) {
	return nil, nil
}

func (s *mockStore) ListPartials(options map[string]interface{}) ([]PartialRef, string, error) {
	return nil, "", nil
}

func (s *mockStore) Save(name string, source string, options map[string]string) error {
	s.prompts[name] = source
	return nil
}

func (s *mockStore) Delete(name string, options map[string]string) error {
	delete(s.prompts, name)
	return nil
}
