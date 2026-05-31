package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kernelpanic09/bedrock-cli/internal/config"
)

// Entry is what we persist to disk for a cached response.
type Entry struct {
	Model        string    `json:"model"`
	Prompt       string    `json:"prompt"`
	Response     string    `json:"response"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	CachedAt     time.Time `json:"cached_at"`
}

// Key computes the SHA256 cache key for a given combination of inputs.
// All parameters are normalized (trimmed, lowercased where appropriate) so that
// minor formatting differences don't bust the cache.
func Key(model, prompt string, temperature float64, maxTokens int) string {
	// Normalize the prompt so trailing whitespace differences don't matter.
	normalized := strings.TrimSpace(prompt)
	raw := fmt.Sprintf("%s|%s|%.2f|%d", model, normalized, temperature, maxTokens)
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

// Get returns a cached entry for the given key, or nil if there's no hit.
// It ignores read errors silently - a cache miss is not a fatal condition.
func Get(key string) *Entry {
	path, err := entryPath(key)
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}
	return &entry
}

// Put writes an entry to the cache.
func Put(key string, entry *Entry) error {
	path, err := entryPath(key)
	if err != nil {
		return fmt.Errorf("resolving cache path: %w", err)
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache entry: %w", err)
	}

	// Write to a temp file then rename so we don't leave partial files on crash.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("committing cache file: %w", err)
	}
	return nil
}

// Delete removes a cache entry.
func Delete(key string) error {
	path, err := entryPath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting cache entry: %w", err)
	}
	return nil
}

// Clear removes all cached responses.
func Clear() error {
	dir, err := config.ResponseCacheDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading cache dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			path := filepath.Join(dir, e.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("removing %s: %w", path, err)
			}
		}
	}
	return nil
}

// Stats returns the number of cached entries and total size in bytes.
func Stats() (count int, totalBytes int64, err error) {
	dir, err := config.ResponseCacheDir()
	if err != nil {
		return 0, 0, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, fmt.Errorf("reading cache dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			count++
			if info, err := e.Info(); err == nil {
				totalBytes += info.Size()
			}
		}
	}
	return count, totalBytes, nil
}

func entryPath(key string) (string, error) {
	dir, err := config.ResponseCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, key+".json"), nil
}
