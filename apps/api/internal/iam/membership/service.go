package membership

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(
	repository Repository,
) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf(
			"membership repository is required",
		)
	}

	return &Service{
		repository: repository,
		now:        time.Now,
	}, nil
}

func (s *Service) Create(
	ctx context.Context,
	input CreateInput,
) (Membership, error) {
	input.TenantID = strings.TrimSpace(
		input.TenantID,
	)

	input.UserID = strings.TrimSpace(
		input.UserID,
	)

	if input.TenantID == "" ||
		input.UserID == "" {
		return Membership{}, fmt.Errorf(
			"%w: tenant ID and user ID are required",
			ErrInvalidInput,
		)
	}

	if !isValidRole(input.Role) {
		return Membership{}, fmt.Errorf(
			"%w: unsupported role %q",
			ErrInvalidInput,
			input.Role,
		)
	}

	if input.Status == "" {
		input.Status = StatusActive
	}

	if !isValidStatus(input.Status) {
		return Membership{}, fmt.Errorf(
			"%w: unsupported status %q",
			ErrInvalidInput,
			input.Status,
		)
	}

	if input.Status == StatusActive &&
		input.JoinedAt == nil {
		joinedAt := s.now().UTC()
		input.JoinedAt = &joinedAt
	}

	createdMembership, err := s.repository.Create(
		ctx,
		input,
	)
	if err != nil {
		if errors.Is(
			err,
			ErrAlreadyExists,
		) {
			return Membership{},
				ErrAlreadyExists
		}

		return Membership{}, fmt.Errorf(
			"create membership: %w",
			err,
		)
	}

	return createdMembership, nil
}

func (s *Service) FindActiveByTenantAndUser(
	ctx context.Context,
	tenantID string,
	userID string,
) (Membership, error) {
	tenantID = strings.TrimSpace(
		tenantID,
	)

	userID = strings.TrimSpace(
		userID,
	)

	if tenantID == "" || userID == "" {
		return Membership{},
			ErrInvalidInput
	}

	foundMembership, err := s.repository.FindByTenantAndUser(
		ctx,
		tenantID,
		userID,
	)
	if err != nil {
		return Membership{}, err
	}

	switch foundMembership.Status {
	case StatusActive:
		return foundMembership, nil

	case StatusSuspended:
		return Membership{}, ErrSuspended

	case StatusInvited, StatusLeft:
		return Membership{}, ErrInactive

	default:
		return Membership{}, fmt.Errorf(
			"unsupported membership status %q",
			foundMembership.Status,
		)
	}
}

func (s *Service) ListByTenantID(
	ctx context.Context,
	tenantID string,
) ([]Membership, error) {
	tenantID = strings.TrimSpace(
		tenantID,
	)

	if tenantID == "" {
		return nil, ErrInvalidInput
	}

	return s.repository.ListByTenantID(
		ctx,
		tenantID,
	)
}

func (s *Service) ListByUserID(
	ctx context.Context,
	userID string,
) ([]Membership, error) {
	userID = strings.TrimSpace(
		userID,
	)

	if userID == "" {
		return nil, ErrInvalidInput
	}

	return s.repository.ListByUserID(
		ctx,
		userID,
	)
}

func (s *Service) ChangeRole(
	ctx context.Context,
	membershipID string,
	role Role,
) (Membership, error) {
	membershipID = strings.TrimSpace(
		membershipID,
	)

	if membershipID == "" ||
		!isValidRole(role) {
		return Membership{},
			ErrInvalidInput
	}

	return s.repository.UpdateRole(
		ctx,
		membershipID,
		role,
	)
}

func (s *Service) ChangeStatus(
	ctx context.Context,
	membershipID string,
	status Status,
) (Membership, error) {
	membershipID = strings.TrimSpace(
		membershipID,
	)

	if membershipID == "" ||
		!isValidStatus(status) {
		return Membership{},
			ErrInvalidInput
	}

	return s.repository.UpdateStatus(
		ctx,
		membershipID,
		status,
	)
}

func isValidRole(
	role Role,
) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember:
		return true

	default:
		return false
	}
}

func isValidStatus(
	status Status,
) bool {
	switch status {
	case StatusActive,
		StatusInvited,
		StatusSuspended,
		StatusLeft:
		return true

	default:
		return false
	}
}
