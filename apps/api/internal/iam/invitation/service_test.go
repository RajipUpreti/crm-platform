package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
)

type fakeRepository struct {
	createInput  CreateInput
	createResult Invitation
	createError  error

	findDigest string
	findResult Invitation
	findError  error

	findByIDResult Invitation
	findByIDError  error

	listResult []Invitation
	listError  error

	acceptInvitationID string
	acceptUserID       string
	acceptResult       membership.Membership
	acceptError        error

	revokeResult   Invitation
	revokeError    error
	revokeID       string
	revokeTenantID string

	replaceResult    Invitation
	replaceError     error
	replaceID        string
	replaceTenantID  string
	replaceDigest    string
	replaceExpiresAt time.Time
}

func (f *fakeRepository) Create(
	ctx context.Context,
	input CreateInput,
) (Invitation, error) {
	f.createInput = input
	return f.createResult, f.createError
}

func (f *fakeRepository) FindByTokenDigest(
	ctx context.Context,
	tokenDigest string,
) (Invitation, error) {
	f.findDigest = tokenDigest
	return f.findResult, f.findError
}

func (f *fakeRepository) ListByTenantID(
	ctx context.Context,
	tenantID string,
	status *Status,
) ([]Invitation, error) {
	return f.listResult, f.listError
}

func (f *fakeRepository) FindByIDAndTenant(
	ctx context.Context,
	invitationID string,
	tenantID string,
) (Invitation, error) {
	return f.findByIDResult, f.findByIDError
}

func (f *fakeRepository) Accept(
	ctx context.Context,
	invitationID string,
	userID string,
) (membership.Membership, error) {
	f.acceptInvitationID = invitationID
	f.acceptUserID = userID

	return f.acceptResult, f.acceptError
}

func (f *fakeRepository) RevokeForTenant(
	ctx context.Context,
	invitationID string,
	tenantID string,
) (Invitation, error) {
	f.revokeID = invitationID
	f.revokeTenantID = tenantID
	return f.revokeResult, f.revokeError
}

func (f *fakeRepository) ReplaceToken(
	ctx context.Context,
	invitationID string,
	tenantID string,
	digest string,
	expiresAt time.Time,
) (Invitation, error) {
	f.replaceID = invitationID
	f.replaceTenantID = tenantID
	f.replaceDigest = digest
	f.replaceExpiresAt = expiresAt
	return f.replaceResult, f.replaceError
}

