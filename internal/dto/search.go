package dto

type PassageResponse struct {
	NoteID    int64   `json:"note_id"`
	NoteTitle string  `json:"note_title"`
	Content   string  `json:"content"`
	Score     float64 `json:"score"`
}
