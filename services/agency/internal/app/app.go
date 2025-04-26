package app

import (
	"log/slog"
	"net/http"
)

func NewServer(config Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, config, logger)

	return mux
}
