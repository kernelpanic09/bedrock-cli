package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	tmpl "github.com/kernelpanic09/bedrock-cli/internal/template"
)

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved templates",
	RunE:  runTemplateList,
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	names, err := tmpl.List()
	if err != nil {
		return err
	}

	if len(names) == 0 {
		out.Dim("No templates saved yet. Use 'bedrock-cli template create' to add one.")
		return nil
	}

	if flagJSON {
		// Include metadata for machine consumers.
		type row struct {
			Name         string `json:"name"`
			Description  string `json:"description"`
			DefaultModel string `json:"default_model,omitempty"`
			Variables    int    `json:"variable_count"`
		}
		var rows []row
		for _, name := range names {
			r := row{Name: name}
			if t, err := tmpl.Load(name); err == nil {
				r.Description = t.Meta.Description
				r.DefaultModel = t.Meta.DefaultModel
				r.Variables = len(t.Meta.Variables)
			}
			rows = append(rows, r)
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	// Human output - show a table with description.
	headers := []string{"Name", "Description", "Model", "Variables"}
	var tableRows [][]string
	for _, name := range names {
		desc := ""
		model := ""
		vars := "0"
		if t, err := tmpl.Load(name); err == nil {
			desc = t.Meta.Description
			model = t.Meta.DefaultModel
			vars = fmt.Sprintf("%d", len(t.Meta.Variables))
		}
		tableRows = append(tableRows, []string{name, desc, model, vars})
	}

	out.Table(headers, tableRows)
	return nil
}
