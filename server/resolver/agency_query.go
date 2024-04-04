package resolver

import (
	"context"

	"github.com/jsmithdenverdev/pager/service"
)

func (query *Query) Agency(ctx context.Context, args struct {
	ID string
}) (*Agency, error) {
	svc := ctx.Value(service.ContextKeyAgencyService).(*service.AgencyService)
	agency, err := svc.Read(args.ID)
	return &Agency{agency}, err
}
