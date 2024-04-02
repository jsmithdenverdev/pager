package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// customClaims contains custom data we want from the token.
type customClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c customClaims) Validate(ctx context.Context) error {
	return nil
}

// ensureValidToken is a middleware that will check the validity of our JWT.
func ensureValidToken(config config, logger *slog.Logger) func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + config.Auth0Domain + "/")
	if err != nil {
		logger.Error(fmt.Sprintf("failed to parse issuer url: %s", err.Error()))
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{config.Auth0Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &customClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to setup jwt validator: %s", err.Error()))
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error(fmt.Sprintf("failed to validate jwt: %s", err.Error()))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

// HasScope checks whether our claims have a specific scope.
func (c customClaims) HasScope(expectedScope string) bool {
	result := strings.Split(c.Scope, " ")
	for i := range result {
		if result[i] == expectedScope {
			return true
		}
	}

	return false
}
