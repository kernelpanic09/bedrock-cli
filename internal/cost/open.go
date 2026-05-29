package cost

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// openPath opens or creates a tracker at the given filesystem path.
// Passing ":memory:" gives a transient in-memory database (useful in tests).
func openPath(path string) (*Tracker, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening cost db at %s: %w", path, err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing schema at %s: %w", path, err)
	}
	return &Tracker{db: db}, nil
}
