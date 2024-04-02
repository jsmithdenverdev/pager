package resolver

import (
	"context"

	"github.com/jsmithdenverdev/pager/service"
)

func (query *Query) UserInfo(ctx context.Context) (*User, error) {
	svc := ctx.Value(service.ContextKeyUserService).(*service.UserService)
	user, err := svc.Info()
	if err != nil {
		return nil, err
	}
	return &User{user}, nil
}
