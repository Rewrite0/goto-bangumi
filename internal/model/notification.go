package model

// Message 表示一个通知消息
type Message struct {
	Title      string // 标题
	Message    string // 消息内容
	Season     string // 番剧季度
	Episode    string // 番剧集数
	PosterLink string // 番剧海报路径
	File       []byte // 文件内容（图片等）
}

// NewMessage 创建一个新的消息实例
func NewMessage(title, message string) *Message {
	return &Message{
		Title:   title,
		Message: message,
	}
}
