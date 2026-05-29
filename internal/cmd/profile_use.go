package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/config"
)

var profileUseCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Set the active AWS profile",
	Long: `Switch to a named AWS profile for subsequent bedrock-cli commands.
The choice is saved in the bedrock-cli config file and does not touch ~/.aws.

Examples:
  bedrock-cli profile use prod
  bedrock-cli profile use default`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileUse,
}

func runProfileUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Verify the profile exists.
	profiles, err := collectAWSProfiles()
	if err != nil {
		return err
	}
	found := false
	for _, p := range profiles {
		if p == name {
			found = true
			break
		}
	}
	if !found {
		out.Warn(fmt.Sprintf("profile %q not found in ~/.aws/config or ~/.aws/credentials - saving anyway", name))
	}

	if err := config.Set("aws-profile", name); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	out.Success(fmt.Sprintf("Active profile set to %q", name))
	return nil
}
