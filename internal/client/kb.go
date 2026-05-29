package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	arttypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

// listKnowledgeBases returns all knowledge bases visible to the current credentials.
func listKnowledgeBases(ctx context.Context, client *bedrockagentruntime.Client) ([]KnowledgeBase, error) {
	// The agent runtime doesn't expose a ListKnowledgeBases API directly;
	// that lives in the bedrock-agent control plane. We use the bedrock agent
	// client here as a best-effort. If the caller needs to list KBs they can
	// use the AWS console or `aws bedrock-agent list-knowledge-bases`.
	//
	// For now we return a helpful error rather than silently returning nothing.
	_ = client
	return nil, fmt.Errorf("listing knowledge bases requires the bedrock-agent control plane client; use 'aws bedrock-agent list-knowledge-bases --region <region>' or upgrade this tool")
}

// queryKnowledgeBase retrieves results from a knowledge base via the agent runtime.
func queryKnowledgeBase(ctx context.Context, client *bedrockagentruntime.Client, kbID, query string, maxResults int) ([]KBResult, error) {
	max := int32(maxResults)
	input := &bedrockagentruntime.RetrieveInput{
		KnowledgeBaseId: aws.String(kbID),
		RetrievalQuery: &arttypes.KnowledgeBaseQuery{
			Text: aws.String(query),
		},
		RetrievalConfiguration: &arttypes.KnowledgeBaseRetrievalConfiguration{
			VectorSearchConfiguration: &arttypes.KnowledgeBaseVectorSearchConfiguration{
				NumberOfResults: &max,
			},
		},
	}

	resp, err := client.Retrieve(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("querying knowledge base %s: %w", kbID, err)
	}

	var results []KBResult
	for _, r := range resp.RetrievalResults {
		result := KBResult{}

		if r.Content != nil && r.Content.Text != nil {
			result.Content = aws.ToString(r.Content.Text)
		}

		if r.Score != nil {
			result.Score = float64(*r.Score)
		}

		if r.Location != nil && r.Location.S3Location != nil {
			result.Source = aws.ToString(r.Location.S3Location.Uri)
		}

		results = append(results, result)
	}

	return results, nil
}
