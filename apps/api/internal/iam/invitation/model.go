package invitation

import (
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusAccepted Status = "ACCEPTED"
	StatusRevoked  Status = "REVOKED"
	StatusExpired  Status = "EXPIRED"
)

type Invitation struct {
	ID string `json:"id" format:"uuid"`

	TenantID string `json:"tenantId" format:"uuid"`

	InvitedByUserID string `json:"invitedByUserId" format:"uuid"`

	Email string `json:"email" format:"email"`

	Role membership.Role `json:"role" enums:"ADMIN,MEMBER"`

	Status Status `json:"status" enums:"PENDING,ACCEPTED,REVOKED,EXPIRED"`

	ExpiresAt time.Time `json:"expiresAt"`

	AcceptedAt *time.Time `json:"acceptedAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`

	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateInput struct {
	TenantID string

	InvitedByUserID string

	Email string

	Role membership.Role

	ExpiresAt time.Time

	TokenDigest string
}

type CreatedInvitation struct {
	Invitation Invitation `json:"invitation"`

	Token string `json:"token"`
}

type AcceptInput struct {
	Token string

	UserID string

	UserEmail string
}
