package app

import (
	"log/slog"
	"net/http"
)

func listAgencies(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "request received", slog.Any("request.headers", r.Header))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("list agencies"))
	})
}
