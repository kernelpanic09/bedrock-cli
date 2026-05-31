package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
	"github.com/kernelpanic09/bedrock-cli/internal/cost"
	"github.com/kernelpanic09/bedrock-cli/internal/models"
)

var (
	flagCompareModels  string
	flagCompareProject string
)

var compareCmd = &cobra.Command{
	Use:   "compare <prompt>",
	Short: "Send the same prompt to multiple models side-by-side",
	Long: `Send the same prompt to two or more models in parallel and print the responses
side-by-side. Good for sanity-checking quality vs. cost tradeoffs.

Examples:
  bedrock-cli compare --models haiku,sonnet,opus "Write a regex for an email"
  bedrock-cli compare --models haiku,llama-3-70b "Explain eventual consistency"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCompare,
}

func init() {
	compareCmd.Flags().StringVar(&flagCompareModels, "models", "haiku,sonnet", "comma-separated list of model aliases or IDs")
	compareCmd.Flags().StringVar(&flagCompareProject, "project", "", "project tag for cost attribution")
}

type compareResult struct {
	modelAlias   string
	modelID      string
	response     string
	inputTokens  int
	outputTokens int
	costUSD      float64
	durationMs   int64
	err          error
}

func runCompare(cmd *cobra.Command, args []string) error {
	if err := maybeRunFirstTime(); err != nil {
		return err
	}

	promptText := strings.Join(args, " ")
	region := resolveRegion()

	aliases := strings.Split(flagCompareModels, ",")
	if len(aliases) < 2 {
		return fmt.Errorf("--models requires at least two models, got %d", len(aliases))
	}

	maxTokens := viper.GetInt("max-tokens")
	temperature := viper.GetFloat64("temperature")

	ctx := context.Background()
	cl, err := client.New(ctx, region)
	if err != nil {
		return fmt.Errorf("creating Bedrock client: %w", err)
	}

	results := make([]compareResult, len(aliases))
	var wg sync.WaitGroup

	for i, alias := range aliases {
		alias = strings.TrimSpace(alias)
		model, err := models.Resolve(alias)
		if err != nil {
			results[i] = compareResult{modelAlias: alias, err: err}
			continue
		}

		wg.Add(1)
		go func(idx int, m *models.Model) {
			defer wg.Done()
			start := time.Now()

			// Collect the full response without streaming (compare mode).
			result, err := cl.Invoke(ctx, m.ID, promptText, temperature, maxTokens)
			elapsed := time.Since(start).Milliseconds()

			if err != nil {
				results[idx] = compareResult{modelAlias: alias, modelID: m.ID, err: err}
				return
			}

			c := cost.Calculate(m.ID, result.InputTokens, result.OutputTokens)
			results[idx] = compareResult{
				modelAlias:   m.Alias,
				modelID:      m.ID,
				response:     result.Response,
				inputTokens:  result.InputTokens,
				outputTokens: result.OutputTokens,
				costUSD:      c,
				durationMs:   elapsed,
			}
		}(i, model)
	}

	wg.Wait()

	// Record each model result in the cost tracker.
	if tracker, err := cost.Open(); err == nil {
		profile := resolveProfile()
		accountID := ""
		if id, idErr := resolveCallerIdentity(ctx, region); idErr == nil {
			accountID = id.AccountID
		}
		for _, r := range results {
			if r.err != nil {
				continue
			}
			_ = tracker.Record(&cost.Invocation{
				Timestamp:    time.Now(),
				Model:        r.modelID,
				InputTokens:  r.inputTokens,
				OutputTokens: r.outputTokens,
				CostUSD:      r.costUSD,
				DurationMs:   r.durationMs,
				Project:      flagCompareProject,
				AWSProfile:   profile,
				AWSAccountID: accountID,
			})
		}
		tracker.Close()
	}

	if flagJSON {
		return printCompareJSON(results)
	}

	// Print each response in a labeled box.
	for _, r := range results {
		label := r.modelAlias
		if label == "" {
			label = r.modelID
		}
		if r.err != nil {
			out.BoxedResponse(label+" [ERROR]", r.err.Error())
			continue
		}
		cost := fmt.Sprintf("$%.4f | %d+%d tokens | %.1fs",
			r.costUSD, r.inputTokens, r.outputTokens, float64(r.durationMs)/1000)
		content := r.response + "\n\n" + cost
		out.BoxedResponse(label, content)
		fmt.Println()
	}

	return nil
}

func printCompareJSON(results []compareResult) error {
	type row struct {
		Model        string  `json:"model"`
		Response     string  `json:"response"`
		InputTokens  int     `json:"input_tokens"`
		OutputTokens int     `json:"output_tokens"`
		CostUSD      float64 `json:"cost_usd"`
		DurationMs   int64   `json:"duration_ms"`
		Error        string  `json:"error,omitempty"`
	}
	var rows []row
	for _, r := range results {
		entry := row{
			Model:        r.modelID,
			Response:     r.response,
			InputTokens:  r.inputTokens,
			OutputTokens: r.outputTokens,
			CostUSD:      r.costUSD,
			DurationMs:   r.durationMs,
		}
		if r.err != nil {
			entry.Error = r.err.Error()
		}
		rows = append(rows, entry)
	}
	enc := json.NewEncoder(out.Writer())
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
