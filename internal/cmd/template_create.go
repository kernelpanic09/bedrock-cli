package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	tmpl "github.com/kernelpanic09/bedrock-cli/internal/template"
)

var flagTemplateFile string

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create or replace a template",
	Long: `Create a named template from a file.

The file can include optional YAML frontmatter describing variables and a default model:

  ---
  description: Describe what this template does
  variables:
    - name: Topic
      required: true
    - name: Style
      default: technical
  defaultModel: sonnet
  ---
  Explain {{.Topic}} in a {{.Style}} way.

Examples:
  bedrock-cli template create code-review --file templates/code-review.txt
  bedrock-cli template create explain --file explain.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateCreate,
}

func init() {
	templateCreateCmd.Flags().StringVar(&flagTemplateFile, "file", "", "path to template file (required)")
	templateCreateCmd.MarkFlagRequired("file")
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	content, err := os.ReadFile(flagTemplateFile)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", flagTemplateFile, err)
	}

	if err := tmpl.Save(name, string(content)); err != nil {
		return err
	}

	out.Success(fmt.Sprintf("Template %q saved.", name))
	return nil
}
