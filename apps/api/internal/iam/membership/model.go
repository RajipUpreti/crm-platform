package membership

import "time"

type Role string

const (
	RoleOwner  Role = "OWNER"
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusInvited   Status = "INVITED"
	StatusSuspended Status = "SUSPENDED"
	StatusLeft      Status = "LEFT"
)

type Membership struct {
	ID string `json:"id" format:"uuid"`

	TenantID string `json:"tenantId" format:"uuid"`

	UserID string `json:"userId" format:"uuid"`

	Role Role `json:"role" enums:"OWNER,ADMIN,MEMBER"`

	Status Status `json:"status" enums:"ACTIVE,INVITED,SUSPENDED,LEFT"`

	JoinedAt *time.Time `json:"joinedAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`

	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateInput struct {
	TenantID string

	UserID string

	Role Role

	Status Status

	JoinedAt *time.Time
}

type UpdateRoleInput struct {
	Role Role
}

type UpdateStatusInput struct {
	Status Status
}
