package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/service"
)

// provisionDeviceRequest represents the data required to provision a new
// Device.
type provisionDeviceRequest struct {
	AgencyID string `json:"agencyId"`
	OwnerID  string `json:"ownerId"`
	Name     string `json:"name"`
}

// Valid performs validations on an provisionDeviceRequest and returns a slice
// of problem structs if issues are encountered.
func (r provisionDeviceRequest) Valid(ctx context.Context) []problem {
	var problems []problem
	return problems
}

// ProvisionDevice is a handler to provision a new device. The caller must have
// the provision_device permission on the requested agency to call this
// endpoint.
func ProvisionDevice(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		svc := ctx.Value(service.ContextKeyDeviceService).(*service.DeviceService)
		req, problems, err := decodeValid[provisionDeviceRequest](ctx, r)

		if err != nil {
			if len(problems) > 0 {
				encodeValidationError(ctx, w, logger, problems)
				return
			} else {
				logger.ErrorContext(
					r.Context(),
					"[in handlers.ProvisionDevice] failed to decode request",
					slog.String("error", err.Error()))

				encodeInternalServerError(ctx, w, logger)
				return
			}
		}

		device, err := svc.ProvisionDevice(req.AgencyID, req.OwnerID, req.Name)

		if err != nil {
			logger.ErrorContext(
				ctx,
				"[in handlers.ProvisionDevice] failed to provision device",
				slog.String("error", err.Error()))

			errUnauthorized := new(authz.AuthzError)
			if errors.As(err, errUnauthorized) {
				encodeUnauthorizedError(ctx, w, logger, errUnauthorized)
				return
			}

			encodeInternalServerError(ctx, w, logger)
			return
		}

		encodeResponse(ctx, w, logger, http.StatusCreated, toDeviceResponse(device))
	})
}
