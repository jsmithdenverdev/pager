package authz

type permission string

const (
	PermissionCreateAgency     permission = "create_agency"
	PermissionCreatePage       permission = "create_page"
	PermissionProvisionDevice  permission = "provision_device"
	PermissionActivateDevice   permission = "activate"
	PermissionDeactivateDevice permission = "deactivate"
)
