package authz

type permission string

const (
	PermissionCreateAgency    permission = "create_agency"
	PermissionProvisionDevice permission = "provision_device"
	PermissionActivateDevice  permission = "activate"
)
