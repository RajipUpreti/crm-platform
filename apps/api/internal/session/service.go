package session

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const maxTokenGenerationAttempts = 3

type Service struct {
	store Store
	ttl   time.Duration
	now   func() time.Time
}

func NewService(
	store Store,
	ttl time.Duration,
) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf(
			"session store is required",
		)
	}

	if ttl <= 0 {
		return nil, fmt.Errorf(
			"session TTL must be positive",
		)
	}

	return &Service{
		store: store,
		ttl:   ttl,
		now:   time.Now,
	}, nil
}

func (s *Service) Create(
	ctx context.Context,
	userID string,
) (CreatedSession, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return CreatedSession{}, fmt.Errorf(
			"%w: user ID is required",
			ErrInvalid,
		)
	}

	now := s.now().UTC()
	expiresAt := now.Add(s.ttl)

	storedSession := Session{
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	for attempt := 0; attempt <
		maxTokenGenerationAttempts; attempt++ {
		token, err := generateToken()
		if err != nil {
			return CreatedSession{}, err
		}

		err = s.store.Create(
			ctx,
			tokenDigest(token),
			storedSession,
			s.ttl,
		)
		if err == nil {
			return CreatedSession{
				Token:     token,
				ExpiresAt: expiresAt,
			}, nil
		}
	}

	return CreatedSession{}, fmt.Errorf(
		"create unique session after %d attempts",
		maxTokenGenerationAttempts,
	)
}

func (s *Service) Find(
	ctx context.Context,
	token string,
) (Session, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Session{}, ErrInvalid
	}

	storedSession, err := s.store.Find(
		ctx,
		tokenDigest(token),
	)
	if err != nil {
		return Session{}, err
	}

	if !storedSession.ExpiresAt.After(
		s.now().UTC(),
	) {
		_ = s.store.Delete(
			ctx,
			tokenDigest(token),
		)

		return Session{}, ErrExpired
	}

	return storedSession, nil
}

func (s *Service) Delete(
	ctx context.Context,
	token string,
) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}

	return s.store.Delete(
		ctx,
		tokenDigest(token),
	)
}
