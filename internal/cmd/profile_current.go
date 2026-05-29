package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var profileCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active profile and AWS identity",
	Long: `Print the active AWS profile, region, account ID, and identity ARN.

Examples:
  bedrock-cli profile current
  bedrock-cli profile current --json`,
	RunE: runProfileCurrent,
}

var profileWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Alias for 'profile current'",
	RunE:  runProfileCurrent,
}

func runProfileCurrent(cmd *cobra.Command, args []string) error {
	profile := viper.GetString("aws-profile")
	region := resolveRegion()

	ctx := context.Background()
	id, err := client.GetCallerIdentity(ctx, region, profile)
	if err != nil {
		// Still print profile/region even if STS fails.
		if flagJSON {
			return printJSON(map[string]any{
				"profile": profile,
				"region":  region,
				"error":   err.Error(),
			})
		}
		out.Header("Active profile")
		out.Printf("  Profile : %s\n", labelOrDefault(profile, "default"))
		out.Printf("  Region  : %s\n", region)
		out.Warn(fmt.Sprintf("could not resolve identity: %v", err))
		return nil
	}

	if flagJSON {
		return printJSON(map[string]any{
			"profile":    labelOrDefault(profile, "default"),
			"region":     region,
			"account_id": id.AccountID,
			"arn":        id.ARN,
			"user_id":    id.UserID,
		})
	}

	out.Header("Active profile")
	out.Printf("  Profile    : %s\n", labelOrDefault(profile, "default"))
	out.Printf("  Region     : %s\n", region)
	out.Printf("  Account ID : %s\n", id.AccountID)
	out.Printf("  ARN        : %s\n", id.ARN)
	out.Printf("  User ID    : %s\n", id.UserID)
	return nil
}

func labelOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
