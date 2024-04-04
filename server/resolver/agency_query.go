package resolver

import (
	"context"

	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

func (query *Query) Agency(ctx context.Context, args struct {
	ID string
}) (*Agency, error) {
	svc := ctx.Value(service.ContextKeyAgencyService).(*service.AgencyService)
	agency, err := svc.Read(args.ID)
	if (models.Agency{}) == agency {
		return nil, err
	}
	return &Agency{agency}, err
}
