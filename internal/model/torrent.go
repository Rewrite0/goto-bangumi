package model

import "time"

// Torrent 种子信息模型
// Torrent 什么时候会创建 1. 发送到下载前, 然后下载后更新download 2. 重命名后更新 renamed字段
type Torrent struct {
	URL             string    `gorm:"primaryKey;column:url" json:"url"`
	Name            string    `gorm:"default:'';column:name" json:"name"`
	CreatedAt       time.Time `gorm:"autoCreateTime;index;column:created_at" json:"created_at"`
	Downloaded      bool      `gorm:"default:false;column:downloaded" json:"downloaded"`
	Renamed         bool      `gorm:"default:false;column:renamed" json:"renamed"`
	EpisodeID       uint      `gorm:"index;column:episode_id" json:"episode_id"`
	DownloadUID     string    `gorm:"index;column:download_uid" json:"download_uid"`
	BangumiParserID uint      `gorm:"index;column:bangumi_parser_id" json:"bangumi_parser_id"`
	BangumiUID      uint      `gorm:"index;column:bangumi_uid" json:"bangumi_uid"`
	Homepage        string    `gorm:"column:homepage" json:"homepage"`
}

// TorrentDownloadInfo 种子下载信息
type TorrentDownloadInfo struct {
	ETA       int    `json:"eta"`
	SavePath  string `json:"save_path"`
	Completed int    `json:"completed"`
}

// TorrentUpdate 种子更新信息
type TorrentUpdate struct {
	Downloaded bool `json:"downloaded"`
}

// EpisodeFile 剧集文件信息（继承 Episode）
type EpisodeFile struct {
	Episode
	TorrentName string `json:"torrent_name"`
	Title       string `json:"title"`
	Suffix      string `json:"suffix"`
}

// SubtitleFile 字幕文件信息（继承 EpisodeFile）
type SubtitleFile struct {
	EpisodeFile
	Language string `json:"language"`
}
