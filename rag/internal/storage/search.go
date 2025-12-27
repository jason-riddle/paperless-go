package storage

import (
	"fmt"
	"sort"
	"time"
)

// SearchSimilar performs a vector similarity search
func (db *DB) SearchSimilar(queryVector []float32, limit int, threshold float64) ([]SearchResult, error) {
	// Query all embeddings and compute similarity in memory
	// In a production system with many embeddings, you would want to use
	// sqlite-vec extension or another vector search solution
	rows, err := db.conn.Query(`
		SELECT
			e.document_id,
			e.vector,
			d.paperless_url,
			d.title,
			d.tags,
			d.last_modified
		FROM embeddings e
		JOIN documents d ON e.document_id = d.id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var (
			documentID   int
			vectorBytes  []byte
			paperlessURL string
			title        string
			tags         string
			lastModified string
		)

		err := rows.Scan(&documentID, &vectorBytes, &paperlessURL, &title, &tags, &lastModified)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Deserialize vector
		vector := deserializeVector(vectorBytes)

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryVector, vector)

		// Filter by threshold
		if similarity >= threshold {
			// Parse timestamp
			lastModTime, err := parseTimestamp(lastModified)
			if err != nil {
				// Log warning but continue with zero time
				lastModTime = time.Time{}
			}

			results = append(results, SearchResult{
				DocumentID:      documentID,
				PaperlessURL:    paperlessURL,
				Title:           title,
				Tags:            tags,
				SimilarityScore: similarity,
				LastModified:    lastModTime,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Sort results by similarity score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	// Limit results
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}
