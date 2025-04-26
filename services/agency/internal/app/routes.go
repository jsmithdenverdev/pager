package app

import (
	"fmt"
	"log/slog"
	"net/http"
)

func addRoutes(mux *http.ServeMux, config Config, logger *slog.Logger) {
	mux.Handle(fmt.Sprintf("GET /%s", config.Environment), listAgencies(logger))
	mux.Handle(fmt.Sprintf("GET /%s/{id}", config.Environment), readAgencyById(logger))
}
