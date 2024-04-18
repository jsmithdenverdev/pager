package models

type Role string

const (
	RoleReader        Role = "READER"
	RoleWriter        Role = "WRITER"
	RolePlatformAdmin Role = "PLATFORM_ADMIN"
)

// UserRole represents the role a user has in a particular agency. A user may
// have the special role of PLATFORM_ADMIN with an empty agency association.
type UserRole struct {
	UserID   string `json:"userId" db:"user_id"`
	AgencyID string `json:"agencyId" db:"agency_id"`
	Role     Role   `json:"role" db:"role"`
}
