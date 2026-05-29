package template

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleTemplate = `---
description: Test template
variables:
  - name: Topic
    required: true
  - name: Length
    default: short
defaultModel: haiku
---
Explain {{.Topic}} in a {{.Length}} way.
`

func TestParse(t *testing.T) {
	tmpl, err := parse("test", sampleTemplate)
	if err != nil {
		t.Fatalf("parse() error: %v", err)
	}

	if tmpl.Meta.Description != "Test template" {
		t.Errorf("Description = %q, want %q", tmpl.Meta.Description, "Test template")
	}
	if tmpl.Meta.DefaultModel != "haiku" {
		t.Errorf("DefaultModel = %q, want %q", tmpl.Meta.DefaultModel, "haiku")
	}
	if len(tmpl.Meta.Variables) != 2 {
		t.Errorf("Variables len = %d, want 2", len(tmpl.Meta.Variables))
	}
	if tmpl.Meta.Variables[0].Required != true {
		t.Error("Topic should be required")
	}
	if tmpl.Meta.Variables[1].Default != "short" {
		t.Errorf("Length default = %q, want %q", tmpl.Meta.Variables[1].Default, "short")
	}
	if tmpl.Body == "" {
		t.Error("Body should not be empty")
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	raw := "Just a plain body {{.Name}}"
	tmpl, err := parse("plain", raw)
	if err != nil {
		t.Fatalf("parse() error: %v", err)
	}
	if tmpl.Body != raw {
		t.Errorf("Body = %q, want %q", tmpl.Body, raw)
	}
	if len(tmpl.Meta.Variables) != 0 {
		t.Error("expected no variables in plain template")
	}
}

func TestParseUnclosedFrontmatter(t *testing.T) {
	raw := "---\ndescription: broken\n"
	_, err := parse("broken", raw)
	if err == nil {
		t.Error("parse() should error on unclosed frontmatter")
	}
}

func TestRender(t *testing.T) {
	tmpl, err := parse("test", sampleTemplate)
	if err != nil {
		t.Fatalf("parse() error: %v", err)
	}

	vars := map[string]string{"Topic": "Go interfaces"}
	result, err := Render(tmpl, vars)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	want := "Explain Go interfaces in a short way."
	if result != want+"\n" && result != want {
		t.Errorf("Render() = %q, want %q (or with trailing newline)", result, want)
	}
}

func TestRenderMissingRequired(t *testing.T) {
	tmpl, err := parse("test", sampleTemplate)
	if err != nil {
		t.Fatalf("parse() error: %v", err)
	}

	// Omit the required Topic variable.
	_, err = Render(tmpl, map[string]string{})
	if err == nil {
		t.Error("Render() should error when required variable is missing")
	}
}

func TestRenderCustomDefault(t *testing.T) {
	tmpl, err := parse("test", sampleTemplate)
	if err != nil {
		t.Fatalf("parse() error: %v", err)
	}

	vars := map[string]string{"Topic": "HCL", "Length": "very detailed"}
	result, err := Render(tmpl, vars)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if result == "" {
		t.Error("Render() returned empty string")
	}
}

func TestMissingRequired(t *testing.T) {
	tmpl, err := parse("test", sampleTemplate)
	if err != nil {
		t.Fatalf("parse() error: %v", err)
	}

	missing := MissingRequired(tmpl, map[string]string{})
	if len(missing) != 1 || missing[0] != "Topic" {
		t.Errorf("MissingRequired() = %v, want [Topic]", missing)
	}

	missing = MissingRequired(tmpl, map[string]string{"Topic": "x"})
	if len(missing) != 0 {
		t.Errorf("MissingRequired() = %v, want []", missing)
	}
}

func TestSaveLoadDelete(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create a templates dir for the test.
	tmplDir := filepath.Join(tmp, ".config", "bedrock-cli", "templates")
	if err := os.MkdirAll(tmplDir, 0o700); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	name := "my-test-tmpl"
	content := "Hello {{.Name}}"

	if err := Save(name, content); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(name)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Body != content {
		t.Errorf("Load().Body = %q, want %q", loaded.Body, content)
	}

	names, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	found := false
	for _, n := range names {
		if n == name {
			found = true
		}
	}
	if !found {
		t.Errorf("List() did not include %q", name)
	}

	if err := Delete(name); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if _, err := Load(name); err == nil {
		t.Error("Load() should error after Delete()")
	}
}
