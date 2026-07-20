package auth

import "time"

type LoginTransaction struct {
	State        string
	Nonce        string
	CodeVerifier string
	ReturnTo     string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

func (t LoginTransaction) Expired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}
