package storage

import (
	"testing"
)

func TestSearchSimilar(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	// Insert test documents with embeddings
	var docs = []struct {
		doc    Document
		vector []float32
	}{
		{
			doc: Document{
				PaperlessID:  1001,
				PaperlessURL: "http://example.com/doc/1001",
				Title:        "Financial Report",
				Tags:         "finance, report",
			},
			vector: []float32{1.0, 0.0, 0.0},
		},
		{
			doc: Document{
				PaperlessID:  1002,
				PaperlessURL: "http://example.com/doc/1002",
				Title:        "Budget Summary",
				Tags:         "finance, budget",
			},
			vector: []float32{0.9, 0.1, 0.0},
		},
		{
			doc: Document{
				PaperlessID:  1003,
				PaperlessURL: "http://example.com/doc/1003",
				Title:        "Recipe Book",
				Tags:         "cooking, recipes",
			},
			vector: []float32{0.0, 1.0, 0.0},
		},
	}

	for _, item := range docs {
		var docID, err = db.InsertDocument(item.doc)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}

		err = db.InsertEmbedding(int(docID), "test content", item.vector)
		if err != nil {
			t.Fatalf("Failed to insert embedding: %v", err)
		}
	}

	// Search with a query similar to first document
	var queryVector = []float32{1.0, 0.0, 0.0}
	var results, err = db.SearchSimilar(queryVector, 10, 0.5)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) < 1 {
		t.Fatal("Expected at least 1 result")
	}

	// First result should be the most similar document
	if results[0].Title != "Financial Report" {
		t.Errorf("Expected first result to be 'Financial Report', got '%s'", results[0].Title)
	}

	// Similarity score should be close to 1.0
	if results[0].SimilarityScore < 0.95 {
		t.Errorf("Expected similarity > 0.95, got %f", results[0].SimilarityScore)
	}
}

func TestSearchSimilarWithThreshold(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc1 = Document{
		PaperlessID:  2001,
		PaperlessURL: "http://example.com/doc/2001",
		Title:        "Similar Document",
		Tags:         "test",
	}
	var vector1 = []float32{1.0, 0.0, 0.0}

	var docID1, err = db.InsertDocument(doc1)
	if err != nil {
		t.Fatalf("Failed to insert document 1: %v", err)
	}
	err = db.InsertEmbedding(int(docID1), "content", vector1)
	if err != nil {
		t.Fatalf("Failed to insert embedding 1: %v", err)
	}

	var doc2 = Document{
		PaperlessID:  2002,
		PaperlessURL: "http://example.com/doc/2002",
		Title:        "Dissimilar Document",
		Tags:         "test",
	}
	var vector2 = []float32{0.0, 1.0, 0.0}

	var docID2, err2 = db.InsertDocument(doc2)
	if err2 != nil {
		t.Fatalf("Failed to insert document 2: %v", err2)
	}
	err2 = db.InsertEmbedding(int(docID2), "content", vector2)
	if err2 != nil {
		t.Fatalf("Failed to insert embedding 2: %v", err2)
	}

	// Search with high threshold - should only return similar document
	var queryVector = []float32{1.0, 0.0, 0.0}
	var results, searchErr = db.SearchSimilar(queryVector, 10, 0.9)
	if searchErr != nil {
		t.Fatalf("Failed to search: %v", searchErr)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result with high threshold, got %d", len(results))
	}

	if len(results) > 0 && results[0].Title != "Similar Document" {
		t.Errorf("Expected 'Similar Document', got '%s'", results[0].Title)
	}
}

func TestSearchSimilarWithLimit(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	// Insert 5 documents
	for i := 1; i <= 5; i++ {
		var doc = Document{
			PaperlessID:  3000 + i,
			PaperlessURL: "http://example.com/doc/3000",
			Title:        "Document",
			Tags:         "test",
		}
		var vector = []float32{1.0, 0.0, 0.0}

		var docID, err = db.InsertDocument(doc)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}
		err = db.InsertEmbedding(int(docID), "content", vector)
		if err != nil {
			t.Fatalf("Failed to insert embedding: %v", err)
		}
	}

	// Search with limit of 3
	var queryVector = []float32{1.0, 0.0, 0.0}
	var results, err = db.SearchSimilar(queryVector, 3, 0.0)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestSearchSimilarNoResults(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	// Search empty database
	var queryVector = []float32{1.0, 0.0, 0.0}
	var results, err = db.SearchSimilar(queryVector, 10, 0.5)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results from empty database, got %d", len(results))
	}
}

func TestSearchSimilarSorting(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var docs = []struct {
		doc    Document
		vector []float32
	}{
		{
			doc: Document{
				PaperlessID:  4001,
				PaperlessURL: "http://example.com/doc/4001",
				Title:        "High Similarity",
				Tags:         "test",
			},
			vector: []float32{1.0, 0.1, 0.0},
		},
		{
			doc: Document{
				PaperlessID:  4002,
				PaperlessURL: "http://example.com/doc/4002",
				Title:        "Medium Similarity",
				Tags:         "test",
			},
			vector: []float32{0.5, 0.5, 0.0},
		},
		{
			doc: Document{
				PaperlessID:  4003,
				PaperlessURL: "http://example.com/doc/4003",
				Title:        "Perfect Match",
				Tags:         "test",
			},
			vector: []float32{1.0, 0.0, 0.0},
		},
	}

	for _, item := range docs {
		var docID, err = db.InsertDocument(item.doc)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}
		err = db.InsertEmbedding(int(docID), "content", item.vector)
		if err != nil {
			t.Fatalf("Failed to insert embedding: %v", err)
		}
	}

	// Search and verify results are sorted by similarity
	var queryVector = []float32{1.0, 0.0, 0.0}
	var results, err = db.SearchSimilar(queryVector, 10, 0.0)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// First result should be perfect match
	if results[0].Title != "Perfect Match" {
		t.Errorf("Expected first result to be 'Perfect Match', got '%s'", results[0].Title)
	}

	// Verify results are sorted in descending order
	for i := 0; i < len(results)-1; i++ {
		if results[i].SimilarityScore < results[i+1].SimilarityScore {
			t.Errorf("Results not sorted: result[%d] score %f < result[%d] score %f",
				i, results[i].SimilarityScore, i+1, results[i+1].SimilarityScore)
		}
	}
}

func TestParseTimestamp(t *testing.T) {
	var tests = []struct {
		name      string
		timestamp string
		shouldErr bool
	}{
		{
			name:      "SQLite format",
			timestamp: "2024-01-15 10:30:45",
			shouldErr: false,
		},
		{
			name:      "ISO8601 format",
			timestamp: "2024-01-15T10:30:45Z",
			shouldErr: false,
		},
		{
			name:      "RFC3339 format",
			timestamp: "2024-01-15T10:30:45+00:00",
			shouldErr: false,
		},
		{
			name:      "invalid format",
			timestamp: "invalid-timestamp",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var _, err = parseTimestamp(tt.timestamp)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
