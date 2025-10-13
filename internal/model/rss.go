package model

// RSSItem RSS订阅项模型
type RSSItem struct {
	ID        uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string `gorm:"column:name" json:"name"`
	URL       string  `gorm:"default:'https://mikanani.me';index;column:url" json:"url"`
	Aggregate bool    `gorm:"default:false;column:aggregate" json:"aggregate"`
	Parser    string  `gorm:"default:'mikan';column:parser" json:"parser"`
	Enabled   bool    `gorm:"default:true;column:enabled" json:"enabled"`
}
