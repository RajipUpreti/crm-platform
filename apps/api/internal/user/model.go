package user

import "time"

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusDeleted   Status = "DELETED"
)

type User struct {
	ID string

	IdentityProvider       string
	IdentityProviderUserID string

	Email         string
	EmailVerified bool

	FirstName string
	LastName  string
	Name      string

	Status Status

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Identity struct {
	Provider       string
	ProviderUserID string

	Email         string
	EmailVerified bool

	FirstName string
	LastName  string
	Name      string
}
