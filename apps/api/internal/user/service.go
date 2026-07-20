package user

import (
	"context"
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
	identity.Provider = strings.TrimSpace(
		identity.Provider,
	)

	identity.ProviderUserID = strings.TrimSpace(
		identity.ProviderUserID,
	)

	identity.Email = strings.ToLower(
		strings.TrimSpace(identity.Email),
	)

	if identity.Provider == "" ||
		identity.ProviderUserID == "" ||
		identity.Email == "" {
		return User{}, ErrInvalidIdentity
	}

	storedUser, err :=
		s.repository.UpsertFromIdentity(
			ctx,
			identity,
		)
	if err != nil {
		return User{}, fmt.Errorf(
			"synchronize user identity: %w",
			err,
		)
	}

	if storedUser.Status == StatusSuspended {
		return User{}, ErrSuspended
	}

	return storedUser, nil
}
