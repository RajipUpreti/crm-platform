package tenant

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(
	`^[a-z0-9]+(?:-[a-z0-9]+)*$`,
)

type Service struct {
	repository Repository
}

func NewService(
	repository Repository,
) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf(
			"tenant repository is required",
		)
	}

	return &Service{
		repository: repository,
	}, nil
}

func (s *Service) Create(
	ctx context.Context,
	input CreateInput,
) (Tenant, error) {
	input.Name = strings.TrimSpace(
		input.Name,
	)

	input.Slug = normalizeSlug(
		input.Slug,
	)

	if input.Name == "" {
		return Tenant{}, fmt.Errorf(
			"%w: tenant name is required",
			ErrInvalidInput,
		)
	}

	if input.Slug == "" {
		return Tenant{}, fmt.Errorf(
			"%w: tenant slug is required",
			ErrInvalidInput,
		)
	}

	if !slugPattern.MatchString(
		input.Slug,
	) {
		return Tenant{}, fmt.Errorf(
			"%w: tenant slug has an invalid format",
			ErrInvalidInput,
		)
	}

	createdTenant, err := s.repository.Create(
		ctx,
		input,
	)
	if err != nil {
		if errors.Is(
			err,
			ErrSlugAlreadyExists,
		) {
			return Tenant{},
				ErrSlugAlreadyExists
		}

		return Tenant{}, fmt.Errorf(
			"create tenant: %w",
			err,
		)
	}

	return createdTenant, nil
}

func (s *Service) FindByID(
	ctx context.Context,
	id string,
) (Tenant, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Tenant{}, ErrNotFound
	}

	foundTenant, err := s.repository.FindByID(
		ctx,
		id,
	)
	if err != nil {
		return Tenant{}, err
	}

	return validateAccessibleTenant(
		foundTenant,
	)
}

func (s *Service) FindBySlug(
	ctx context.Context,
	slug string,
) (Tenant, error) {
	slug = normalizeSlug(slug)
	if slug == "" {
		return Tenant{}, ErrNotFound
	}

	foundTenant, err := s.repository.FindBySlug(
		ctx,
		slug,
	)
	if err != nil {
		return Tenant{}, err
	}

	return validateAccessibleTenant(
		foundTenant,
	)
}

func (s *Service) ListByUserID(
	ctx context.Context,
	userID string,
) ([]Tenant, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidInput
	}

	tenants, err := s.repository.ListByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list user tenants: %w",
			err,
		)
	}

	return tenants, nil
}

func normalizeSlug(
	value string,
) string {
	return strings.ToLower(
		strings.TrimSpace(value),
	)
}

func validateAccessibleTenant(
	foundTenant Tenant,
) (Tenant, error) {
	switch foundTenant.Status {
	case StatusActive:
		return foundTenant, nil

	case StatusSuspended:
		return Tenant{}, ErrSuspended

	case StatusDeleted:
		return Tenant{}, ErrDeleted

	default:
		return Tenant{}, fmt.Errorf(
			"unsupported tenant status %q",
			foundTenant.Status,
		)
	}
}
