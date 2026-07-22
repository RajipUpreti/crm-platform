package auth

import (
	"context"

	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
)

type SessionCreator interface {
	Create(
		ctx context.Context,
		input session.CreateInput,
	) (session.CreatedSession, error)
}
