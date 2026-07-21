package session

import "time"

type Session struct {
	UserID string `json:"userId"`

	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type CreatedSession struct {
	Token     string
	ExpiresAt time.Time
}
