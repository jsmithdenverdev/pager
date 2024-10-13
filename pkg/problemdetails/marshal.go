package problemdetails

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func MarshalToAPIGatewayProxyResponse(details ProblemDetails) (events.APIGatewayProxyResponse, error) {
	// Marshal the ProblemDetails object into JSON
	body, err := json.Marshal(details)
	if err != nil {
		// Return a 500 Internal Server Error if marshalling fails
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"title":"Internal Server Error"}`,
			Headers: map[string]string{
				"Content-Type": "application/problem+json",
			},
		}, err
	}

	// Return the marshaled ProblemDetails in the APIGatewayProxyResponse
	return events.APIGatewayProxyResponse{
		StatusCode: details.Status,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/problem+json",
		},
	}, nil
}
