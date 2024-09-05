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

type createAgencyRequest struct {
	Name string `json:"name"`
}

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

func (r createAgencyRequest) MapTo() models.Agency {
	var m models.Agency
	m.Name = r.Name
	return m
}

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
