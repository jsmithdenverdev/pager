package app

import (
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func NewServer(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, config, logger, dynamoClient, snsClient)

	return mux
}
