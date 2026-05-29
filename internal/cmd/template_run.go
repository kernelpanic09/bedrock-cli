package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kernelpanic09/bedrock-cli/internal/cache"
	"github.com/kernelpanic09/bedrock-cli/internal/client"
	"github.com/kernelpanic09/bedrock-cli/internal/cost"
	"github.com/kernelpanic09/bedrock-cli/internal/models"
	tmpl "github.com/kernelpanic09/bedrock-cli/internal/template"
)

var flagTemplateVars []string

var templateRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Render and send a template",
	Long: `Render a named template with the given variable values and send it to a model.

Variables are provided with --var Name=Value. Required variables without a value
will cause an error. Optional variables use their declared defaults.

Examples:
  bedrock-cli template run code-review --var Diff="$(git diff)" --var Focus=security
  bedrock-cli template run explain --var Topic="mutual TLS"
  bedrock-cli template run summarize --var Text="$(cat article.txt)" --var Length="one paragraph"`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateRun,
}

func init() {
	templateRunCmd.Flags().StringArrayVar(&flagTemplateVars, "var", nil, "variable as Name=Value (repeatable)")
	templateRunCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "bypass the response cache")
	templateRunCmd.Flags().BoolVar(&flagNoStream, "no-stream", false, "wait for full response instead of streaming")
}

func runTemplateRun(cmd *cobra.Command, args []string) error {
	name := args[0]

	t, err := tmpl.Load(name)
	if err != nil {
		return err
	}

	// Parse --var flags into a map.
	vars := make(map[string]string)
	for _, v := range flagTemplateVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --var %q: expected Name=Value format", v)
		}
		vars[parts[0]] = parts[1]
	}

	// Check for missing required vars.
	missing := tmpl.MissingRequired(t, vars)
	if len(missing) != 0 {
		return fmt.Errorf("missing required variable(s): %s\nProvide them with --var Name=Value", strings.Join(missing, ", "))
	}

	rendered, err := tmpl.Render(t, vars)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	// Determine which model to use.
	modelAlias := resolveModel()
	if t.Meta.DefaultModel != "" && !cmd.Flags().Changed("model") {
		modelAlias = t.Meta.DefaultModel
	}

	model, err := models.Resolve(modelAlias)
	if err != nil {
		return err
	}

	maxTokens := viper.GetInt("max-tokens")
	temperature := viper.GetFloat64("temperature")
	noStream := flagNoStream || viper.GetBool("no-stream")

	cacheKey := cache.Key(model.ID, rendered, temperature, maxTokens)

	if !flagNoCache {
		if entry := cache.Get(cacheKey); entry != nil {
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
	cl, err := client.New(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating Bedrock client: %w", err)
	}

	start := time.Now()
	var result *client.InvokeResult

	if noStream {
		result, err = cl.Invoke(ctx, model.ID, rendered, temperature, maxTokens)
		if err != nil {
			return err
		}
		fmt.Println(result.Response)
	} else {
		var sb strings.Builder
		result, err = cl.InvokeStream(ctx, model.ID, rendered, temperature, maxTokens, func(token string) {
			fmt.Print(token)
			sb.WriteString(token)
		})
		if err != nil {
			return err
		}
		fmt.Println()
		result.Response = sb.String()
	}

	durationMs := time.Since(start).Milliseconds()
	costUSD := cost.Calculate(model.ID, result.InputTokens, result.OutputTokens)

	if viper.GetBool("show-cost") {
		out.CostLine(costUSD, result.InputTokens, result.OutputTokens, durationMs, false)
	}

	if !flagNoCache {
		entry := &cache.Entry{
			Model:        model.ID,
			Prompt:       rendered,
			Response:     result.Response,
			InputTokens:  result.InputTokens,
			OutputTokens: result.OutputTokens,
			CachedAt:     time.Now(),
		}
		_ = cache.Put(cacheKey, entry)
	}

	tracker, err := cost.Open()
	if err == nil {
		defer tracker.Close()
		_ = tracker.Record(&cost.Invocation{
			Timestamp:    time.Now(),
			Model:        model.ID,
			InputTokens:  result.InputTokens,
			OutputTokens: result.OutputTokens,
			CostUSD:      costUSD,
			TemplateName: name,
			DurationMs:   durationMs,
		})
	}

	return nil
}
