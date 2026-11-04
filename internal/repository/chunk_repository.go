package repository

import (
	"context"
	"database/sql"

	"github.com/sashank33v/recallengine/internal/model"
)

type ChunkRepository struct {
	db *sql.DB
}

func NewChunkRepository(db *sql.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}
func (r *ChunkRepository) Search(ctx context.Context, userID int64, searchQuery string, limit int) ([]model.Chunk, error) {
	query := `
		SELECT c.id, c.note_id, n.title, c.content, ts_rank(c.search_vector, q)::float8 AS rank
		FROM note_chunks c
		JOIN notes n ON n.id = c.note_id,
		     websearch_to_tsquery('english', $2) q
		WHERE c.user_id = $1 AND c.search_vector @@ q
		ORDER BY rank DESC
		LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, userID, searchQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := []model.Chunk{}
	for rows.Next() {
		var c model.Chunk
		err := rows.Scan(&c.ID, &c.NoteID, &c.NoteTitle, &c.Content, &c.Rank)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}
