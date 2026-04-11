package notification

import (
	"context"
	"fmt"
	"log/slog"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
)

// TelegramNotifier sends notifications via Telegram Bot API.
type TelegramNotifier struct {
	token   string
	chatID  string
	baseURL string
}

// NewTelegramNotifier creates a ready-to-use Telegram notifier.
func NewTelegramNotifier(config *model.NotificationConfig) (*TelegramNotifier, error) {
	if config.Token == "" || config.ChatID == "" {
		return nil, fmt.Errorf("telegram requires token and chat_id")
	}
	return &TelegramNotifier{
		token:   config.Token,
		chatID:  config.ChatID,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s/", config.Token),
	}, nil
}

// Send sends a notification via Telegram.
func (t *TelegramNotifier) Send(ctx context.Context, message *Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	var err error
	if len(message.Image) > 0 {
		err = t.sendPhoto(ctx, message.Text, message.Image)
	} else {
		err = t.sendText(ctx, message.Text)
	}

	if err != nil {
		slog.Error("[Telegram] Failed to send notification", "error", err)
		return err
	}

	return nil
}

func (t *TelegramNotifier) sendPhoto(ctx context.Context, text string, photo []byte) error {
	url := fmt.Sprintf("%ssendPhoto", t.baseURL)

	formData := map[string]string{
		"chat_id":              t.chatID,
		"caption":              text,
		"disable_notification": "true",
	}

	files := map[string][]byte{
		"photo": photo,
	}

	client := network.GetRequestClient()
	_, err := client.PostData(ctx, url, formData, files)
	if err != nil {
		return fmt.Errorf("failed to send photo message: %w", err)
	}

	return nil
}

func (t *TelegramNotifier) sendText(ctx context.Context, text string) error {
	url := fmt.Sprintf("%ssendMessage", t.baseURL)

	formData := map[string]string{
		"chat_id":              t.chatID,
		"text":                 text,
		"disable_notification": "true",
	}

	client := network.GetRequestClient()
	_, err := client.PostData(ctx, url, formData, nil)
	if err != nil {
		return fmt.Errorf("failed to send text message: %w", err)
	}

	return nil
}
