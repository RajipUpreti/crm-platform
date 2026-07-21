package session

import (
	"context"
	"time"
)

type Store interface {
	Create(
		ctx context.Context,
		tokenDigest string,
		session Session,
		ttl time.Duration,
	) error

	Find(
		ctx context.Context,
		tokenDigest string,
	) (Session, error)

	Delete(
		ctx context.Context,
		tokenDigest string,
	) error
}
