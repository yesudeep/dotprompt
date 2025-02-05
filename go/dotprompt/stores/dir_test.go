// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/dotprompt/go/dotprompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDir(t *testing.T) (string, func()) {
	dir, err := os.MkdirTemp("", "dotprompt-test-*")
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestNewDirStore(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	store := NewDirStore(DirStoreOptions{Directory: dir})
	assert.NotNil(t, store)
	assert.Equal(t, dir, store.directory)
}

func TestDirStore_Load(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	store := NewDirStore(DirStoreOptions{Directory: dir})

	t.Run("loads simple prompt", func(t *testing.T) {
		content := "Test prompt content"
		err := os.WriteFile(filepath.Join(dir, "test.prompt"), []byte(content), 0644)
		require.NoError(t, err)

		prompt, err := store.Load("test", nil)
		require.NoError(t, err)
		assert.Equal(t, "test", prompt.Name)
		assert.Equal(t, content, prompt.Source)
		assert.NotEmpty(t, prompt.Version)
	})

	t.Run("loads prompt with variant", func(t *testing.T) {
		content := "Test variant content"
		err := os.WriteFile(filepath.Join(dir, "test.v1.prompt"), []byte(content), 0644)
		require.NoError(t, err)

		prompt, err := store.Load("test", map[string]string{"variant": "v1"})
		require.NoError(t, err)
		assert.Equal(t, "test", prompt.Name)
		assert.Equal(t, "v1", prompt.Variant)
		assert.Equal(t, content, prompt.Source)
	})

	t.Run("fails on missing prompt", func(t *testing.T) {
		_, err := store.Load("missing", nil)
		assert.Error(t, err)
	})
}

func TestDirStore_LoadPartial(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	store := NewDirStore(DirStoreOptions{Directory: dir})

	t.Run("loads simple partial", func(t *testing.T) {
		content := "Test partial content"
		err := os.WriteFile(filepath.Join(dir, "_header.prompt"), []byte(content), 0644)
		require.NoError(t, err)

		partial, err := store.LoadPartial("header", nil)
		require.NoError(t, err)
		assert.Equal(t, "header", partial.Name)
		assert.Equal(t, content, partial.Source)
	})

	t.Run("loads partial with variant", func(t *testing.T) {
		content := "Test variant partial"
		err := os.WriteFile(filepath.Join(dir, "_header.v1.prompt"), []byte(content), 0644)
		require.NoError(t, err)

		partial, err := store.LoadPartial("header", map[string]string{"variant": "v1"})
		require.NoError(t, err)
		assert.Equal(t, "header", partial.Name)
		assert.Equal(t, "v1", partial.Variant)
		assert.Equal(t, content, partial.Source)
	})
}

func TestDirStore_List(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	store := NewDirStore(DirStoreOptions{Directory: dir})

	// Create test prompts
	files := map[string]string{
		"test1.prompt":     "Content 1",
		"test2.v1.prompt": "Content 2",
		"_partial.prompt": "Partial content",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		require.NoError(t, err)
	}

	t.Run("lists all prompts", func(t *testing.T) {
		prompts, cursor, err := store.List(nil)
		require.NoError(t, err)
		assert.Empty(t, cursor)
		assert.Len(t, prompts, 2)

		// Verify prompts are listed correctly
		var found bool
		for _, p := range prompts {
			if p.Name == "test1" && p.Variant == "" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find test1 prompt")
	})

	t.Run("handles pagination", func(t *testing.T) {
		prompts, cursor, err := store.List(map[string]interface{}{"limit": 1})
		require.NoError(t, err)
		assert.NotEmpty(t, cursor)
		assert.Len(t, prompts, 1)
	})
}

func TestDirStore_ListPartials(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	store := NewDirStore(DirStoreOptions{Directory: dir})

	// Create test files
	files := map[string]string{
		"_partial1.prompt":     "Partial 1",
		"_partial2.v1.prompt": "Partial 2",
		"test.prompt":         "Not a partial",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		require.NoError(t, err)
	}

	t.Run("lists all partials", func(t *testing.T) {
		partials, cursor, err := store.ListPartials(nil)
		require.NoError(t, err)
		assert.Empty(t, cursor)
		assert.Len(t, partials, 2)

		// Verify partials are listed correctly
		var found bool
		for _, p := range partials {
			if p.Name == "partial1" && p.Variant == "" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find partial1")
	})
}

func TestDirStore_Save(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	store := NewDirStore(DirStoreOptions{Directory: dir})

	t.Run("saves new prompt", func(t *testing.T) {
		prompt := &dotprompt.PromptData{
			Name:   "test",
			Source: "Test content",
		}

		err := store.Save(prompt)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "test.prompt"))
		require.NoError(t, err)
		assert.Equal(t, prompt.Source, string(content))
	})

	t.Run("saves prompt with variant", func(t *testing.T) {
		prompt := &dotprompt.PromptData{
			Name:    "test",
			Variant: "v1",
			Source:  "Variant content",
		}

		err := store.Save(prompt)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "test.v1.prompt"))
		require.NoError(t, err)
		assert.Equal(t, prompt.Source, string(content))
	})

	t.Run("fails without name", func(t *testing.T) {
		prompt := &dotprompt.PromptData{
			Source: "Test content",
		}

		err := store.Save(prompt)
		assert.Error(t, err)
	})
}

func TestDirStore_Delete(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	store := NewDirStore(DirStoreOptions{Directory: dir})

	t.Run("deletes existing prompt", func(t *testing.T) {
		// Create a test prompt
		err := os.WriteFile(filepath.Join(dir, "test.prompt"), []byte("content"), 0644)
		require.NoError(t, err)

		err = store.Delete("test", nil)
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(dir, "test.prompt"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("deletes prompt with variant", func(t *testing.T) {
		// Create a test prompt with variant
		err := os.WriteFile(filepath.Join(dir, "test.v1.prompt"), []byte("content"), 0644)
		require.NoError(t, err)

		err = store.Delete("test", map[string]string{"variant": "v1"})
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(dir, "test.v1.prompt"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("fails on missing prompt", func(t *testing.T) {
		err := store.Delete("missing", nil)
		assert.Error(t, err)
	})
}
