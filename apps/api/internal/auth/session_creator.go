package auth

import (
	"context"

	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
)

type SessionCreator interface {
	Create(
		ctx context.Context,
		userID string,
	) (session.CreatedSession, error)
}
