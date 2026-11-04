package dto

import "time"

type CreateQuestionRequest struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Topic    string `json:"topic"`
}

type QuestionResponse struct {
	ID        int64     `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
