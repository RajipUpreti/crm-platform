package membership

import (
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

// Member combines a tenant membership with its CRM user.
type Member struct {
	Membership DetailedMembership `json:"membership"`

	User DetailedUser `json:"user"`
}

type DetailedMembership struct {
	ID string `json:"id" format:"uuid"`

	TenantID string `json:"-" swaggerignore:"true"`

	UserID string `json:"-" swaggerignore:"true"`

	Role Role `json:"role" enums:"OWNER,ADMIN,MEMBER"`

	Status Status `json:"status" enums:"ACTIVE,INVITED,SUSPENDED,LEFT"`

	JoinedAt *time.Time `json:"joinedAt,omitempty"`
}

type DetailedUser struct {
	ID string `json:"id" format:"uuid"`

	Email string `json:"email" format:"email"`

	DisplayName string `json:"displayName,omitempty"`

	Status user.Status `json:"status" enums:"ACTIVE,SUSPENDED,DELETED"`
}
