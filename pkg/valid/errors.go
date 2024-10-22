package valid

import "fmt"

type FailedValidationError struct {
	Problems []Problem
}

func (err FailedValidationError) Error() string {
	message := "validation failed with the following problems: "
	for i, problem := range err.Problems {
		if i == len(err.Problems)-1 {
			message += fmt.Sprintf("%s: %s", problem.Name, problem.Description)
		}
		message += fmt.Sprintf("%s: %s, ", problem.Name, problem.Description)
	}
	return message
}

func NewFailedValidationError(problems []Problem) FailedValidationError {
	return FailedValidationError{
		Problems: problems,
	}
}
