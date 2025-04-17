package app

type UserStatus = string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusPending  UserStatus = "PENDING"
)

type user struct {
	model
	Email  string     `dynamodbav:"email"`
	IDPID  string     `dynamodbav:"idpId"`
	Status UserStatus `dynamodbav:"status"`
}
