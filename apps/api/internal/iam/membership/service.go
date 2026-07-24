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

func (s *Service) ListDetailedByTenantID(
	ctx context.Context,
	tenantID string,
) ([]Member, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrInvalidInput
	}

	return s.repository.ListDetailedByTenantID(
		ctx,
		tenantID,
	)
}

func (s *Service) ChangeRoleForTenant(
	ctx context.Context,
	tenantID string,
	actorUserID string,
	actorRole Role,
	membershipID string,
	role Role,
) (Membership, error) {
	target, err := s.validateMutation(
		ctx,
		tenantID,
		actorUserID,
		actorRole,
		membershipID,
	)
	if err != nil {
		return Membership{}, err
	}

	if role != RoleAdmin && role != RoleMember {
		return Membership{}, ErrInvalidInput
	}

	if target.Membership.Status == StatusLeft {
		return Membership{}, ErrNotFound
	}

	if target.Membership.Status != StatusActive &&
		target.Membership.Status != StatusSuspended {
		return Membership{}, ErrForbidden
	}

	if target.Membership.Role == RoleOwner &&
		target.Membership.Status == StatusActive {
		if err := s.requireAdditionalActiveOwner(
			ctx,
			target.Membership.TenantID,
		); err != nil {
			return Membership{}, err
		}
	}

	return s.repository.UpdateRoleForTenant(
		ctx,
		target.Membership.ID,
		target.Membership.TenantID,
		role,
	)
}

func (s *Service) ChangeStatusForTenant(
	ctx context.Context,
	tenantID string,
	actorUserID string,
	actorRole Role,
	membershipID string,
	status Status,
) (Membership, error) {
	target, err := s.validateMutation(
		ctx,
		tenantID,
		actorUserID,
		actorRole,
		membershipID,
	)
	if err != nil {
		return Membership{}, err
	}

	if status != StatusActive &&
		status != StatusSuspended {
		return Membership{}, ErrInvalidInput
	}

	if target.Membership.Status != StatusActive &&
		target.Membership.Status != StatusSuspended {
		return Membership{}, ErrForbidden
	}

	if target.Membership.Role == RoleOwner &&
		target.Membership.Status == StatusActive &&
		status == StatusSuspended {
		if err := s.requireAdditionalActiveOwner(
			ctx,
			target.Membership.TenantID,
		); err != nil {
			return Membership{}, err
		}
	}

	return s.repository.UpdateStatusForTenant(
		ctx,
		target.Membership.ID,
		target.Membership.TenantID,
		status,
	)
}

func (s *Service) RemoveForTenant(
	ctx context.Context,
	tenantID string,
	actorUserID string,
	actorRole Role,
	membershipID string,
) error {
	target, err := s.validateMutation(
		ctx,
		tenantID,
		actorUserID,
		actorRole,
		membershipID,
	)
	if err != nil {
		return err
	}

	if target.Membership.Status == StatusLeft {
		return ErrNotFound
	}

	if target.Membership.Role == RoleOwner &&
		target.Membership.Status == StatusActive {
		if err := s.requireAdditionalActiveOwner(
			ctx,
			target.Membership.TenantID,
		); err != nil {
			return err
		}
	}

	_, err = s.repository.UpdateStatusForTenant(
		ctx,
		target.Membership.ID,
		target.Membership.TenantID,
		StatusLeft,
	)
	return err
}

func (s *Service) validateMutation(
	ctx context.Context,
	tenantID string,
	actorUserID string,
	actorRole Role,
	membershipID string,
) (Member, error) {
	tenantID = strings.TrimSpace(tenantID)
	actorUserID = strings.TrimSpace(actorUserID)
	membershipID = strings.TrimSpace(membershipID)

	if tenantID == "" ||
		actorUserID == "" ||
		membershipID == "" {
		return Member{}, ErrInvalidInput
	}

	target, err := s.repository.FindDetailedByID(
		ctx,
		membershipID,
	)
	if err != nil {
		return Member{}, err
	}

	if target.Membership.TenantID != tenantID {
		return Member{}, ErrNotFound
	}

	if target.Membership.UserID == actorUserID {
		return Member{}, ErrSelfModification
	}

	switch actorRole {
	case RoleOwner:
		return target, nil

	case RoleAdmin:
		if target.Membership.Role != RoleMember {
			return Member{}, ErrForbidden
		}

		return target, nil

	default:
		return Member{}, ErrForbidden
	}
}

func (s *Service) requireAdditionalActiveOwner(
	ctx context.Context,
	tenantID string,
) error {
	count, err := s.repository.CountActiveOwners(
		ctx,
		tenantID,
	)
	if err != nil {
		return fmt.Errorf(
			"count active owners: %w",
			err,
		)
	}

	if count <= 1 {
		return ErrLastOwner
	}

	return nil
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
