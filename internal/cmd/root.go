package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kernelpanic09/bedrock-cli/internal/config"
	"github.com/kernelpanic09/bedrock-cli/internal/output"
)

var (
	// Flags shared across commands.
	flagRegion  string
	flagModel   string
	flagVerbose bool
	flagJSON    bool

	out = output.Stdout()
)

// rootCmd is the top-level command. Running `bedrock-cli` with no subcommand
// and arguments treats it as `bedrock-cli prompt`.
var rootCmd = &cobra.Command{
	Use:   "bedrock-cli",
	Short: "A friendlier CLI for AWS Bedrock",
	Long: `bedrock-cli lets you send prompts, compare models, manage templates, and track costs
without hand-crafting JSON payloads or memorizing model IDs.

Run 'bedrock-cli help <command>' for details on any command.`,
	// If called with no subcommand and at least one argument, treat it as a prompt.
	Args: cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if flagVerbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// Delegate to promptCmd with the args joined as the prompt text.
		promptCmd.SetArgs(args)
		return promptCmd.RunE(promptCmd, args)
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&flagRegion, "region", "", "AWS region (overrides config and env)")
	rootCmd.PersistentFlags().StringVarP(&flagModel, "model", "m", "", "model alias or full ID (overrides config default)")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as JSON (on commands that support it)")

	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(costCmd)
	rootCmd.AddCommand(kbCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	// Load the config file and apply defaults. Errors here are non-fatal
	// (e.g. missing config file on first run).
	_, _ = config.Load()

	// Flag overrides take highest precedence.
	if flagRegion != "" {
		viper.Set("region", flagRegion)
	}
	if flagModel != "" {
		viper.Set("default-model", flagModel)
	}
}

// resolveModel returns the effective model ID, checking:
//  1. --model flag
//  2. BEDROCK_CLI_DEFAULT_MODEL env var
//  3. Config file default-model
//  4. Hardcoded default (sonnet)
func resolveModel() string {
	m := viper.GetString("default-model")
	if m == "" {
		m = config.DefaultModel
	}
	return m
}

// resolveRegion returns the effective region.
func resolveRegion() string {
	r := viper.GetString("region")
	if r == "" {
		r = config.DefaultRegion
	}
	return r
}

// maybeRunFirstTime checks whether this is the first time the user has run
// bedrock-cli. If so, it walks them through minimal setup.
func maybeRunFirstTime() error {
	cfgPath, err := config.Path()
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfgPath); err == nil {
		// Config exists - nothing to do.
		return nil
	}

	// First run.
	out.Header("Welcome to bedrock-cli!")
	out.Println("")
	out.Println("Looks like this is your first time. Let's set a couple defaults.")
	out.Println("Press Enter to keep the bracketed default.")
	out.Println("")

	region := prompt("AWS region", config.DefaultRegion)
	model := prompt("Default model (haiku/sonnet/opus or full ID)", config.DefaultModel)

	viper.Set("region", region)
	viper.Set("default-model", model)

	if err := config.Save(); err != nil {
		return fmt.Errorf("saving initial config: %w", err)
	}

	out.Success("Config saved. You're good to go.")
	out.Println("")
	return nil
}

// prompt reads a line from stdin with a default fallback.
func prompt(label, defaultVal string) string {
	fmt.Printf("  %s [%s]: ", label, defaultVal)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}
