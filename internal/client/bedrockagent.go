package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	batypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	arttypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AgentClient wraps the Bedrock control-plane, bedrockagent, agentruntime, bedrockruntime, and S3 clients.
type AgentClient struct {
	control      *bedrock.Client      // bedrock control plane (guardrails, models)
	agent        *bedrockagent.Client // bedrockagent control plane (KBs, agents)
	agentRuntime *bedrockagentruntime.Client
	brt          *bedrockruntime.Client
	s3           *s3.Client
	region       string
}

// NewAgentClient creates a new AgentClient using the standard credential chain.
func NewAgentClient(ctx context.Context, region string) (*AgentClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return &AgentClient{
		control:      bedrock.NewFromConfig(cfg),
		agent:        bedrockagent.NewFromConfig(cfg),
		agentRuntime: bedrockagentruntime.NewFromConfig(cfg),
		brt:          bedrockruntime.NewFromConfig(cfg),
		s3:           s3.NewFromConfig(cfg),
		region:       region,
	}, nil
}

// NewAgentClientWithProfile creates a new AgentClient using a named AWS profile.
func NewAgentClientWithProfile(ctx context.Context, region, profile string) (*AgentClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config for profile %q: %w", profile, err)
	}
	return &AgentClient{
		control:      bedrock.NewFromConfig(cfg),
		agent:        bedrockagent.NewFromConfig(cfg),
		agentRuntime: bedrockagentruntime.NewFromConfig(cfg),
		brt:          bedrockruntime.NewFromConfig(cfg),
		s3:           s3.NewFromConfig(cfg),
		region:       region,
	}, nil
}

// KnowledgeBaseDetail holds full details about a knowledge base.
type KnowledgeBaseDetail struct {
	ID             string
	Name           string
	Description    string
	Status         string
	RoleARN        string
	EmbeddingModel string
	VectorStore    string
	DataSources    []DataSource
}

// DataSource describes a single data source attached to a knowledge base.
type DataSource struct {
	ID       string
	Name     string
	Status   string
	Type     string
	S3Bucket string
	S3Prefix string
}

// IngestionJob describes a knowledge base ingestion job.
type IngestionJob struct {
	JobID          string
	DataSourceID   string
	Status         string
	StartedAt      string
	UpdatedAt      string
	FailureReasons []string
}

// AgentSummary is a simplified view of a Bedrock Agent.
type AgentSummary struct {
	AgentID   string
	AgentName string
	Status    string
}

// AgentDetail holds full details about a Bedrock Agent.
type AgentDetail struct {
	AgentID        string
	AgentName      string
	Status         string
	Description    string
	Instruction    string
	Model          string
	ActionGroups   []string
	KnowledgeBases []string
}

// AgentInvokeResult is the result of a single agent invocation turn.
type AgentInvokeResult struct {
	SessionID string
	Response  string
}

// GuardrailSummary is a simplified view of a Bedrock Guardrail.
type GuardrailSummary struct {
	ID      string
	Name    string
	Version string
	Status  string
}

// GuardrailDetail holds full details about a Bedrock Guardrail.
type GuardrailDetail struct {
	ID             string
	Name           string
	Version        string
	Status         string
	Description    string
	TopicPolicies  []string
	ContentFilters []ContentFilter
	PIIRedactions  []string
	WordFilters    []string
}

// ContentFilter describes a content filter entry.
type ContentFilter struct {
	Type           string
	InputStrength  string
	OutputStrength string
}

// GuardrailTestResult holds the outcome of applying a guardrail to a prompt.
type GuardrailTestResult struct {
	Action      string // NONE, GUARDRAIL_INTERVENED
	Outputs     []string
	Assessments []GuardrailAssessment
}

// GuardrailAssessment describes why a guardrail fired.
type GuardrailAssessment struct {
	TopicPolicy   string
	ContentPolicy string
	PIIPolicy     string
	WordPolicy    string
}

