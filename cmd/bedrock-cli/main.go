// Command bedrock-cli is a friendlier CLI for AWS Bedrock: send prompts,
// compare models, manage templates and knowledge bases, and track costs.
//
// The command tree lives in internal/cmd; this package is the thin entry point.
package main

import "github.com/kernelpanic09/bedrock-cli/internal/cmd"

func main() {
	cmd.Execute()
}
