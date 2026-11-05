package dto

import "time"

type AskRequest struct {
	Question       string `json:"question"`
	ConversationID int64  `json:"conversation_id"`
}

type AskResponse struct {
	Answer  string            `json:"answer"`
	Sources []PassageResponse `json:"sources"`
}

type ConversationResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageResponse struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
