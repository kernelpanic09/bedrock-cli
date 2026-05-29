package cost

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/kernelpanic09/bedrock-cli/internal/config"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS invocations (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp     TEXT    NOT NULL,
    model         TEXT    NOT NULL,
    input_tokens  INTEGER,
    output_tokens INTEGER,
    cost_usd      REAL,
    cached        BOOLEAN DEFAULT 0,
    template_name TEXT,
    duration_ms   INTEGER
);
`

// Invocation is a single row in the cost tracking database.
type Invocation struct {
	ID           int64
	Timestamp    time.Time
	Model        string
	InputTokens  int
	OutputTokens int
	CostUSD      float64
	Cached       bool
	TemplateName string
	DurationMs   int64
}

// ModelSummary holds aggregated cost stats for a single model.
type ModelSummary struct {
	Model        string
	Calls        int
	InputTokens  int
	OutputTokens int
	TotalCost    float64
	CachedCalls  int
}

// Tracker wraps the SQLite cost database.
type Tracker struct {
	db *sql.DB
}

// Open opens (or creates) the usage database.
func Open() (*Tracker, error) {
	dir, err := config.CacheDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "usage.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening cost db: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing cost db schema: %w", err)
	}

	return &Tracker{db: db}, nil
}

// Close closes the underlying database connection.
func (t *Tracker) Close() error {
	return t.db.Close()
}

// Record writes an invocation record to the database.
func (t *Tracker) Record(inv *Invocation) error {
	_, err := t.db.Exec(
		`INSERT INTO invocations
			(timestamp, model, input_tokens, output_tokens, cost_usd, cached, template_name, duration_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		inv.Timestamp.UTC().Format(time.RFC3339),
		inv.Model,
		inv.InputTokens,
		inv.OutputTokens,
		inv.CostUSD,
		inv.Cached,
		inv.TemplateName,
		inv.DurationMs,
	)
	if err != nil {
		return fmt.Errorf("recording invocation: %w", err)
	}
	return nil
}

// Summary returns aggregated stats over a time window.
// If since is the zero value, all records are included.
func (t *Tracker) Summary(since time.Time) ([]Invocation, error) {
	query := `SELECT id, timestamp, model, input_tokens, output_tokens, cost_usd, cached, template_name, duration_ms
	          FROM invocations`
	args := []any{}

	if !since.IsZero() {
		query += " WHERE timestamp >= ?"
		args = append(args, since.UTC().Format(time.RFC3339))
	}
	query += " ORDER BY timestamp DESC"

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying invocations: %w", err)
	}
	defer rows.Close()

	return scanInvocations(rows)
}

// ByModel returns per-model aggregated stats.
func (t *Tracker) ByModel(since time.Time) ([]ModelSummary, error) {
	query := `
		SELECT
			model,
			COUNT(*) AS calls,
			SUM(input_tokens) AS input_tokens,
			SUM(output_tokens) AS output_tokens,
			SUM(cost_usd) AS total_cost,
			SUM(CASE WHEN cached = 1 THEN 1 ELSE 0 END) AS cached_calls
		FROM invocations`
	args := []any{}

	if !since.IsZero() {
		query += " WHERE timestamp >= ?"
		args = append(args, since.UTC().Format(time.RFC3339))
	}
	query += " GROUP BY model ORDER BY total_cost DESC"

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by-model: %w", err)
	}
	defer rows.Close()

	var results []ModelSummary
	for rows.Next() {
		var s ModelSummary
		if err := rows.Scan(
			&s.Model, &s.Calls, &s.InputTokens, &s.OutputTokens, &s.TotalCost, &s.CachedCalls,
		); err != nil {
			return nil, fmt.Errorf("scanning model summary: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// TotalCost returns the sum of all recorded costs.
func (t *Tracker) TotalCost(since time.Time) (float64, error) {
	query := "SELECT COALESCE(SUM(cost_usd), 0) FROM invocations"
	args := []any{}
	if !since.IsZero() {
		query += " WHERE timestamp >= ?"
		args = append(args, since.UTC().Format(time.RFC3339))
	}

	var total float64
	if err := t.db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("computing total cost: %w", err)
	}
	return total, nil
}

func scanInvocations(rows *sql.Rows) ([]Invocation, error) {
	var results []Invocation
	for rows.Next() {
		var inv Invocation
		var tsStr string
		var templateName sql.NullString
		if err := rows.Scan(
			&inv.ID, &tsStr, &inv.Model,
			&inv.InputTokens, &inv.OutputTokens, &inv.CostUSD,
			&inv.Cached, &templateName, &inv.DurationMs,
		); err != nil {
			return nil, fmt.Errorf("scanning invocation: %w", err)
		}
		if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
			inv.Timestamp = ts
		}
		inv.TemplateName = templateName.String
		results = append(results, inv)
	}
	return results, rows.Err()
}
