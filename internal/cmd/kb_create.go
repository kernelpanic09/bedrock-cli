package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var (
	flagKBBucket        string
	flagKBEmbedding     string
	flagKBRoleARN       string
	flagKBCollectionARN string
)

var kbCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a Knowledge Base with an S3 data source",
	Long: `Create a Bedrock Knowledge Base backed by an OpenSearch Serverless vector store.

You need to supply:
  --bucket         S3 bucket that holds the source documents
  --role           IAM role ARN that Bedrock will assume
  --collection-arn OpenSearch Serverless collection ARN (must exist beforehand)

The embedding model defaults to amazon.titan-embed-text-v2:0.

Examples:
  bedrock-cli kb create my-kb \
    --bucket my-docs-bucket \
    --role arn:aws:iam::123456789012:role/BedrockKBRole \
    --collection-arn arn:aws:aoss:us-east-1:123456789012:collection/abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runKBCreate,
}

func init() {
	kbCreateCmd.Flags().StringVar(&flagKBBucket, "bucket", "", "S3 bucket for the data source (required)")
	kbCreateCmd.Flags().StringVar(&flagKBEmbedding, "embedding", "amazon.titan-embed-text-v2:0", "embedding model ID")
	kbCreateCmd.Flags().StringVar(&flagKBRoleARN, "role", "", "IAM role ARN for the knowledge base (required)")
	kbCreateCmd.Flags().StringVar(&flagKBCollectionARN, "collection-arn", "", "OpenSearch Serverless collection ARN (required)")
	_ = kbCreateCmd.MarkFlagRequired("bucket")
	_ = kbCreateCmd.MarkFlagRequired("role")
	_ = kbCreateCmd.MarkFlagRequired("collection-arn")
}

func runKBCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	out.Dim(fmt.Sprintf("Creating knowledge base %q with bucket %s...", name, flagKBBucket))
	detail, err := cl.CreateKnowledgeBase(ctx, name, flagKBBucket, flagKBEmbedding, flagKBRoleARN, flagKBCollectionARN)
	if err != nil {
		return err
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(detail)
	}

	out.Success(fmt.Sprintf("Created knowledge base %s (%s)", detail.Name, detail.ID))
	out.Printf("  Status      : %s\n", detail.Status)
	out.Printf("  Vector store: %s\n", detail.VectorStore)
	out.Printf("  Embedding   : %s\n", detail.EmbeddingModel)
	out.Println("")
	out.Dim("Next: upload docs with 'bedrock-cli kb add-docs " + detail.ID + " <path>'")
	return nil
}