// ListKnowledgeBases returns all knowledge bases via the bedrockagent control-plane client.
func (c *AgentClient) ListKnowledgeBases(ctx context.Context) ([]KnowledgeBase, error) {
	var results []KnowledgeBase
	paginator := bedrockagent.NewListKnowledgeBasesPaginator(c.agent, &bedrockagent.ListKnowledgeBasesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing knowledge bases: %w", err)
		}
		for _, kb := range page.KnowledgeBaseSummaries {
			results = append(results, KnowledgeBase{
				ID:          aws.ToString(kb.KnowledgeBaseId),
				Name:        aws.ToString(kb.Name),
				Description: aws.ToString(kb.Description),
				Status:      string(kb.Status),
			})
		}
	}
	return results, nil
}

// DescribeKnowledgeBase returns full details for a single knowledge base.
func (c *AgentClient) DescribeKnowledgeBase(ctx context.Context, kbID string) (*KnowledgeBaseDetail, error) {
	resp, err := c.agent.GetKnowledgeBase(ctx, &bedrockagent.GetKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return nil, fmt.Errorf("describing knowledge base %s: %w", kbID, err)
	}
	kb := resp.KnowledgeBase
	detail := &KnowledgeBaseDetail{
		ID:          aws.ToString(kb.KnowledgeBaseId),
		Name:        aws.ToString(kb.Name),
		Description: aws.ToString(kb.Description),
		Status:      string(kb.Status),
		RoleARN:     aws.ToString(kb.RoleArn),
	}
	if kb.KnowledgeBaseConfiguration != nil && kb.KnowledgeBaseConfiguration.VectorKnowledgeBaseConfiguration != nil {
		detail.EmbeddingModel = aws.ToString(kb.KnowledgeBaseConfiguration.VectorKnowledgeBaseConfiguration.EmbeddingModelArn)
	}
	if kb.StorageConfiguration != nil {
		detail.VectorStore = string(kb.StorageConfiguration.Type)
	}
	ds, err := c.ListDataSources(ctx, kbID)
	if err == nil {
		detail.DataSources = ds
	}
	return detail, nil
}

