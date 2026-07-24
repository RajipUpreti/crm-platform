package tenant

import (
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
)

type Access struct {
	Tenant Tenant `json:"tenant"`

	Membership membership.Membership `json:"membership"`

	Current bool `json:"current"`
}
