package storage

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// initialSchema contains the SQL for creating tables
const initialSchema = `-- Documents table stores metadata from Paperless
CREATE TABLE IF NOT EXISTS documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    paperless_id INTEGER UNIQUE NOT NULL,
    paperless_url TEXT NOT NULL,
    title TEXT,
    tags TEXT,
    embedded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_modified TIMESTAMP
);

-- Embeddings table stores vector embeddings
CREATE TABLE IF NOT EXISTS embeddings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    document_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    vector BLOB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

-- Index state tracks the last processed Paperless document ID
CREATE TABLE IF NOT EXISTS index_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_paperless_id INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO index_state (id, last_paperless_id) VALUES (1, 0);

-- Failures are tracked per Paperless document ID
CREATE TABLE IF NOT EXISTS index_failures (
    paperless_id INTEGER PRIMARY KEY,
    error TEXT NOT NULL,
    failed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_paperless_id ON documents(paperless_id);
CREATE INDEX IF NOT EXISTS idx_document_id ON embeddings(document_id);
`

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection and runs migrations
func NewDB(dbPath string) (*DB, error) {
	// Ensure the data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database connection
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations executes the SQL schema
func (db *DB) runMigrations() error {
	// Execute schema
	if _, err := db.conn.Exec(initialSchema); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// serializeVector converts a float32 slice to bytes for storage
func serializeVector(vector []float32) []byte {
	buf := make([]byte, len(vector)*4)
	for i, v := range vector {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

// deserializeVector converts bytes back to a float32 slice
func deserializeVector(data []byte) []float32 {
	vector := make([]float32, len(data)/4)
	for i := range vector {
		bits := binary.LittleEndian.Uint32(data[i*4:])
		vector[i] = math.Float32frombits(bits)
	}
	return vector
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// parseTimestamp parses SQLite timestamp strings.
func parseTimestamp(ts string) (time.Time, error) {
	// Try common SQLite timestamp formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05.999999 -0700 -0700",
		"2006-01-02 15:04:05.999999999 -0700 -0700",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05.999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", ts)
}
