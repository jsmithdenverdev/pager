package app

import (
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func readAgencyById(logger *slog.Logger, client *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "request received", slog.Any("request.headers", r.Header))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("read agency by id"))
	})
}
