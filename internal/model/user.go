package model

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Timezone     string
	CreatedAt    time.Time
}
