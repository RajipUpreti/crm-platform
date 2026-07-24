package invitation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
)

const invitationTokenBytes = 32

func generateToken() (string, error) {
	randomBytes := make([]byte, invitationTokenBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf(
			"generate invitation token: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(
		randomBytes,
	), nil
}

func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type Service struct {
	repository Repository
	ttl        time.Duration
	now        func() time.Time
}

func NewService(
	repository Repository,
	ttl time.Duration,
) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf(
			"invitation repository is required",
		)
	}

	if ttl <= 0 {
		return nil, fmt.Errorf(
			"invitation TTL must be positive",
		)
	}

	return &Service{
		repository: repository,
		ttl:        ttl,
		now:        time.Now,
	}, nil
}

func (s *Service) Create(
	ctx context.Context,
	tenantID string,
	invitedByUserID string,
	email string,
	role membership.Role,
) (CreatedInvitation, error) {
	tenantID = strings.TrimSpace(tenantID)
	invitedByUserID = strings.TrimSpace(invitedByUserID)
	email = strings.ToLower(
		strings.TrimSpace(email),
	)

	if tenantID == "" ||
		invitedByUserID == "" ||
		email == "" {
		return CreatedInvitation{}, ErrInvalidInput
	}

	if role != membership.RoleAdmin &&
		role != membership.RoleMember {
		return CreatedInvitation{}, fmt.Errorf(
			"%w: unsupported invitation role %q",
			ErrInvalidInput,
			role,
		)
	}

	token, err := generateToken()
	if err != nil {
		return CreatedInvitation{}, err
	}

	now := s.now().UTC()

	created, err := s.repository.Create(
		ctx,
		CreateInput{
			TenantID:        tenantID,
			InvitedByUserID: invitedByUserID,
			Email:           email,
			Role:            role,
			ExpiresAt:       now.Add(s.ttl),
			TokenDigest:     tokenDigest(token),
		},
	)
	if err != nil {
		if errors.Is(err, ErrAlreadyPending) {
			return CreatedInvitation{},
				ErrAlreadyPending
		}

		return CreatedInvitation{}, fmt.Errorf(
			"create invitation: %w",
			err,
		)
	}

	return CreatedInvitation{
		Invitation: created,
		Token:      token,
	}, nil
}

func (s *Service) ListByTenantID(
	ctx context.Context,
	tenantID string,
	status *Status,
) ([]Invitation, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrInvalidInput
	}

	if status != nil && !isValidStatus(*status) {
		return nil, ErrInvalidInput
	}

	found, err := s.repository.ListByTenantID(
		ctx,
		tenantID,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list invitations: %w",
			err,
		)
	}

	now := s.now().UTC()
	result := make([]Invitation, 0, len(found))
	for _, current := range found {
		current = invitationWithEffectiveStatus(
			current,
			now,
		)
		if status == nil || current.Status == *status {
			result = append(result, current)
		}
	}

	return result, nil
}

func (s *Service) RevokeForTenant(
	ctx context.Context,
	invitationID string,
	tenantID string,
) (Invitation, error) {
	found, err := s.findForTenant(
		ctx,
		invitationID,
		tenantID,
	)
	if err != nil {
		return Invitation{}, err
	}

	switch found.Status {
	case StatusPending:
	case StatusAccepted:
		return Invitation{}, ErrAlreadyAccepted
	case StatusRevoked:
		return Invitation{}, ErrRevoked
	case StatusExpired:
		return Invitation{}, ErrExpired
	default:
		return Invitation{}, ErrInvalidInput
	}

	revoked, err := s.repository.RevokeForTenant(
		ctx,
		found.ID,
		found.TenantID,
	)
	if err != nil {
		return Invitation{}, fmt.Errorf(
			"revoke invitation: %w",
			err,
		)
	}

	return revoked, nil
}

func (s *Service) ResendForTenant(
	ctx context.Context,
	invitationID string,
	tenantID string,
) (CreatedInvitation, error) {
	found, err := s.findForTenant(
		ctx,
		invitationID,
		tenantID,
	)
	if err != nil {
		return CreatedInvitation{}, err
	}

	switch found.Status {
	case StatusPending, StatusExpired:
	case StatusAccepted:
		return CreatedInvitation{}, ErrAlreadyAccepted
	case StatusRevoked:
		return CreatedInvitation{}, ErrRevoked
	default:
		return CreatedInvitation{}, ErrInvalidInput
	}

	token, err := generateToken()
	if err != nil {
		return CreatedInvitation{}, err
	}

	expiresAt := s.now().UTC().Add(s.ttl)
	replaced, err := s.repository.ReplaceToken(
		ctx,
		found.ID,
		found.TenantID,
		tokenDigest(token),
		expiresAt,
	)
	if err != nil {
		return CreatedInvitation{}, fmt.Errorf(
			"resend invitation: %w",
			err,
		)
	}

	return CreatedInvitation{
		Invitation: replaced,
		Token:      token,
	}, nil
}

func (s *Service) findForTenant(
	ctx context.Context,
	invitationID string,
	tenantID string,
) (Invitation, error) {
	invitationID = strings.TrimSpace(invitationID)
	tenantID = strings.TrimSpace(tenantID)
	if invitationID == "" || tenantID == "" {
		return Invitation{}, ErrInvalidInput
	}

	found, err := s.repository.FindByIDAndTenant(
		ctx,
		invitationID,
		tenantID,
	)
	if err != nil {
		return Invitation{}, err
	}

	return invitationWithEffectiveStatus(
		found,
		s.now().UTC(),
	), nil
}

func (s *Service) Accept(
	ctx context.Context,
	input AcceptInput,
) (membership.Membership, error) {
	input.Token = strings.TrimSpace(input.Token)
	input.UserID = strings.TrimSpace(input.UserID)
	input.UserEmail = strings.ToLower(
		strings.TrimSpace(input.UserEmail),
	)

	if input.Token == "" ||
		input.UserID == "" ||
		input.UserEmail == "" {
		return membership.Membership{},
			ErrInvalidInput
	}

	found, err := s.repository.FindByTokenDigest(
		ctx,
		tokenDigest(input.Token),
	)
	if err != nil {
		return membership.Membership{}, err
	}

	switch found.Status {
	case StatusAccepted:
		return membership.Membership{},
			ErrAlreadyAccepted

	case StatusRevoked:
		return membership.Membership{},
			ErrRevoked

	case StatusExpired:
		return membership.Membership{},
			ErrExpired
	}

	if !found.ExpiresAt.After(s.now().UTC()) {
		return membership.Membership{},
			ErrExpired
	}

	if !strings.EqualFold(
		found.Email,
		input.UserEmail,
	) {
		return membership.Membership{},
			ErrEmailMismatch
	}

	createdMembership, err := s.repository.Accept(
		ctx,
		found.ID,
		input.UserID,
	)
	if err != nil {
		return membership.Membership{},
			fmt.Errorf(
				"accept invitation: %w",
				err,
			)
	}

	return createdMembership, nil
}

func invitationWithEffectiveStatus(
	value Invitation,
	now time.Time,
) Invitation {
	if value.Status == StatusPending &&
		!value.ExpiresAt.After(now) {
		value.Status = StatusExpired
	}

	return value
}

func isValidStatus(status Status) bool {
	switch status {
	case StatusPending,
		StatusAccepted,
		StatusRevoked,
		StatusExpired:
		return true
	default:
		return false
	}
}
