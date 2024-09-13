package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// createAgencyRequest represents the data required to create a new Agency.
type createAgencyRequest struct {
	Name string `json:"name"`
}

// Valid performs validations on a createAgencyRequest and returns a slice of
// problem structs if issues are encountered.
func (r createAgencyRequest) Valid(ctx context.Context) []problem {
	var problems []problem
	if r.Name == "" {
		problems = append(problems, problem{
			Name:        "name",
			Description: "Name must be at least 1 character",
		})
	}
	return problems
}

// MapTo maps a createAgencyRequest to a models.Agency.
func (r createAgencyRequest) MapTo() models.Agency {
	var m models.Agency
	m.Name = r.Name
	return m
}

// CreateAgency is a handler to create a new agency. The user must have the
// create_agency permission on the platform to call this endpoint.
//
// @Summary Create a new Pager Agency
// @Description Creates a new Pager Agency. The calling user must have the create_agency permission on the platform to call this endpoint.
// @Tags create,agency
// @Accept json
// @Produce json
// @Success 200 {object} handlers.agencyResponse
// @Router /health-check [GET]
func CreateAgency(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		svc := ctx.Value(service.ContextKeyAgencyService).(*service.AgencyService)
		req, problems, err := decodeValid[createAgencyRequest](ctx, r)

		if err != nil {
			if len(problems) > 0 {
				encodeValidationError(ctx, w, logger, problems)
				return
			} else {
				logger.ErrorContext(
					r.Context(),
					"[in handlers.CreateAgency] failed to decode request",
					slog.String("error", err.Error()))

				encodeInternalServerError(ctx, w, logger)
				return
			}
		}

		agency, err := svc.CreateAgency(req.Name)

		if err != nil {
			logger.ErrorContext(
				r.Context(),
				"[in handlers.CreateAgency] failed to create agency",
				slog.String("error", err.Error()))

			errUnauthorized := new(authz.AuthzError)
			if errors.As(err, errUnauthorized) {
				encodeUnauthorizedError(ctx, w, logger, errUnauthorized)
				return
			}

			encodeInternalServerError(ctx, w, logger)
			return
		}

		encodeResponse(ctx, w, logger, http.StatusCreated, toAgencyResponse(agency))
	})
}
