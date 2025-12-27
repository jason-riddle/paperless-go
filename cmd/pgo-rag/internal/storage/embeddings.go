package storage

import (
	"database/sql"
	"fmt"
)

// InsertDocument inserts a new document into the database
func (db *DB) InsertDocument(doc Document) (int64, error) {
	result, err := db.conn.Exec(`
		INSERT INTO documents (paperless_id, paperless_url, title, tags, last_modified)
		VALUES (?, ?, ?, ?, ?)
	`, doc.PaperlessID, doc.PaperlessURL, doc.Title, doc.Tags, doc.LastModified)
	if err != nil {
		return 0, fmt.Errorf("failed to insert document: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// UpdateDocument updates an existing document
func (db *DB) UpdateDocument(doc Document) error {
	_, err := db.conn.Exec(`
		UPDATE documents
		SET paperless_url = ?, title = ?, tags = ?, last_modified = ?, embedded_at = CURRENT_TIMESTAMP
		WHERE paperless_id = ?
	`, doc.PaperlessURL, doc.Title, doc.Tags, doc.LastModified, doc.PaperlessID)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

// InsertEmbedding inserts a new embedding into the database
func (db *DB) InsertEmbedding(docID int, content string, vector []float32) error {
	vectorBytes := serializeVector(vector)
	_, err := db.conn.Exec(`
		INSERT INTO embeddings (document_id, content, vector)
		VALUES (?, ?, ?)
	`, docID, content, vectorBytes)
	if err != nil {
		return fmt.Errorf("failed to insert embedding: %w", err)
	}
	return nil
}

// GetDocumentByPaperlessID retrieves a document by its Paperless ID
func (db *DB) GetDocumentByPaperlessID(paperlessID int) (*Document, error) {
	var doc Document
	var embeddedAt sql.NullString
	var lastModified sql.NullString
	err := db.conn.QueryRow(`
		SELECT id, paperless_id, paperless_url, title, tags, embedded_at, last_modified
		FROM documents
		WHERE paperless_id = ?
	`, paperlessID).Scan(
		&doc.ID,
		&doc.PaperlessID,
		&doc.PaperlessURL,
		&doc.Title,
		&doc.Tags,
		&embeddedAt,
		&lastModified,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if embeddedAt.Valid {
		parsed, err := parseTimestamp(embeddedAt.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedded_at: %w", err)
		}
		doc.EmbeddedAt = parsed
	}
	if lastModified.Valid {
		parsed, err := parseTimestamp(lastModified.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_modified: %w", err)
		}
		doc.LastModified = parsed
	}
	return &doc, nil
}

// DeleteDocument deletes a document and its embeddings
func (db *DB) DeleteDocument(paperlessID int) error {
	_, err := db.conn.Exec(`DELETE FROM documents WHERE paperless_id = ?`, paperlessID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// DeleteEmbeddingsByDocumentID deletes all embeddings for a document
func (db *DB) DeleteEmbeddingsByDocumentID(documentID int) error {
	_, err := db.conn.Exec(`DELETE FROM embeddings WHERE document_id = ?`, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete embeddings: %w", err)
	}
	return nil
}

// ListDocuments returns all documents in the database
func (db *DB) ListDocuments() ([]Document, error) {
	rows, err := db.conn.Query(`
		SELECT id, paperless_id, paperless_url, title, tags, embedded_at, last_modified
		FROM documents
		ORDER BY paperless_id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var embeddedAt sql.NullString
		var lastModified sql.NullString
		err := rows.Scan(
			&doc.ID,
			&doc.PaperlessID,
			&doc.PaperlessURL,
			&doc.Title,
			&doc.Tags,
			&embeddedAt,
			&lastModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		if embeddedAt.Valid {
			parsed, err := parseTimestamp(embeddedAt.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse embedded_at: %w", err)
			}
			doc.EmbeddedAt = parsed
		}
		if lastModified.Valid {
			parsed, err := parseTimestamp(lastModified.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse last_modified: %w", err)
			}
			doc.LastModified = parsed
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, nil
}

// CountDocuments returns the total number of documents
func (db *DB) CountDocuments() (int, error) {
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	return count, nil
}
