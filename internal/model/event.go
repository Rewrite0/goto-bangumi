package model

import "time"

// Event 事件数据结构，携带事件的通用信息
// 事件流大概是 新的 下载->Check->重命名->通知

type RenameEvent struct {
	Torrent *Torrent
	Bangumi *Bangumi
	Key   string // 用于防止重复处理的键值
}

// DownloadCheckEvent 下载检查事件, 用以检查真正的Guid 和是否下载成功
type DownloadCheckEvent struct {
	Guids   []string
	Torrent *Torrent
	Bangumi *Bangumi
	Key     string // 用于防止重复处理的键值
}

// DownloadingCheckEvent 下载中检查事件
type DownloadingCheckEvent struct {
	Torrent   *Torrent
	Bangumi   *Bangumi
	StartTime time.Time `json:"start_time"` // 开始检查下载的时间
	Key       string    // 用于防止重复处理的键值
}

type NotificationEvent struct {
	Message Message
	Key     string // 用于防止重复处理的键值
}
