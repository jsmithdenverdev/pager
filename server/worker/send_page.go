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
		deviceID   = message.Payload["deviceId"].(string)
		pageID     = message.Payload["pageId"].(string)
		deliveryID = message.Payload["deliveryId"].(string)
	)

	handler.logger.InfoContext(
		handler.ctx,
		"SendPageHander::Handle",
		"deviceID", deviceID,
		"pageID", pageID,
		"deliveryID", deliveryID)

	return nil
}
