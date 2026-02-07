package model

// RSSItem RSS订阅项模型
type RSSItem struct {
	ID        uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string `gorm:"column:name" json:"name"`
	Link       string  `gorm:"default:'https://mikanani.me';index;column:link" json:"link"`
	Aggregate bool    `gorm:"default:false;column:aggregate" json:"aggregate"`
	Parse    string  `gorm:"default:'tmdb';column:parser" json:"parser"`
	IncludeFilter string `json:"include_filter" gorm:"default:'';comment:'番剧包含过滤器'"`
	ExcludeFilter string `json:"exclude_filter" gorm:"default:'';comment:'番剧排除过滤器'"`
	Enabled   bool    `gorm:"default:true;column:enabled" json:"enabled"`
}
