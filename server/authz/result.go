package authz

type Result struct {
	Error      error
	Authorized bool
}
