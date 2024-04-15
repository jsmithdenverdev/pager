package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type UserProvisioner struct {
	ctx    context.Context
	ticker *time.Ticker
	db     *sqlx.DB
	logger *slog.Logger
}

func NewUserProvisioner(
	ctx context.Context,
	intervalS int,
	db *sqlx.DB,
	logger *slog.Logger,
) *UserProvisioner {
	return &UserProvisioner{
		ctx:    ctx,
		ticker: time.NewTicker(time.Duration(intervalS) * time.Second),
		db:     db,
		logger: logger,
	}
}

func (worker *UserProvisioner) Start() error {
	for {
		select {
		case <-worker.ctx.Done():
			return worker.ctx.Err()
		case <-worker.ticker.C:
			if err := worker.work(); err != nil {
				worker.ticker.Stop()
				return err
			}
		}
	}
}

func (worker *UserProvisioner) work() error {
	return nil
}
