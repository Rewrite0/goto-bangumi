package model

import (
	"fmt"
	"time"
)

// DownloadStatus 种子下载状态
type DownloadStatus int

const (
	DownloadNone    DownloadStatus = 0 // 未下载
	DownloadSending DownloadStatus = 1 // 已发送到下载器
	DownloadDone    DownloadStatus = 2 // 下载完成
	DownloadError   DownloadStatus = 4 // 异常/手动停止下载
)

// Torrent 种子信息模型
// Torrent 什么时候会创建 1. 发送到下载前, 然后下载后更新download 2. 重命名后更新 renamed字段
type Torrent struct {
	Link         string    `gorm:"primaryKey;column:Link" json:"link"`
	DownloadUID string    `gorm:"index;column:download_uid" json:"download_uid"`
	Name        string    `gorm:"default:'';column:name" json:"name"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index;column:created_at" json:"created_at"`
	Downloaded  DownloadStatus `gorm:"default:0;column:downloaded" json:"downloaded"`
	Renamed     bool      `gorm:"default:false;column:renamed" json:"renamed"`
	// torrent 属于一个 bangumi
	BangumiID int    `gorm:"index;column:bangumi_id" json:"bangumi_id"`
	Homepage  string `gorm:"column:homepage" json:"homepage"`

	// GORM 关联对象（用于预加载）
	Bangumi *Bangumi `gorm:"foreignKey:BangumiID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
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
	EpisodeMetadata
	TorrentName string `json:"torrent_name"`
	Title       string `json:"title"`
	Suffix      string `json:"suffix"`
}

// SubtitleFile 字幕文件信息（继承 EpisodeFile）
type SubtitleFile struct {
	EpisodeFile
	Language string `json:"language"`
}

// TorrentInfo 种子解析信息
type TorrentInfo struct {
	Name       string
	InfoHashV1 string
	InfoHashV2 string
	MagnetURI  string
	File       []byte
}

func (t TorrentInfo) String() string {
	return fmt.Sprintf("Name: %s\n InfoHashV1: %s\n InfoHashV2: %s\n MagnetURI: %s",
		t.Name, t.InfoHashV1, t.InfoHashV2, t.MagnetURI)
}
