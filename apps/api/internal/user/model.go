package user

import "time"

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusDeleted   Status = "DELETED"
)

type User struct {
	ID string `json:"id" format:"uuid" example:"3ea55be2-3e57-4c85-9d19-f6048c405e44"`

	IdentityProvider       string `json:"-" swaggerignore:"true"`
	IdentityProviderUserID string `json:"-" swaggerignore:"true"`

	Email string `json:"email" format:"email" example:"developer@example.com"`

	EmailVerified bool `json:"emailVerified" example:"true"`

	FirstName string `json:"firstName,omitempty" example:"Local"`

	LastName string `json:"lastName,omitempty" example:"Developer"`

	DisplayName string `json:"displayName,omitempty" example:"Local Developer"`

	Status Status `json:"status" enums:"ACTIVE,SUSPENDED,DELETED" example:"ACTIVE"`

	CreatedAt time.Time `json:"createdAt" example:"2026-07-21T01:00:00Z"`

	UpdatedAt time.Time `json:"updatedAt" example:"2026-07-21T01:00:00Z"`
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
