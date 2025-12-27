package storage

import (
	"testing"
	"time"
)

func TestInsertDocument(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc = Document{
		PaperlessID:  123,
		PaperlessURL: "http://example.com/doc/123",
		Title:        "Test Document",
		Tags:         "tag1, tag2",
		LastModified: time.Now(),
	}

	var id, err = db.InsertDocument(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}
}

func TestInsertDocumentDuplicate(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc = Document{
		PaperlessID:  123,
		PaperlessURL: "http://example.com/doc/123",
		Title:        "Test Document",
		Tags:         "tag1, tag2",
		LastModified: time.Now(),
	}

	var _, err = db.InsertDocument(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Try to insert duplicate
	var _, err2 = db.InsertDocument(doc)
	if err2 == nil {
		t.Error("Expected error when inserting duplicate document, got nil")
	}
}

func TestGetDocumentByPaperlessID(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc = Document{
		PaperlessID:  456,
		PaperlessURL: "http://example.com/doc/456",
		Title:        "Test Document",
		Tags:         "test, document",
		LastModified: time.Now(),
	}

	var _, err = db.InsertDocument(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	var retrieved, err2 = db.GetDocumentByPaperlessID(456)
	if err2 != nil {
		t.Fatalf("Failed to get document: %v", err2)
	}

	if retrieved == nil {
		t.Fatal("Retrieved document is nil")
	}

	if retrieved.PaperlessID != doc.PaperlessID {
		t.Errorf("Expected PaperlessID %d, got %d", doc.PaperlessID, retrieved.PaperlessID)
	}
	if retrieved.Title != doc.Title {
		t.Errorf("Expected Title %s, got %s", doc.Title, retrieved.Title)
	}
	if retrieved.Tags != doc.Tags {
		t.Errorf("Expected Tags %s, got %s", doc.Tags, retrieved.Tags)
	}
}

func TestGetDocumentByPaperlessIDNotFound(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc, err = db.GetDocumentByPaperlessID(999)
	if err != nil {
		t.Fatalf("Expected no error for non-existent document, got: %v", err)
	}

	if doc != nil {
		t.Error("Expected nil document for non-existent ID")
	}
}

func TestUpdateDocument(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc = Document{
		PaperlessID:  789,
		PaperlessURL: "http://example.com/doc/789",
		Title:        "Original Title",
		Tags:         "original",
		LastModified: time.Now(),
	}

	var _, err = db.InsertDocument(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Update the document
	doc.Title = "Updated Title"
	doc.Tags = "updated, new"
	doc.PaperlessURL = "http://example.com/doc/789/updated"

	err = db.UpdateDocument(doc)
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	// Retrieve and verify
	var retrieved, err2 = db.GetDocumentByPaperlessID(789)
	if err2 != nil {
		t.Fatalf("Failed to get updated document: %v", err2)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected Title 'Updated Title', got %s", retrieved.Title)
	}
	if retrieved.Tags != "updated, new" {
		t.Errorf("Expected Tags 'updated, new', got %s", retrieved.Tags)
	}
}

func TestDeleteDocument(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc = Document{
		PaperlessID:  321,
		PaperlessURL: "http://example.com/doc/321",
		Title:        "To Be Deleted",
		Tags:         "delete",
		LastModified: time.Now(),
	}

	var _, err = db.InsertDocument(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	err = db.DeleteDocument(321)
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	// Verify document is gone
	var retrieved, err2 = db.GetDocumentByPaperlessID(321)
	if err2 != nil {
		t.Fatalf("Error checking deleted document: %v", err2)
	}

	if retrieved != nil {
		t.Error("Expected deleted document to be nil")
	}
}

func TestInsertEmbedding(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc = Document{
		PaperlessID:  111,
		PaperlessURL: "http://example.com/doc/111",
		Title:        "Document with Embedding",
		Tags:         "embedding",
		LastModified: time.Now(),
	}

	var docID, err = db.InsertDocument(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	var vector = []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	err = db.InsertEmbedding(int(docID), "test content", vector)
	if err != nil {
		t.Fatalf("Failed to insert embedding: %v", err)
	}

	// Verify embedding was inserted
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM embeddings WHERE document_id = ?", docID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count embeddings: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 embedding, got %d", count)
	}
}

func TestDeleteEmbeddingsByDocumentID(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var doc = Document{
		PaperlessID:  222,
		PaperlessURL: "http://example.com/doc/222",
		Title:        "Document with Multiple Embeddings",
		Tags:         "embeddings",
		LastModified: time.Now(),
	}

	var docID, err = db.InsertDocument(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	var vector = []float32{0.1, 0.2, 0.3}
	err = db.InsertEmbedding(int(docID), "content 1", vector)
	if err != nil {
		t.Fatalf("Failed to insert first embedding: %v", err)
	}

	err = db.InsertEmbedding(int(docID), "content 2", vector)
	if err != nil {
		t.Fatalf("Failed to insert second embedding: %v", err)
	}

	err = db.DeleteEmbeddingsByDocumentID(int(docID))
	if err != nil {
		t.Fatalf("Failed to delete embeddings: %v", err)
	}

	// Verify embeddings are gone
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM embeddings WHERE document_id = ?", docID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count embeddings: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 embeddings after deletion, got %d", count)
	}
}

func TestListDocuments(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var docs = []Document{
		{
			PaperlessID:  1,
			PaperlessURL: "http://example.com/doc/1",
			Title:        "Doc 1",
			Tags:         "tag1",
			LastModified: time.Now(),
		},
		{
			PaperlessID:  2,
			PaperlessURL: "http://example.com/doc/2",
			Title:        "Doc 2",
			Tags:         "tag2",
			LastModified: time.Now(),
		},
		{
			PaperlessID:  3,
			PaperlessURL: "http://example.com/doc/3",
			Title:        "Doc 3",
			Tags:         "tag3",
			LastModified: time.Now(),
		},
	}

	for _, doc := range docs {
		var _, err = db.InsertDocument(doc)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}
	}

	var list, err = db.ListDocuments()
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(list))
	}

	// Verify documents are ordered by PaperlessID
	for i := range list {
		if list[i].PaperlessID != docs[i].PaperlessID {
			t.Errorf("Document %d: expected PaperlessID %d, got %d", i, docs[i].PaperlessID, list[i].PaperlessID)
		}
	}
}

func TestCountDocuments(t *testing.T) {
	var db = setupTestDB(t)
	defer db.Close()

	var count, err = db.CountDocuments()
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 documents initially, got %d", count)
	}

	var doc = Document{
		PaperlessID:  555,
		PaperlessURL: "http://example.com/doc/555",
		Title:        "Counter Test",
		Tags:         "test",
		LastModified: time.Now(),
	}

	var _, insertErr = db.InsertDocument(doc)
	if insertErr != nil {
		t.Fatalf("Failed to insert document: %v", insertErr)
	}

	count, err = db.CountDocuments()
	if err != nil {
		t.Fatalf("Failed to count documents after insert: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 document, got %d", count)
	}
}
