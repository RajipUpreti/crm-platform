package invitation

import (
	"context"
	"time"

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

	FindByIDAndTenant(
		ctx context.Context,
		invitationID string,
		tenantID string,
	) (Invitation, error)

	ListByTenantID(
		ctx context.Context,
		tenantID string,
		status *Status,
	) ([]Invitation, error)

	Accept(
		ctx context.Context,
		invitationID string,
		userID string,
	) (membership.Membership, error)

	RevokeForTenant(
		ctx context.Context,
		invitationID string,
		tenantID string,
	) (Invitation, error)

	ReplaceToken(
		ctx context.Context,
		invitationID string,
		tenantID string,
		tokenDigest string,
		expiresAt time.Time,
	) (Invitation, error)
}
