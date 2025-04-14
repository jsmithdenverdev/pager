package authz

// Entitlement represents a type of access or privilege within the system.
type Entitlement string

const (
	// EntPlatformAdmin represents the platform administrator entitlement.
	EntPlatformAdmin Entitlement = "PLATFORM_ADMIN"
)
