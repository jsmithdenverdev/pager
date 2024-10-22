package valid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func DecodeAPIGatewayRequest[V Validator](ctx context.Context, event events.APIGatewayProxyRequest) (V, []Problem, error) {
	var model V

	// decode to JSON
	if err := json.NewDecoder(bytes.NewReader([]byte(event.Body))).Decode(&model); err != nil {
		return *new(V), nil, fmt.Errorf("[in decodeValidateBody] decode json: %w", err)
	}

	// validate
	if problems := model.Valid(ctx); len(problems) > 0 {
		return *new(V), problems, fmt.Errorf(
			"[in decodeValidateBody] invalid %T: %d problems", model, len(problems),
		)
	}

	return model, nil, nil
}
