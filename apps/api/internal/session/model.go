package session

import "time"

type Session struct {
	UserID string `json:"userId"`

	TenantID string `json:"tenantId"`

	MembershipID string `json:"membershipId"`

	CreatedAt time.Time `json:"createdAt"`

	ExpiresAt time.Time `json:"expiresAt"`
}

type CreateInput struct {
	UserID string

	TenantID string

	MembershipID string
}

type CreatedSession struct {
	Token string

	ExpiresAt time.Time
}
