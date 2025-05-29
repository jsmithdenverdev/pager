package models

import (
	"fmt"
	"strings"
	"time"
)

type AgencyStatus = string

const (
	AgencyStatusActive   AgencyStatus = "ACTIVE"
	AgencyStatusInactive AgencyStatus = "INACTIVE"
)

// agency represents an agency in the database.
type Agency struct {
	ID         string       `dynamodbav:"id"`
	Name       string       `dynamodbav:"name"`
	Status     AgencyStatus `dynamodbav:"status"`
	Created    time.Time    `dynamodbav:"created"`
	Modified   time.Time    `dynamodbav:"modified"`
	CreatedBy  string       `dynamodbav:"createdBy"`
	ModifiedBy string       `dynamodbav:"modifiedBy"`
}

func (a Agency) Type() string {
	return EntityTypeAgency
}

func (a Agency) EncodeKey() attributevalue.Key {
	return attributevalue.Key{
		PK: fmt.Sprintf("agency#%s", a.ID),
		SK: "meta",
	}
}

func (a Agency) DecodeKey(key attributevalue.Key) error {
	pkSplit := strings.Split(key.PK, "#")
	if len(pkSplit) != 2 {
		return fmt.Errorf("invalid PK: %s", key.PK)
	}
	a.ID = pkSplit[1]
	return nil
}
