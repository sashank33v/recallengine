package repository

import (
	"context"
	"database/sql"

	"github.com/sashank33v/recallengine/internal/model"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(ctx context.Context, userID int64, title string) (int64, error) {
	var id int64
	query := `INSERT INTO conversations (user_id, title) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, userID, title).Scan(&id)
	return id, err
}

func (r *ConversationRepository) Owns(ctx context.Context, userID, conversationID int64) (bool, error) {
	var owns bool
	query := `SELECT EXISTS (SELECT 1 FROM conversations WHERE id = $1 AND user_id = $2)`
	err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(&owns)
	return owns, err
}

func (r *ConversationRepository) AddMessage(ctx context.Context, conversationID int64, role, content string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query1 := `INSERT INTO chat_messages (conversation_id, role, content) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, query1, conversationID, role, content)
	if err != nil {
		return err
	}
	query2 := `UPDATE conversations SET updated_at = now() WHERE id = $1`
	_, err = tx.ExecContext(ctx, query2, conversationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ConversationRepository) List(ctx context.Context, userID int64) ([]model.Conversation, error) {
	query := `SELECT id, title, created_at, updated_at FROM conversations WHERE user_id = $1 ORDER BY updated_at DESC LIMIT 50`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := []model.Conversation{}
	for rows.Next() {
		var c model.Conversation
		err := rows.Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func (r *ConversationRepository) Messages(ctx context.Context, conversationID int64) ([]model.ChatMessage, error) {
	query := `SELECT id, role, content, created_at FROM chat_messages WHERE conversation_id = $1 ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query,
		conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []model.ChatMessage{}
	for rows.Next() {
		var m model.ChatMessage
		err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *ConversationRepository) CountMessages(ctx context.Context, conversationID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM chat_messages WHERE conversation_id = $1`
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(&count)
	return count, err
}
