package main

import (
	"log/slog"
	"net/http"

	"github.com/authzed/authzed-go/v1"
	"github.com/jmoiron/sqlx"
)

// handleReadUserInfo returns a handler for reading the user info of the
// currently authenticated user. The handler returns a `user` model.
func handleReadUserInfo(logger *slog.Logger, authz *authzed.Client, db *sqlx.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userId := ctx.Value(contextKeyUser).(string)

		logger.InfoContext(ctx, "readUserInfo", "userId", userId)

		var user user
		if err := db.QueryRowxContext(
			ctx,
			`SELECT id, email, idp_id, status, created, created_by, modified, modified_by
			 FROM users 
			 WHERE idp_id = $1`,
			userId,
		).StructScan(&user); err != nil {
			encode(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := encode(w, r, http.StatusOK, user); err != nil {
			encode(w, r, http.StatusInternalServerError, err)
		}
	})
}
