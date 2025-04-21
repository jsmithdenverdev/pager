package identity_test

import (
	"context"
	"testing"

	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/stretchr/testify/assert"
)

func TestWithUser(t *testing.T) {
	ctx := context.Background()
	user := identity.User{}

	ctxWithUser := identity.WithUser(ctx, user)
	retrievedUser, ok := identity.UserFrom(ctxWithUser)

	assert.True(t, ok)
	assert.Equal(t, user, retrievedUser)
}

func TestUserFrom(t *testing.T) {
	ctx := context.Background()
	user := identity.User{}

	ctxWithUser := identity.WithUser(ctx, user)
	retrievedUser, ok := identity.UserFrom(ctxWithUser)

	assert.True(t, ok)
	assert.Equal(t, user, retrievedUser)
}
