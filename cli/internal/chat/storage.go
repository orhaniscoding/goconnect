package chat

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for database/sql
)

// Storage provides persistent chat message storage using SQLite
type Storage struct {
	db     *sql.DB
	dbPath string
	mu     sync.RWMutex
}

// NewStorage creates a new chat storage instance
func NewStorage(dataDir string) (*Storage, error) {
	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "chat.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open chat database: %w", err)
	}

	s := &Storage{
		db:     db,
		dbPath: dbPath,
	}

	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// initSchema creates the chat tables if they don't exist
func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		from_peer TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		network_id TEXT DEFAULT '',
		read INTEGER DEFAULT 0,
		created_at INTEGER DEFAULT (strftime('%s', 'now'))
	);

	CREATE INDEX IF NOT EXISTS idx_messages_network ON messages(network_id);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_messages_from ON messages(from_peer);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize chat schema: %w", err)
	}

	return nil
}

// Close closes the database connection after checkpointing WAL
func (s *Storage) Close() error {
	// Checkpoint WAL to ensure all data is written to main database
	// This also helps with cleanup of WAL files on Windows
	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return s.db.Close()
}

// SaveMessage persists a message to the database
func (s *Storage) SaveMessage(msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not set
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("%d-%s", msg.Time.UnixNano(), msg.From)
	}

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO messages (id, from_peer, content, timestamp, network_id)
		VALUES (?, ?, ?, ?, ?)
	`, msg.ID, msg.From, msg.Content, msg.Time.Unix(), msg.NetworkID)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// GetMessages retrieves messages with optional filtering and pagination
func (s *Storage) GetMessages(networkID string, limit int, beforeID string) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	var rows *sql.Rows
	var err error

	if beforeID != "" {
		// Get timestamp of the beforeID message for cursor-based pagination
		var beforeTimestamp int64
		err = s.db.QueryRow("SELECT timestamp FROM messages WHERE id = ?", beforeID).Scan(&beforeTimestamp)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get before message: %w", err)
		}

		if networkID != "" {
			rows, err = s.db.Query(`
				SELECT id, from_peer, content, timestamp, network_id 
				FROM messages 
				WHERE network_id = ? AND timestamp < ?
				ORDER BY timestamp DESC 
				LIMIT ?
			`, networkID, beforeTimestamp, limit)
		} else {
			rows, err = s.db.Query(`
				SELECT id, from_peer, content, timestamp, network_id 
				FROM messages 
				WHERE timestamp < ?
				ORDER BY timestamp DESC 
				LIMIT ?
			`, beforeTimestamp, limit)
		}
	} else {
		if networkID != "" {
			rows, err = s.db.Query(`
				SELECT id, from_peer, content, timestamp, network_id 
				FROM messages 
				WHERE network_id = ?
				ORDER BY timestamp DESC 
				LIMIT ?
			`, networkID, limit)
		} else {
			rows, err = s.db.Query(`
				SELECT id, from_peer, content, timestamp, network_id 
				FROM messages 
				ORDER BY timestamp DESC 
				LIMIT ?
			`, limit)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var timestamp int64
		if err := rows.Scan(&msg.ID, &msg.From, &msg.Content, &timestamp, &msg.NetworkID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.Time = time.Unix(timestamp, 0)
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return messages, nil
}

// GetMessagesByPeer retrieves messages from a specific peer
func (s *Storage) GetMessagesByPeer(peerID string, limit int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`
		SELECT id, from_peer, content, timestamp, network_id 
		FROM messages 
		WHERE from_peer = ?
		ORDER BY timestamp DESC 
		LIMIT ?
	`, peerID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var timestamp int64
		if err := rows.Scan(&msg.ID, &msg.From, &msg.Content, &timestamp, &msg.NetworkID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.Time = time.Unix(timestamp, 0)
		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteMessage removes a message by ID
func (s *Storage) DeleteMessage(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// DeleteOldMessages removes messages older than the specified duration
func (s *Storage) DeleteOldMessages(olderThan time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-olderThan).Unix()
	result, err := s.db.Exec("DELETE FROM messages WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old messages: %w", err)
	}

	return result.RowsAffected()
}

// GetMessageCount returns the total number of messages
func (s *Storage) GetMessageCount() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// SearchMessages performs a full-text search on message content
func (s *Storage) SearchMessages(query string, limit int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	// Use LIKE for simple search (FTS could be added for better performance)
	rows, err := s.db.Query(`
		SELECT id, from_peer, content, timestamp, network_id 
		FROM messages 
		WHERE content LIKE ?
		ORDER BY timestamp DESC 
		LIMIT ?
	`, "%"+query+"%", limit)

	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var timestamp int64
		if err := rows.Scan(&msg.ID, &msg.From, &msg.Content, &timestamp, &msg.NetworkID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.Time = time.Unix(timestamp, 0)
		messages = append(messages, msg)
	}

	return messages, nil
}
