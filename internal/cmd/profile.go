package cmd

import "github.com/spf13/cobra"

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage AWS profiles for bedrock-cli",
	Long: `List available AWS profiles, switch the active one, and inspect the
current identity without leaving the terminal.

The active profile is stored in the bedrock-cli config file and does not
modify your ~/.aws files.`,
}

func init() {
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileCurrentCmd)
	profileCmd.AddCommand(profileWhoamiCmd)
}
