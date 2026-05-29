package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kernelpanic09/bedrock-cli/internal/config"
)

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value (or all values)",
	Long: `Get the value of a single config key, or all keys if none is provided.

Examples:
  bedrock-cli config get
  bedrock-cli config get default-model
  bedrock-cli config get region`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			key := args[0]
			val := config.Get(key)
			if val == "" {
				out.Dim("(not set, using default)")
			} else {
				fmt.Println(val)
			}
			return nil
		}
		return runConfigShow(cmd, args)
	},
}

// runConfigShow is also the default action for `bedrock-cli config`.
func runConfigShow(cmd *cobra.Command, args []string) error {
	cfgPath, _ := config.Path()

	if flagJSON {
		settings := map[string]any{
			"default-model": viper.GetString("default-model"),
			"region":        viper.GetString("region"),
			"max-tokens":    viper.GetInt("max-tokens"),
			"temperature":   viper.GetFloat64("temperature"),
			"cache-ttl":     viper.GetInt("cache-ttl"),
			"no-color":      viper.GetBool("no-color"),
			"no-stream":     viper.GetBool("no-stream"),
			"show-cost":     viper.GetBool("show-cost"),
			"config-file":   cfgPath,
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(settings)
	}

	out.Header("bedrock-cli config")
	out.Dim("  Config file: " + cfgPath)
	out.Println("")

	out.Table([]string{"Key", "Value"}, [][]string{
		{"default-model", viper.GetString("default-model")},
		{"region", viper.GetString("region")},
		{"max-tokens", fmt.Sprintf("%d", viper.GetInt("max-tokens"))},
		{"temperature", fmt.Sprintf("%.2f", viper.GetFloat64("temperature"))},
		{"cache-ttl", fmt.Sprintf("%d", viper.GetInt("cache-ttl"))},
		{"no-color", fmt.Sprintf("%v", viper.GetBool("no-color"))},
		{"no-stream", fmt.Sprintf("%v", viper.GetBool("no-stream"))},
		{"show-cost", fmt.Sprintf("%v", viper.GetBool("show-cost"))},
	})
	out.Println("")
	out.Dim("Change any value with 'bedrock-cli config set <key> <value>'")
	return nil
}
