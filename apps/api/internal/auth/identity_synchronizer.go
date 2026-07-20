package auth

import (
	"context"

	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type IdentitySynchronizer interface {
	SynchronizeIdentity(
		ctx context.Context,
		identity user.Identity,
	) (user.User, error)
}
