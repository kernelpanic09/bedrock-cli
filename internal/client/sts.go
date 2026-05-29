package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// CallerIdentity holds the result of STS GetCallerIdentity.
type CallerIdentity struct {
	AccountID string
	UserID    string
	ARN       string
}

// identityCache is a in-memory cache for CallerIdentity results keyed by region+profile.
var identityCache struct {
	mu      sync.Mutex
	entries map[string]identityCacheEntry
}

type identityCacheEntry struct {
	identity  *CallerIdentity
	expiresAt time.Time
}

const identityCacheTTL = time.Hour

// GetCallerIdentity returns the AWS identity for the current credential chain.
// Results are cached in-process for one hour to avoid redundant STS calls.
func GetCallerIdentity(ctx context.Context, region, profile string) (*CallerIdentity, error) {
	key := region + ":" + profile

	identityCache.mu.Lock()
	if identityCache.entries == nil {
		identityCache.entries = make(map[string]identityCacheEntry)
	}
	if e, ok := identityCache.entries[key]; ok && time.Now().Before(e.expiresAt) {
		identityCache.mu.Unlock()
		return e.identity, nil
	}
	identityCache.mu.Unlock()

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	svc := sts.NewFromConfig(cfg)
	resp, err := svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("getting caller identity: %w", err)
	}

	id := &CallerIdentity{
		AccountID: aws.ToString(resp.Account),
		UserID:    aws.ToString(resp.UserId),
		ARN:       aws.ToString(resp.Arn),
	}

	identityCache.mu.Lock()
	identityCache.entries[key] = identityCacheEntry{
		identity:  id,
		expiresAt: time.Now().Add(identityCacheTTL),
	}
	identityCache.mu.Unlock()

	return id, nil
}
