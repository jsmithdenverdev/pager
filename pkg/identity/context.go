package identity

import "context"

type contextKey string

const (
	// contextKeyClient is the key used to store the Client in a context.
	contextKeyClient contextKey = "client"
	// contextKeyUserInfo is the key used to store the User information in a context.
	contextKeyUserInfo contextKey = "user"
)

// WithUser returns a new context with the provided User information stored in it.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, contextKeyUserInfo, user)
}

// UserFrom extracts the User information from the context, if present.
func UserFrom(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(contextKeyUserInfo).(User)
	return user, ok
}
