package requestcontext

import (
	"context"
	"errors"

	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type authenticationContextKey struct{}

type Authentication struct {
	Session session.Session
	User    user.User
}

var ErrAuthenticationMissing = errors.New("authentication context missing")

func WithAuthentication(
	ctx context.Context,
	authentication Authentication,
) context.Context {
	return context.WithValue(
		ctx,
		authenticationContextKey{},
		authentication,
	)
}

func AuthenticationFromContext(
	ctx context.Context,
) (Authentication, error) {
	authentication, ok :=
		ctx.Value(
			authenticationContextKey{},
		).(Authentication)

	if !ok {
		return Authentication{},
			ErrAuthenticationMissing
	}

	return authentication, nil
}
