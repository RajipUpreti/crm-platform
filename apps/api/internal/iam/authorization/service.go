package authorization

import (
	"fmt"
	"sort"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/permission"
)

type Service struct {
	rolePermissions map[membership.Role]map[permission.Permission]struct{}

	knownPermissions map[permission.Permission]struct{}
}

func NewService() *Service {
	knownPermissions := make(
		map[permission.Permission]struct{},
		len(permission.All()),
	)

	for _, requiredPermission := range permission.All() {
		knownPermissions[requiredPermission] = struct{}{}
	}

	return &Service{
		rolePermissions:  defaultRolePermissions(),
		knownPermissions: knownPermissions,
	}
}

func (s *Service) Authorize(
	role membership.Role,
	requiredPermission permission.Permission,
) error {
	if _, exists := s.knownPermissions[requiredPermission]; !exists {
		return fmt.Errorf(
			"%w: %q",
			ErrInvalidPermission,
			requiredPermission,
		)
	}

	rolePermissions, exists := s.rolePermissions[role]
	if !exists {
		return fmt.Errorf(
			"%w: %q",
			ErrInvalidRole,
			role,
		)
	}

	if _, allowed := rolePermissions[requiredPermission]; !allowed {
		return fmt.Errorf(
			"%w: role %q lacks %q",
			ErrPermissionDenied,
			role,
			requiredPermission,
		)
	}

	return nil
}

func (s *Service) HasPermission(
	role membership.Role,
	requiredPermission permission.Permission,
) bool {
	return s.Authorize(
		role,
		requiredPermission,
	) == nil
}

func (s *Service) PermissionsForRole(
	role membership.Role,
) ([]permission.Permission, error) {
	rolePermissions, exists := s.rolePermissions[role]
	if !exists {
		return nil, fmt.Errorf(
			"%w: %q",
			ErrInvalidRole,
			role,
		)
	}

	result := make(
		[]permission.Permission,
		0,
		len(rolePermissions),
	)

	for currentPermission := range rolePermissions {
		result = append(
			result,
			currentPermission,
		)
	}

	sort.Slice(
		result,
		func(i int, j int) bool {
			return result[i] < result[j]
		},
	)

	return result, nil
}

func defaultRolePermissions() map[membership.Role]map[permission.Permission]struct{} {
	ownerPermissions := permissionSet(
		permission.All()...,
	)

	adminPermissions := permissionSet(
		permission.TenantRead,
		permission.TenantUpdate,

		permission.MemberRead,
		permission.MemberInvite,
		permission.MemberUpdateRole,
		permission.MemberSuspend,
		permission.MemberRemove,

		permission.InvitationRead,
		permission.InvitationRevoke,

		permission.ContactRead,
		permission.ContactCreate,
		permission.ContactUpdate,
		permission.ContactDelete,

		permission.CompanyRead,
		permission.CompanyCreate,
		permission.CompanyUpdate,
		permission.CompanyDelete,

		permission.DealRead,
		permission.DealCreate,
		permission.DealUpdate,
		permission.DealDelete,
	)

	memberPermissions := permissionSet(
		permission.TenantRead,
		permission.MemberRead,

		permission.ContactRead,
		permission.ContactCreate,
		permission.ContactUpdate,

		permission.CompanyRead,
		permission.CompanyCreate,
		permission.CompanyUpdate,

		permission.DealRead,
		permission.DealCreate,
		permission.DealUpdate,
	)

	return map[membership.Role]map[permission.Permission]struct{}{
		membership.RoleOwner:  ownerPermissions,
		membership.RoleAdmin:  adminPermissions,
		membership.RoleMember: memberPermissions,
	}
}

func permissionSet(
	permissions ...permission.Permission,
) map[permission.Permission]struct{} {
	result := make(
		map[permission.Permission]struct{},
		len(permissions),
	)

	for _, currentPermission := range permissions {
		result[currentPermission] = struct{}{}
	}

	return result
}
