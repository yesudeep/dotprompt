# Dotprompt Go Implementation

This is the Go implementation of dotprompt, providing equivalent functionality to the JavaScript version.

## Features

- Handlebars templating using [raymond](https://github.com/aymerick/raymond)
- File-based prompt storage
- Support for prompt variants and partials
- Message history management
- Media and section support
- Tool integration

## Installation

```bash
go get github.com/google/dotprompt/go
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/google/dotprompt/go/src/dotprompt"
    "github.com/google/dotprompt/go/src/dotprompt/stores"
)

func main() {
    // Create a new store
    store := stores.NewDirStore(stores.DirStoreOptions{
        Directory: "./prompts",
    })

    // Create a new dotprompt instance
    dp := dotprompt.New(&dotprompt.DotpromptOptions{
        DefaultModel: "gpt-4",
        Store: store,
    })

    // Load and render a prompt
    fn, err := dp.Get("example", nil)
    if err != nil {
        panic(err)
    }

    result, err := fn(map[string]any{
        "name": "World",
    }, nil)
    if err != nil {
        panic(err)
    }

    // Print the messages
    for _, msg := range result.Messages {
        fmt.Printf("%s: %s\n", msg.Role, msg.Content)
    }
}
```

## Example Prompt

```handlebars
{{role "system"}}
You are a helpful assistant.

{{role "user"}}
Hello {{name}}!
```

## Testing

```bash
go test ./...
```

## License

Apache 2.0 - See LICENSE file for details.
