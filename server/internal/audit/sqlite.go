package audit

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
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

// SqliteAuditor persists audit events to SQLite with a verifiable hash chain and optional anchors.
type SqliteAuditor struct {
	db             *sql.DB
	hasher         func(string) string
	anchorInterval int
	maxAge         time.Duration
	maxRows        int
	signingKey     ed25519.PrivateKey
	signingKeyID   string
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
func WithMaxRows(n int) SqliteOption {
	return func(a *SqliteAuditor) {
		if n > 0 {
			a.maxRows = n
		}
	}
}
func WithAnchorInterval(n int) SqliteOption {
	return func(a *SqliteAuditor) {
		if n > 0 {
			a.anchorInterval = n
		}
	}
}
func WithSqliteHashing(secret []byte) SqliteOption {
	return func(a *SqliteAuditor) { a.hasher = newHasher(secret) }
}
func WithSqliteHashSecrets(secrets ...[]byte) SqliteOption {
	return func(a *SqliteAuditor) {
		h, _, _ := multiSecretHasher(secrets)
		if h != nil {
			a.hasher = h
		}
	}
}
func WithIntegritySigningKey(priv ed25519.PrivateKey) SqliteOption {
	return func(a *SqliteAuditor) {
		if len(priv) == ed25519.PrivateKeySize {
			a.signingKey = priv
		}
	}
}
func WithIntegritySigningKeyID(kid string, priv ed25519.PrivateKey) SqliteOption {
	return func(a *SqliteAuditor) {
		if len(priv) == ed25519.PrivateKeySize {
			a.signingKey = priv
			a.signingKeyID = kid
		}
	}
}

// NewSqliteAuditor opens the SQLite database and ensures schema exists.
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

// Close releases database resources.
func (a *SqliteAuditor) Close() error { return a.db.Close() }

func (a *SqliteAuditor) migrate() error {
	schema := `CREATE TABLE IF NOT EXISTS audit_events (
        seq INTEGER PRIMARY KEY AUTOINCREMENT,
        ts TEXT NOT NULL,
        tenant_id TEXT NOT NULL DEFAULT '',
        action TEXT NOT NULL,
        actor TEXT NOT NULL,
        object TEXT NOT NULL,
        details TEXT,
        request_id TEXT,
        chain_hash TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_audit_events_action_ts ON audit_events(action, ts);
    CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_ts ON audit_events(tenant_id, ts);
    CREATE TABLE IF NOT EXISTS audit_chain_anchors (
        seq INTEGER PRIMARY KEY,
        ts TEXT NOT NULL,
        chain_hash TEXT NOT NULL
    );`
	if _, err := a.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate audit_events: %w", err)
	}
	// If legacy table existed without chain_hash, ignore error.
	_, _ = a.db.Exec(`ALTER TABLE audit_events ADD COLUMN chain_hash TEXT`)
	// If legacy table existed without tenant_id, ignore error.
	_, _ = a.db.Exec(`ALTER TABLE audit_events ADD COLUMN tenant_id TEXT NOT NULL DEFAULT ''`)
	return nil
}

