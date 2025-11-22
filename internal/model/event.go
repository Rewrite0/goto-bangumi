package model

import "time"

// Event 事件数据结构，携带事件的通用信息
// 事件流大概是 新的 下载->Check->重命名->通知

type RenameEvent struct {
	Torrent *Torrent
	Bangumi *Bangumi
}

type DownloadCheckEvent struct {
	Guids   []string
	Torrent *Torrent
	Bangumi *Bangumi
}

type DownloadingCheckEvent struct {
	Torrent   *Torrent
	Bangumi   *Bangumi
	StartTime time.Time `json:"start_time"` // 开始检查下载的时间
}

type NotificationEvent struct {
	Message Message
}
