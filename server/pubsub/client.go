package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Client struct {
	db       *sqlx.DB
	ln       *pq.Listener
	ctx      context.Context
	handlers map[Topic]Handler
	topics   map[Topic]topicConfig
	logger   *slog.Logger
}

func NewClient(ctx context.Context, db *sqlx.DB, ln *pq.Listener, logger *slog.Logger) *Client {
	return &Client{
		ctx:      ctx,
		db:       db,
		ln:       ln,
		logger:   logger,
		handlers: make(map[Topic]Handler),
		topics:   make(map[Topic]topicConfig),
	}
}

func Subscribe(client *Client, topic Topic, handler Handler) error {
	if _, ok := client.handlers[topic]; ok {
		return fmt.Errorf("topic already registered: %s", topic)
	}
	if err := client.ln.Listen(string(topic)); err != nil {
		return err
	}

	var config topicConfig
	if err := client.db.QueryRowxContext(
		client.ctx,
		`SELECT topic, retries_enabled, retries
		FROM topics
		WHERE topic = $1`,
		topic,
	).StructScan(&config); err != nil {
		return err
	}
	client.handlers[topic] = handler
	client.topics[topic] = config
	return nil
}

func (client *Client) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			if err := client.ln.Close(); err != nil {
				return err
			}
			return ctx.Err()
		case notification := <-client.ln.Notify:
			topic := notification.Channel
			var (
				config, cOk  = client.topics[Topic(topic)]
				handler, hOk = client.handlers[Topic(topic)]
			)

			// If we foudn both config and a handler for this topic
			if cOk && hOk {
				go func() {
					var message Message
					if err := json.Unmarshal([]byte(notification.Extra), &message); err != nil {
						client.logger.ErrorContext(
							client.ctx,
							"failed to unmarshal message body",
							"message", notification.Extra,
							"error", err)
						if err := client.markUnprocessable(Topic(topic), notification); err != nil {
							client.logger.ErrorContext(
								client.ctx,
								"failed to record unprocessable message",
								"error", err)
						}
						return
					}

					if config.RetriesEnabled && message.Retries == config.Retries {
						if err := client.addToDeadletter(message.ID); err != nil {
							client.logger.ErrorContext(
								client.ctx,
								"failed to add to deadletter",
								"topic", config,
								"message", message,
								"notification", notification,
								"error", err)
						}
						return
					}

					if err := handler.Handle(message); err != nil {
						if config.RetriesEnabled {
							if err := client.requeue(message); err != nil {
								client.logger.ErrorContext(
									client.ctx,
									"failed to requeue",
									"topic", config,
									"message", message,
									"notification", notification,
									"error", err)
							}
						} else {
							if err := client.addToDeadletter(message.ID); err != nil {
								client.logger.ErrorContext(
									client.ctx,
									"failed to add to deadletter",
									"topic", config,
									"message", message,
									"notification", notification,
									"error", err)
							}
						}
					}

				}()
			} else {
				continue
			}
		}
	}
}

func (client *Client) Publish(tx *sqlx.Tx, messages []Message) error {

	if _, err := tx.NamedExecContext(
		client.ctx,
		`INSERT INTO messages (topic, payload, created_by, modified_by)
		VALUES (:topic, :payload, :created_by, :modified_by)`,
		messages); err != nil {
		return err
	}

	return nil
}

func (client *Client) markUnprocessable(topic Topic, notification *pq.Notification) error {
	if _, err := client.db.ExecContext(
		client.ctx,
		`INSERT INTO unprocessable_messages (topic, payload, created_by, modified_by)
		VALUES ($1, $2, $3, $4)`,
		topic,
		notification.Extra,
		"SYSTEM",
		"SYSTEM"); err != nil {
		return err
	}

	return nil
}

func (client *Client) addToDeadletter(id string) error {
	if _, err := client.db.ExecContext(
		client.ctx,
		`INSERT INTO messages_dl (message_id, created_by, modified_by)
		VALUES ($1, $2, $3)`,
		id,
		"SYSTEM",
		"SYSTEM"); err != nil {
		return err
	}
	return nil
}

func (client *Client) requeue(message Message) error {
	if _, err := client.db.ExecContext(
		client.ctx,
		`UPDATE messages
		SET retries = $1, modified = $2
		WHERE id = $3`,
		message.Retries+1,
		time.Now().UTC(),
		message.ID,
	); err != nil {
		return err
	}

	return nil
}
