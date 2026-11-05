package model

import "time"

type Conversation struct {
	ID        int64
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChatMessage struct {
	ID        int64
	Role      string
	Content   string
	CreatedAt time.Time
}
