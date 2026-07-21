package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Service struct {
	repository Repository
}

func NewService(
	repository Repository,
) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf(
			"user repository is required",
		)
	}

	return &Service{
		repository: repository,
	}, nil
}

func (s *Service) SynchronizeIdentity(
	ctx context.Context,
	identity Identity,
) (User, error) {
	identity = normalizeIdentity(identity)

	if err := validateIdentity(identity); err != nil {
		return User{}, err
	}

	storedUser, err := s.repository.UpsertFromIdentity(
		ctx,
		identity,
	)
	if err != nil {
		return User{}, fmt.Errorf(
			"synchronize user identity: %w",
			err,
		)
	}

	switch storedUser.Status {
	case StatusActive:
		return storedUser, nil

	case StatusSuspended:
		return User{}, ErrSuspended

	case StatusDeleted:
		return User{}, ErrNotFound

	default:
		return User{}, fmt.Errorf(
			"unsupported user status %q",
			storedUser.Status,
		)
	}
}

func (s *Service) FindByID(
	ctx context.Context,
	id string,
) (User, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return User{}, ErrNotFound
	}

	storedUser, err := s.repository.FindByID(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return User{}, ErrNotFound
		}

		return User{}, fmt.Errorf(
			"find user: %w",
			err,
		)
	}

	switch storedUser.Status {
	case StatusActive:
		return storedUser, nil

	case StatusSuspended:
		return User{}, ErrSuspended

	case StatusDeleted:
		return User{}, ErrNotFound

	default:
		return User{}, fmt.Errorf(
			"unsupported user status %q",
			storedUser.Status,
		)
	}
}

func normalizeIdentity(
	identity Identity,
) Identity {
	identity.Provider = strings.ToLower(
		strings.TrimSpace(identity.Provider),
	)

	identity.ProviderUserID = strings.TrimSpace(
		identity.ProviderUserID,
	)

	identity.Email = strings.ToLower(
		strings.TrimSpace(identity.Email),
	)

	identity.FirstName = strings.TrimSpace(
		identity.FirstName,
	)

	identity.LastName = strings.TrimSpace(
		identity.LastName,
	)

	identity.DisplayName = strings.TrimSpace(
		identity.DisplayName,
	)

	return identity
}

func validateIdentity(
	identity Identity,
) error {
	if identity.Provider == "" {
		return fmt.Errorf(
			"%w: provider is required",
			ErrInvalidIdentity,
		)
	}

	if identity.ProviderUserID == "" {
		return fmt.Errorf(
			"%w: provider user ID is required",
			ErrInvalidIdentity,
		)
	}

	if identity.Email == "" {
		return fmt.Errorf(
			"%w: email is required",
			ErrInvalidIdentity,
		)
	}

	if !identity.EmailVerified {
		return fmt.Errorf(
			"%w: verified email is required",
			ErrInvalidIdentity,
		)
	}

	return nil
}
