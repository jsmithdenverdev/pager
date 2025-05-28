package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func addRoutes(mux *http.ServeMux, config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) {
	mux.Handle(fmt.Sprintf("POST /%s", config.Environment), createPage(config, logger, dynamoClient, snsClient))
}
