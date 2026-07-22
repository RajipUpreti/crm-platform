package auth

import (
	"context"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/onboarding"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type TenantOnboarder interface {
	EnsureTenantAccess(
		ctx context.Context,
		currentUser user.User,
	) (onboarding.Result, error)
}
