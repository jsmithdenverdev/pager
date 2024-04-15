package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type PageDeliverer struct {
	ctx    context.Context
	ticker *time.Ticker
	db     *sqlx.DB
	logger *slog.Logger
}

func NewPageDeliverer(
	ctx context.Context,
	intervalS int,
	db *sqlx.DB,
	logger *slog.Logger) *PageDeliverer {
	return &PageDeliverer{
		ctx:    ctx,
		ticker: time.NewTicker(time.Duration(intervalS) * time.Second),
		db:     db,
		logger: logger,
	}
}

func (worker *PageDeliverer) Start() error {
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

func (worker *PageDeliverer) work() error {
	rows, err := worker.db.QueryxContext(
		worker.ctx,
		`SELECT DISTINCT (p.id)
		 FROM pages p
		 INNER JOIN page_deliveries pd on pd.page_id = p.id
		 WHERE pd.status = 'PENDING'`)

	if err != nil {
		return err
	}

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	worker.logger.InfoContext(worker.ctx,
		"DeliveryWorker",
		"pages needing delivery", ids)
	return nil
}