func TestServiceCreateCreatesInvitation(
	t *testing.T,
) {
	t.Parallel()

	fixedNow := time.Date(
		2026,
		time.July,
		22,
		4,
		0,
		0,
		0,
		time.UTC,
	)

	repository := &fakeRepository{
		createResult: Invitation{
			ID:              "invitation-id",
			TenantID:        "tenant-id",
			InvitedByUserID: "inviter-id",
			Email:           "newuser@example.com",
			Role:            membership.RoleMember,
			Status:          StatusPending,
			ExpiresAt:       fixedNow.Add(7 * 24 * time.Hour),
		},
	}

	service, err := NewService(
		repository,
		7*24*time.Hour,
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

	result, err := service.Create(
		context.Background(),
		" tenant-id ",
		" inviter-id ",
		" NewUser@Example.COM ",
		membership.RoleMember,
	)
	if err != nil {
		t.Fatalf(
			"Create() error = %v",
			err,
		)
	}

	if result.Token == "" {
		t.Fatal(
			"Create() returned an empty token",
		)
	}

	if repository.createInput.TenantID !=
		"tenant-id" {
		t.Fatalf(
			"tenant ID = %q; expected tenant-id",
			repository.createInput.TenantID,
		)
	}

	if repository.createInput.InvitedByUserID !=
		"inviter-id" {
		t.Fatalf(
			"invited by user ID = %q; expected inviter-id",
			repository.createInput.InvitedByUserID,
		)
	}

	if repository.createInput.Email !=
		"newuser@example.com" {
		t.Fatalf(
			"email = %q; expected normalized email",
			repository.createInput.Email,
		)
	}

	if repository.createInput.Role !=
		membership.RoleMember {
		t.Fatalf(
			"role = %q; expected MEMBER",
			repository.createInput.Role,
		)
	}

	if repository.createInput.TokenDigest == "" {
		t.Fatal(
			"token digest was not provided",
		)
	}

	if repository.createInput.TokenDigest ==
		result.Token {
		t.Fatal(
			"raw invitation token was sent to repository",
		)
	}

	if len(repository.createInput.TokenDigest) != 64 {
		t.Fatalf(
			"token digest length = %d; expected 64",
			len(repository.createInput.TokenDigest),
		)
	}

	expectedExpiration := fixedNow.Add(7 * 24 * time.Hour)

	if !repository.createInput.ExpiresAt.Equal(
		expectedExpiration,
	) {
		t.Fatalf(
			"expiresAt = %v; expected %v",
			repository.createInput.ExpiresAt,
			expectedExpiration,
		)
	}
}

func TestServiceCreateRejectsMissingInput(
	t *testing.T,
) {
	t.Parallel()

	service, err := NewService(
		&fakeRepository{},
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Create(
		context.Background(),
		"",
		"inviter-id",
		"newuser@example.com",
		membership.RoleMember,
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"error = %v; expected ErrInvalidInput",
			err,
		)
	}
}

func TestServiceCreateRejectsOwnerRole(
	t *testing.T,
) {
	t.Parallel()

	service, err := NewService(
		&fakeRepository{},
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Create(
		context.Background(),
		"tenant-id",
		"inviter-id",
		"newuser@example.com",
		membership.RoleOwner,
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"error = %v; expected ErrInvalidInput",
			err,
		)
	}
}

func TestServiceCreateReturnsAlreadyPending(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		createError: ErrAlreadyPending,
	}

	service, err := NewService(
		repository,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Create(
		context.Background(),
		"tenant-id",
		"inviter-id",
		"newuser@example.com",
		membership.RoleAdmin,
	)

	if !errors.Is(
		err,
		ErrAlreadyPending,
	) {
		t.Fatalf(
			"error = %v; expected ErrAlreadyPending",
			err,
		)
	}
}

func TestServiceAcceptCreatesMembership(
	t *testing.T,
) {
	t.Parallel()

	fixedNow := time.Date(
		2026,
		time.July,
		22,
		4,
		0,
		0,
		0,
		time.UTC,
	)

	repository := &fakeRepository{
		findResult: Invitation{
			ID:        "invitation-id",
			TenantID:  "tenant-id",
			Email:     "newuser@example.com",
			Role:      membership.RoleMember,
			Status:    StatusPending,
			ExpiresAt: fixedNow.Add(time.Hour),
		},
		acceptResult: membership.Membership{
			ID:       "membership-id",
			TenantID: "tenant-id",
			UserID:   "user-id",
			Role:     membership.RoleMember,
			Status:   membership.StatusActive,
		},
	}

	service, err := NewService(
		repository,
		7*24*time.Hour,
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

	result, err := service.Accept(
		context.Background(),
		AcceptInput{
			Token:     "raw-invitation-token",
			UserID:    "user-id",
			UserEmail: " NewUser@Example.COM ",
		},
	)
	if err != nil {
		t.Fatalf(
			"Accept() error = %v",
			err,
		)
	}

	if repository.findDigest == "" {
		t.Fatal(
			"token digest was not used for lookup",
		)
	}

	if repository.findDigest ==
		"raw-invitation-token" {
		t.Fatal(
			"raw invitation token was used for lookup",
		)
	}

	if len(repository.findDigest) != 64 {
		t.Fatalf(
			"digest length = %d; expected 64",
			len(repository.findDigest),
		)
	}

	if repository.acceptInvitationID !=
		"invitation-id" {
		t.Fatalf(
			"invitation ID = %q; expected invitation-id",
			repository.acceptInvitationID,
		)
	}

	if repository.acceptUserID !=
		"user-id" {
		t.Fatalf(
			"user ID = %q; expected user-id",
			repository.acceptUserID,
		)
	}

	if result.ID != "membership-id" {
		t.Fatalf(
			"membership ID = %q; expected membership-id",
			result.ID,
		)
	}
}

func TestServiceAcceptRejectsEmailMismatch(
	t *testing.T,
) {
	t.Parallel()

	fixedNow := time.Date(
		2026,
		time.July,
		22,
		4,
		0,
		0,
		0,
		time.UTC,
	)

	repository := &fakeRepository{
		findResult: Invitation{
			ID:        "invitation-id",
			Email:     "invited@example.com",
			Status:    StatusPending,
			ExpiresAt: fixedNow.Add(time.Hour),
		},
	}

	service, err := NewService(
		repository,
		7*24*time.Hour,
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

	_, err = service.Accept(
		context.Background(),
		AcceptInput{
			Token:     "invitation-token",
			UserID:    "user-id",
			UserEmail: "different@example.com",
		},
	)

	if !errors.Is(
		err,
		ErrEmailMismatch,
	) {
		t.Fatalf(
			"error = %v; expected ErrEmailMismatch",
			err,
		)
	}

	if repository.acceptInvitationID != "" {
		t.Fatal(
			"repository Accept() should not be called",
		)
	}
}

func TestServiceAcceptRejectsExpiredInvitation(
	t *testing.T,
) {
	t.Parallel()

	fixedNow := time.Date(
		2026,
		time.July,
		22,
		4,
		0,
		0,
		0,
		time.UTC,
	)

	repository := &fakeRepository{
		findResult: Invitation{
			ID:        "invitation-id",
			Email:     "newuser@example.com",
			Status:    StatusPending,
			ExpiresAt: fixedNow.Add(-time.Minute),
		},
	}

	service, err := NewService(
		repository,
		7*24*time.Hour,
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

	_, err = service.Accept(
		context.Background(),
		AcceptInput{
			Token:     "invitation-token",
			UserID:    "user-id",
			UserEmail: "newuser@example.com",
		},
	)

	if !errors.Is(err, ErrExpired) {
		t.Fatalf(
			"error = %v; expected ErrExpired",
			err,
		)
	}
}

func TestServiceAcceptRejectsAcceptedInvitation(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findResult: Invitation{
			ID:     "invitation-id",
			Status: StatusAccepted,
		},
	}

	service, err := NewService(
		repository,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Accept(
		context.Background(),
		AcceptInput{
			Token:     "invitation-token",
			UserID:    "user-id",
			UserEmail: "newuser@example.com",
		},
	)

	if !errors.Is(
		err,
		ErrAlreadyAccepted,
	) {
		t.Fatalf(
			"error = %v; expected ErrAlreadyAccepted",
			err,
		)
	}
}

func TestServiceAcceptRejectsRevokedInvitation(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findResult: Invitation{
			ID:     "invitation-id",
			Status: StatusRevoked,
		},
	}

	service, err := NewService(
		repository,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Accept(
		context.Background(),
		AcceptInput{
			Token:     "invitation-token",
			UserID:    "user-id",
			UserEmail: "newuser@example.com",
		},
	)

	if !errors.Is(
		err,
		ErrRevoked,
	) {
		t.Fatalf(
			"error = %v; expected ErrRevoked",
			err,
		)
	}
}

func TestServiceListAppliesExpiryAndStatusFilter(
	t *testing.T,
) {
	t.Parallel()

	fixedNow := time.Date(
		2026,
		time.July,
		24,
		4,
		0,
		0,
		0,
		time.UTC,
	)
	repository := &fakeRepository{
		listResult: []Invitation{
			{
				ID:        "expired-invitation",
				Status:    StatusPending,
				ExpiresAt: fixedNow,
			},
			{
				ID:        "pending-invitation",
				Status:    StatusPending,
				ExpiresAt: fixedNow.Add(time.Hour),
			},
		},
	}
	service, _ := NewService(repository, 7*24*time.Hour)
	service.now = func() time.Time { return fixedNow }
	status := StatusExpired

	result, err := service.ListByTenantID(
		context.Background(),
		"tenant-id",
		&status,
	)
	if err != nil {
		t.Fatalf("ListByTenantID() error = %v", err)
	}
	if len(result) != 1 ||
		result[0].ID != "expired-invitation" ||
		result[0].Status != StatusExpired {
		t.Fatalf("result = %#v; expected expired invitation", result)
	}
}

func TestServiceRevokeUsesTenantScopedMutation(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findByIDResult: Invitation{
			ID:        "invitation-id",
			TenantID:  "tenant-id",
			Status:    StatusPending,
			ExpiresAt: time.Now().Add(time.Hour),
		},
		revokeResult: Invitation{
			ID:       "invitation-id",
			TenantID: "tenant-id",
			Status:   StatusRevoked,
		},
	}
	service, _ := NewService(repository, 7*24*time.Hour)

	_, err := service.RevokeForTenant(
		context.Background(),
		"invitation-id",
		"tenant-id",
	)
	if err != nil {
		t.Fatalf("RevokeForTenant() error = %v", err)
	}
	if repository.revokeID != "invitation-id" ||
		repository.revokeTenantID != "tenant-id" {
		t.Fatalf(
			"revoke scope = (%q, %q)",
			repository.revokeID,
			repository.revokeTenantID,
		)
	}
}

func TestServiceRevokeRejectsAcceptedInvitation(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findByIDResult: Invitation{
			ID:       "invitation-id",
			TenantID: "tenant-id",
			Status:   StatusAccepted,
		},
	}
	service, _ := NewService(repository, 7*24*time.Hour)

	_, err := service.RevokeForTenant(
		context.Background(),
		"invitation-id",
		"tenant-id",
	)
	if !errors.Is(err, ErrAlreadyAccepted) {
		t.Fatalf("error = %v; expected ErrAlreadyAccepted", err)
	}
	if repository.revokeID != "" {
		t.Fatal("accepted invitation was revoked")
	}
}

func TestServiceTenantIsolationReturnsNotFound(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findByIDError: ErrNotFound,
	}
	service, _ := NewService(repository, 7*24*time.Hour)

	_, err := service.RevokeForTenant(
		context.Background(),
		"other-tenant-invitation",
		"tenant-id",
	)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v; expected ErrNotFound", err)
	}
}

