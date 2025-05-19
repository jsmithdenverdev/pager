package models

// OwnershipLink represents the OwnershipLink of an endpoint by a user.
// The model is a simple relationship binding that doesn't include other
// metadata. The relationship is encoded within the pk and sk.
type OwnershipLink struct {
	KeyFields
	AuditableFields
}
