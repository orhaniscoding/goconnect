package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	_ "modernc.org/sqlite"
)

// SqliteAuditor persists audit events to a SQLite database.
// It optionally hashes actor/object identifiers using HMAC-SHA256 (same truncation as in-memory store).
type SqliteAuditor struct {
	db     *sql.DB
	hasher func(string) string
}

// SqliteOption configures the SqliteAuditor.
type SqliteOption func(*SqliteAuditor)

// WithSqliteHashing enables pseudonymous hashing for actor/object fields.
func WithSqliteHashing(secret []byte) SqliteOption {
	return func(a *SqliteAuditor) { a.hasher = newHasher(secret) }
}

// NewSqliteAuditor opens (or creates) the SQLite database at dsn (e.g. file path) and ensures schema.
func NewSqliteAuditor(dsn string, opts ...SqliteOption) (*SqliteAuditor, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	a := &SqliteAuditor{db: db}
	for _, o := range opts {
		o(a)
	}
	if err := a.migrate(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *SqliteAuditor) migrate() error {
	schema := `CREATE TABLE IF NOT EXISTS audit_events (
        seq INTEGER PRIMARY KEY AUTOINCREMENT,
        ts TEXT NOT NULL,
        action TEXT NOT NULL,
        actor TEXT NOT NULL,
        object TEXT NOT NULL,
        details TEXT,
        request_id TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_audit_events_action_ts ON audit_events(action, ts);`
	_, err := a.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate audit_events: %w", err)
	}
	return nil
}

// Close releases database resources.
func (a *SqliteAuditor) Close() error { return a.db.Close() }

// Event implements the Auditor interface.
func (a *SqliteAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	rid, _ := ctx.Value("request_id").(string)
	actOut := "[redacted]"
	objOut := "[redacted]"
	if a.hasher != nil {
		actOut = a.hasher(actor)
		objOut = a.hasher(object)
	}
	b, _ := json.Marshal(details)
	if _, err := a.db.ExecContext(ctx, `INSERT INTO audit_events(ts, action, actor, object, details, request_id) VALUES(?,?,?,?,?,?)`,
		time.Now().UTC().Format(time.RFC3339Nano), action, actOut, objOut, string(b), rid); err != nil {
		metrics.IncAuditFailure()
	}
}

// ListRecent returns up to limit most recent events (descending ts). Intended for tests/diagnostics.
func (a *SqliteAuditor) ListRecent(ctx context.Context, limit int) ([]EventRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := a.db.QueryContext(ctx, `SELECT ts, action, actor, object, details, request_id FROM audit_events ORDER BY seq DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EventRecord{}
	for rows.Next() {
		var (
			tsStr                                   string
			action, actor, object, detailsJSON, rid sql.NullString
		)
		if err := rows.Scan(&tsStr, &action, &actor, &object, &detailsJSON, &rid); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339Nano, tsStr)
		det := map[string]any{}
		if detailsJSON.Valid && detailsJSON.String != "" {
			_ = json.Unmarshal([]byte(detailsJSON.String), &det)
		}
		out = append(out, EventRecord{TS: t, Action: action.String, Actor: actor.String, Object: object.String, Details: det, RequestID: rid.String})
	}
	return out, nil
}

// Count returns total number of audit events (for tests).
func (a *SqliteAuditor) Count(ctx context.Context) (int64, error) {
	row := a.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM audit_events`)
	var c int64
	if err := row.Scan(&c); err != nil {
		return 0, err
	}
	return c, nil
}

// ErrNotSupported indicates the sqlite auditor is not configured.
var ErrNotSupported = errors.New("sqlite auditor not configured")
