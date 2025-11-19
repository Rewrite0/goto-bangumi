package notification

import (
	"fmt"
	"log/slog"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
)

// TeleConfig Telegram 通知器配置
type TeleConfig struct {
	Token   string `json:"token" mapstructure:"token"`
	ChatID  string `json:"chat_id" mapstructure:"chat_id"`
	BaseURL string
}

// TelegramNotifier Telegram 通知器
type TelegramNotifier struct {
	config *TeleConfig
}

// NewTelegramNotifier 创建一个新的 Telegram 通知器实例
func NewTelegramNotifier() *TelegramNotifier {
	config := &TeleConfig{}
	return &TelegramNotifier{
		config: config,
	}
}

// Init 初始化通知器
func (t *TelegramNotifier) Init(config *model.NotificationConfig)  {
	t.config.Token = config.Token
	t.config.ChatID = config.ChatID
	t.config.BaseURL = fmt.Sprintf("https://api.telegram.org/bot%s/", t.config.Token)
	slog.Debug("[Telegram] Telegram 通知器初始化成功")
}

// PostMsg 发送 Telegram 通知
func (t *TelegramNotifier) PostMsg(message *model.Message) (error) {
	if message == nil {
		return fmt.Errorf("消息不能为空")
	}

	var err error
	if len(message.File) > 0 {
		// 发送带图片的消息
		err = t.sendPhoto(message.Message, message.File)
	} else {
		// 发送纯文本消息
		err = t.sendText(message.Message)
	}

	if err != nil {
		slog.Error("[Telegram] 通知发送失败", "error", err)
		return  err
	}

	return nil
}

// sendPhoto 发送带图片的消息
func (t *TelegramNotifier) sendPhoto(text string, photo []byte) error {
	url := fmt.Sprintf("%ssendPhoto", t.config.BaseURL)

	// 准备表单数据
	formData := map[string]string{
		"chat_id":              t.config.ChatID,
		"caption":              text,
		"disable_notification": "true",
	}

	// 准备文件数据
	files := map[string][]byte{
		"photo": photo,
	}

	// 使用 network.PostData 发送请求
	client := network.GetRequestClient()
	resp, err := client.PostData(url, formData, files)
	if err != nil {
		return fmt.Errorf("发送图片消息失败: %w", err)
	}

	slog.Debug("[Telegram] 图片消息发送成功", "response_size", len(resp))
	return nil
}

// sendText 发送纯文本消息
func (t *TelegramNotifier) sendText(text string) error {
	url := fmt.Sprintf("%ssendMessage", t.config.BaseURL)

	// 准备表单数据
	formData := map[string]string{
		"chat_id":              t.config.ChatID,
		"text":                 text,
		"disable_notification": "true",
	}

	// 使用 network.PostData 发送请求,不需要文件
	client := network.GetRequestClient()
	resp, err := client.PostData(url, formData, nil)
	if err != nil {
		return fmt.Errorf("发送文本消息失败: %w", err)
	}

	slog.Debug("[Telegram] 文本消息发送成功", "response_size", len(resp))
	return nil
}
