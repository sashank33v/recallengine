package model

type Chunk struct {
	ID        int64
	NoteID    int64
	NoteTitle string
	Content   string
	Rank      float64
}
