package worker

import (
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/pubsub"
)

type SendPageHandler struct {
	ctx    context.Context
	db     *sqlx.DB
	logger *slog.Logger
}

func NewSendPageHandler(
	ctx context.Context,
	db *sqlx.DB,
	logger *slog.Logger) *SendPageHandler {
	return &SendPageHandler{
		ctx:    ctx,
		db:     db,
		logger: logger,
	}
}

func (handler *SendPageHandler) Handle(message pubsub.Message) error {
	var (
		payload pubsub.PayloadSendPage
	)

	if err := pubsub.Unmarshal(message, &payload); err != nil {
		return err
	}

	handler.logger.InfoContext(
		handler.ctx,
		"SendPageHander::Handle",
		"deliveryID", payload.PageDeliveryID,
	)

	return nil
}
