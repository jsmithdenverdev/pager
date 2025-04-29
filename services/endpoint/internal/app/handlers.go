package app

import (
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func listEndpoints(config Config, logger *slog.Logger, client *dynamodb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("list endpoints"))
	}
}

func readEndpoint(config Config, logger *slog.Logger, client *dynamodb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("read endpoint"))
	}
}
