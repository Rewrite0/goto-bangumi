// Package notification provides functionalities to send notifications via different channels.
package notification

import (
	"log/slog"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"

)

type Client struct {
	notifier Notifier
}

var NotificationClient = &Client{}

func (c *Client) Init(config *model.NotificationConfig) {
	switch config.Type {
	case "telegram":
		c.notifier = NewTelegramNotifier()
		c.notifier.Init(config)
		slog.Info("[Notification] Telegram notifier initialized")
	default:
		slog.Warn("[Notification] Unknown notification type, no notifier initialized", "type", config.Type)
	}
}


// processMsg 处理为一个统一的发送文字
func (c *Client) processMsg(message *model.Message) {
	// 如果有集数信息, 那么就是一个更新通知
	if message.Episode != "" {
		message.Message = "番剧名称：" + message.Title +
			"\n季度：第" + message.Season + "季" +
			"\n更新集数：第" + message.Episode + "集"
	}

	// 对海报进行处理, 拿到海报的数据
	if message.PosterLink != "" {

		resp, err := network.LoadImage(message.PosterLink)
		if err != nil {
			slog.Error("[Notification] Failed to download poster image", "error", err)
			return
		}
		message.File = resp
	}
}

// Send 发送通知
func (c *Client) Send(message *model.Message) {
	if c.notifier == nil {
		slog.Warn("[Notification] No notifier initialized, skipping notification")
		return
	}

	// 处理消息内容
	c.processMsg(message)

	// 发送通知
	if c.notifier.PostMsg == nil {
		slog.Warn("[Notification] Notifier does not implement PostMsg, skipping notification")
		return
	}
	slog.Debug("[Notification] Sending notification", "title", message.Title, "episode", message.Episode)
}
