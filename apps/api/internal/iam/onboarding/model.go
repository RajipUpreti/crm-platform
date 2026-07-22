package onboarding

import (
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
)

type Result struct {
	Tenant     tenant.Tenant         `json:"tenant"`
	Membership membership.Membership `json:"membership"`
	Created    bool                  `json:"created"`
}
