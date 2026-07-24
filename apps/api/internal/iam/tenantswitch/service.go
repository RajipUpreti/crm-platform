package tenantswitch

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
)

type TenantReader interface {
	FindByID(
		ctx context.Context,
		id string,
	) (tenant.Tenant, error)
}

type MembershipReader interface {
	FindActiveByTenantAndUser(
		ctx context.Context,
		tenantID string,
		userID string,
	) (membership.Membership, error)
}

type SessionCreator interface {
	Create(
		ctx context.Context,
		input session.CreateInput,
	) (session.CreatedSession, error)
}

type SessionDestroyer interface {
	Delete(
		ctx context.Context,
		token string,
	) error
}

type Result struct {
	Tenant tenant.Tenant `json:"tenant"`

	Membership membership.Membership `json:"membership"`

	Session session.CreatedSession `json:"-"`
}

type Service struct {
	tenants TenantReader

	memberships MembershipReader

	sessionCreator SessionCreator

	sessionDestroyer SessionDestroyer
}

func NewService(
	tenants TenantReader,
	memberships MembershipReader,
	sessionCreator SessionCreator,
	sessionDestroyer SessionDestroyer,
) (*Service, error) {
	if tenants == nil {
		return nil, errors.New(
			"tenant reader is required",
		)
	}

	if memberships == nil {
		return nil, errors.New(
			"membership reader is required",
		)
	}

	if sessionCreator == nil {
		return nil, errors.New(
			"session creator is required",
		)
	}

	if sessionDestroyer == nil {
		return nil, errors.New(
			"session destroyer is required",
		)
	}

	return &Service{
		tenants:          tenants,
		memberships:      memberships,
		sessionCreator:   sessionCreator,
		sessionDestroyer: sessionDestroyer,
	}, nil
}

func (s *Service) Switch(
	ctx context.Context,
	userID string,
	currentToken string,
	targetTenantID string,
) (Result, error) {
	userID = strings.TrimSpace(userID)

	currentToken = strings.TrimSpace(
		currentToken,
	)

	targetTenantID = strings.TrimSpace(
		targetTenantID,
	)

	if userID == "" ||
		currentToken == "" ||
		targetTenantID == "" {
		return Result{},
			fmt.Errorf(
				"invalid tenant switch input",
			)
	}

	targetTenant, err := s.tenants.FindByID(
		ctx,
		targetTenantID,
	)
	if err != nil {
		return Result{}, err
	}

	targetMembership, err := s.memberships.
		FindActiveByTenantAndUser(
			ctx,
			targetTenant.ID,
			userID,
		)
	if err != nil {
		return Result{}, err
	}

	createdSession, err := s.sessionCreator.Create(
		ctx,
		session.CreateInput{
			UserID: userID,

			TenantID: targetTenant.ID,

			MembershipID: targetMembership.ID,
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"create switched tenant session: %w",
			err,
		)
	}

	if err := s.sessionDestroyer.Delete(
		ctx,
		currentToken,
	); err != nil {
		_ = s.sessionDestroyer.Delete(
			ctx,
			createdSession.Token,
		)

		return Result{}, fmt.Errorf(
			"delete previous session: %w",
			err,
		)
	}

	return Result{
		Tenant:     targetTenant,
		Membership: targetMembership,
		Session:    createdSession,
	}, nil
}
