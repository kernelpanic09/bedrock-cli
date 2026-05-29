package cost

import (
	"testing"
	"time"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		model        string
		inputTokens  int
		outputTokens int
	}{
		{"haiku", 1000, 500},
		{"sonnet", 1000, 1000},
		{"unknown-model", 1000, 1000},
	}

	for _, tt := range tests {
		cost := Calculate(tt.model, tt.inputTokens, tt.outputTokens)
		if cost < 0 {
			t.Errorf("Calculate(%q, %d, %d) returned negative cost: %.6f",
				tt.model, tt.inputTokens, tt.outputTokens, cost)
		}
	}
}

func TestCalculateHaikuPricing(t *testing.T) {
	// haiku: $0.00025/1k input, $0.00125/1k output
	got := Calculate("haiku", 1000, 1000)
	want := 0.00025 + 0.00125
	if abs(got-want) > 1e-9 {
		t.Errorf("Calculate(haiku, 1000, 1000) = %.6f, want %.6f", got, want)
	}
}

func TestCalculateUnknownModelReturnsZero(t *testing.T) {
	got := Calculate("not-a-real-model", 100, 100)
	if got != 0.0 {
		t.Errorf("Calculate(unknown) = %.6f, want 0.0", got)
	}
}

func TestRecordAndSummary(t *testing.T) {
	tracker := newInMemoryTracker(t)

	now := time.Now().UTC()
	invs := []Invocation{
		{
			Timestamp:    now.Add(-2 * time.Hour),
			Model:        "haiku",
			InputTokens:  100,
			OutputTokens: 200,
			CostUSD:      0.001,
		},
		{
			Timestamp:    now.Add(-1 * time.Hour),
			Model:        "sonnet",
			InputTokens:  500,
			OutputTokens: 300,
			CostUSD:      0.006,
			Cached:       true,
		},
	}

	for i := range invs {
		if err := tracker.Record(&invs[i]); err != nil {
			t.Fatalf("Record() error: %v", err)
		}
	}

	all, err := tracker.Summary(time.Time{})
	if err != nil {
		t.Fatalf("Summary() error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("Summary() returned %d rows, want 2", len(all))
	}
}

func TestByModel(t *testing.T) {
	tracker := newInMemoryTracker(t)

	now := time.Now().UTC()
	records := []Invocation{
		{Timestamp: now, Model: "haiku", InputTokens: 100, OutputTokens: 100, CostUSD: 0.001},
		{Timestamp: now, Model: "haiku", InputTokens: 200, OutputTokens: 200, CostUSD: 0.002, Cached: true},
		{Timestamp: now, Model: "sonnet", InputTokens: 500, OutputTokens: 300, CostUSD: 0.006},
	}
	for i := range records {
		if err := tracker.Record(&records[i]); err != nil {
			t.Fatalf("Record() error: %v", err)
		}
	}

	summaries, err := tracker.ByModel(time.Time{})
	if err != nil {
		t.Fatalf("ByModel() error: %v", err)
	}

	if len(summaries) != 2 {
		t.Errorf("ByModel() returned %d rows, want 2", len(summaries))
	}

	for _, s := range summaries {
		if s.Model == "haiku" {
			if s.Calls != 2 {
				t.Errorf("haiku calls = %d, want 2", s.Calls)
			}
			if s.CachedCalls != 1 {
				t.Errorf("haiku cached = %d, want 1", s.CachedCalls)
			}
		}
	}
}

func TestSummaryWithSinceFilter(t *testing.T) {
	tracker := newInMemoryTracker(t)

	now := time.Now().UTC()
	old := now.Add(-48 * time.Hour)
	recent := now.Add(-1 * time.Hour)

	if err := tracker.Record(&Invocation{Timestamp: old, Model: "haiku", CostUSD: 0.001}); err != nil {
		t.Fatalf("Record() error: %v", err)
	}
	if err := tracker.Record(&Invocation{Timestamp: recent, Model: "sonnet", CostUSD: 0.002}); err != nil {
		t.Fatalf("Record() error: %v", err)
	}

	since := now.Add(-24 * time.Hour)
	results, err := tracker.Summary(since)
	if err != nil {
		t.Fatalf("Summary(since) error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Summary(since 24h) returned %d rows, want 1", len(results))
	}
	if len(results) > 0 && results[0].Model != "sonnet" {
		t.Errorf("expected sonnet in filtered results, got %q", results[0].Model)
	}
}

func TestTotalCost(t *testing.T) {
	tracker := newInMemoryTracker(t)

	now := time.Now().UTC()
	records := []Invocation{
		{Timestamp: now, Model: "haiku", CostUSD: 0.001},
		{Timestamp: now, Model: "sonnet", CostUSD: 0.005},
	}
	for i := range records {
		if err := tracker.Record(&records[i]); err != nil {
			t.Fatalf("Record() error: %v", err)
		}
	}

	total, err := tracker.TotalCost(time.Time{})
	if err != nil {
		t.Fatalf("TotalCost() error: %v", err)
	}
	want := 0.006
	if abs(total-want) > 1e-9 {
		t.Errorf("TotalCost() = %.6f, want %.6f", total, want)
	}
}

// newInMemoryTracker creates a tracker backed by an in-memory SQLite db.
func newInMemoryTracker(t *testing.T) *Tracker {
	t.Helper()
	tracker, err := openPath(":memory:")
	if err != nil {
		t.Fatalf("openPath(:memory:) error: %v", err)
	}
	t.Cleanup(func() { tracker.Close() })
	return tracker
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
