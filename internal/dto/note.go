package dto

import "time"

type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type NoteResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
