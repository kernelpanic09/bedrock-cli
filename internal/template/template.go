package template

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kernelpanic09/bedrock-cli/internal/config"
	"gopkg.in/yaml.v3"
)

// Variable describes a template variable with optional default and required flag.
type Variable struct {
	Name     string `yaml:"name"`
	Required bool   `yaml:"required"`
	Default  string `yaml:"default"`
}

// Meta is the parsed YAML frontmatter from a template file.
type Meta struct {
	Description  string     `yaml:"description"`
	Variables    []Variable `yaml:"variables"`
	DefaultModel string     `yaml:"defaultModel"`
}

// Template is the parsed form of a user template.
type Template struct {
	Name string
	Meta Meta
	Body string // raw body without frontmatter
}

const frontmatterDelim = "---"

// Load reads a named template from the user's templates directory.
func Load(name string) (*Template, error) {
	dir, err := config.TemplatesDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, name+".txt")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template %q not found - run 'bedrock-cli template list' to see available templates", name)
		}
		return nil, fmt.Errorf("reading template %q: %w", name, err)
	}

	return parse(name, string(data))
}

// Save writes a template to disk.
func Save(name, content string) error {
	dir, err := config.TemplatesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name+".txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("saving template %q: %w", name, err)
	}
	return nil
}

// Delete removes a template from disk.
func Delete(name string) error {
	dir, err := config.TemplatesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name+".txt")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template %q not found", name)
		}
		return fmt.Errorf("deleting template %q: %w", name, err)
	}
	return nil
}

// List returns the names of all saved templates.
func List() ([]string, error) {
	dir, err := config.TemplatesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing templates: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			names = append(names, strings.TrimSuffix(e.Name(), ".txt"))
		}
	}
	return names, nil
}

// parse splits a raw file into frontmatter and body, then unmarshals the meta.
func parse(name, raw string) (*Template, error) {
	t := &Template{Name: name}

	// Check for frontmatter delimiters.
	if strings.HasPrefix(raw, frontmatterDelim+"\n") {
		// Find the closing delimiter.
		rest := raw[len(frontmatterDelim)+1:]
		end := strings.Index(rest, "\n"+frontmatterDelim)
		if end == -1 {
			return nil, errors.New("template has opening frontmatter delimiter but no closing one")
		}
		fm := rest[:end]
		if err := yaml.Unmarshal([]byte(fm), &t.Meta); err != nil {
			return nil, fmt.Errorf("parsing template frontmatter: %w", err)
		}
		// Body starts after the closing delimiter and the trailing newline.
		body := rest[end+len("\n"+frontmatterDelim):]
		t.Body = strings.TrimPrefix(body, "\n")
	} else {
		// No frontmatter - the whole file is the body.
		t.Body = raw
	}

	return t, nil
}
