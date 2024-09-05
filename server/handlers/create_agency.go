package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/service"
)

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
func CreateAgency(logger *slog.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svc := r.Context().Value(service.ContextKeyAgencyService).(*service.AgencyService)
		agency, problems, err := decodeRequest[createAgencyRequest](r.Context(), r)

		if err != nil {
			logger.ErrorContext(r.Context(), "[in handlers.CreateAgency] failed to decode request", slog.String("error", err.Error()))
			if len(problems) > 0 {
				encodeValidationError(w, logger, problems)
				return
			}
			http.Error(w, `{"Error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		agency, err = svc.CreateAgency(agency.Name)

		if err != nil {
			logger.ErrorContext(r.Context(), "[in handlers.CreateAgency] failed to create agency", slog.String("error", err.Error()))
			errUnauthorized := new(authz.AuthzError)
			if errors.As(err, errUnauthorized) {
				encodeUnauthorizedError(w, logger, errUnauthorized)
				return
			}
			http.Error(w, `{"Error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		encodeResponse(w, logger, http.StatusCreated, toAgencyResponse(agency))
	})
}
