package authorization

import (
	"errors"
	"testing"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/permission"
)

func TestOwnerHasEveryPermission(
	t *testing.T,
) {
	t.Parallel()

	service := NewService()

	for _, requiredPermission := range permission.All() {
		if err := service.Authorize(
			membership.RoleOwner,
			requiredPermission,
		); err != nil {
			t.Fatalf(
				"OWNER denied %q: %v",
				requiredPermission,
				err,
			)
		}
	}
}

func TestAdminCanInviteMembers(
	t *testing.T,
) {
	t.Parallel()

	service := NewService()

	err := service.Authorize(
		membership.RoleAdmin,
		permission.MemberInvite,
	)
	if err != nil {
		t.Fatalf(
			"Authorize() error = %v",
			err,
		)
	}
}

func TestAdminCannotDeleteTenant(
	t *testing.T,
) {
	t.Parallel()

	service := NewService()

	err := service.Authorize(
		membership.RoleAdmin,
		permission.TenantDelete,
	)

	if !errors.Is(
		err,
		ErrPermissionDenied,
	) {
		t.Fatalf(
			"error = %v; expected ErrPermissionDenied",
			err,
		)
	}
}

func TestMemberCannotInviteMembers(
	t *testing.T,
) {
	t.Parallel()

	service := NewService()

	err := service.Authorize(
		membership.RoleMember,
		permission.MemberInvite,
	)

	if !errors.Is(
		err,
		ErrPermissionDenied,
	) {
		t.Fatalf(
			"error = %v; expected ErrPermissionDenied",
			err,
		)
	}
}

func TestMemberCanCreateContact(
	t *testing.T,
) {
	t.Parallel()

	service := NewService()

	err := service.Authorize(
		membership.RoleMember,
		permission.ContactCreate,
	)
	if err != nil {
		t.Fatalf(
			"Authorize() error = %v",
			err,
		)
	}
}

func TestUnknownRoleIsRejected(
	t *testing.T,
) {
	t.Parallel()

	service := NewService()

	err := service.Authorize(
		membership.Role("SUPER_ADMIN"),
		permission.TenantRead,
	)

	if !errors.Is(
		err,
		ErrInvalidRole,
	) {
		t.Fatalf(
			"error = %v; expected ErrInvalidRole",
			err,
		)
	}
}

func TestUnknownPermissionIsRejected(
	t *testing.T,
) {
	t.Parallel()

	service := NewService()

	err := service.Authorize(
		membership.RoleOwner,
		permission.Permission(
			"unknown.permission",
		),
	)

	if !errors.Is(
		err,
		ErrInvalidPermission,
	) {
		t.Fatalf(
			"error = %v; expected ErrInvalidPermission",
			err,
		)
	}
}
