package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/jsmithdenverdev/pager/models"
)

type Agency struct {
	models.Agency
}

func (a *Agency) ID() graphql.ID {
	return graphql.ID(a.Agency.ID)
}

func (a *Agency) Name() string {
	return a.Agency.Name
}

func (a *Agency) Status() string {
	var status = string(a.Agency.Status)
	return status
}

func (a *Agency) Created() graphql.Time {
	return graphql.Time{
		Time: a.Agency.Created,
	}
}

func (a *Agency) CreatedBy() string {
	return a.Agency.CreatedBy
}

func (a *Agency) Modified() graphql.Time {
	return graphql.Time{
		Time: a.Agency.Modified,
	}
}

func (a *Agency) ModifiedBy() string {
	return a.Agency.ModifiedBy
}
