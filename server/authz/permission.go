package authz

type Permission struct {
	Relationship string
	Resource     Resource
	Subject      Resource
}
