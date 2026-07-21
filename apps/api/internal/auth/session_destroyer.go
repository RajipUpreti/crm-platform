package auth

import "context"

type SessionDestroyer interface {
	Delete(
		ctx context.Context,
		token string,
	) error
}
