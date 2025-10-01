package audit

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	_ "modernc.org/sqlite"
)

// SqliteAuditor persists audit events to a SQLite database.
// It optionally hashes actor/object identifiers using HMAC-SHA256 (same truncation as in-memory store).
type SqliteAuditor struct {
	db             *sql.DB
	hasher         func(string) string
	maxRows        int           // 0 means unbounded (no pruning)
	anchorInterval int           // if >0 create anchor snapshot every anchorInterval events
	maxAge         time.Duration // if >0 events older than now-maxAge pruned
}

// IntegrityExport represents a snapshot of chain integrity state for external verification.
type IntegrityExport struct {
	Head struct {
		Seq int64  `json:"seq"`
		Hash string `json:"hash"`
		TS string   `json:"ts"`
	} `json:"head"`
	Anchors []struct {
		Seq int64  `json:"seq"`
		Hash string `json:"hash"`
		TS string   `json:"ts"`
	} `json:"anchors"`
	LatestSeq   int64  `json:"latest_seq"`
	EarliestSeq int64  `json:"earliest_seq"`
	GeneratedAt string `json:"generated_at"`
}

// ExportIntegrity gathers current head and up to limit most recent anchors (ascending order).
// limit <=0 defaults to 20.
func (a *SqliteAuditor) ExportIntegrity(ctx context.Context, limit int) (IntegrityExport, error) {
	if limit <= 0 { limit = 20 }
	var exp IntegrityExport
	now := time.Now().UTC().Format(time.RFC3339Nano)
	exp.GeneratedAt = now
	// Head (seq, ts, hash)
	row := a.db.QueryRowContext(ctx, `SELECT seq, ts, chain_hash FROM audit_events ORDER BY seq DESC LIMIT 1`)
	var headSeq int64
	var headTS, headHash sql.NullString
	_ = row.Scan(&headSeq, &headTS, &headHash)
	exp.Head.Seq = headSeq
	if headTS.Valid { exp.Head.TS = headTS.String }
	if headHash.Valid { exp.Head.Hash = headHash.String }
	// Earliest seq
	row2 := a.db.QueryRowContext(ctx, `SELECT seq FROM audit_events ORDER BY seq ASC LIMIT 1`)
	_ = row2.Scan(&exp.EarliestSeq)
	exp.LatestSeq = headSeq
	// Anchors newest first limited, then reverse to ascending
	rows, err := a.db.QueryContext(ctx, `SELECT seq, ts, chain_hash FROM audit_chain_anchors ORDER BY seq DESC LIMIT ?`, limit)
	if err != nil { return exp, err }
	defer rows.Close()
	tmp := []struct{Seq int64; TS, Hash string}{}
	for rows.Next() {
		var s int64; var ts, h string
		if err := rows.Scan(&s, &ts, &h); err != nil { return exp, err }
		tmp = append(tmp, struct{Seq int64; TS, Hash string}{Seq:s, TS:ts, Hash:h})
	}
	// reverse
	for i := len(tmp)-1; i >=0; i-- {
		aRec := struct{ Seq int64 `json:"seq"`; Hash string `json:"hash"`; TS string `json:"ts"` }{Seq: tmp[i].Seq, Hash: tmp[i].Hash, TS: tmp[i].TS}
		exp.Anchors = append(exp.Anchors, struct{ Seq int64 `json:"seq"`; Hash string `json:"hash"`; TS string `json:"ts"` }{Seq: aRec.Seq, Hash: aRec.Hash, TS: aRec.TS})
	}
	return exp, nil
}
// SqliteOption configures the SqliteAuditor.
type SqliteOption func(*SqliteAuditor)

func WithMaxAge(d time.Duration) SqliteOption {
	return func(a *SqliteAuditor) {
		if d > 0 {
			a.maxAge = d
		}
	}
}

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

