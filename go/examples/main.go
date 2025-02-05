package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/dotprompt/go/src/dotprompt"
	"github.com/google/dotprompt/go/src/dotprompt/stores"
)

func main() {
	// Create prompts directory
	promptsDir := filepath.Join(".", "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Create example prompt file
	promptContent := `{{role "system"}}
You are a helpful assistant that provides concise responses.

{{role "user"}}
Hello {{name}}! How are you?

{{> footer}}`

	if err := os.WriteFile(filepath.Join(promptsDir, "greeting.prompt"), []byte(promptContent), 0644); err != nil {
		log.Fatal(err)
	}

	// Create footer partial
	footerContent := `{{role "assistant"}}
I'm doing great, thank you for asking! How can I help you today?`

	if err := os.WriteFile(filepath.Join(promptsDir, "_footer.prompt"), []byte(footerContent), 0644); err != nil {
		log.Fatal(err)
	}

	// Initialize dotprompt
	store := stores.NewDirStore(stores.DirStoreOptions{
		Directory: promptsDir,
	})

	dp := dotprompt.New(&dotprompt.DotpromptOptions{
		DefaultModel: "gpt-4",
		Store:       store,
	})

	// Load and render the prompt
	fn, err := dp.Get("greeting", nil)
	if err != nil {
		log.Fatal(err)
	}

	result, err := fn(map[string]any{
		"name": "Alice",
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
