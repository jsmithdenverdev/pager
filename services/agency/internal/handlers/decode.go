package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// decodeValid decodes a JSON string into a validatorMapper, validates
// it, and maps it to the output type. If decoding, validation, or mapping
// fails, it returns the appropriate errors and problems.
func decodeValid[I validator](ctx context.Context, event events.APIGatewayProxyRequest) (I, []problem, error) {
	var inputModel I

	// decode to JSON
	if err := json.NewDecoder(bytes.NewReader([]byte(event.Body))).Decode(&inputModel); err != nil {
		return *new(I), nil, fmt.Errorf("[in decodeValidateBody] decode json: %w", err)
	}

	// validate
	if problems := inputModel.Valid(ctx); len(problems) > 0 {
		return *new(I), problems, fmt.Errorf(
			"[in decodeValidateBody] invalid %T: %d problems", inputModel, len(problems),
		)
	}

	return inputModel, nil, nil
}