func TestServiceResendRotatesTokenAndExtendsExpiry(
	t *testing.T,
) {
	t.Parallel()

	fixedNow := time.Date(
		2026,
		time.July,
		24,
		4,
		0,
		0,
		0,
		time.UTC,
	)
	repository := &fakeRepository{
		findByIDResult: Invitation{
			ID:        "invitation-id",
			TenantID:  "tenant-id",
			Status:    StatusPending,
			ExpiresAt: fixedNow.Add(time.Hour),
		},
		replaceResult: Invitation{
			ID:        "invitation-id",
			TenantID:  "tenant-id",
			Status:    StatusPending,
			ExpiresAt: fixedNow.Add(7 * 24 * time.Hour),
		},
	}
	service, _ := NewService(repository, 7*24*time.Hour)
	service.now = func() time.Time { return fixedNow }

	result, err := service.ResendForTenant(
		context.Background(),
		"invitation-id",
		"tenant-id",
	)
	if err != nil {
		t.Fatalf("ResendForTenant() error = %v", err)
	}
	if result.Token == "" {
		t.Fatal("resend returned an empty token")
	}
	if repository.replaceDigest != tokenDigest(result.Token) {
		t.Fatal("repository did not receive the new token digest")
	}
	if repository.replaceDigest == tokenDigest("old-token") {
		t.Fatal("old token digest was not replaced")
	}

	expectedExpiry := fixedNow.Add(7 * 24 * time.Hour)
	if !repository.replaceExpiresAt.Equal(expectedExpiry) {
		t.Fatalf(
			"expiry = %v; expected %v",
			repository.replaceExpiresAt,
			expectedExpiry,
		)
	}
}

func TestServiceResendRejectsRevokedInvitation(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findByIDResult: Invitation{
			ID:       "invitation-id",
			TenantID: "tenant-id",
			Status:   StatusRevoked,
		},
	}
	service, _ := NewService(repository, 7*24*time.Hour)

	_, err := service.ResendForTenant(
		context.Background(),
		"invitation-id",
		"tenant-id",
	)
	if !errors.Is(err, ErrRevoked) {
		t.Fatalf("error = %v; expected ErrRevoked", err)
	}
	if repository.replaceID != "" {
		t.Fatal("revoked invitation token was replaced")
	}
}
