package apigateway

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func writeInternalServerError(response *events.APIGatewayProxyResponse) {
	response.StatusCode = http.StatusInternalServerError
	response.Body = `{"type": "internal-server-error", "title": "Internal Server Error", "status": 500}`
}
