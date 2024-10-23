package authz

import (
	"context"
)

type contextKey string

const (
	contextKeyClient   contextKey = "client"
	contextKeyUserInfo contextKey = "user"
)

func AddClientToContext(ctx context.Context, authorizer Authorizer) context.Context {
	return context.WithValue(ctx, contextKeyClient, authorizer)
}

func RetrieveClientFromContext(ctx context.Context) (Authorizer, bool) {
	client, ok := ctx.Value(contextKeyClient).(*client)
	return client, ok
}

func AddUserInfoToContext(ctx context.Context, userInfo UserInfo) context.Context {
	return context.WithValue(ctx, contextKeyUserInfo, userInfo)
}

func RetrieveUserInfoFromContext(ctx context.Context) (UserInfo, bool) {
	userInfo, ok := ctx.Value(contextKeyUserInfo).(UserInfo)
	return userInfo, ok
}
