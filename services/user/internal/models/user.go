package models

type UserStatus = string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusPending  UserStatus = "PENDING"
)

type User struct {
	Model
	Email  string     `dynamodbav:"email"`
	IDPID  string     `dynamodbav:"idpId"`
	Status UserStatus `dynamodbav:"status"`
}
