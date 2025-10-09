package model

import "time"

// Torrent 种子信息模型
type Torrent struct {
	URL                  string     `gorm:"primaryKey;column:url" json:"url"`
	Name                 string     `gorm:"default:'';column:name" json:"name"`
	CreatedAt            time.Time  `gorm:"autoCreateTime;index;column:created_at" json:"created_at"`
	Downloaded           bool       `gorm:"default:false;column:downloaded" json:"downloaded"`
	Renamed              bool       `gorm:"default:false;column:renamed" json:"renamed"`
	DownloadUID          *string    `gorm:"index;column:download_uid" json:"download_uid"`
	BangumiOfficialTitle string     `gorm:"default:'';index;column:bangumi_official_title" json:"bangumi_official_title"`
	BangumiSeason        int        `gorm:"default:1;index;column:bangumi_season" json:"bangumi_season"`
	RssLink              string     `gorm:"default:'';index;column:rss_link" json:"rss_link"`
	Homepage             string    `gorm:"column:homepage" json:"homepage"`
}

// TorrentDownloadInfo 种子下载信息
type TorrentDownloadInfo struct {
	ETA       *int   `json:"eta"`
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
