package cmd

import (
	"context"
	"log/slog"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

// resolveCallerIdentity returns the AWS caller identity, using the cached STS client.
// Errors are non-fatal - callers should log and continue.
func resolveCallerIdentity(ctx context.Context, region string) (*client.CallerIdentity, error) {
	profile := resolveProfile()
	id, err := client.GetCallerIdentity(ctx, region, profile)
	if err != nil {
		slog.Debug("could not resolve caller identity", "error", err)
		return nil, err
	}
	return id, nil
}
