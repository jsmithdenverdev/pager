package problemdetails

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// EncodeResponse encodes data as a JSON response.
func EncodeResponse(
	ctx context.Context,
	response *events.APIGatewayProxyResponse,
	status int,
	data ProblemDetails) {
	response.Headers = make(map[string]string)
	response.Headers["Content-Type"] = "application/json"
	response.StatusCode = status

	b, err := json.Marshal(data)

	if err != nil {
		response.Body = `{"message": "internal server error"}`
		return
	}

	response.Body = string(b)
}
