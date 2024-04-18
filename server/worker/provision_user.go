package worker

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/auth0/go-auth0/management"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/pubsub"
)

type ProvisionUserHandler struct {
	ctx    context.Context
	db     *sqlx.DB
	auth0  *management.Management
	logger *slog.Logger
}

func NewProvisionUserHandler(
	ctx context.Context,
	db *sqlx.DB,
	auth0 *management.Management,
	logger *slog.Logger) *ProvisionUserHandler {
	return &ProvisionUserHandler{
		ctx:    ctx,
		db:     db,
		auth0:  auth0,
		logger: logger,
	}
}

func (handler *ProvisionUserHandler) Handle(message pubsub.Message) error {
	var (
		email               = message.Payload["email"].(string)
		allAuth0Users       []*management.User
		pagerAuth0Users     []*management.User
		pagerAuth0User      *management.User
		pagerConnectionName = "Username-Password-Authentication"
	)

	// Search for the user in auth0 by their email. This method searches all
	// auth connections for our account. To be on the safe side we'll filter
	// down to just the users who belong to our pager auth0 database connection.
	allAuth0Users, err := handler.auth0.User.ListByEmail(handler.ctx, email)
	if err != nil {
		return err
	}

	// Filter the users by their connection name
	for _, auth0User := range allAuth0Users {
		for _, identity := range auth0User.Identities {
			if *identity.Connection == pagerConnectionName {
				pagerAuth0Users = append(pagerAuth0Users, auth0User)
			}
		}
	}

	// By this point, we can be confident that this array will have 0 or 1
	// entries. 0 means this user is not in our pager database connection and
	// needs to be created.
	if len(pagerAuth0Users) > 0 {
		pagerAuth0User = pagerAuth0Users[0]
	} else {
		var (
			tmpPassword   = randPassword(25)
			emailVerified = false
		)

		// Create the user
		if err := handler.auth0.User.Create(handler.ctx, &management.User{
			Connection:    &pagerConnectionName,
			Email:         &email,
			Password:      &tmpPassword,
			EmailVerified: &emailVerified,
		}); err != nil {
			return err
		}

		// Unfortunately the method above doesn't return the user, so we have to
		// call ListByEmail once more to fetch the created user.
		users, err := handler.auth0.User.ListByEmail(handler.ctx, email)
		if err != nil {
			return err
		}
		if len(users) > 0 {
			pagerAuth0User = users[0]
		} else {
			return fmt.Errorf("failed to find user after creation: %s", email)
		}
	}

	if _, err := handler.db.ExecContext(
		handler.ctx,
		`UPDATE users
	SET idp_id = $1, modified = $2, modified_by = $3
	WHERE email = $4`,
		*pagerAuth0User.ID,
		time.Now().UTC(),
		"SYSTEM",
		email); err != nil {
		return err
	}

	return nil
}

func randPassword(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!?#%&*")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
