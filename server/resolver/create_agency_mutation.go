package resolver

import (
	"context"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	jwtvalidator "github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/jsmithdenverdev/pager/service"
)

func (muation *Mutation) CreateAgency(ctx context.Context, args struct {
	Input struct {
		Name string
	}
}) (*Agency, error) {
	claims := ctx.Value(jwtmiddleware.ContextKey{}).(*jwtvalidator.ValidatedClaims)
	svc := ctx.Value(service.ContextKeyAgencyService).(*service.AgencyService)
	agency, err := svc.Create(args.Input.Name, claims.RegisteredClaims.Subject)
	if err != nil {
		return &Agency{}, err
	}
	return &Agency{agency}, nil
}
