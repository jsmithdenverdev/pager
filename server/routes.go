package main

import (
	"log/slog"
	"net/http"

	"github.com/authzed/authzed-go/v1"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

// addRoutes adds application routes to an http.ServeMux instances, these routes
// are attached to handlers that handle the request.
// All handler dependencies are passed into addRoutes as a flat set of arguments
// allowing us to see the full set of dependencies for all request handlers.
func addRoutes(mux *http.ServeMux,
	config config,
	logger *slog.Logger,
	validate *validator.Validate,
	authz *authzed.Client,
	db *sqlx.DB) {
	// authRouter is an instance of serve mux that will be wrapped with auth
	// middleware. Any routes that require authentication need to be registered
	// on it.
	authRouter := http.NewServeMux()

	// GET /health returns the application health check.
	// Authentication is not required.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authRouter.Handle("GET /userinfo", handleReadUserInfo(logger, authz, db))
	authRouter.Handle("GET /agencies", handleListAgencies(logger, authz, db))

	mux.Handle("/", AuthMiddleware(config, logger)(withContext(authRouter)))
}
