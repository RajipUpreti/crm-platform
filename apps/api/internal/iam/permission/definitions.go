package permission

const (
	TenantRead   Permission = "tenant.read"
	TenantUpdate Permission = "tenant.update"
	TenantDelete Permission = "tenant.delete"

	MemberRead       Permission = "member.read"
	MemberInvite     Permission = "member.invite"
	MemberUpdateRole Permission = "member.update_role"
	MemberSuspend    Permission = "member.suspend"
	MemberRemove     Permission = "member.remove"

	InvitationRead   Permission = "invitation.read"
	InvitationCreate Permission = "invitation.create"
	InvitationRevoke Permission = "invitation.revoke"

	ContactRead   Permission = "contact.read"
	ContactCreate Permission = "contact.create"
	ContactUpdate Permission = "contact.update"
	ContactDelete Permission = "contact.delete"

	CompanyRead   Permission = "company.read"
	CompanyCreate Permission = "company.create"
	CompanyUpdate Permission = "company.update"
	CompanyDelete Permission = "company.delete"

	DealRead   Permission = "deal.read"
	DealCreate Permission = "deal.create"
	DealUpdate Permission = "deal.update"
	DealDelete Permission = "deal.delete"
)

func All() []Permission {
	return []Permission{
		TenantRead,
		TenantUpdate,
		TenantDelete,

		MemberRead,
		MemberInvite,
		MemberUpdateRole,
		MemberSuspend,
		MemberRemove,

		InvitationRead,
		InvitationCreate,
		InvitationRevoke,

		ContactRead,
		ContactCreate,
		ContactUpdate,
		ContactDelete,

		CompanyRead,
		CompanyCreate,
		CompanyUpdate,
		CompanyDelete,

		DealRead,
		DealCreate,
		DealUpdate,
		DealDelete,
	}
}
