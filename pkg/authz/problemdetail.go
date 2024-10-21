package authz

import (
	"fmt"
	"net/http"

	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
)

func NewProblemDetail(err UnauthorizedError) problemdetail.ProblemDetailer {
	pd := problemdetail.New(
		"auth/authorization",
		problemdetail.WithTitle("Unauthorized"),
		problemdetail.WithDetail(err.Error()),
		problemdetail.WithInstance(fmt.Sprintf("%s::%s", err.Resource.Type, err.Resource.ID)))

	pd.WriteStatus(http.StatusBadRequest)

	return pd
}
