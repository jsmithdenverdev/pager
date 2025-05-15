package models

import "time"

type RegistrationStatus = string

const (
	RegistrationStatusPending  RegistrationStatus = "PENDING"
	RegistrationStatusComplete RegistrationStatus = "COMPLETE"
	RegistrationStatusDeclined RegistrationStatus = "DECLINED"
	RegistrationStatusExpired  RegistrationStatus = "EXPIRED"
)

type EndpointRegistration struct {
	PK         string             `dynamodbav:"pk"`
	SK         string             `dynamodbav:"sk"`
	Type       EntityType         `dynamodbav:"type"`
	Status     RegistrationStatus `dynamodbav:"status"`
	EndpointID string             `dynamodbav:"endpointId"`
	Created    time.Time          `dynamodbav:"created"`
	Modified   time.Time          `dynamodbav:"modified"`
	CreatedBy  string             `dynamodbav:"createdBy"`
	ModifiedBy string             `dynamodbav:"modifiedBy"`
}
