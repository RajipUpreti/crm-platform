package tenant

import "time"

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusDeleted   Status = "DELETED"
)

type Tenant struct {
	ID string `json:"id" format:"uuid"`

	Name string `json:"name" example:"Acme Corporation"`

	Slug string `json:"slug" example:"acme-corporation"`

	Status Status `json:"status" enums:"ACTIVE,SUSPENDED,DELETED"`

	CreatedAt time.Time `json:"createdAt"`

	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateInput struct {
	Name string

	Slug string
}

type UpdateInput struct {
	Name *string

	Status *Status
}
