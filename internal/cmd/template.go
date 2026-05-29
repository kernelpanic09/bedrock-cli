package cmd

import "github.com/spf13/cobra"

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage prompt templates",
	Long: `Create, list, run, and delete named prompt templates.

Templates are stored in ~/.config/bedrock-cli/templates/ as text files with optional
YAML frontmatter. Variables are Go template syntax: {{.VarName}}.`,
}

func init() {
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateRunCmd)
	templateCmd.AddCommand(templateDeleteCmd)
}
