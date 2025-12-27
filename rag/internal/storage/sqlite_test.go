package storage

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDB(t *testing.T) {
	var tmpDir = t.TempDir()
	var dbPath = filepath.Join(tmpDir, "test.db")

	var db, err = NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.conn == nil {
		t.Fatal("Database connection is nil")
	}

	// Verify schema was created
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('documents', 'embeddings')").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query schema: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 tables, got %d", count)
	}
}

func TestSerializeDeserializeVector(t *testing.T) {
	var tests = []struct {
		name   string
		vector []float32
	}{
		{
			name:   "small vector",
			vector: []float32{0.1, 0.2, 0.3},
		},
		{
			name:   "negative values",
			vector: []float32{-0.5, 0.0, 0.5},
		},
		{
			name:   "large values",
			vector: []float32{1000.0, -1000.0, 0.0},
		},
		{
			name:   "edge cases",
			vector: []float32{float32(math.MaxFloat32), float32(-math.MaxFloat32), 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serialized = serializeVector(tt.vector)
			var deserialized = deserializeVector(serialized)

			if len(deserialized) != len(tt.vector) {
				t.Errorf("Length mismatch: expected %d, got %d", len(tt.vector), len(deserialized))
			}

			for i := range tt.vector {
				if deserialized[i] != tt.vector[i] {
					t.Errorf("Value mismatch at index %d: expected %f, got %f", i, tt.vector[i], deserialized[i])
				}
			}
		})
	}
}

func TestCosineSimilarity(t *testing.T) {
	var tests = []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
		epsilon  float64
	}{
		{
			name:     "identical vectors",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1.0, 0.0},
			b:        []float32{0.0, 1.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1.0, 0.0},
			b:        []float32{-1.0, 0.0},
			expected: -1.0,
			epsilon:  0.0001,
		},
		{
			name:     "similar vectors",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{2.0, 4.0, 6.0},
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "different length vectors",
			a:        []float32{1.0, 2.0},
			b:        []float32{1.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "zero vector",
			a:        []float32{0.0, 0.0},
			b:        []float32{1.0, 2.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result = cosineSimilarity(tt.a, tt.b)
			if math.Abs(result-tt.expected) > tt.epsilon {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestDBClose(t *testing.T) {
	var tmpDir = t.TempDir()
	var dbPath = filepath.Join(tmpDir, "test.db")

	var db, err = NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestNewDBWithNestedDirectory(t *testing.T) {
	var tmpDir = t.TempDir()
	var dbPath = filepath.Join(tmpDir, "nested", "dir", "test.db")

	var db, err = NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database with nested directory: %v", err)
	}
	defer db.Close()

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("Nested directory was not created")
	}
}
