package template

import (
	"bytes"
	"fmt"
	"strings"
	gotemplate "text/template"
)

// Render fills in the template body with the given variables.
// vars maps variable name to value.
func Render(t *Template, vars map[string]string) (string, error) {
	// Fill in defaults for any variables not explicitly provided.
	for _, v := range t.Meta.Variables {
		if _, ok := vars[v.Name]; !ok {
			if v.Required {
				return "", fmt.Errorf("required variable %q not provided", v.Name)
			}
			if v.Default != "" {
				vars[v.Name] = v.Default
			}
		}
	}

	tmpl, err := gotemplate.New(t.Name).Option("missingkey=error").Parse(t.Body)
	if err != nil {
		return "", fmt.Errorf("parsing template body: %w", err)
	}

	// Convert the map to a generic any map for the template engine.
	data := make(map[string]any, len(vars))
	for k, v := range vars {
		data[k] = v
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// Make the error message friendlier.
		errMsg := err.Error()
		if strings.Contains(errMsg, "map has no entry for key") {
			// Extract the key name from the error.
			return "", fmt.Errorf("template references undefined variable: %w", err)
		}
		return "", fmt.Errorf("rendering template: %w", err)
	}

	return buf.String(), nil
}

// MissingRequired returns the names of required variables not present in vars.
func MissingRequired(t *Template, vars map[string]string) []string {
	var missing []string
	for _, v := range t.Meta.Variables {
		if v.Required {
			if _, ok := vars[v.Name]; !ok {
				missing = append(missing, v.Name)
			}
		}
	}
	return missing
}
