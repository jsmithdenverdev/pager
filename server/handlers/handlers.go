package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// validator defines a method for validating an object. It returns a slice of
// problems found during validation.
type validator interface {
	Valid(ctx context.Context) (problems []problem)
}

// fromMapper is a generic interface that defines a method for mapping an object to
// another type. The MapTo method returns the mapped object and an error if the
// mapping fails.
type mapper[T any] interface {
	MapTo() T
}

// validatorMapper is a generic interface that combines Validator and
// Mapper interfaces. It requires implementing both validation and mapping
// methods.
type validatorMapper[T any] interface {
	validator
	mapper[T]
}

// problem represents an issue found during validation.
type problem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// decodeRequest decodes a JSON string into a validatorMapper, validates
// it, and maps it to the output type. If decoding, validation, or mapping
// fails, it returns the appropriate errors and problems.
func decodeRequest[I validatorMapper[O], O any](ctx context.Context, r *http.Request) (O, []problem, error) {
	var inputModel I

	// decode to JSON
	if err := json.NewDecoder(r.Body).Decode(&inputModel); err != nil {
		return *new(O), nil, fmt.Errorf("[in decodeValidateBody] decode json: %w", err)
	}

	// validate
	if problems := inputModel.Valid(ctx); len(problems) > 0 {
		return *new(O), problems, fmt.Errorf(
			"[in decodeValidateBody] invalid %T: %d problems", inputModel, len(problems),
		)
	}

	// map to return type and return
	return inputModel.MapTo(), nil, nil
}

// encodeResponse encodes data as a JSON response.
func encodeResponse(w http.ResponseWriter, logger *slog.Logger, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Error while marshaling data", "err", err, "data", data)
		http.Error(w, `{"Error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

// encodeValidationError encodes data as an HTTP 400 bad request response.
func encodeValidationError(w http.ResponseWriter, logger *slog.Logger, problems []problem) {
	encodeResponse(w, logger, http.StatusBadRequest, struct {
		Problems []problem `json:"problems"`
	}{
		Problems: problems,
	})
}

// encodeUnauthorizedError encodes data as an HTTP 401 unauthorized response.
func encodeUnauthorizedError(w http.ResponseWriter, logger *slog.Logger, err error) {
	encodeResponse(w, logger, http.StatusUnauthorized, err.Error())
}
