package dto

type TopicMasteryResponse struct {
	Topic          string `json:"topic"`
	TotalAnswers   int    `json:"total_answers"`
	CorrectAnswers int    `json:"correct_answers"`
	MasteryPercent int    `json:"mastery_percent"`
}
