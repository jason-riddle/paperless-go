package storage

import (
	"path/filepath"
	"testing"
)

// setupTestDB creates a temporary test database for testing
func setupTestDB(t *testing.T) *DB {
	var tmpDir = t.TempDir()
	var dbPath = filepath.Join(tmpDir, "test.db")

	var db, err = NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}
