package authz

import (
	"fmt"
	"net/http"

	"github.com/jsmithdenverdev/pager/pkg/problemdetails"
)

func ProblemDetails(err UnauthorizedError) problemdetails.ProblemDetails {
	return problemdetails.ProblemDetails{
		Type:     "auth/authorization",
		Status:   http.StatusInternalServerError,
		Title:    "Unauthorized",
		Detail:   err.Error(),
		Instance: fmt.Sprintf("%s::%s", err.Resource.Type, err.Resource.ID),
	}
}
