package cmd

import "github.com/spf13/cobra"

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Explore available Bedrock models",
	Long:  `List models from the local catalog or get details about a specific model.`,
}

func init() {
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsInfoCmd)
}
