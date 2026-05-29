package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify bedrock-cli configuration",
	Long: `View the current configuration or change individual settings.

Config is stored in ~/.config/bedrock-cli/config.toml.
Environment variables (BEDROCK_CLI_REGION, BEDROCK_CLI_DEFAULT_MODEL, etc.) take precedence.

Examples:
  bedrock-cli config
  bedrock-cli config set default-model haiku
  bedrock-cli config set region us-west-2
  bedrock-cli config get default-model`,
	RunE: runConfigShow,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}
