package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kernelpanic09/bedrock-cli/internal/cache"
	"github.com/kernelpanic09/bedrock-cli/internal/client"
	"github.com/kernelpanic09/bedrock-cli/internal/cost"
	"github.com/kernelpanic09/bedrock-cli/internal/models"
)

var (
	flagNoCache     bool
	flagNoStream    bool
	flagMaxTokens   int
	flagTemperature float64
	flagSystem      string
	flagProject     string
)

var promptCmd = &cobra.Command{
	Use:   "prompt <text>",
	Short: "Send a prompt to a Bedrock model",
	Long: `Send a prompt to a Bedrock model and print the response.

Streaming is on by default. Use --no-stream to wait for the full response.
Responses are cached by default; --no-cache bypasses the cache.

Examples:
  bedrock-cli prompt "Explain the difference between OIDC and SAML"
  bedrock-cli prompt --model haiku "What is HCL?"
  bedrock-cli prompt --model sonnet --no-cache "Latest features of Kubernetes 1.30"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runPrompt,
}

func init() {
	promptCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "bypass the response cache")
	promptCmd.Flags().BoolVar(&flagNoStream, "no-stream", false, "wait for full response instead of streaming")
	promptCmd.Flags().IntVar(&flagMaxTokens, "max-tokens", 0, "maximum tokens in response (0 = use config default)")
	promptCmd.Flags().Float64Var(&flagTemperature, "temperature", -1, "sampling temperature 0-1 (-1 = use config default)")
	promptCmd.Flags().StringVar(&flagSystem, "system", "", "system prompt to prepend")
	promptCmd.Flags().StringVar(&flagProject, "project", "", "project tag for cost attribution")
}

func runPrompt(cmd *cobra.Command, args []string) error {
	if err := maybeRunFirstTime(); err != nil {
		return err
	}

	promptText := strings.Join(args, " ")
	if flagSystem != "" {
		promptText = flagSystem + "\n\n" + promptText
	}

	modelAlias := resolveModel()
	model, err := models.Resolve(modelAlias)
	if err != nil {
		return err
	}

	maxTokens := viper.GetInt("max-tokens")
	if flagMaxTokens > 0 {
		maxTokens = flagMaxTokens
	}

	temperature := viper.GetFloat64("temperature")
	if flagTemperature >= 0 {
		temperature = flagTemperature
	}

	noStream := flagNoStream || viper.GetBool("no-stream")

	cacheKey := cache.Key(model.ID, promptText, temperature, maxTokens)

	// Check cache first unless explicitly skipped.
	if !flagNoCache {
		if entry := cache.Get(cacheKey); entry != nil {
			if flagJSON {
				return printJSON(map[string]any{
					"response":      entry.Response,
					"model":         entry.Model,
					"input_tokens":  entry.InputTokens,
					"output_tokens": entry.OutputTokens,
					"cached":        true,
				})
			}
			fmt.Print(entry.Response)
			if !strings.HasSuffix(entry.Response, "\n") {
				fmt.Println()
			}
			if viper.GetBool("show-cost") {
				c := cost.Calculate(model.ID, entry.InputTokens, entry.OutputTokens)
				out.CostLine(c, entry.InputTokens, entry.OutputTokens, 0, true)
			}
			return nil
		}
	}

	ctx := context.Background()
	region := resolveRegion()

	cl, err := client.New(ctx, region)
	if err != nil {
		return fmt.Errorf("creating Bedrock client: %w", err)
	}

	start := time.Now()
	var result *client.InvokeResult

	if noStream {
		result, err = cl.Invoke(ctx, model.ID, promptText, temperature, maxTokens)
		if err != nil {
			return err
		}
		if !flagJSON {
			fmt.Println(result.Response)
		}
	} else {
		var sb strings.Builder
		result, err = cl.InvokeStream(ctx, model.ID, promptText, temperature, maxTokens, func(token string) {
			fmt.Print(token)
			sb.WriteString(token)
		})
		if err != nil {
			return err
		}
		fmt.Println() // trailing newline after stream
		result.Response = sb.String()
	}

	durationMs := time.Since(start).Milliseconds()

	if flagJSON {
		return printJSON(map[string]any{
			"response":      result.Response,
			"model":         model.ID,
			"input_tokens":  result.InputTokens,
			"output_tokens": result.OutputTokens,
			"cached":        false,
			"duration_ms":   durationMs,
		})
	}

	// Record cost.
	costUSD := cost.Calculate(model.ID, result.InputTokens, result.OutputTokens)
	if viper.GetBool("show-cost") {
		out.CostLine(costUSD, result.InputTokens, result.OutputTokens, durationMs, false)
	}

	// Cache the response for next time.
	if !flagNoCache {
		entry := &cache.Entry{
			Model:        model.ID,
			Prompt:       promptText,
			Response:     result.Response,
			InputTokens:  result.InputTokens,
			OutputTokens: result.OutputTokens,
			CachedAt:     time.Now(),
		}
		if err := cache.Put(cacheKey, entry); err != nil {
			// Non-fatal: log and continue.
			slogWarn("failed to cache response", "error", err)
		}
	}

	// Record in cost tracker, capturing identity for account-level attribution.
	tracker, err := cost.Open()
	if err == nil {
		defer tracker.Close()
		inv := &cost.Invocation{
			Timestamp:    time.Now(),
			Model:        model.ID,
			InputTokens:  result.InputTokens,
			OutputTokens: result.OutputTokens,
			CostUSD:      costUSD,
			DurationMs:   durationMs,
			Project:      flagProject,
			AWSProfile:   viper.GetString("aws-profile"),
		}
		if id, idErr := resolveCallerIdentity(ctx, region); idErr == nil {
			inv.AWSAccountID = id.AccountID
		}
		_ = tracker.Record(inv)
	}

	return nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(out.Writer())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
