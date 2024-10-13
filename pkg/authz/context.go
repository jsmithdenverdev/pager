package authz

import "context"

type contextKey string

const (
	contextKeyClient contextKey = "client"
)

func ClientFromContext(ctx context.Context) (Authorizer, bool) {
	client, ok := ctx.Value(contextKeyClient).(*client)
	return client, ok
}
