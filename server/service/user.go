package service

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/models"
)

type UserService struct {
	ctx    context.Context
	db     *sqlx.DB
	userId string
}

func NewUserService(ctx context.Context, db *sqlx.DB, userId string) *UserService {
	return &UserService{
		ctx:    ctx,
		db:     db,
		userId: userId,
	}
}

func (service *UserService) Info() (*models.User, error) {
	var user models.User
	if err := service.db.QueryRowxContext(
		service.ctx,
		`SELECT id, email, idp_id, status, created, created_by, modified, modified_by
		 FROM users
		 WHERE idp_id = $1`,
		service.userId,
	).StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
