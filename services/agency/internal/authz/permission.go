package authz

type permission string

const (
	PermissionCreateAgency     permission = "create_agency"
	PermissionCreatePage       permission = "create_page"
	PermissionProvisionDevice  permission = "provision_device"
	PermissionActivateDevice   permission = "activate"
	PermissionDeactivateDevice permission = "deactivate"
	PermissionInviteUser       permission = "invite_user"
)

type Permission struct {
	Relationship string
	Resource     Resource
	Subject      Resource
}