// ListDataSources returns the data sources for a knowledge base.
func (c *AgentClient) ListDataSources(ctx context.Context, kbID string) ([]DataSource, error) {
	resp, err := c.agent.ListDataSources(ctx, &bedrockagent.ListDataSourcesInput{
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return nil, fmt.Errorf("listing data sources for %s: %w", kbID, err)
	}
	var sources []DataSource
	for _, ds := range resp.DataSourceSummaries {
		sources = append(sources, DataSource{
			ID:     aws.ToString(ds.DataSourceId),
			Name:   aws.ToString(ds.Name),
			Status: string(ds.Status),
		})
	}
	// Fetch full config to get bucket info.
	for i := range sources {
		full, err := c.agent.GetDataSource(ctx, &bedrockagent.GetDataSourceInput{
			KnowledgeBaseId: aws.String(kbID),
			DataSourceId:    aws.String(sources[i].ID),
		})
		if err != nil {
			continue
		}
		ds := full.DataSource
		if ds.DataSourceConfiguration != nil {
			sources[i].Type = string(ds.DataSourceConfiguration.Type)
			if ds.DataSourceConfiguration.S3Configuration != nil {
				arn := aws.ToString(ds.DataSourceConfiguration.S3Configuration.BucketArn)
				// Convert ARN like arn:aws:s3:::bucket-name to bucket-name.
				if strings.HasPrefix(arn, "arn:") {
					parts := strings.Split(arn, ":")
					arn = parts[len(parts)-1]
				}
				sources[i].S3Bucket = arn
				if ds.DataSourceConfiguration.S3Configuration.InclusionPrefixes != nil {
					sources[i].S3Prefix = strings.Join(ds.DataSourceConfiguration.S3Configuration.InclusionPrefixes, ", ")
				}
			}
		}
	}
	return sources, nil
}

// CreateKnowledgeBase creates a new KB with an OpenSearch Serverless vector store.
// collectionARN must already exist (OpenSearch Serverless provisioning is outside scope).
func (c *AgentClient) CreateKnowledgeBase(ctx context.Context, name, bucket, embeddingModel, roleARN, collectionARN string) (*KnowledgeBaseDetail, error) {
	if roleARN == "" {
		return nil, fmt.Errorf(`role ARN required to create a knowledge base.

The IAM role needs:
  bedrock:InvokeModel
  aoss:APIAccessAll on the collection
  s3:GetObject, s3:ListBucket on the source bucket

See: https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base-setup.html`)
	}
	if collectionARN == "" {
		return nil, fmt.Errorf(`OpenSearch Serverless collection ARN required (--collection-arn).

Create one first:
  aws opensearchserverless create-collection --name %s --type VECTORSEARCH

Then pass the returned ARN with --collection-arn.`, sanitizeName(name))
	}

	resp, err := c.agent.CreateKnowledgeBase(ctx, &bedrockagent.CreateKnowledgeBaseInput{
		Name:    aws.String(name),
		RoleArn: aws.String(roleARN),
		KnowledgeBaseConfiguration: &batypes.KnowledgeBaseConfiguration{
			Type: batypes.KnowledgeBaseTypeVector,
			VectorKnowledgeBaseConfiguration: &batypes.VectorKnowledgeBaseConfiguration{
				EmbeddingModelArn: aws.String(embeddingModel),
			},
		},
		StorageConfiguration: &batypes.StorageConfiguration{
			Type: batypes.KnowledgeBaseStorageTypeOpensearchServerless,
			OpensearchServerlessConfiguration: &batypes.OpenSearchServerlessConfiguration{
				CollectionArn:   aws.String(collectionARN),
				VectorIndexName: aws.String(sanitizeName(name) + "-index"),
				FieldMapping: &batypes.OpenSearchServerlessFieldMapping{
					VectorField:   aws.String("embedding"),
					TextField:     aws.String("text"),
					MetadataField: aws.String("metadata"),
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating knowledge base: %w", err)
	}

	kb := resp.KnowledgeBase
	kbID := aws.ToString(kb.KnowledgeBaseId)

	// Create S3 data source.
	_, err = c.agent.CreateDataSource(ctx, &bedrockagent.CreateDataSourceInput{
		KnowledgeBaseId: aws.String(kbID),
		Name:            aws.String(name + "-s3"),
		DataSourceConfiguration: &batypes.DataSourceConfiguration{
			Type: batypes.DataSourceTypeS3,
			S3Configuration: &batypes.S3DataSourceConfiguration{
				BucketArn: aws.String(bucketARN(bucket)),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating S3 data source for bucket %s: %w", bucket, err)
	}

	return &KnowledgeBaseDetail{
		ID:             kbID,
		Name:           aws.ToString(kb.Name),
		Status:         string(kb.Status),
		RoleARN:        aws.ToString(kb.RoleArn),
		EmbeddingModel: embeddingModel,
		VectorStore:    "OPENSEARCH_SERVERLESS",
	}, nil
}

// DeleteKnowledgeBase removes a KB and its data sources.
func (c *AgentClient) DeleteKnowledgeBase(ctx context.Context, kbID string) error {
	dsList, err := c.agent.ListDataSources(ctx, &bedrockagent.ListDataSourcesInput{
		KnowledgeBaseId: aws.String(kbID),
	})
	if err == nil {
		for _, ds := range dsList.DataSourceSummaries {
			_, _ = c.agent.DeleteDataSource(ctx, &bedrockagent.DeleteDataSourceInput{
				KnowledgeBaseId: aws.String(kbID),
				DataSourceId:    ds.DataSourceId,
			})
		}
	}
	_, err = c.agent.DeleteKnowledgeBase(ctx, &bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return fmt.Errorf("deleting knowledge base %s: %w", kbID, err)
	}
	return nil
}

// UploadDocsToKB uploads local files to the KB's first S3 data source bucket.
func (c *AgentClient) UploadDocsToKB(ctx context.Context, kbID string, paths []string, progress func(name string)) (int, error) {
	sources, err := c.ListDataSources(ctx, kbID)
	if err != nil {
		return 0, err
	}
	bucket := ""
	for _, ds := range sources {
		if ds.S3Bucket != "" {
			bucket = ds.S3Bucket
			break
		}
	}
	if bucket == "" {
		return 0, fmt.Errorf("no S3 data source found on knowledge base %s", kbID)
	}

	count := 0
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			return count, fmt.Errorf("stat %s: %w", p, err)
		}
		if fi.IsDir() {
			n, err := c.uploadDir(ctx, bucket, p, p, progress)
			count += n
			if err != nil {
				return count, err
			}
		} else {
			if err := c.uploadFile(ctx, bucket, p, filepath.Base(p)); err != nil {
				return count, err
			}
			progress(p)
			count++
		}
	}
	return count, nil
}

func (c *AgentClient) uploadDir(ctx context.Context, bucket, dir, base string, progress func(string)) (int, error) {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("reading dir %s: %w", dir, err)
	}
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			n, err := c.uploadDir(ctx, bucket, full, base, progress)
			count += n
			if err != nil {
				return count, err
			}
			continue
		}
		rel, _ := filepath.Rel(base, full)
		if err := c.uploadFile(ctx, bucket, full, rel); err != nil {
			return count, err
		}
		progress(full)
		count++
	}
	return count, nil
}

func (c *AgentClient) uploadFile(ctx context.Context, bucket, localPath, key string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", localPath, err)
	}
	defer f.Close()

	body, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("reading %s: %w", localPath, err)
	}

	_, err = c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(string(body)),
	})
	if err != nil {
		return fmt.Errorf("uploading %s to s3://%s/%s: %w", localPath, bucket, key, err)
	}
	return nil
}

