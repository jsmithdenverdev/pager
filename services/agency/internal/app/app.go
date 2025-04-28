package app

import (
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewServer(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, config, logger, client)

	return mux
}
