package invitation

import (
	"context"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
)

type Repository interface {
	Create(
		ctx context.Context,
		input CreateInput,
	) (Invitation, error)

	FindByTokenDigest(
		ctx context.Context,
		tokenDigest string,
	) (Invitation, error)

	ListByTenantID(
		ctx context.Context,
		tenantID string,
	) ([]Invitation, error)

	Accept(
		ctx context.Context,
		invitationID string,
		userID string,
	) (membership.Membership, error)

	Revoke(
		ctx context.Context,
		invitationID string,
	) (Invitation, error)
}
