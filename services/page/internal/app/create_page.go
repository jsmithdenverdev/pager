package app

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log/slog"
	"net/http"
)

func createPage(conf Config, logger *slog.Logger, dynamoClient *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
