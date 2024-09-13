package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// createAgencyRequest represents the data required to create a new Agency.
type inviteUserRequest struct {
	Email    string `json:"email"`
	AgencyID string `json:"agencyId"`
	Role     string `json:"role"`
}

// Valid performs validations on an inviteUserRequest and returns a slice of
// problem structs if issues are encountered.
func (r inviteUserRequest) Valid(ctx context.Context) []problem {
	var problems []problem
	if r.Email == "" {
		problems = append(problems, problem{
			Name:        "name",
			Description: "Name must be at least 1 character",
		})
	}
	if r.AgencyID == "" {
		problems = append(problems, problem{
			Name:        "agencyId",
			Description: "Agency ID must be provided",
		})
	}
	if r.Role == "" {
		problems = append(problems, problem{
			Name:        "role",
			Description: "Role must be provided",
		})
	} else {
		// Check if role is within allowed roles
		if _, ok := map[string]any{
			string(models.RoleReader): struct{}{},
			string(models.RoleWriter): struct{}{},
		}[r.Role]; !ok {
			problems = append(problems, problem{
				Name: "role",
				Description: fmt.Sprintf("Role must be one of %s", strings.Join([]string{
					string(models.RoleReader),
					string(models.RoleWriter),
				}, ",")),
			})
		}
	}
	return problems
}

// InviteUser is a handler to invite a new user. The user must have the
// invite_user permission on the platform to call this endpoint.
func InviteUser(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		svc := ctx.Value(service.ContextKeyAgencyService).(*service.AgencyService)
		req, problems, err := decodeValid[inviteUserRequest](ctx, r)

		if err != nil {
			if len(problems) > 0 {
				encodeValidationError(ctx, w, logger, problems)
				return
			} else {
				logger.ErrorContext(
					r.Context(),
					"[in handlers.InviteUser] failed to decode request",
					slog.String("error", err.Error()))

				encodeInternalServerError(ctx, w, logger)
				return
			}
		}

		user, err := svc.InviteUser(req.Email, req.AgencyID, models.Role(req.Role))

		if err != nil {
			logger.ErrorContext(
				ctx,
				"[in handlers.InviteUser] failed to invite user",
				slog.String("error", err.Error()))

			errUnauthorized := new(authz.AuthzError)
			if errors.As(err, errUnauthorized) {
				encodeUnauthorizedError(ctx, w, logger, errUnauthorized)
				return
			}

			encodeInternalServerError(ctx, w, logger)
			return
		}

		encodeResponse(ctx, w, logger, http.StatusCreated, toUserResponse(user))
	})
}
