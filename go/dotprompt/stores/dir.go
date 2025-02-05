package stores

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/dotprompt/go/dotprompt"
)

// DirStoreOptions represents options for creating a directory-based store
type DirStoreOptions struct {
	Directory string
}

// DirStore implements a file system based prompt store
type DirStore struct {
	directory string
}

// NewDirStore creates a new directory-based store
func NewDirStore(options DirStoreOptions) *DirStore {
	return &DirStore{
		directory: options.Directory,
	}
}

// readPromptFile reads and returns the contents of a prompt file
func (d *DirStore) readPromptFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}
	return string(content), nil
}

// calculateVersion generates a version hash for the content
func (d *DirStore) calculateVersion(content string) string {
	hash := sha1.New()
	hash.Write([]byte(content))
	return hex.EncodeToString(hash.Sum(nil))[:8]
}

// parsePromptFilename extracts name and variant from a filename
func (d *DirStore) parsePromptFilename(filename string) (name string, variant string, err error) {
	match := strings.Split(strings.TrimSuffix(filename, ".prompt"), ".")
	if len(match) == 0 {
		return "", "", fmt.Errorf("invalid prompt filename: %s", filename)
	}

	name = match[0]
	if len(match) > 1 {
		variant = match[1]
	}
	return name, variant, nil
}

// isPartial checks if a filename represents a partial
func (d *DirStore) isPartial(filename string) bool {
	return strings.HasPrefix(filepath.Base(filename), "_")
}

// Load loads a prompt by name
func (d *DirStore) Load(name string, options map[string]string) (*dotprompt.PromptData, error) {
	variant := options["variant"]
	fileName := name
	if variant != "" {
		fileName = fmt.Sprintf("%s.%s", name, variant)
	}
	fileName += ".prompt"

	filePath := filepath.Join(d.directory, fileName)
	content, err := d.readPromptFile(filePath)
	if err != nil {
		return nil, err
	}

	version := d.calculateVersion(content)
	return &dotprompt.PromptData{
		Name:    name,
		Variant: variant,
		Version: version,
		Source:  content,
	}, nil
}

// LoadPartial loads a partial by name
func (d *DirStore) LoadPartial(name string, options map[string]string) (*dotprompt.PromptData, error) {
	variant := options["variant"]
	fileName := "_" + name
	if variant != "" {
		fileName = fmt.Sprintf("%s.%s", fileName, variant)
	}
	fileName += ".prompt"

	filePath := filepath.Join(d.directory, fileName)
	content, err := d.readPromptFile(filePath)
	if err != nil {
		return nil, err
	}

	version := d.calculateVersion(content)
	return &dotprompt.PromptData{
		Name:    name,
		Variant: variant,
		Version: version,
		Source:  content,
	}, nil
}

// scanDirectory recursively scans the directory for prompt files
func (d *DirStore) scanDirectory(dir string) ([]string, error) {
	var results []string
	err := filepath.WalkDir(filepath.Join(d.directory, dir), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".prompt") {
			relPath, err := filepath.Rel(d.directory, path)
			if err != nil {
				return err
			}
			results = append(results, relPath)
		}
		return nil
	})
	return results, err
}

// List returns a list of all prompts
func (d *DirStore) List(options map[string]interface{}) ([]dotprompt.PromptRef, string, error) {
	files, err := d.scanDirectory("")
	if err != nil {
		return nil, "", err
	}

	var prompts []dotprompt.PromptRef
	for _, file := range files {
		if !d.isPartial(file) {
			name, variant, err := d.parsePromptFilename(filepath.Base(file))
			if err != nil {
				continue
			}
			prompts = append(prompts, dotprompt.PromptRef{
				Name:    name,
				Variant: variant,
			})
		}
	}

	// Handle pagination if needed
	cursor := ""
	if limit, ok := options["limit"].(int); ok && limit > 0 && len(prompts) > limit {
		if len(prompts) > limit {
			cursor = prompts[limit-1].Name
			prompts = prompts[:limit]
		}
	}

	return prompts, cursor, nil
}

// ListPartials returns a list of all partials
func (d *DirStore) ListPartials(options map[string]interface{}) ([]dotprompt.PartialRef, string, error) {
	files, err := d.scanDirectory("")
	if err != nil {
		return nil, "", err
	}

	var partials []dotprompt.PartialRef
	for _, file := range files {
		if d.isPartial(file) {
			name, variant, err := d.parsePromptFilename(strings.TrimPrefix(filepath.Base(file), "_"))
			if err != nil {
				continue
			}
			partials = append(partials, dotprompt.PartialRef{
				Name:    name,
				Variant: variant,
			})
		}
	}

	// Handle pagination if needed
	cursor := ""
	if limit, ok := options["limit"].(int); ok && limit > 0 && len(partials) > limit {
		if len(partials) > limit {
			cursor = partials[limit-1].Name
			partials = partials[:limit]
		}
	}

	return partials, cursor, nil
}

// Save saves a prompt to the store
func (d *DirStore) Save(prompt *dotprompt.PromptData) error {
	if prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}

	dirName := filepath.Dir(prompt.Name)
	baseName := filepath.Base(prompt.Name)
	fileName := baseName
	if prompt.Variant != "" {
		fileName = fmt.Sprintf("%s.%s", baseName, prompt.Variant)
	}
	fileName += ".prompt"

	filePath := filepath.Join(d.directory, dirName, fileName)

	// Create directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(prompt.Source), 0644); err != nil {
		return fmt.Errorf("failed to write prompt file: %w", err)
	}

	return nil
}

// Delete deletes a prompt from the store
func (d *DirStore) Delete(name string, options map[string]string) error {
	variant := options["variant"]
	fileName := name
	if variant != "" {
		fileName = fmt.Sprintf("%s.%s", name, variant)
	}
	fileName += ".prompt"

	filePath := filepath.Join(d.directory, fileName)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete prompt file: %w", err)
	}

	return nil
}
