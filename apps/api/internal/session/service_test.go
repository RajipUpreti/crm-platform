package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTokenDigest(t *testing.T) {
	t.Parallel()

	const expected = "34d328009b123fbbb0dc93f18b3e6de1ecf7b1a5783c33dff7ffe1926f09e943"

	if actual := tokenDigest("raw-token"); actual != expected {
		t.Fatalf(
			"tokenDigest() = %q; expected %q",
			actual,
			expected,
		)
	}
}

type fakeStore struct {
	createdDigest  string
	createdSession Session
	createdTTL     time.Duration
	createError    error

	findResult Session
	findError  error

	deletedDigest string
}

func (f *fakeStore) Create(
	ctx context.Context,
	digest string,
	storedSession Session,
	ttl time.Duration,
) error {
	f.createdDigest = digest
	f.createdSession = storedSession
	f.createdTTL = ttl

	return f.createError
}

func (f *fakeStore) Find(
	ctx context.Context,
	digest string,
) (Session, error) {
	return f.findResult, f.findError
}

func (f *fakeStore) Delete(
	ctx context.Context,
	digest string,
) error {
	f.deletedDigest = digest
	return nil
}

func TestServiceCreate(
	t *testing.T,
) {
	t.Parallel()

	store := &fakeStore{}
	ttl := 24 * time.Hour

	service, err := NewService(
		store,
		ttl,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	fixedNow := time.Date(
		2026,
		time.July,
		20,
		6,
		0,
		0,
		0,
		time.UTC,
	)

	service.now = func() time.Time {
		return fixedNow
	}

	created, err := service.Create(
		context.Background(),
		"user-id",
	)
	if err != nil {
		t.Fatalf(
			"Create() error = %v",
			err,
		)
	}

	if created.Token == "" {
		t.Fatal(
			"Create() returned an empty token",
		)
	}

	if store.createdDigest == created.Token {
		t.Fatal(
			"raw session token was stored",
		)
	}

	if len(store.createdDigest) != 64 {
		t.Fatalf(
			"digest length = %d; expected 64",
			len(store.createdDigest),
		)
	}

	if store.createdSession.UserID != "user-id" {
		t.Fatalf(
			"user ID = %q; expected user-id",
			store.createdSession.UserID,
		)
	}

	expectedExpiration :=
		fixedNow.Add(ttl)

	if !created.ExpiresAt.Equal(
		expectedExpiration,
	) {
		t.Fatalf(
			"expiration = %v; expected %v",
			created.ExpiresAt,
			expectedExpiration,
		)
	}
}

func TestServiceCreateRejectsMissingUserID(
	t *testing.T,
) {
	t.Parallel()

	service, err := NewService(
		&fakeStore{},
		time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Create(
		context.Background(),
		" ",
	)

	if !errors.Is(err, ErrInvalid) {
		t.Fatalf(
			"error = %v; expected ErrInvalid",
			err,
		)
	}
}

func TestServiceFindRejectsExpiredSession(
	t *testing.T,
) {
	t.Parallel()

	fixedNow := time.Date(
		2026,
		time.July,
		20,
		6,
		0,
		0,
		0,
		time.UTC,
	)

	store := &fakeStore{
		findResult: Session{
			UserID:    "user-id",
			CreatedAt: fixedNow.Add(-2 * time.Hour),
			ExpiresAt: fixedNow.Add(-time.Hour),
		},
	}

	service, err := NewService(
		store,
		time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	service.now = func() time.Time {
		return fixedNow
	}

	_, err = service.Find(
		context.Background(),
		"raw-token",
	)

	if !errors.Is(err, ErrExpired) {
		t.Fatalf(
			"error = %v; expected ErrExpired",
			err,
		)
	}

	if store.deletedDigest == "" {
		t.Fatal(
			"expired session was not deleted",
		)
	}
}
