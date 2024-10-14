package valid

import (
	"net/http"

	"github.com/jsmithdenverdev/pager/pkg/problemdetails"
)

func ProblemDetails(problems []Problem) problemdetails.ProblemDetails {
	return problemdetails.ProblemDetails{
		Type:     "auth/authorization",
		Status:   http.StatusInternalServerError,
		Title:    "Unauthorized",
		Detail:   "",
		Instance: "",
	}
}
