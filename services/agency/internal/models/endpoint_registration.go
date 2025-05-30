package models

import (
	"fmt"
	"github.com/jsmithdenverdev/pager/pkg/dynarow"
	"strings"
	"time"
)

type RegistrationStatus = string

const (
	RegistrationStatusPending  RegistrationStatus = "PENDING"
	RegistrationStatusComplete RegistrationStatus = "COMPLETE"
	RegistrationStatusFailed   RegistrationStatus = "FAILED"
)

type EndpointRegistration struct {
	AgencyID   string             `dynamodbav:"agency_id"`
	EndpointID string             `dynamodbav:"endpoint_id"`
	Status     RegistrationStatus `dynamodbav:"status"`
	Created    time.Time          `dynamodbav:"created"`
	Modified   time.Time          `dynamodbav:"modified"`
	CreatedBy  string             `dynamodbav:"created_by"`
	ModifiedBy string             `dynamodbav:"modified_by"`
}

func (er *EndpointRegistration) Type() string {
	return EntityTypeRegistration
}

func (er *EndpointRegistration) EncodeKey() dynarow.Key {
	return dynarow.Key{
		PK: fmt.Sprintf("agency#%s", er.AgencyID),
		SK: fmt.Sprintf("registration#%s", er.EndpointID),
	}
}

func (er *EndpointRegistration) DecodeKey(key dynarow.Key) error {
	pkParts := strings.Split(key.PK, "#")
	if len(pkParts) != 2 {
		return fmt.Errorf("invalid pk: %s", key.PK)
	}
	skParts := strings.Split(key.SK, "#")
	if len(skParts) != 2 {
		return fmt.Errorf("invalid sk: %s", key.PK)
	}
	er.AgencyID = pkParts[1]
	er.EndpointID = skParts[1]
	return nil
}
