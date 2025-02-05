package helpers

import (
	"fmt"
	"html/template"

	"github.com/aymerick/raymond"
)

// RegisterDefaultHelpers registers the default set of helpers
func RegisterDefaultHelpers() {
	raymond.RegisterHelper("section", func(name string, options *raymond.Options) template.HTML {
		return template.HTML(fmt.Sprintf("<<<dotprompt:section %s>>>", name))
	})

	raymond.RegisterHelper("media", func(options *raymond.Options) template.HTML {
		url := options.HashStr("url")
		contentType := options.HashStr("contentType")
		if contentType != "" {
			return template.HTML(fmt.Sprintf("<<<dotprompt:media:url %s %s>>>", url, contentType))
		}
		return template.HTML(fmt.Sprintf("<<<dotprompt:media:url %s>>>", url))
	})

	raymond.RegisterHelper("role", func(role string, options *raymond.Options) template.HTML {
		return template.HTML(fmt.Sprintf("<<<dotprompt:role %s>>>", role))
	})

	raymond.RegisterHelper("history", func(options *raymond.Options) template.HTML {
		return template.HTML("<<<dotprompt:history>>>")
	})

	raymond.RegisterHelper("eq", func(a, b interface{}, options *raymond.Options) bool {
		return a == b
	})

	raymond.RegisterHelper("not", func(value interface{}, options *raymond.Options) bool {
		if b, ok := value.(bool); ok {
			return !b
		}
		return false
	})

	raymond.RegisterHelper("and", func(args ...interface{}) bool {
		if len(args) <= 1 { // Last arg is *raymond.Options
			return false
		}
		for _, arg := range args[:len(args)-1] {
			if b, ok := arg.(bool); !ok || !b {
				return false
			}
		}
		return true
	})

	raymond.RegisterHelper("or", func(args ...interface{}) bool {
		if len(args) <= 1 { // Last arg is *raymond.Options
			return false
		}
		for _, arg := range args[:len(args)-1] {
			if b, ok := arg.(bool); ok && b {
				return true
			}
		}
		return false
	})

	raymond.RegisterHelper("json", func(value interface{}, options *raymond.Options) template.HTML {
		if value == nil {
			return template.HTML("null")
		}
		if s, ok := value.(string); ok {
			return template.HTML(fmt.Sprintf("%q", s))
		}
		return template.HTML(fmt.Sprintf("%v", value))
	})
}
