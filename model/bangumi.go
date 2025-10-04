package model

import (
	"gorm.io/gorm"
)

// Bangumi 番剧信息模型
type Bangumi struct {
	gorm.Model
	OfficialTitle string  `json:"official_title" gorm:"default:'';comment:'番剧中文名'"`
	Year          *string `json:"year" gorm:"comment:'番剧年份'"`
	TitleRaw      string  `json:"title_raw" gorm:"default:'title_raw';comment:'番剧原名'"`
	Season        int     `json:"season" gorm:"default:1;comment:'番剧季度'"`
	SeasonRaw     *string `json:"season_raw" gorm:"comment:'番剧季度原名'"`
	GroupName     *string `json:"group_name" gorm:"comment:'字幕组'"`
	DPI           *string `json:"dpi" gorm:"comment:'分辨率'"`
	Source        *string `json:"source" gorm:"comment:'来源'"`
	Subtitle      *string `json:"subtitle" gorm:"comment:'字幕'"`
	EpsCollect    bool    `json:"eps_collect" gorm:"default:false;comment:'是否已收集'"`
	Offset        int     `json:"offset" gorm:"default:0;comment:'番剧偏移量'"`
	IncludeFilter string  `json:"include_filter" gorm:"default:'';comment:'番剧包含过滤器'"`
	ExcludeFilter string  `json:"exclude_filter" gorm:"default:'';comment:'番剧排除过滤器'"`
	Parser        string  `json:"parser" gorm:"default:'mikan';comment:'番剧解析器'"`
	TmdbID        *string `json:"tmdb_id" gorm:"comment:'番剧TMDB ID'"`
	BangumiID     *string `json:"bangumi_id" gorm:"comment:'番剧Bangumi ID'"`
	MikanID       *string `json:"mikan_id" gorm:"comment:'番剧Mikan ID'"`
	RssLink       string  `json:"rss_link" gorm:"default:'';comment:'番剧RSS链接'"`
	PosterLink    string  `json:"poster_link" gorm:"default:'';comment:'番剧海报链接'"`
	RuleName      *string `json:"rule_name" gorm:"comment:'番剧规则名'"`
	Added         bool    `json:"added" gorm:"default:false;comment:'是否已添加'"`
	Deleted       bool    `json:"deleted" gorm:"default:false;comment:'是否已删除'"`
}
