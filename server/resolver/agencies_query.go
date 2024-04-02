package resolver

import (
	"context"

	"github.com/jsmithdenverdev/pager/service"
)

func (query *Query) Agencies(ctx context.Context) (*[]*Agency, error) {
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
