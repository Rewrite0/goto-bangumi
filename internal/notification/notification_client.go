package notification

import (
	"context"
	"log/slog"

	"goto-bangumi/internal/model"
)

// Client wraps a single Notifier selected by configuration.
type Client struct {
	notifier Notifier
}

// NotificationClient is the global notification client.
var NotificationClient = &Client{}

// Init initializes the notification client with the configured channel.
func (c *Client) Init(config *model.NotificationConfig) {
	if !config.Enable {
		slog.Info("[Notification] Notification disabled")
		return
	}

	switch config.Type {
	case "telegram":
		notifier, err := NewTelegramNotifier(config)
		if err != nil {
			slog.Error("[Notification] Failed to init Telegram", "error", err)
			return
		}
		c.notifier = notifier
	default:
		slog.Warn("[Notification] Unknown notification type", "type", config.Type)
	}
}

// Send sends a notification message. Errors are logged but not returned.
func (c *Client) Send(ctx context.Context, message *Message) {
	if c.notifier == nil {
		slog.Warn("[Notification] No notifier initialized, skipping")
		return
	}

	if err := c.notifier.Send(ctx, message); err != nil {
		slog.Error("[Notification] Send failed", "error", err)
	}
}
