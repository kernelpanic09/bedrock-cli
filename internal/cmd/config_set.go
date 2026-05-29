package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/config"
)

var validConfigKeys = map[string]string{
	"default-model": "alias or full model ID to use by default",
	"region":        "AWS region (e.g. us-east-1)",
	"max-tokens":    "default max tokens per response",
	"temperature":   "default sampling temperature (0.0-1.0)",
	"cache-ttl":     "cache TTL in hours (0 = forever, -1 = disable)",
	"no-color":      "disable color output (true/false)",
	"no-stream":     "disable streaming by default (true/false)",
	"show-cost":     "show cost line after each invocation (true/false)",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a single configuration key to a value and save to disk.

Valid keys:
  default-model    alias or full model ID to use by default
  region           AWS region (e.g. us-east-1)
  max-tokens       default max tokens per response
  temperature      default sampling temperature (0.0-1.0)
  cache-ttl        cache TTL in hours (0 = forever, -1 = disable)
  no-color         disable color output (true/false)
  no-stream        disable streaming by default (true/false)
  show-cost        show cost line after each invocation (true/false)

Examples:
  bedrock-cli config set default-model haiku
  bedrock-cli config set region us-west-2
  bedrock-cli config set show-cost false`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if _, ok := validConfigKeys[key]; !ok {
			return fmt.Errorf("unknown config key %q\n\nValid keys: %s", key, listValidKeys())
		}

		if err := config.Set(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}

		out.Success(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	},
}

func listValidKeys() string {
	var keys string
	for k := range validConfigKeys {
		keys += "\n  " + k
	}
	return keys
}
