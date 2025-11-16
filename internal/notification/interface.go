package notification

import (

	"goto-bangumi/internal/model"
)

// Notifier 定义通知器的统一接口
type Notifier interface {
	// PostMsg 发送通知消息
	// 返回 true 表示发送成功，false 表示失败
	PostMsg(message *model.Message) (error)

	// Init 初始化通知器
	Init(config *model.NotificationConfig)
}
