package dotprompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToParts(t *testing.T) {
	t.Run("parses text only", func(t *testing.T) {
		source := "Simple text"
		parts, err := ToParts(source)
		require.NoError(t, err)
		assert.Len(t, parts, 1)

		textPart, ok := parts[0].(TextPart)
		require.True(t, ok)
		assert.Equal(t, "Simple text", textPart.Text)
	})

	t.Run("parses media", func(t *testing.T) {
		source := "Before<<<dotprompt:media:url test.jpg image/jpeg>>>After"
		parts, err := ToParts(source)
		require.NoError(t, err)
		assert.Len(t, parts, 3)

		textPart1, ok := parts[0].(TextPart)
		require.True(t, ok)
		assert.Equal(t, "Before", textPart1.Text)

		mediaPart, ok := parts[1].(MediaPart)
		require.True(t, ok)
		assert.Equal(t, "test.jpg", mediaPart.Media.URL)
		assert.Equal(t, "image/jpeg", mediaPart.Media.ContentType)

		textPart2, ok := parts[2].(TextPart)
		require.True(t, ok)
		assert.Equal(t, "After", textPart2.Text)
	})

	t.Run("parses section", func(t *testing.T) {
		source := "Start<<<dotprompt:section test>>>End"
		parts, err := ToParts(source)
		require.NoError(t, err)
		assert.Len(t, parts, 3)

		textPart1, ok := parts[0].(TextPart)
		require.True(t, ok)
		assert.Equal(t, "Start", textPart1.Text)

		metaPart, ok := parts[1].(MetadataPart)
		require.True(t, ok)
		assert.Equal(t, "test", metaPart.Metadata.Purpose)
		assert.True(t, metaPart.Metadata.Pending)

		textPart2, ok := parts[2].(TextPart)
		require.True(t, ok)
		assert.Equal(t, "End", textPart2.Text)
	})

	t.Run("handles invalid media directive", func(t *testing.T) {
		source := "<<<dotprompt:media:url>>>"
		_, err := ToParts(source)
		assert.Error(t, err)
	})

	t.Run("handles invalid section directive", func(t *testing.T) {
		source := "<<<dotprompt:section>>>"
		_, err := ToParts(source)
		assert.Error(t, err)
	})
}

func TestToMessages(t *testing.T) {
	t.Run("converts single role", func(t *testing.T) {
		source := "<<<dotprompt:role system>>>Hello"
		messages, err := ToMessages(source, nil)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, "system", messages[0].Role)
		assert.Equal(t, "Hello", messages[0].Content)
	})

	t.Run("converts multiple roles", func(t *testing.T) {
		source := "<<<dotprompt:role system>>>Hello\n<<<dotprompt:role user>>>Hi"
		messages, err := ToMessages(source, nil)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, "system", messages[0].Role)
		assert.Equal(t, "Hello", messages[0].Content)
		assert.Equal(t, "user", messages[1].Role)
		assert.Equal(t, "Hi", messages[1].Content)
	})

	t.Run("handles history", func(t *testing.T) {
		source := "<<<dotprompt:history>>><<<dotprompt:role user>>>Hello"
		data := map[string]any{
			"history": []Message{
				{Role: "system", Content: "Previous"},
			},
		}

		messages, err := ToMessages(source, data)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, "system", messages[0].Role)
		assert.Equal(t, "Previous", messages[0].Content)
		assert.Equal(t, "user", messages[1].Role)
		assert.Equal(t, "Hello", messages[1].Content)
	})

	t.Run("ignores empty content", func(t *testing.T) {
		source := "<<<dotprompt:role system>>>\n\n<<<dotprompt:role user>>>Hello"
		messages, err := ToMessages(source, nil)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, "Hello", messages[0].Content)
	})
}

func TestInsertHistory(t *testing.T) {
	t.Run("inserts history at marker", func(t *testing.T) {
		messages := []Message{
			{Role: "system", Content: "Start"},
			{Role: "history"},
			{Role: "user", Content: "End"},
		}
		history := []Message{
			{Role: "user", Content: "Previous1"},
			{Role: "assistant", Content: "Previous2"},
		}

		result := InsertHistory(messages, history)
		assert.Len(t, result, 4)
		assert.Equal(t, "Start", result[0].Content)
		assert.Equal(t, "Previous1", result[1].Content)
		assert.Equal(t, "Previous2", result[2].Content)
		assert.Equal(t, "End", result[3].Content)
	})

	t.Run("handles empty history", func(t *testing.T) {
		messages := []Message{
			{Role: "system", Content: "Start"},
			{Role: "history"},
			{Role: "user", Content: "End"},
		}

		result := InsertHistory(messages, nil)
		assert.Equal(t, messages, result)
	})
}

func TestParseDocument(t *testing.T) {
	t.Run("parses complete document", func(t *testing.T) {
		source := `<<<dotprompt:role system>>>
Instructions
<<<dotprompt:section test>>>
<<<dotprompt:media:url image.jpg image/jpeg>>>
<<<dotprompt:role user>>>
Query`

		doc, err := ParseDocument(source)
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.NotNil(t, doc.Document)
		assert.NotEmpty(t, doc.Document.Content)
	})
}
