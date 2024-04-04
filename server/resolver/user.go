package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

type User struct {
	*models.User
}

func (u *User) ID() graphql.ID {
	return graphql.ID(u.User.ID)
}

func (u *User) IdpID() string {
	return u.User.IdpID
}

func (u *User) Email() string {
	return u.User.Email
}

func (u *User) Status() string {
	return string(u.User.Status)
}

func (u *User) Created() graphql.Time {
	return graphql.Time{
		Time: u.User.Created,
	}
}

func (u *User) CreatedBy() string {
	return u.User.CreatedBy
}

func (u *User) Modified() graphql.Time {
	return graphql.Time{
		Time: u.User.Created,
	}
}

func (u *User) ModifiedBy() string {
	return u.User.ModifiedBy
}

func (u *User) Agencies(ctx context.Context) (*[]*Agency, error) {
	svc := ctx.Value(service.ContextKeyAgencyService).(*service.AgencyService)
	agencies, err := svc.List(service.AgenciesPagination{})
	if err != nil {
		return nil, err
	}
	var results []*Agency
	for _, agency := range agencies {
		results = append(results, &Agency{agency})
	}
	return &results, nil
}
