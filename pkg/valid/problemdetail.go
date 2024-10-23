package valid

import (
	"net/http"

	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
)

type ProblemDetail struct {
	*problemdetail.ProblemDetail
	Problems []Problem `json:"problems"`
}

func NewProblemDetail(problems []Problem) problemdetail.ProblemDetailer {
	pd := problemdetail.New(
		"validation",
		problemdetail.WithTitle("Invalid request"),
		problemdetail.WithDetail("The request you supplied didn't pass validation."))

	pd.WriteStatus(http.StatusBadRequest)

	return &ProblemDetail{
		ProblemDetail: pd,
		Problems:      problems,
	}
}
