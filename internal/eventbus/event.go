// Package eventbus provides a simple event bus for decoupling communication between modules.
package eventbus

// EventHandler 事件处理函数类型
type EventHandler func(data any) error

// 预定义的事件类型常量
const (
	// EventDownloadCompleted 下载完成事件
	EventDownloadCompleted = "download.completed"

	// EventRenameCompleted 重命名完成事件
	EventRenameCompleted = "rename.completed"

	// EventRenameFailed 重命名失败事件
	EventRenameFailed = "rename.failed"

	// EventNotificationSent 通知发送事件
	EventNotificationSent = "notification.sent"
)

// Event 事件数据结构，携带事件的通用信息
type Event struct {
	Type string // 事件类型
	Data any    // 事件数据
}

// NewEvent 创建一个新的事件实例
func NewEvent(eventType string, data any) *Event {
	return &Event{
		Type: eventType,
		Data: data,
	}
}