// Event records an audit event, building a verifiable hash chain and optional anchor.
func (a *SqliteAuditor) Event(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	rid, _ := ctx.Value("request_id").(string)
	actOut, objOut := "[redacted]", "[redacted]"
	if a.hasher != nil {
		actOut = a.hasher(actor)
		objOut = a.hasher(object)
	}
	b, _ := json.Marshal(details)
	start := time.Now()
	status := "success"
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	canonicalDetails := canonicalizeDetails(details)
	prevHash := a.getLastChainHash(ctx)
	// Include tenantID in chain input for integrity
	chainInput := prevHash + "|" + ts + "|" + tenantID + "|" + action + "|" + actOut + "|" + objOut + "|" + canonicalDetails + "|" + rid
	chainHash := sha256Hex(chainInput)
	if _, err := a.db.ExecContext(ctx, `INSERT INTO audit_events(ts, tenant_id, action, actor, object, details, request_id, chain_hash) VALUES(?,?,?,?,?,?,?,?)`, ts, tenantID, action, actOut, objOut, string(b), rid, chainHash); err != nil {
		metrics.IncAuditFailure("exec")
		status = "failure"
	}
	metrics.ObserveAuditInsert("sqlite", status, time.Since(start).Seconds())
	if status == "success" {
		metrics.IncChainHead()
	}

	if a.maxRows > 0 {
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
		if _, err := a.db.ExecContext(ctx, `DELETE FROM audit_chain_anchors WHERE seq NOT IN (SELECT seq FROM audit_events)`); err != nil {
			metrics.IncAuditFailure("prune_anchor")
		}
	}

	if status == "success" && a.anchorInterval > 0 {
		row := a.db.QueryRowContext(ctx, `SELECT seq FROM audit_events ORDER BY seq DESC LIMIT 1`)
		var seq int64
		if err := row.Scan(&seq); err == nil {
			if seq%int64(a.anchorInterval) == 0 {
				if _, err := a.db.ExecContext(ctx, `INSERT OR IGNORE INTO audit_chain_anchors(seq, ts, chain_hash) VALUES(?,?,?)`, seq, ts, chainHash); err != nil {
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

func sha256Hex(s string) string { sum := sha256.Sum256([]byte(s)); return hex.EncodeToString(sum[:]) }

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

// ListRecent returns recent events for diagnostics/tests.
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
		var tsStr string
		var action, actor, object, detailsJSON, rid sql.NullString
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

// Count returns total number of audit events.
func (a *SqliteAuditor) Count(ctx context.Context) (int64, error) {
	row := a.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM audit_events`)
	var c int64
	if err := row.Scan(&c); err != nil {
		return 0, err
	}
	return c, nil
}

// IntegrityExport is a compact integrity snapshot, optionally signed.
type IntegrityExport struct {
	Head struct {
		Seq  int64  `json:"seq"`
		Hash string `json:"hash"`
		TS   string `json:"ts"`
	} `json:"head"`
	Anchors []struct {
		Seq  int64  `json:"seq"`
		Hash string `json:"hash"`
		TS   string `json:"ts"`
	} `json:"anchors"`
	LatestSeq   int64  `json:"latest_seq"`
	EarliestSeq int64  `json:"earliest_seq"`
	GeneratedAt string `json:"generated_at"`
	Signature   string `json:"signature,omitempty"`
	KeyID       string `json:"kid,omitempty"`
}

// ExportIntegrity builds an integrity snapshot and optionally signs it.
func (a *SqliteAuditor) ExportIntegrity(ctx context.Context, limit int) (IntegrityExport, error) {
	if limit <= 0 {
		limit = 20
	}
	var exp IntegrityExport
	exp.GeneratedAt = time.Now().UTC().Format(time.RFC3339Nano)
	row := a.db.QueryRowContext(ctx, `SELECT seq, ts, chain_hash FROM audit_events ORDER BY seq DESC LIMIT 1`)
	var headSeq int64
	var headTS, headHash sql.NullString
	_ = row.Scan(&headSeq, &headTS, &headHash)
	exp.Head.Seq = headSeq
	if headTS.Valid {
		exp.Head.TS = headTS.String
	}
	if headHash.Valid {
		exp.Head.Hash = headHash.String
	}
	row2 := a.db.QueryRowContext(ctx, `SELECT seq FROM audit_events ORDER BY seq ASC LIMIT 1`)
	_ = row2.Scan(&exp.EarliestSeq)
	exp.LatestSeq = headSeq
	rows, err := a.db.QueryContext(ctx, `SELECT seq, ts, chain_hash FROM audit_chain_anchors ORDER BY seq DESC LIMIT ?`, limit)
	if err != nil {
		return exp, err
	}
	defer rows.Close()
	tmp := []struct {
		Seq      int64
		TS, Hash string
	}{}
	for rows.Next() {
		var s int64
		var ts, h string
		if err := rows.Scan(&s, &ts, &h); err != nil {
			return exp, err
		}
		tmp = append(tmp, struct {
			Seq      int64
			TS, Hash string
		}{Seq: s, TS: ts, Hash: h})
	}
	for i := len(tmp) - 1; i >= 0; i-- {
		aRec := struct {
			Seq  int64  `json:"seq"`
			Hash string `json:"hash"`
			TS   string `json:"ts"`
		}{Seq: tmp[i].Seq, Hash: tmp[i].Hash, TS: tmp[i].TS}
		exp.Anchors = append(exp.Anchors, struct {
			Seq  int64  `json:"seq"`
			Hash string `json:"hash"`
			TS   string `json:"ts"`
		}{Seq: aRec.Seq, Hash: aRec.Hash, TS: aRec.TS})
	}
	if len(a.signingKey) == ed25519.PrivateKeySize {
		if a.signingKeyID != "" {
			exp.KeyID = a.signingKeyID
		}
		tmpExp := exp
		tmpExp.Signature = ""
		payload, err := json.Marshal(tmpExp)
		if err == nil {
			sig := ed25519.Sign(a.signingKey, payload)
			exp.Signature = base64.RawURLEncoding.EncodeToString(sig)
			metrics.IncIntegritySigned()
		} else {
			metrics.IncAuditFailure("integrity_sign")
		}
	}
	return exp, nil
}

// VerifyChain recomputes the chain and detects the first mismatch.
func (a *SqliteAuditor) VerifyChain(ctx context.Context) error {
	start := time.Now()
	rows, err := a.db.QueryContext(ctx, `SELECT ts, tenant_id, action, actor, object, details, request_id, chain_hash FROM audit_events ORDER BY seq ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	prev := ""
	index := 0
	for rows.Next() {
		var ts, tenantID, action, actor, object, detailsJSON, rid, storedHash sql.NullString
		if err := rows.Scan(&ts, &tenantID, &action, &actor, &object, &detailsJSON, &rid, &storedHash); err != nil {
			return err
		}
		detMap := map[string]any{}
		if detailsJSON.Valid && detailsJSON.String != "" {
			_ = json.Unmarshal([]byte(detailsJSON.String), &detMap)
		}
		canonicalDetails := canonicalizeDetails(detMap)
		chainInput := prev + "|" + ts.String + "|" + tenantID.String + "|" + action.String + "|" + actor.String + "|" + object.String + "|" + canonicalDetails + "|" + rid.String
		recomputed := sha256Hex(chainInput)
		if index == 0 && prev == "" {
			if !storedHash.Valid {
				metrics.ObserveChainVerification(time.Since(start).Seconds(), false)
				return fmt.Errorf("chain baseline missing hash at first retained row")
			}
			prev = storedHash.String
		} else {
			if !storedHash.Valid || storedHash.String != recomputed {
				metrics.ObserveChainVerification(time.Since(start).Seconds(), false)
				return fmt.Errorf("chain mismatch at index %d hash=%s expected=%s", index, storedHash.String, recomputed)
			}
			prev = recomputed
		}
		index++
	}
	metrics.ObserveChainVerification(time.Since(start).Seconds(), true)
	return nil
}

// ListAnchors returns anchors ordered ascending.
func (a *SqliteAuditor) ListAnchors(ctx context.Context) ([]struct {
	Seq      int64
	TS, Hash string
}, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT seq, ts, chain_hash FROM audit_chain_anchors ORDER BY seq ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Seq      int64
		TS, Hash string
	}
	for rows.Next() {
		var seq int64
		var ts, h string
		if err := rows.Scan(&seq, &ts, &h); err != nil {
			return nil, err
		}
		out = append(out, struct {
			Seq      int64
			TS, Hash string
		}{Seq: seq, TS: ts, Hash: h})
	}
	return out, nil
}

// VerifyFromAnchor verifies starting at anchorSeq.
func (a *SqliteAuditor) VerifyFromAnchor(ctx context.Context, anchorSeq int64) error {
	if anchorSeq <= 0 {
		return a.VerifyChain(ctx)
	}
	start := time.Now()
	var prev string
	if anchorSeq > 1 {
		row := a.db.QueryRowContext(ctx, `SELECT chain_hash FROM audit_events WHERE seq = ?`, anchorSeq-1)
		_ = row.Scan(&prev)
	}
	rows, err := a.db.QueryContext(ctx, `SELECT seq, ts, tenant_id, action, actor, object, details, request_id, chain_hash FROM audit_events WHERE seq >= ? ORDER BY seq ASC`, anchorSeq)
	if err != nil {
		return err
	}
	defer rows.Close()
	index := 0
	for rows.Next() {
		var seq int64
		var ts, tenantID, action, actor, object, detailsJSON, rid, storedHash sql.NullString
		if err := rows.Scan(&seq, &ts, &tenantID, &action, &actor, &object, &detailsJSON, &rid, &storedHash); err != nil {
			return err
		}
		detMap := map[string]any{}
		if detailsJSON.Valid && detailsJSON.String != "" {
			_ = json.Unmarshal([]byte(detailsJSON.String), &detMap)
		}
		canonicalDetails := canonicalizeDetails(detMap)
		chainInput := prev + "|" + ts.String + "|" + tenantID.String + "|" + action.String + "|" + actor.String + "|" + object.String + "|" + canonicalDetails + "|" + rid.String
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

// LogEntry represents a single audit log entry
type LogEntry struct {
	Seq       int64          `json:"seq"`
	Timestamp string         `json:"timestamp"`
	TenantID  string         `json:"tenant_id"`
	Action    string         `json:"action"`
	Actor     string         `json:"actor"`
	Object    string         `json:"object"`
	Details   map[string]any `json:"details"`
	RequestID string         `json:"request_id"`
}

// AuditFilter contains filtering options for audit log queries
type AuditFilter struct {
	Actor      string
	Action     string
	ObjectType string
	From       *time.Time
	To         *time.Time
}

// QueryLogs retrieves audit logs with filtering and pagination
func (a *SqliteAuditor) QueryLogs(ctx context.Context, tenantID string, limit, offset int) ([]LogEntry, int, error) {
	return a.QueryLogsFiltered(ctx, tenantID, AuditFilter{}, limit, offset)
}

// QueryLogsFiltered retrieves audit logs with optional filters and pagination
func (a *SqliteAuditor) QueryLogsFiltered(ctx context.Context, tenantID string, filter AuditFilter, limit, offset int) ([]LogEntry, int, error) {
	// Build WHERE clause
	conditions := []string{"tenant_id = ?"}
	args := []interface{}{tenantID}

	if filter.Actor != "" {
		conditions = append(conditions, "actor = ?")
		args = append(args, filter.Actor)
	}
	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}
	if filter.ObjectType != "" {
		// Object is stored as JSON, so we need to check if it contains the type
		conditions = append(conditions, "json_extract(object, '$.type') = ?")
		args = append(args, filter.ObjectType)
	}
	if filter.From != nil {
		conditions = append(conditions, "ts >= ?")
		args = append(args, filter.From.Format(time.RFC3339))
	}
	if filter.To != nil {
		conditions = append(conditions, "ts <= ?")
		args = append(args, filter.To.Format(time.RFC3339))
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_events WHERE %s`, whereClause)
	if err := a.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Query logs
	query := fmt.Sprintf(`
		SELECT seq, ts, tenant_id, action, actor, object, details, request_id
		FROM audit_events
		WHERE %s
		ORDER BY seq DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, limit, offset)

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		var detailsJSON sql.NullString
		if err := rows.Scan(&l.Seq, &l.Timestamp, &l.TenantID, &l.Action, &l.Actor, &l.Object, &detailsJSON, &l.RequestID); err != nil {
			return nil, 0, err
		}
		if detailsJSON.Valid && detailsJSON.String != "" {
			_ = json.Unmarshal([]byte(detailsJSON.String), &l.Details)
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}
