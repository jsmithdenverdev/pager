package authz_test

import (
	"context"
	"testing"

	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/stretchr/testify/assert"
)

func TestWithClient(t *testing.T) {
	ctx := context.Background()
	client := &authz.Client{}

	ctxWithClient := authz.WithClient(ctx, client)
	retrievedClient, ok := authz.ClientFrom(ctxWithClient)

	assert.True(t, ok)
	assert.Equal(t, client, retrievedClient)
}

func TestClientFrom(t *testing.T) {
	ctx := context.Background()
	client := &authz.Client{}

	ctxWithClient := authz.WithClient(ctx, client)
	retrievedClient, ok := authz.ClientFrom(ctxWithClient)

	assert.True(t, ok)
	assert.Equal(t, client, retrievedClient)
}

func TestWithUser(t *testing.T) {
	ctx := context.Background()
	user := authz.User{}

	ctxWithUser := authz.WithUser(ctx, user)
	retrievedUser, ok := authz.UserFrom(ctxWithUser)

	assert.True(t, ok)
	assert.Equal(t, user, retrievedUser)
}

func TestUserFrom(t *testing.T) {
	ctx := context.Background()
	user := authz.User{}

	ctxWithUser := authz.WithUser(ctx, user)
	retrievedUser, ok := authz.UserFrom(ctxWithUser)

	assert.True(t, ok)
	assert.Equal(t, user, retrievedUser)
}
