package dto

import "time"

type AnswerRequest struct {
	QuestionID int `json:"question_id"`
	Quality    int `json:"quality"`
}
type AnswerResponse struct {
	QuestionID   int64     `json:"question_id"`
	Quality      int       `json:"quality"`
	IntervalDays int       `json:"interval_days"`
	Repetitions  int       `json:"repetitions"`
	EaseFactor   float64   `json:"ease_factor"`
	NextDueAt    time.Time `json:"next_due_at"`
}
