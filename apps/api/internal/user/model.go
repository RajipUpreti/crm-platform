package user

import "time"

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusDeleted   Status = "DELETED"
)

type User struct {
	ID string `json:"id"`

	IdentityProvider       string `json:"-"`
	IdentityProviderUserID string `json:"-"`

	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`

	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	DisplayName string `json:"displayName,omitempty"`

	Status Status `json:"status"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Identity struct {
	Provider       string
	ProviderUserID string

	Email         string
	EmailVerified bool

	FirstName   string
	LastName    string
	DisplayName string
}
