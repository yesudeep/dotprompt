package dotprompt

// Part represents a section of a prompt document
type Part interface {
	isPartType()
}

// TextPart represents a text section
type TextPart struct {
	Text string `json:"text"`
}

func (TextPart) isPartType() {}

// MediaPart represents a media section
type MediaPart struct {
	Media struct {
		URL         string `json:"url"`
		ContentType string `json:"contentType,omitempty"`
	} `json:"media"`
}

func (MediaPart) isPartType() {}

// MetadataPart represents metadata in the prompt
type MetadataPart struct {
	Metadata struct {
		Purpose string          `json:"purpose"`
		Pending bool           `json:"pending"`
		Data    map[string]any `json:"data,omitempty"`
	} `json:"metadata"`
}

func (MetadataPart) isPartType() {}

// Document represents a complete prompt document
type Document struct {
	Content []Part `json:"content"`
}

// PromptData represents a complete prompt with metadata
type PromptData struct {
	Name        string                 `json:"name,omitempty"`
	Variant     string                 `json:"variant,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Source      string                 `json:"source"`
	Description string                 `json:"description,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Tools       []string               `json:"tools,omitempty"`
	ToolDefs    []ToolDefinition       `json:"toolDefs,omitempty"`
	Config      map[string]any         `json:"config,omitempty"`
	Input       *InputConfig           `json:"input,omitempty"`
	Output      *OutputConfig          `json:"output,omitempty"`
	Default     map[string]any         `json:"default,omitempty"`
	Schema      *JSONSchema            `json:"schema,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// InputConfig defines input configuration for a prompt
type InputConfig struct {
	Default map[string]any `json:"default,omitempty"`
	Schema  *JSONSchema   `json:"schema,omitempty"`
}

// OutputConfig defines output configuration for a prompt
type OutputConfig struct {
	Format string      `json:"format,omitempty"`
	Schema *JSONSchema `json:"schema,omitempty"`
}

// JSONSchema represents a JSON schema definition
type JSONSchema struct {
	Type                 string                 `json:"type,omitempty"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`
	Required            []string               `json:"required,omitempty"`
	AdditionalProperties bool                  `json:"additionalProperties,omitempty"`
	Items               *JSONSchema            `json:"items,omitempty"`
	Description         string                 `json:"description,omitempty"`
}

// ToolDefinition represents a tool that can be used in prompts
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Schema      *JSONSchema `json:"schema,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string         `json:"role"`
	Content string         `json:"content"`
	Data    map[string]any `json:"data,omitempty"`
}

// PromptRef represents a reference to a prompt
type PromptRef struct {
	Name    string `json:"name"`
	Variant string `json:"variant,omitempty"`
	Version string `json:"version,omitempty"`
}

// PartialRef represents a reference to a partial
type PartialRef struct {
	Name    string `json:"name"`
	Variant string `json:"variant,omitempty"`
	Version string `json:"version,omitempty"`
}

// PromptFunction represents a compiled prompt that can be executed
type PromptFunction func(data map[string]any, options *PromptMetadata) (*RenderedPrompt, error)

// PromptMetadata represents metadata for a prompt
type PromptMetadata struct {
	Name        string                 `json:"name,omitempty"`
	Variant     string                 `json:"variant,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Description string                 `json:"description,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Tools       []string               `json:"tools,omitempty"`
	ToolDefs    []ToolDefinition       `json:"toolDefs,omitempty"`
	Config      map[string]any         `json:"config,omitempty"`
	Input       *InputConfig           `json:"input,omitempty"`
	Output      *OutputConfig          `json:"output,omitempty"`
	Default     map[string]any         `json:"default,omitempty"`
	Schema      *JSONSchema            `json:"schema,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RenderedPrompt represents a rendered prompt with its metadata
type RenderedPrompt struct {
	Messages  []Message      `json:"messages"`
	Metadata  PromptMetadata `json:"metadata"`
	Document  Document       `json:"document"`
	Functions []Message      `json:"functions,omitempty"`
}

// ParsedPrompt represents a parsed prompt with its metadata
type ParsedPrompt struct {
	Document Document      `json:"document"`
	Metadata PromptMetadata `json:"metadata"`
}

// PromptStore defines the interface for prompt storage
type PromptStore interface {
	Load(name string, options map[string]string) (*PromptData, error)
	LoadPartial(name string, options map[string]string) (*PromptData, error)
	List(options map[string]interface{}) ([]PromptRef, string, error)
	ListPartials(options map[string]interface{}) ([]PartialRef, string, error)
}

// PromptStoreWritable extends PromptStore with write operations
type PromptStoreWritable interface {
	PromptStore
	Save(prompt *PromptData) error
	Delete(name string, options map[string]string) error
}

// ToolResolver is a function type that resolves tool references
type ToolResolver func(name string) (*ToolDefinition, error)

// SchemaResolver is a function type that resolves schema references
type SchemaResolver func(ref string) (*JSONSchema, error)

// PartialResolver is a function type that resolves partial references
type PartialResolver func(name string) (string, error)

// DataArgument represents the data passed to a prompt
type DataArgument map[string]any
