package dotprompt

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aymerick/raymond"
)

// DotpromptOptions represents options for creating a new Dotprompt instance
type DotpromptOptions struct {
	DefaultModel    string                 `json:"defaultModel,omitempty"`
	ModelConfigs    map[string]any         `json:"modelConfigs,omitempty"`
	Helpers        map[string]interface{}  `json:"helpers,omitempty"`
	Partials       map[string]string       `json:"partials,omitempty"`
	Tools          map[string]ToolDefinition  `json:"tools,omitempty"`
	ToolResolver   ToolResolver           `json:"toolResolver,omitempty"`
	Schemas        map[string]*JSONSchema  `json:"schemas,omitempty"`
	SchemaResolver SchemaResolver         `json:"schemaResolver,omitempty"`
	PartialResolver PartialResolver       `json:"partialResolver,omitempty"`
	Store          PromptStore            `json:"store,omitempty"`
}

// Dotprompt represents the main environment for handling prompts
type Dotprompt struct {
	handlebars     *raymond.Template
	knownHelpers   map[string]bool
	defaultModel   string
	modelConfigs   map[string]any
	tools          map[string]ToolDefinition
	toolResolver   ToolResolver
	schemas        map[string]*JSONSchema
	schemaResolver SchemaResolver
	partialResolver PartialResolver
	store          PromptStore
	mu             sync.RWMutex
}

// New creates a new Dotprompt instance
func New(options *DotpromptOptions) *Dotprompt {
	d := &Dotprompt{
		knownHelpers:   make(map[string]bool),
		modelConfigs:   make(map[string]any),
		tools:          make(map[string]ToolDefinition),
		schemas:        make(map[string]*JSONSchema),
	}

	if options != nil {
		d.modelConfigs = options.ModelConfigs
		d.defaultModel = options.DefaultModel
		d.tools = options.Tools
		d.toolResolver = options.ToolResolver
		d.schemas = options.Schemas
		d.schemaResolver = options.SchemaResolver
		d.partialResolver = options.PartialResolver
		d.store = options.Store

		// Register helpers
		for name, helper := range options.Helpers {
			d.DefineHelper(name, helper)
		}

		// Register partials
		for name, source := range options.Partials {
			d.DefinePartial(name, source)
		}
	}

	// Register default helpers
	d.registerDefaultHelpers()

	return d
}

// DefineHelper registers a new helper function
func (d *Dotprompt) DefineHelper(name string, helper interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.knownHelpers[name] = true
	raymond.RegisterHelper(name, helper)
}

// DefinePartial registers a new partial template
func (d *Dotprompt) DefinePartial(name, source string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	raymond.RegisterPartial(name, source)
}

// DefineTool registers a new tool
func (d *Dotprompt) DefineTool(def ToolDefinition) *Dotprompt {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.tools[def.Name] = def
	return d
}

// Get returns a function that can load and execute a prompt
func (d *Dotprompt) Get(name string, options map[string]string) (PromptFunction, error) {
	if d.store == nil {
		return nil, fmt.Errorf("must supply a store option when initializing dotprompt to use Get()")
	}

	return func(data map[string]any, metadata *PromptMetadata) (*RenderedPrompt, error) {
		fn, err := d.Load(name, map[string]string{
			"variant": metadata.Variant,
			"version": metadata.Version,
		})
		if err != nil {
			return nil, err
		}
		return fn(data, metadata)
	}, nil
}

// Load loads a prompt from the store and compiles it
func (d *Dotprompt) Load(name string, options map[string]string) (PromptFunction, error) {
	if d.store == nil {
		return nil, fmt.Errorf("no store configured")
	}

	prompt, err := d.store.Load(name, options)
	if err != nil {
		return nil, err
	}

	return d.Compile(prompt.Source, &PromptMetadata{
		Name:    prompt.Name,
		Variant: prompt.Variant,
		Version: prompt.Version,
	})
}

// Compile compiles a prompt source into an executable function
func (d *Dotprompt) Compile(source string, metadata *PromptMetadata) (PromptFunction, error) {
	// Parse the document
	parsed, err := ParseDocument(source)
	if err != nil {
		return nil, err
	}

	// Merge metadata
	if metadata != nil {
		// Merge metadata logic here
	}

	// Compile the template
	tmpl, err := raymond.Parse(source)
	if err != nil {
		return nil, err
	}

	return func(data map[string]any, options *PromptMetadata) (*RenderedPrompt, error) {
		// Execute the template
		rendered, err := tmpl.Exec(data)
		if err != nil {
			return nil, err
		}

		// Convert to messages
		messages, err := ToMessages(rendered, data)
		if err != nil {
			return nil, err
		}

		return &RenderedPrompt{
			Messages: messages,
			Metadata: *metadata,
			Document: parsed.Document,
		}, nil
	}, nil
}

// registerDefaultHelpers registers the built-in helper functions
func (d *Dotprompt) registerDefaultHelpers() {
	d.DefineHelper("section", func(name string, options *raymond.Options) string {
		return fmt.Sprintf("<<<dotprompt:section %s>>>", name)
	})

	d.DefineHelper("media", func(options *raymond.Options) string {
		url := options.HashStr("url")
		contentType := options.HashStr("contentType")
		if contentType != "" {
			return fmt.Sprintf("<<<dotprompt:media:url %s %s>>>", url, contentType)
		}
		return fmt.Sprintf("<<<dotprompt:media:url %s>>>", url)
	})

	d.DefineHelper("role", func(role string, options *raymond.Options) string {
		return fmt.Sprintf("<<<dotprompt:role %s>>>", role)
	})

	d.DefineHelper("history", func(options *raymond.Options) string {
		return "<<<dotprompt:history>>>"
	})
}

// Render renders a prompt with the given data
func (d *Dotprompt) Render(source string, data map[string]any, metadata *PromptMetadata) (*RenderedPrompt, error) {
	fn, err := d.Compile(source, metadata)
	if err != nil {
		return nil, err
	}
	return fn(data, metadata)
}
