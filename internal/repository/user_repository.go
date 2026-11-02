package repository

import (
	"context"
	"database/sql"

	"github.com/sashank33v/recallengine/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (*model.User, error) {
	query := `
	INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, timezone, created_at
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email, passwordHash).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Timezone, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
	SELECT id, email, password_hash, timezone, created_at FROM users WHERE email = $1
	`
	user := &model.User{}

	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Timezone, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
