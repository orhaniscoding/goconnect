package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	_ "modernc.org/sqlite"
)

// SqliteAuditor persists audit events to a SQLite database.
// It optionally hashes actor/object identifiers using HMAC-SHA256 (same truncation as in-memory store).
type SqliteAuditor struct {
	db      *sql.DB
	hasher  func(string) string
	maxRows int // 0 means unbounded (no pruning)
}

// SqliteOption configures the SqliteAuditor.
type SqliteOption func(*SqliteAuditor)

// WithSqliteHashing enables pseudonymous hashing for actor/object fields.
func WithSqliteHashing(secret []byte) SqliteOption {
	return func(a *SqliteAuditor) { a.hasher = newHasher(secret) }
}

// WithMaxRows sets a hard cap on stored rows; rows beyond the cap are pruned after insert.
// Pruning strategy: delete oldest rows so that total <= maxRows (single DELETE with subquery).
func WithMaxRows(n int) SqliteOption {
	return func(a *SqliteAuditor) {
		if n > 0 {
			a.maxRows = n
		}
	}
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
		request_id TEXT,
		chain_hash TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_audit_events_action_ts ON audit_events(action, ts);`
	if _, err := a.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate audit_events: %w", err)
	}
	// attempt add column if legacy table (ignore error if already exists)
	_, _ = a.db.Exec(`ALTER TABLE audit_events ADD COLUMN chain_hash TEXT`)
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
	start := time.Now()
	status := "success"
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	// Build canonical JSON for hashing chain: ordered keys for details to ensure deterministic serialization
	canonicalDetails := canonicalizeDetails(details)
	prevHash := a.getLastChainHash(ctx)
	chainInput := prevHash + "|" + ts + "|" + action + "|" + actOut + "|" + objOut + "|" + canonicalDetails + "|" + rid
	chainHash := sha256Hex(chainInput)
	if _, err := a.db.ExecContext(ctx, `INSERT INTO audit_events(ts, action, actor, object, details, request_id, chain_hash) VALUES(?,?,?,?,?,?,?)`,
		ts, action, actOut, objOut, string(b), rid, chainHash); err != nil {
		metrics.IncAuditFailure("exec")
		status = "failure"
	}
	metrics.ObserveAuditInsert("sqlite", status, time.Since(start).Seconds())
	if status == "success" { metrics.IncChainHead() }

	// Post-insert pruning if retention limit configured.
	if a.maxRows > 0 {
		// Delete excess oldest rows keeping newest maxRows based on seq ordering.
		// Use a CTE to select seq values to delete.
		res, err := a.db.ExecContext(ctx, `WITH to_delete AS (
			SELECT seq FROM audit_events ORDER BY seq ASC
			LIMIT (SELECT CASE WHEN COUNT(1) > ? THEN COUNT(1) - ? ELSE 0 END FROM audit_events)
		) DELETE FROM audit_events WHERE seq IN (SELECT seq FROM to_delete);`, a.maxRows, a.maxRows)
		if err == nil {
			if rows, _ := res.RowsAffected(); rows > 0 {
				metrics.AddAuditEviction("sqlite", int(rows))
			}
		} else {
			metrics.IncAuditFailure("prune")
		}
	}
}

func (a *SqliteAuditor) getLastChainHash(ctx context.Context) string {
	row := a.db.QueryRowContext(ctx, `SELECT chain_hash FROM audit_events ORDER BY seq DESC LIMIT 1`)
	var h sql.NullString
	_ = row.Scan(&h)
	if h.Valid { return h.String }
	return ""
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func canonicalizeDetails(m map[string]any) string {
	if m == nil || len(m) == 0 { return "{}" }
	keys := make([]string,0,len(m))
	for k := range m { keys = append(keys,k) }
	sort.Strings(keys)
	var b strings.Builder
	b.WriteByte('{')
	for i,k := range keys {
		valBytes, _ := json.Marshal(m[k])
		b.WriteString(fmt.Sprintf("%q:%s", k, string(valBytes)))
		if i < len(keys)-1 { b.WriteByte(',') }
	}
	b.WriteByte('}')
	return b.String()
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

// VerifyChain walks the chain sequentially validating each chain_hash. Returns first mismatch error.
func (a *SqliteAuditor) VerifyChain(ctx context.Context) error {
	start := time.Now()
	rows, err := a.db.QueryContext(ctx, `SELECT ts, action, actor, object, details, request_id, chain_hash FROM audit_events ORDER BY seq ASC`)
	if err != nil { return err }
	defer rows.Close()
	prev := ""
	index := 0
	for rows.Next() {
		var ts, action, actor, object, detailsJSON, rid, storedHash sql.NullString
		if err := rows.Scan(&ts, &action, &actor, &object, &detailsJSON, &rid, &storedHash); err != nil { return err }
		detMap := map[string]any{}
		if detailsJSON.Valid && detailsJSON.String != "" { _ = json.Unmarshal([]byte(detailsJSON.String), &detMap) }
		canonicalDetails := canonicalizeDetails(detMap)
		chainInput := prev + "|" + ts.String + "|" + action.String + "|" + actor.String + "|" + object.String + "|" + canonicalDetails + "|" + rid.String
		recomputed := sha256Hex(chainInput)
		if !storedHash.Valid || storedHash.String != recomputed {
			metrics.ObserveChainVerification(time.Since(start).Seconds(), false)
			return fmt.Errorf("chain mismatch at index %d seq hash=%s expected=%s", index, storedHash.String, recomputed)
		}
		prev = recomputed
		index++
	}
	metrics.ObserveChainVerification(time.Since(start).Seconds(), true)
	return nil
}

// ErrNotSupported indicates the sqlite auditor is not configured.
var ErrNotSupported = errors.New("sqlite auditor not configured")
