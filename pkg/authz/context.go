package authz

import (
	"context"
)

type contextKey string

const (
	contextKeyClient   contextKey = "client"
	contextKeyUserInfo contextKey = "user"
)

func WithClient(ctx context.Context, authorizer *Client) context.Context {
	return context.WithValue(ctx, contextKeyClient, authorizer)
}

func ClientFrom(ctx context.Context) (*Client, bool) {
	client, ok := ctx.Value(contextKeyClient).(*Client)
	return client, ok
}

func WithUser(ctx context.Context, userInfo User) context.Context {
	return context.WithValue(ctx, contextKeyUserInfo, userInfo)
}

func UserFrom(ctx context.Context) (User, bool) {
	userInfo, ok := ctx.Value(contextKeyUserInfo).(User)
	return userInfo, ok
}
