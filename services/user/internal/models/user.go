package models

type UserStatus = string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusPending  UserStatus = "PENDING"
)

type User struct {
	ID     string     `dynamodbav:"id"`
	Email  string     `dynamodbav:"email"`
	IDPID  string     `dynamodbav:"idpId"`
	Status UserStatus `dynamodbav:"status"`
}
