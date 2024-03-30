package main

import (
	"log/slog"
	"net/http"

	"github.com/authzed/authzed-go/v1"
	"github.com/jmoiron/sqlx"
)

func handleListAgencies(logger *slog.Logger, authz *authzed.Client, db *sqlx.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
