package onboarding

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

const maxSlugAttempts = 5

var nonSlugCharacters = regexp.MustCompile(
	`[^a-z0-9]+`,
)

type Service struct {
	repository Repository
}

func NewService(
	repository Repository,
) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf(
			"onboarding repository is required",
		)
	}

	return &Service{
		repository: repository,
	}, nil
}

func (s *Service) EnsureTenantAccess(
	ctx context.Context,
	currentUser user.User,
) (Result, error) {
	if strings.TrimSpace(currentUser.ID) == "" {
		return Result{}, ErrInvalidUser
	}

	existing, err := s.repository.FindPrimaryAccess(
		ctx,
		currentUser.ID,
	)
	if err == nil {
		return existing, nil
	}

	if !errors.Is(
		err,
		membership.ErrNotFound,
	) {
		return Result{}, fmt.Errorf(
			"find existing tenant access: %w",
			err,
		)
	}

	tenantName := personalTenantName(currentUser)

	baseSlug := personalTenantSlug(currentUser)

	for attempt := 0; attempt <
		maxSlugAttempts; attempt++ {
		candidateSlug := baseSlug

		if attempt > 0 {
			candidateSlug = fmt.Sprintf(
				"%s-%d",
				baseSlug,
				attempt+1,
			)
		}

		result, err := s.repository.ProvisionPersonalTenant(
			ctx,
			currentUser.ID,
			tenantName,
			candidateSlug,
		)
		if err == nil {
			return result, nil
		}

		if errors.Is(
			err,
			tenant.ErrSlugAlreadyExists,
		) {
			continue
		}

		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrProvisioningFailed,
			err,
		)
	}

	return Result{}, fmt.Errorf(
		"%w: could not generate a unique tenant slug",
		ErrProvisioningFailed,
	)
}

func personalTenantName(
	currentUser user.User,
) string {
	displayName := strings.TrimSpace(
		currentUser.DisplayName,
	)

	if displayName == "" {
		displayName = strings.TrimSpace(
			strings.Join(
				[]string{
					currentUser.FirstName,
					currentUser.LastName,
				},
				" ",
			),
		)
	}

	if displayName == "" {
		displayName = strings.TrimSpace(
			strings.Split(
				currentUser.Email,
				"@",
			)[0],
		)
	}

	if displayName == "" {
		displayName = "Personal"
	}

	return displayName + "'s Workspace"
}

func personalTenantSlug(
	currentUser user.User,
) string {
	source := strings.TrimSpace(
		currentUser.DisplayName,
	)

	if source == "" {
		source = strings.TrimSpace(
			currentUser.FirstName,
		)
	}

	if source == "" {
		source = strings.Split(
			currentUser.Email,
			"@",
		)[0]
	}

	source = strings.ToLower(source)

	slug := nonSlugCharacters.
		ReplaceAllString(
			source,
			"-",
		)

	slug = strings.Trim(
		slug,
		"-",
	)

	if slug == "" {
		slug = "personal"
	}

	userSuffix := strings.ToLower(
		strings.ReplaceAll(
			currentUser.ID,
			"-",
			"",
		),
	)

	if len(userSuffix) > 8 {
		userSuffix = userSuffix[:8]
	}

	return slug + "-" + userSuffix
}
