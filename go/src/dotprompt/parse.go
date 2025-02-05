package dotprompt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	partRegex = regexp.MustCompile(`<<<dotprompt:([^>]+)>>>`)
)

// ParseDocument parses a prompt source into a document
func ParseDocument(source string) (*ParsedPrompt, error) {
	// Extract metadata from frontmatter if present
	metadata := &PromptMetadata{}
	
	// Parse the document content into parts
	parts, err := ToParts(source)
	if err != nil {
		return nil, err
	}

	return &ParsedPrompt{
		Document: Document{Content: parts},
		Metadata: *metadata,
	}, nil
}

// ToParts converts a source string into a slice of parts
func ToParts(source string) ([]Part, error) {
	var parts []Part
	pieces := partRegex.Split(source, -1)
	matches := partRegex.FindAllStringSubmatch(source, -1)

	var matchIndex int
	for i, piece := range pieces {
		if strings.TrimSpace(piece) != "" {
			parts = append(parts, TextPart{Text: piece})
		}

		if i < len(matches) {
			directive := matches[matchIndex][1]
			matchIndex++

			if strings.HasPrefix(directive, "media:url") {
				fields := strings.Fields(directive)
				if len(fields) < 2 {
					return nil, fmt.Errorf("invalid media directive: %s", directive)
				}

				part := MediaPart{}
				part.Media.URL = fields[1]
				if len(fields) > 2 {
					part.Media.ContentType = fields[2]
				}
				parts = append(parts, part)
			} else if strings.HasPrefix(directive, "section") {
				fields := strings.Fields(directive)
				if len(fields) < 2 {
					return nil, fmt.Errorf("invalid section directive: %s", directive)
				}

				part := MetadataPart{}
				part.Metadata.Purpose = fields[1]
				part.Metadata.Pending = true
				parts = append(parts, part)
			}
		}
	}

	return parts, nil
}

// ToMessages converts a rendered string into a slice of messages
func ToMessages(renderedString string, data map[string]any) ([]Message, error) {
	var messages []Message
	var currentRole string
	var currentContent strings.Builder

	// Split the rendered string into pieces
	pieces := partRegex.Split(renderedString, -1)
	matches := partRegex.FindAllStringSubmatch(renderedString, -1)

	var matchIndex int
	for i, piece := range pieces {
		if strings.TrimSpace(piece) != "" {
			currentContent.WriteString(piece)
		}

		if i < len(matches) {
			directive := matches[matchIndex][1]
			matchIndex++

			if strings.HasPrefix(directive, "role") {
				// If we have content from a previous role, add it as a message
				if currentContent.Len() > 0 && currentRole != "" {
					messages = append(messages, Message{
						Role:    currentRole,
						Content: strings.TrimSpace(currentContent.String()),
					})
					currentContent.Reset()
				}

				// Set the new role
				fields := strings.Fields(directive)
				if len(fields) > 1 {
					currentRole = fields[1]
				}
			} else if directive == "history" {
				// Handle history insertion
				if history, ok := data["history"].([]Message); ok {
					messages = append(messages, history...)
				}
			}
		}
	}

	// Add the final message if there's content
	if currentContent.Len() > 0 && currentRole != "" {
		messages = append(messages, Message{
			Role:    currentRole,
			Content: strings.TrimSpace(currentContent.String()),
		})
	}

	return messages, nil
}

// InsertHistory inserts history messages into the message stream
func InsertHistory(messages []Message, history []Message) []Message {
	if len(history) == 0 {
		return messages
	}

	var result []Message
	for _, msg := range messages {
		if msg.Role == "history" {
			result = append(result, history...)
		} else {
			result = append(result, msg)
		}
	}
	return result
}
