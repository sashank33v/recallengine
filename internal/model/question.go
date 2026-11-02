package model

import "time"

type Question struct {
	ID           int64
	UserID       int64
	QuestionText string
	AnswerText   string
	Source       string
	Status       string
	CreatedAt    time.Time
}
