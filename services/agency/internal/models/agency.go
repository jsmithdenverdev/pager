package models

import (
	"fmt"
	"github.com/jsmithdenverdev/pager/pkg/dynarow"
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
	CreatedBy  string       `dynamodbav:"created_by"`
	ModifiedBy string       `dynamodbav:"modified_by"`
}

func (a *Agency) Type() string {
	return EntityTypeAgency
}

func (a *Agency) EncodeKey() dynarow.Key {
	return dynarow.Key{
		PK: fmt.Sprintf("agency#%s", a.ID),
		SK: "meta",
	}
}

func (a *Agency) DecodeKey(key dynarow.Key) error {
	idParts := strings.Split(key.PK, "#")
	if len(idParts) != 2 {
		return fmt.Errorf("invalid pk: %s", key.PK)
	}
	a.ID = idParts[1]
	return nil
}
