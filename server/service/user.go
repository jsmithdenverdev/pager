package service

import (
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
)

type UserService struct {
	ctx        context.Context
	user       string
	authclient *authz.Client
	db         *sqlx.DB
	logger     *slog.Logger
}

func NewUserService(ctx context.Context, user string, authz *authz.Client, db *sqlx.DB, logger *slog.Logger) *UserService {
	return &UserService{
		ctx:        ctx,
		user:       user,
		authclient: authz,
		db:         db,
		logger:     logger,
	}
}

func (service *UserService) Info() (models.User, error) {
	var user models.User
	if err := service.db.QueryRowxContext(
		service.ctx,
		`SELECT id, email, idp_id, status, created, created_by, modified, modified_by
		 FROM users
		 WHERE idp_id = $1`,
		service.user,
	).StructScan(&user); err != nil {
		return user, err
	}

	return user, nil
}
