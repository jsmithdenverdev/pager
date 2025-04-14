// Package authz provides context utilities for authorization.
package authz

import (
	"context"
)

type contextKey string

const (
	// contextKeyClient is the key used to store the Client in a context.
	contextKeyClient contextKey = "client"
	// contextKeyUserInfo is the key used to store the User information in a context.
	contextKeyUserInfo contextKey = "user"
)

// WithClient returns a new context with the provided Client stored in it.
func WithClient(ctx context.Context, authorizer *Client) context.Context {
	return context.WithValue(ctx, contextKeyClient, authorizer)
}

// ClientFrom extracts the Client from the context, if present.
func ClientFrom(ctx context.Context) (*Client, bool) {
	client, ok := ctx.Value(contextKeyClient).(*Client)
	return client, ok
}

// WithUser returns a new context with the provided User information stored in it.
func WithUser(ctx context.Context, userInfo User) context.Context {
	return context.WithValue(ctx, contextKeyUserInfo, userInfo)
}

// UserFrom extracts the User information from the context, if present.
func UserFrom(ctx context.Context) (User, bool) {
	userInfo, ok := ctx.Value(contextKeyUserInfo).(User)
	return userInfo, ok
}
