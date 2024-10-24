package authz

import (
	"fmt"
	"net/http"

	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
)

func NewProblemDetail(err UnauthorizedError) problemdetail.ProblemDetailer {
	pd := problemdetail.New(
		"authorization",
		problemdetail.WithTitle("Unauthorized"),
		problemdetail.WithDetail(err.Error()),
		problemdetail.WithInstance(fmt.Sprintf("%s::%s", *err.Entity.EntityType, *err.Entity.EntityId)))

	pd.WriteStatus(http.StatusUnauthorized)

	return pd
}
