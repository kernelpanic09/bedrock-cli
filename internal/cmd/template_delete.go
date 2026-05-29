package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	tmpl "github.com/kernelpanic09/bedrock-cli/internal/template"
)

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a saved template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := tmpl.Delete(name); err != nil {
			return err
		}
		out.Success(fmt.Sprintf("Template %q deleted.", name))
		return nil
	},
}
