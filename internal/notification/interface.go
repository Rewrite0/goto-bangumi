package notification

import "context"

// Notifier defines the interface for a notification channel.
type Notifier interface {
	Send(ctx context.Context, message *Message) error
}
