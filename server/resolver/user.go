package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/jsmithdenverdev/pager/models"
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
