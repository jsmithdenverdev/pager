package main

import (
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	jwtvalidator "github.com/auth0/go-jwt-middleware/v2/validator"
)

// pagerContextKey is a unique identifier for a pager request context
type contextKey string

var (
	// contextKeyUser is the key used to retrieve the current user from the
	// request context.
	contextKeyUser = contextKey("user")
)

// withContext is middleware that attaches pager context values to the request
// context.
func withContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := ctx.Value(jwtmiddleware.ContextKey{}).(*jwtvalidator.ValidatedClaims)
		ctx = context.WithValue(ctx, contextKeyUser, claims.RegisteredClaims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
