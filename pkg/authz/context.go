package authz

import "context"

type contextKey string

const (
	contextKeyClient contextKey = "client"
)

func AddClientToContext(ctx context.Context, authorizer Authorizer) context.Context {
	return context.WithValue(ctx, contextKeyClient, authorizer)
}

func RetrieveClientFromContext(ctx context.Context) (Authorizer, bool) {
	client, ok := ctx.Value(contextKeyClient).(*client)
	return client, ok
}