// StartIngestionJob triggers a sync on a KB data source.
// If dsID is empty, it uses the first data source.
func (c *AgentClient) StartIngestionJob(ctx context.Context, kbID, dsID string) (*IngestionJob, error) {
	if dsID == "" {
		sources, err := c.ListDataSources(ctx, kbID)
		if err != nil {
			return nil, err
		}
		if len(sources) == 0 {
			return nil, fmt.Errorf("no data sources found on knowledge base %s", kbID)
		}
		dsID = sources[0].ID
	}

	resp, err := c.agent.StartIngestionJob(ctx, &bedrockagent.StartIngestionJobInput{
		KnowledgeBaseId: aws.String(kbID),
		DataSourceId:    aws.String(dsID),
	})
	if err != nil {
		return nil, fmt.Errorf("starting ingestion job: %w", err)
	}
	job := resp.IngestionJob
	result := &IngestionJob{
		JobID:        aws.ToString(job.IngestionJobId),
		DataSourceID: aws.ToString(job.DataSourceId),
		Status:       string(job.Status),
	}
	if job.StartedAt != nil {
		result.StartedAt = job.StartedAt.Format("2006-01-02 15:04:05")
	}
	for _, r := range job.FailureReasons {
		result.FailureReasons = append(result.FailureReasons, r)
	}
	return result, nil
}