// WithAnchorInterval enables periodic anchor snapshots (phase 2 hash chain). Every N events an anchor will
// record the seq, ts and chain_hash in a separate anchor table for logarithmic verification windows.
// If n <= 0 anchors are disabled.
func WithAnchorInterval(n int) SqliteOption {
	return func(a *SqliteAuditor) {
		if n > 0 {
			a.anchorInterval = n
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
	CREATE INDEX IF NOT EXISTS idx_audit_events_action_ts ON audit_events(action, ts);
	CREATE TABLE IF NOT EXISTS audit_chain_anchors (
		seq INTEGER PRIMARY KEY, -- references audit_events.seq
		ts TEXT NOT NULL,
		chain_hash TEXT NOT NULL
	);`
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
	if status == "success" {
		metrics.IncChainHead()
	}

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

	// Time-based pruning (age retention) independent of row cap.
	if a.maxAge > 0 {
		cutoff := time.Now().Add(-a.maxAge).UTC().Format(time.RFC3339Nano)
		res, err := a.db.ExecContext(ctx, `DELETE FROM audit_events WHERE ts < ?`, cutoff)
		if err == nil {
			if rows, _ := res.RowsAffected(); rows > 0 {
				metrics.AddAuditEviction("sqlite", int(rows))
			}
		} else {
			metrics.IncAuditFailure("prune_age")
		}
		// Prune orphan anchors whose seq no longer exists
		_, err = a.db.ExecContext(ctx, `DELETE FROM audit_chain_anchors WHERE seq NOT IN (SELECT seq FROM audit_events)`)
		if err != nil {
			metrics.IncAuditFailure("prune_anchor")
		}
	}

	// Anchor snapshot creation: retrieve seq of inserted row if anchors enabled and success.
	if status == "success" && a.anchorInterval > 0 {
		// Get last seq (the one we just inserted)
		row := a.db.QueryRowContext(ctx, `SELECT seq FROM audit_events ORDER BY seq DESC LIMIT 1`)
		var seq int64
		if err := row.Scan(&seq); err == nil {
			if seq%int64(a.anchorInterval) == 0 {
				// Insert anchor snapshot (ignore unique conflict in rare race or re-entry)
				_, err := a.db.ExecContext(ctx, `INSERT OR IGNORE INTO audit_chain_anchors(seq, ts, chain_hash) VALUES(?,?,?)`, seq, ts, chainHash)
				if err != nil {
					metrics.IncAuditFailure("anchor_insert")
				} else {
					metrics.IncChainAnchor()
				}
			}
		}
	}
}

func (a *SqliteAuditor) getLastChainHash(ctx context.Context) string {
	row := a.db.QueryRowContext(ctx, `SELECT chain_hash FROM audit_events ORDER BY seq DESC LIMIT 1`)
	var h sql.NullString
	_ = row.Scan(&h)
	if h.Valid {
		return h.String
	}
	return ""
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func canonicalizeDetails(m map[string]any) string {
	if len(m) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		valBytes, _ := json.Marshal(m[k])
		b.WriteString(fmt.Sprintf("%q:%s", k, string(valBytes)))
		if i < len(keys)-1 {
			b.WriteByte(',')
		}
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
	if err != nil {
		return err
	}
	defer rows.Close()
	prev := "" // baseline (empty) accepted for first retained segment
	index := 0
	for rows.Next() {
		var ts, action, actor, object, detailsJSON, rid, storedHash sql.NullString
		if err := rows.Scan(&ts, &action, &actor, &object, &detailsJSON, &rid, &storedHash); err != nil {
			return err
		}
		detMap := map[string]any{}
		if detailsJSON.Valid && detailsJSON.String != "" {
			_ = json.Unmarshal([]byte(detailsJSON.String), &detMap)
		}
		canonicalDetails := canonicalizeDetails(detMap)
		chainInput := prev + "|" + ts.String + "|" + action.String + "|" + actor.String + "|" + object.String + "|" + canonicalDetails + "|" + rid.String
		recomputed := sha256Hex(chainInput)
		if index == 0 && prev == "" {
			// Baseline acceptance: first retained row after pruning may reference a missing prior hash.
			// Accept stored hash without recomputation match requirement; seed prev with stored value.
			if !storedHash.Valid {
				metrics.ObserveChainVerification(time.Since(start).Seconds(), false)
				return fmt.Errorf("chain baseline missing hash at first retained row")
			}
			prev = storedHash.String
		} else {
			if !storedHash.Valid || storedHash.String != recomputed {
				metrics.ObserveChainVerification(time.Since(start).Seconds(), false)
				return fmt.Errorf("chain mismatch at index %d seq hash=%s expected=%s", index, storedHash.String, recomputed)
			}
			prev = recomputed
		}
		index++
	}
	metrics.ObserveChainVerification(time.Since(start).Seconds(), true)
	return nil
}

// ListAnchors returns anchor (seq, ts, chain_hash) ordered ascending.
func (a *SqliteAuditor) ListAnchors(ctx context.Context) ([]struct {
	Seq  int64
	TS   string
	Hash string
}, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT seq, ts, chain_hash FROM audit_chain_anchors ORDER BY seq ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Seq  int64
		TS   string
		Hash string
	}
	for rows.Next() {
		var seq int64
		var ts, h string
		if err := rows.Scan(&seq, &ts, &h); err != nil {
			return nil, err
		}
		out = append(out, struct {
			Seq  int64
			TS   string
			Hash string
		}{Seq: seq, TS: ts, Hash: h})
	}
	return out, nil
}

// VerifyFromAnchor verifies chain starting at the first event after (or equal) anchorSeq until head.
// If anchorSeq == 0, behaves like full VerifyChain. It recomputes hashes starting with the hash stored at anchorSeq-1.
func (a *SqliteAuditor) VerifyFromAnchor(ctx context.Context, anchorSeq int64) error {
	if anchorSeq <= 0 {
		return a.VerifyChain(ctx)
	}
	start := time.Now()
	// Reuse existing chain verification metrics (duration + failure counter) for partial verification
	// to avoid metric cardinality explosion; a label could be added in future if separation needed.
	// Get previous hash (hash at last event < anchorSeq)
	var prev string // baseline if prior event pruned
	if anchorSeq > 1 {
		row := a.db.QueryRowContext(ctx, `SELECT chain_hash FROM audit_events WHERE seq = ?`, anchorSeq-1)
		_ = row.Scan(&prev) // if missing treat empty
	}
	rows, err := a.db.QueryContext(ctx, `SELECT seq, ts, action, actor, object, details, request_id, chain_hash FROM audit_events WHERE seq >= ? ORDER BY seq ASC`, anchorSeq)
	if err != nil {
		return err
	}
	defer rows.Close()
	index := 0
	for rows.Next() {
		var seq int64
		var ts, action, actor, object, detailsJSON, rid, storedHash sql.NullString
		if err := rows.Scan(&seq, &ts, &action, &actor, &object, &detailsJSON, &rid, &storedHash); err != nil {
			return err
		}
		detMap := map[string]any{}
		if detailsJSON.Valid && detailsJSON.String != "" {
			_ = json.Unmarshal([]byte(detailsJSON.String), &detMap)
		}
		canonicalDetails := canonicalizeDetails(detMap)
		chainInput := prev + "|" + ts.String + "|" + action.String + "|" + actor.String + "|" + object.String + "|" + canonicalDetails + "|" + rid.String
		recomputed := sha256Hex(chainInput)
		if index == 0 && prev == "" {
			if !storedHash.Valid {
				metrics.ObserveChainVerification(time.Since(start).Seconds(), false)
				return fmt.Errorf("partial chain baseline missing hash at first retained row seq %d", seq)
			}
			prev = storedHash.String
		} else {
			if !storedHash.Valid || storedHash.String != recomputed {
				metrics.ObserveChainVerification(time.Since(start).Seconds(), false)
				return fmt.Errorf("partial chain mismatch at index %d seq %d hash=%s expected=%s", index, seq, storedHash.String, recomputed)
			}
			prev = recomputed
		}
		index++
	}
	metrics.ObserveChainVerification(time.Since(start).Seconds(), true)
	return nil
}

// ErrNotSupported indicates the sqlite auditor is not configured.
var ErrNotSupported = errors.New("sqlite auditor not configured")
