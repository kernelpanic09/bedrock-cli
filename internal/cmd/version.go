package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is injected at build time via -ldflags.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the bedrock-cli version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bedrock-cli %s\n", Version)
	},
}
