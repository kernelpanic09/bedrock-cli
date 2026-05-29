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
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp      TEXT    NOT NULL,
    model          TEXT    NOT NULL,
    input_tokens   INTEGER,
    output_tokens  INTEGER,
    cost_usd       REAL,
    cached         BOOLEAN DEFAULT 0,
    template_name  TEXT,
    duration_ms    INTEGER,
    aws_account_id TEXT,
    aws_profile    TEXT,
    project        TEXT
);
`

// columns that may be missing on databases created before this schema version.
var migrateColumns = []string{
	"aws_account_id",
	"aws_profile",
	"project",
}

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
	AWSAccountID string
	AWSProfile   string
	Project      string
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

	if err := migrateSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Tracker{db: db}, nil
}

// migrateSchema adds any columns that exist in the current schema but are
// missing from an older database created before the column was added.
func migrateSchema(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(invocations)")
	if err != nil {
		return fmt.Errorf("reading table info: %w", err)
	}
	defer rows.Close()

	existing := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var defaultVal sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
			continue
		}
		existing[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scanning table info: %w", err)
	}

	for _, col := range migrateColumns {
		if !existing[col] {
			if _, err := db.Exec("ALTER TABLE invocations ADD COLUMN " + col + " TEXT"); err != nil {
				return fmt.Errorf("adding column %s: %w", col, err)
			}
		}
	}
	return nil
}

// Close closes the underlying database connection.
func (t *Tracker) Close() error {
	return t.db.Close()
}

// Record writes an invocation record to the database.
func (t *Tracker) Record(inv *Invocation) error {
	_, err := t.db.Exec(
		`INSERT INTO invocations
			(timestamp, model, input_tokens, output_tokens, cost_usd, cached, template_name, duration_ms, aws_account_id, aws_profile, project)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inv.Timestamp.UTC().Format(time.RFC3339),
		inv.Model,
		inv.InputTokens,
		inv.OutputTokens,
		inv.CostUSD,
		inv.Cached,
		inv.TemplateName,
		inv.DurationMs,
		inv.AWSAccountID,
		inv.AWSProfile,
		inv.Project,
	)
	if err != nil {
		return fmt.Errorf("recording invocation: %w", err)
	}
	return nil
}

// ProjectSummary holds aggregated cost stats for a single project tag.
type ProjectSummary struct {
	Project      string
	Calls        int
	InputTokens  int
	OutputTokens int
	TotalCost    float64
}

// AccountSummary holds aggregated cost stats for a single AWS account.
type AccountSummary struct {
	AccountID    string
	Calls        int
	InputTokens  int
	OutputTokens int
	TotalCost    float64
}

// Summary returns aggregated stats over a time window.
// If since is the zero value, all records are included.
func (t *Tracker) Summary(since time.Time) ([]Invocation, error) {
	query := `SELECT id, timestamp, model, input_tokens, output_tokens, cost_usd, cached, template_name, duration_ms, aws_account_id, aws_profile, project
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
		var templateName, accountID, profile, project sql.NullString
		if err := rows.Scan(
			&inv.ID, &tsStr, &inv.Model,
			&inv.InputTokens, &inv.OutputTokens, &inv.CostUSD,
			&inv.Cached, &templateName, &inv.DurationMs,
			&accountID, &profile, &project,
		); err != nil {
			return nil, fmt.Errorf("scanning invocation: %w", err)
		}
		if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
			inv.Timestamp = ts
		}
		inv.TemplateName = templateName.String
		inv.AWSAccountID = accountID.String
		inv.AWSProfile = profile.String
		inv.Project = project.String
		results = append(results, inv)
	}
	return results, rows.Err()
}

// ByProject returns per-project aggregated stats.
func (t *Tracker) ByProject(since time.Time) ([]ProjectSummary, error) {
	query := `
		SELECT
			COALESCE(project, '') AS project,
			COUNT(*) AS calls,
			SUM(input_tokens) AS input_tokens,
			SUM(output_tokens) AS output_tokens,
			SUM(cost_usd) AS total_cost
		FROM invocations`
	args := []any{}
	if !since.IsZero() {
		query += " WHERE timestamp >= ?"
		args = append(args, since.UTC().Format(time.RFC3339))
	}
	query += " GROUP BY project ORDER BY total_cost DESC"

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by-project: %w", err)
	}
	defer rows.Close()

	var results []ProjectSummary
	for rows.Next() {
		var s ProjectSummary
		if err := rows.Scan(&s.Project, &s.Calls, &s.InputTokens, &s.OutputTokens, &s.TotalCost); err != nil {
			return nil, fmt.Errorf("scanning project summary: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// ByAccount returns per-account aggregated stats.
func (t *Tracker) ByAccount(since time.Time) ([]AccountSummary, error) {
	query := `
		SELECT
			COALESCE(aws_account_id, 'unknown') AS account_id,
			COUNT(*) AS calls,
			SUM(input_tokens) AS input_tokens,
			SUM(output_tokens) AS output_tokens,
			SUM(cost_usd) AS total_cost
		FROM invocations`
	args := []any{}
	if !since.IsZero() {
		query += " WHERE timestamp >= ?"
		args = append(args, since.UTC().Format(time.RFC3339))
	}
	query += " GROUP BY aws_account_id ORDER BY total_cost DESC"

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by-account: %w", err)
	}
	defer rows.Close()

	var results []AccountSummary
	for rows.Next() {
		var s AccountSummary
		if err := rows.Scan(&s.AccountID, &s.Calls, &s.InputTokens, &s.OutputTokens, &s.TotalCost); err != nil {
			return nil, fmt.Errorf("scanning account summary: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}