// ListIngestionJobs returns recent ingestion jobs across all data sources for a KB.
func (c *AgentClient) ListIngestionJobs(ctx context.Context, kbID string) ([]IngestionJob, error) {
	sources, err := c.ListDataSources(ctx, kbID)
	if err != nil {
		return nil, err
	}
	var jobs []IngestionJob
	for _, ds := range sources {
		resp, err := c.agent.ListIngestionJobs(ctx, &bedrockagent.ListIngestionJobsInput{
			KnowledgeBaseId: aws.String(kbID),
			DataSourceId:    aws.String(ds.ID),
		})
		if err != nil {
			continue
		}
		for _, j := range resp.IngestionJobSummaries {
			job := IngestionJob{
				JobID:        aws.ToString(j.IngestionJobId),
				DataSourceID: aws.ToString(j.DataSourceId),
				Status:       string(j.Status),
			}
			if j.StartedAt != nil {
				job.StartedAt = j.StartedAt.Format("2006-01-02 15:04:05")
			}
			if j.UpdatedAt != nil {
				job.UpdatedAt = j.UpdatedAt.Format("2006-01-02 15:04:05")
			}
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

// QueryKB queries a knowledge base via the agent runtime.
func (c *AgentClient) QueryKB(ctx context.Context, kbID, query string, maxResults int) ([]KBResult, error) {
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
	resp, err := c.agentRuntime.Retrieve(ctx, input)
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

// ListAgents returns all Bedrock Agents.
func (c *AgentClient) ListAgents(ctx context.Context) ([]AgentSummary, error) {
	resp, err := c.agent.ListAgents(ctx, &bedrockagent.ListAgentsInput{})
	if err != nil {
		return nil, fmt.Errorf("listing agents: %w", err)
	}
	var agents []AgentSummary
	for _, a := range resp.AgentSummaries {
		agents = append(agents, AgentSummary{
			AgentID:   aws.ToString(a.AgentId),
			AgentName: aws.ToString(a.AgentName),
			Status:    string(a.AgentStatus),
		})
	}
	return agents, nil
}

// DescribeAgent returns full details for a Bedrock Agent.
func (c *AgentClient) DescribeAgent(ctx context.Context, agentID string) (*AgentDetail, error) {
	resp, err := c.agent.GetAgent(ctx, &bedrockagent.GetAgentInput{
		AgentId: aws.String(agentID),
	})
	if err != nil {
		return nil, fmt.Errorf("describing agent %s: %w", agentID, err)
	}
	a := resp.Agent
	detail := &AgentDetail{
		AgentID:     aws.ToString(a.AgentId),
		AgentName:   aws.ToString(a.AgentName),
		Status:      string(a.AgentStatus),
		Description: aws.ToString(a.Description),
		Instruction: aws.ToString(a.Instruction),
		Model:       aws.ToString(a.FoundationModel),
	}

	ags, err := c.agent.ListAgentActionGroups(ctx, &bedrockagent.ListAgentActionGroupsInput{
		AgentId:      aws.String(agentID),
		AgentVersion: aws.String("DRAFT"),
	})
	if err == nil {
		for _, ag := range ags.ActionGroupSummaries {
			detail.ActionGroups = append(detail.ActionGroups, aws.ToString(ag.ActionGroupName))
		}
	}

	kbs, err := c.agent.ListAgentKnowledgeBases(ctx, &bedrockagent.ListAgentKnowledgeBasesInput{
		AgentId:      aws.String(agentID),
		AgentVersion: aws.String("DRAFT"),
	})
	if err == nil {
		for _, kb := range kbs.AgentKnowledgeBaseSummaries {
			detail.KnowledgeBases = append(detail.KnowledgeBases, aws.ToString(kb.KnowledgeBaseId))
		}
	}

	return detail, nil
}

// InvokeAgent calls a Bedrock Agent and streams the response.
func (c *AgentClient) InvokeAgent(ctx context.Context, agentID, sessionID, input string, onToken func(string)) (*AgentInvokeResult, error) {
	resp, err := c.agentRuntime.InvokeAgent(ctx, &bedrockagentruntime.InvokeAgentInput{
		AgentId:      aws.String(agentID),
		AgentAliasId: aws.String("TSTALIASID"),
		SessionId:    aws.String(sessionID),
		InputText:    aws.String(input),
		EnableTrace:  aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("invoking agent %s: %w", agentID, err)
	}

	result := &AgentInvokeResult{SessionID: sessionID}
	var sb strings.Builder

	stream := resp.GetStream()
	for event := range stream.Events() {
		switch v := event.(type) {
		case *arttypes.ResponseStreamMemberChunk:
			if v.Value.Bytes != nil {
				text := string(v.Value.Bytes)
				onToken(text)
				sb.WriteString(text)
			}
		}
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("reading agent stream: %w", err)
	}

	result.Response = sb.String()
	return result, nil
}

// ListGuardrails returns all guardrails via the bedrock control-plane.
func (c *AgentClient) ListGuardrails(ctx context.Context) ([]GuardrailSummary, error) {
	resp, err := c.control.ListGuardrails(ctx, &bedrock.ListGuardrailsInput{})
	if err != nil {
		return nil, fmt.Errorf("listing guardrails: %w", err)
	}
	var gs []GuardrailSummary
	for _, g := range resp.Guardrails {
		gs = append(gs, GuardrailSummary{
			ID:      aws.ToString(g.Id),
			Name:    aws.ToString(g.Name),
			Version: aws.ToString(g.Version),
			Status:  string(g.Status),
		})
	}
	return gs, nil
}

// DescribeGuardrail returns full config for a guardrail.
func (c *AgentClient) DescribeGuardrail(ctx context.Context, guardrailID string) (*GuardrailDetail, error) {
	resp, err := c.control.GetGuardrail(ctx, &bedrock.GetGuardrailInput{
		GuardrailIdentifier: aws.String(guardrailID),
	})
	if err != nil {
		return nil, fmt.Errorf("describing guardrail %s: %w", guardrailID, err)
	}

	detail := &GuardrailDetail{
		ID:          aws.ToString(resp.GuardrailId),
		Name:        aws.ToString(resp.Name),
		Version:     aws.ToString(resp.Version),
		Status:      string(resp.Status),
		Description: aws.ToString(resp.Description),
	}

	if resp.TopicPolicy != nil {
		for _, t := range resp.TopicPolicy.Topics {
			detail.TopicPolicies = append(detail.TopicPolicies, aws.ToString(t.Name))
		}
	}
	if resp.ContentPolicy != nil {
		for _, f := range resp.ContentPolicy.Filters {
			detail.ContentFilters = append(detail.ContentFilters, ContentFilter{
				Type:           string(f.Type),
				InputStrength:  string(f.InputStrength),
				OutputStrength: string(f.OutputStrength),
			})
		}
	}
	if resp.SensitiveInformationPolicy != nil {
		for _, p := range resp.SensitiveInformationPolicy.PiiEntities {
			detail.PIIRedactions = append(detail.PIIRedactions, string(p.Type))
		}
	}
	if resp.WordPolicy != nil {
		for _, w := range resp.WordPolicy.Words {
			detail.WordFilters = append(detail.WordFilters, aws.ToString(w.Text))
		}
	}
	return detail, nil
}

// TestGuardrail applies a guardrail to a prompt and returns the assessment.
func (c *AgentClient) TestGuardrail(ctx context.Context, guardrailID, guardrailVersion, text string) (*GuardrailTestResult, error) {
	if guardrailVersion == "" {
		guardrailVersion = "DRAFT"
	}
	resp, err := c.brt.ApplyGuardrail(ctx, &bedrockruntime.ApplyGuardrailInput{
		GuardrailIdentifier: aws.String(guardrailID),
		GuardrailVersion:    aws.String(guardrailVersion),
		Source:              brtypes.GuardrailContentSourceInput,
		Content: []brtypes.GuardrailContentBlock{
			&brtypes.GuardrailContentBlockMemberText{
				Value: brtypes.GuardrailTextBlock{
					Text: aws.String(text),
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("applying guardrail %s: %w", guardrailID, err)
	}

	result := &GuardrailTestResult{
		Action: string(resp.Action),
	}
	for _, o := range resp.Outputs {
		if o.Text != nil {
			result.Outputs = append(result.Outputs, aws.ToString(o.Text))
		}
	}
	for _, a := range resp.Assessments {
		assessment := GuardrailAssessment{}
		if a.TopicPolicy != nil {
			var topics []string
			for _, t := range a.TopicPolicy.Topics {
				topics = append(topics, string(t.Type)+":"+aws.ToString(t.Name))
			}
			assessment.TopicPolicy = strings.Join(topics, ", ")
		}
		if a.ContentPolicy != nil {
			var filters []string
			for _, f := range a.ContentPolicy.Filters {
				filters = append(filters, string(f.Type)+"("+string(f.Action)+")")
			}
			assessment.ContentPolicy = strings.Join(filters, ", ")
		}
		if a.SensitiveInformationPolicy != nil {
			var pii []string
			for _, p := range a.SensitiveInformationPolicy.PiiEntities {
				pii = append(pii, string(p.Type)+"("+string(p.Action)+")")
			}
			assessment.PIIPolicy = strings.Join(pii, ", ")
		}
		if a.WordPolicy != nil {
			var words []string
			for _, w := range a.WordPolicy.CustomWords {
				words = append(words, aws.ToString(w.Match)+"("+string(w.Action)+")")
			}
			assessment.WordPolicy = strings.Join(words, ", ")
		}
		result.Assessments = append(result.Assessments, assessment)
	}
	return result, nil
}

// bucketARN converts a bucket name to an S3 ARN.
func bucketARN(bucket string) string {
	if strings.HasPrefix(bucket, "arn:") {
		return bucket
	}
	return "arn:aws:s3:::" + bucket
}

// sanitizeName returns a name safe for OpenSearch index names.
func sanitizeName(name string) string {
	result := strings.ToLower(name)
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, result)
}
