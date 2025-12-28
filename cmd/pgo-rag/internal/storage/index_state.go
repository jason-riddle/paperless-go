package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// IndexState tracks the last processed Paperless document ID.
type IndexState struct {
	LastPaperlessID int
	UpdatedAt       time.Time
}

// IndexFailure tracks indexing failures for a document.
type IndexFailure struct {
	PaperlessID int
	Error       string
	FailedAt    time.Time
}

// GetIndexState returns the current index state.
func (db *DB) GetIndexState() (IndexState, error) {
	var state IndexState
	var updatedAt sql.NullString
	err := db.conn.QueryRow(`
		SELECT last_paperless_id, updated_at
		FROM index_state
		WHERE id = 1
	`).Scan(&state.LastPaperlessID, &updatedAt)
	if err != nil {
		return state, fmt.Errorf("failed to get index state: %w", err)
	}
	if updatedAt.Valid {
		parsed, err := parseTimestamp(updatedAt.String)
		if err != nil {
			return state, fmt.Errorf("failed to parse index_state.updated_at: %w", err)
		}
		state.UpdatedAt = parsed
	}
	return state, nil
}

// UpdateIndexState sets the last processed Paperless ID.
func (db *DB) UpdateIndexState(lastPaperlessID int) error {
	_, err := db.conn.Exec(`
		UPDATE index_state
		SET last_paperless_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, lastPaperlessID)
	if err != nil {
		return fmt.Errorf("failed to update index state: %w", err)
	}
	return nil
}

// ResetIndexState clears the last processed Paperless ID.
func (db *DB) ResetIndexState() error {
	return db.UpdateIndexState(0)
}

// ClearIndexData removes documents, embeddings, failures, and resets state.
func (db *DB) ClearIndexData() error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin clear transaction: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM embeddings`); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to clear embeddings: %v (rollback error: %w)", err, rollbackErr)
		}
		return fmt.Errorf("failed to clear embeddings: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM documents`); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to clear documents: %v (rollback error: %w)", err, rollbackErr)
		}
		return fmt.Errorf("failed to clear documents: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM index_failures`); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to clear failures: %v (rollback error: %w)", err, rollbackErr)
		}
		return fmt.Errorf("failed to clear failures: %w", err)
	}
	if _, err := tx.Exec(`UPDATE index_state SET last_paperless_id = 0, updated_at = CURRENT_TIMESTAMP WHERE id = 1`); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to reset index state: %v (rollback error: %w)", err, rollbackErr)
		}
		return fmt.Errorf("failed to reset index state: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit clear transaction: %w", err)
	}

	return nil
}

// RecordIndexFailure stores the latest error for a Paperless document.
func (db *DB) RecordIndexFailure(paperlessID int, err error) error {
	if err == nil {
		return nil
	}
	_, execErr := db.conn.Exec(`
		INSERT INTO index_failures (paperless_id, error)
		VALUES (?, ?)
		ON CONFLICT(paperless_id) DO UPDATE SET
			error = excluded.error,
			failed_at = CURRENT_TIMESTAMP
	`, paperlessID, err.Error())
	if execErr != nil {
		return fmt.Errorf("failed to record index failure: %w", execErr)
	}
	return nil
}

// ClearIndexFailure removes any recorded failure for a document.
func (db *DB) ClearIndexFailure(paperlessID int) error {
	_, err := db.conn.Exec(`DELETE FROM index_failures WHERE paperless_id = ?`, paperlessID)
	if err != nil {
		return fmt.Errorf("failed to clear index failure: %w", err)
	}
	return nil
}

// GetIndexFailure returns the failure for a specific document.
func (db *DB) GetIndexFailure(paperlessID int) (*IndexFailure, error) {
	var failure IndexFailure
	var failedAt sql.NullString
	err := db.conn.QueryRow(`
		SELECT paperless_id, error, failed_at
		FROM index_failures
		WHERE paperless_id = ?
	`, paperlessID).Scan(&failure.PaperlessID, &failure.Error, &failedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get index failure: %w", err)
	}
	if failedAt.Valid {
		parsed, err := parseTimestamp(failedAt.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse index_failures.failed_at: %w", err)
		}
		failure.FailedAt = parsed
	}
	return &failure, nil
}
