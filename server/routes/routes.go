package routes

import (
	"context"
	"log/slog"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	jwtvalidator "github.com/auth0/go-jwt-middleware/v2/validator"

	"github.com/auth0/go-auth0/management"

	"github.com/authzed/authzed-go/v1"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/graphql-go/handler"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/config"
	"github.com/jsmithdenverdev/pager/handlers"
	"github.com/jsmithdenverdev/pager/middleware"
	"github.com/jsmithdenverdev/pager/pubsub"
	"github.com/jsmithdenverdev/pager/schema"
	"github.com/jsmithdenverdev/pager/service"
)

// Register registers application routes on an instance of a chi.Mux.
func Register(
	router *chi.Mux,
	logger *slog.Logger,
	cfg config.Config,
	authzedClient *authzed.Client,
	db *sqlx.DB,
	auth0 *management.Management,
	pubsubClient *pubsub.Client) error {
	var (
		apiRouter     = chi.NewRouter()
		graphQLRouter = chi.NewRouter()
	)

	// Add recovery midleware globally
	router.Use(chimiddleware.Recoverer)
	router.Mount("/api", apiRouter)
	router.Mount("/graphql", graphQLRouter)

	// Health check
	router.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	// Add token validation middleware to api and graphql routes
	apiRouter.Use(middleware.EnsureValidToken(cfg, logger))
	graphQLRouter.Use(middleware.EnsureValidToken(cfg, logger))

	// Attach services to routers. This must be done after token middleware which
	// must run first to parse an incoming JWT and fetch the sub claim.
	apiRouter.Use(attachServicesToRequestContext(
		logger,
		cfg,
		authzedClient,
		db,
		auth0,
		pubsubClient))

	graphQLRouter.Use(attachServicesToRequestContext(
		logger,
		cfg,
		authzedClient,
		db,
		auth0,
		pubsubClient))

	// Register individual routes
	registerAgencyRoutes(apiRouter, logger)
	registerDeviceRoutes(apiRouter, logger)
	registerUserRoutes(apiRouter, logger)
	registerPageRoutes(apiRouter, logger)

	// registerGraphQLRoutes may return an error if the graphql schema is invalid
	if err := registerGraphQLRotues(graphQLRouter, logger); err != nil {
		return err
	}

	return nil
}

// registerAgencyRoutes registers agency routes on an instance of a chi.Mux.
func registerAgencyRoutes(router *chi.Mux, logger *slog.Logger) {
	agencyRouter := chi.NewRouter()
	router.Mount("/agencies", agencyRouter)

	agencyRouter.Post("/", handlers.CreateAgency(logger).(http.HandlerFunc))
}

// registerDeviceRoutes registers device routes on an instance of a chi.Mux.
func registerDeviceRoutes(router *chi.Mux, logger *slog.Logger) {
	deviceRouter := chi.NewRouter()
	router.Mount("/devices", deviceRouter)

	deviceRouter.Post("/", handlers.ProvisionDevice(logger).(http.HandlerFunc))
}

// registerUserRoutes registers user routes on an instance of a chi.Mux.
func registerUserRoutes(router *chi.Mux, logger *slog.Logger) {
	userRouter := chi.NewRouter()
	router.Mount("/users", userRouter)

	userRouter.Post("/invites", handlers.InviteUser(logger).(http.HandlerFunc))
}

// registerPageRoutes registers page routes on an instance of a chi.Mux.
func registerPageRoutes(router *chi.Mux, logger *slog.Logger) {
	pagesRouter := chi.NewRouter()
	router.Mount("/pages", pagesRouter)
}

// registerPageRoutes registers graphql route on an instance of a chi.Mux.
func registerGraphQLRotues(router *chi.Mux, logger *slog.Logger) error {
	schema, err := schema.New()
	if err != nil {
		return err
	}

	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	router.Handle("/", h)

	return nil
}

// attachServicesToRequestContext attaches request scoped services to a request
// context.
// It should probably be defined as middleware, but shrug...
func attachServicesToRequestContext(
	logger *slog.Logger,
	cfg config.Config,
	authzedClient *authzed.Client,
	db *sqlx.DB,
	auth0 *management.Management,
	pubsubClient *pubsub.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Fetch the request context
			ctx := r.Context()
			// Fetch the user id from the request context, this is in the form of a
			// subject claim that is parsed from a JWT validated in middleware.
			user := ctx.
				Value(jwtmiddleware.ContextKey{}).(*jwtvalidator.ValidatedClaims).
				RegisteredClaims.
				Subject

			// Create a new request scoped authz client using the userID. This allows
			// us to send authorization requests on behalf of this user without
			// passing an id around.
			authz := authz.NewClient(ctx, authzedClient, logger, user)

			// Add UserService to context
			ctx = context.WithValue(ctx,
				service.ContextKeyUserService,
				service.NewUserService(ctx, user, authz, db, logger))

			// Add AgencyService to context
			ctx = context.WithValue(ctx,
				service.ContextKeyAgencyService,
				service.NewAgencyService(ctx, user, authz, db, auth0, pubsubClient, logger))

			// Add DeviceService to context
			ctx = context.WithValue(ctx,
				service.ContextKeyDeviceService,
				service.NewDeviceService(ctx, user, authz, db, logger))

			// Add PageService to context
			ctx = context.WithValue(ctx,
				service.ContextKeyPageService,
				service.NewPageService(ctx, user, authz, db, pubsubClient, logger))

			// Re assign the new context to the request
			r = r.WithContext(ctx)

			// Move onto the next piece of middleware
			next.ServeHTTP(w, r)
		})
	}

}
