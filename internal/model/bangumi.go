package model

// import "fmt"

// Bangumi 番剧信息模型
// type Bangumi struct {
// 	ID          uint        `gorm:"primaryKey"`
// 	BangumiUID  string      `json:"bangumi_id" gorm:"default:'';comment:'番剧UID'"`
// 	BangumiInfo BangumiInfo `gorm:"foreignKey:BangumiUID;references:UID;constraint:OnUpdate:CASCADE"`
// 	// MikanID     string      `json:"mikan_id" gorm:"default:'';comment:'Mikan ID'"`
// 	// MikanItem   MikanItem   `gorm:"foreignKey:MikanID;references:MikanID"`
// 	// TmdbID      string      `json:"tmdb_id" gorm:"default:'';comment:'TMDB ID'"`
// 	// TmdbItem    TmdbItem    `gorm:"foreignKey:TmdbID;references:TmdbID"`
// }
//
// // TableName 指定表名
// func (Bangumi) TableName() string {
// 	return "bangumi"
// }

// String 格式化输出番剧信息
// func (b *Bangumi) String() string {
// 	return fmt.Sprintf("Bangumi{ID: %d, OfficialTitle: %s, Season: %d, RssLink: %s}",
// 		b.ID, b.OfficialTitle, b.Season, b.RssLink)
// }

type MikanItem struct {
	ID            uint   `gorm:"primaryKey"`
	MikanID       string `json:"mikan_id" gorm:"default:'';comment:'Mikan ID'"`
	OfficialTitle string `json:"official_title" gorm:"default:'';comment:'番剧中文名'"`
	Season        int    `json:"season" gorm:"default:1;comment:'季度'"`
	PosterLink    string `json:"poster_link" gorm:"default:'';comment:'海报链接'"`
}

type MikanBangumiMapping struct {
	ID         uint   `gorm:"primaryKey"`
	MikanID    string `json:"mikan_id" gorm:"default:'';comment:'Mikan ID'"`
	BangumiUID uint   `json:"bangumi_uid" gorm:"index;comment:'关联的番剧UID'"`
}

type TmdbItem struct {
	ID            uint    `gorm:"primaryKey"`
	TmdbID        string  `json:"tmdb_id" gorm:"default:'';comment:'TMDB ID'"`
	Year          string  `json:"year" gorm:"default:'';comment:'番剧年份'"`
	OriginalTitle string  `json:"original_title" gorm:"default:'';comment:'番剧原名'"`
	EpisodeCount  int     `json:"episode_count" gorm:"default:0;comment:'总集数'"`
	Title         string  `json:"title" gorm:"default:'';comment:'番剧名称'"`
	Season        int     `json:"season" gorm:"default:1;comment:'季度'"`
	PosterURL     string  `json:"poster_url" gorm:"default:'';comment:'海报链接'"`
	VoteAverage   float64 `json:"vote_average" gorm:"default:0;comment:'评分'"`
}

type TmdbBangumiMapping struct {
	ID         uint     `gorm:"primaryKey"`
	TmdbID     string   `json:"tmdb_id" gorm:"default:'';comment:'TMDB ID'"`
	TmdbItem   TmdbItem `gorm:"foreignKey:TmdbID;references:TmdbID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BangumiUID uint     `json:"bangumi_uid" gorm:"index;comment:'关联的番剧UID'"`
}

// BangumiParser 用来存储番剧解析器的原始信息
type BangumiParser struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"default:'';comment:'番剧名称'"`
	Group     string `gorm:"default:'';comment:'字幕组'"`
	Season    int    `gorm:"default:1;comment:'季度'"`
	SeasonRaw string `gorm:"default:'';comment:'季度原名'"`
	DPI       string `gorm:"default:'';comment:'分辨率'"`
	Sub       string `gorm:"default:'';comment:'字幕语言'"`
	SubType   string `gorm:"default:'';comment:'字幕类型'"`
	Source    string `gorm:"default:'';comment:'来源'"`
	AudioInfo string `gorm:"default:'';comment:'音频信息'"`
	VideoInfo string `gorm:"default:'';comment:'视频信息'"`
}

type BangumiParserMapping struct {
	ID              uint `gorm:"primaryKey"`
	BangumiParserID uint `json:"bangumi_parser_id" gorm:"index;comment:'关联的番剧解析器ID'"`
	BangumiUID      uint `json:"bangumi_uid" gorm:"index;comment:'关联的番剧UID'"`
}

// Bangumi 用于存储一些可配置的番剧信息
type Bangumi struct {
	UID           uint   `gorm:"primaryKey"`
	OfficialTitle string `json:"official_title" gorm:"default:'';comment:'番剧中文名'"`
	Year          string `json:"year" gorm:"default:'';comment:'番剧年份'"`
	Season        int    `json:"season" gorm:"default:1;comment:'番剧季度'"`
	EpsCollect    bool   `json:"eps_collect" gorm:"default:false;comment:'是否已收集'"`
	Offset        int    `json:"offset" gorm:"default:0;comment:'番剧偏移量'"`
	IncludeFilter string `json:"include_filter" gorm:"default:'';comment:'番剧包含过滤器'"`
	ExcludeFilter string `json:"exclude_filter" gorm:"default:'';comment:'番剧排除过滤器'"`
	Parser        string `json:"parser" gorm:"default:'mikan';comment:'番剧解析器'"`
	RssLink       string `json:"rss_link" gorm:"default:'';comment:'番剧RSS链接'"`
	PosterLink    string `json:"poster_link" gorm:"default:'';comment:'番剧海报链接'"`
	// RuleName      string `json:"rule_name" gorm:"default:'';comment:'番剧规则名'"`
	// Added         bool   `json:"added" gorm:"default:false;comment:'是否已添加'"`
	Deleted bool `json:"deleted" gorm:"default:false;comment:'是否已删除'"`
}

// 重新设计几个表来确定 bangumi 和 mikanid, tmdbid , bangumiid 的关系
// mikanid -> id, mikan_id,rss_link
// tmdbid -> id, tmdb_id,rss_link, 这个 id 要加个 #season
// bangumiid -> id, bangumi_id,rss_link
// 当解析到相同的, ID, 重建 ID
// bangumi 里面要保留的项: 前端要: title,year, seasion , group,  group_name,poster_link,parser,

// 流程如下:
// 1. 解析 rss, 先排除已经下载的 torrent , 这是通过 torrent 表来做的
// 2. 对于没有下载的 torrent, 通过 BangumiParser 的 Title 得到对应的 BangumiParser ID
// 3. 通过 BangumiParserMapping 得到对应的 BangumiUID
// 4. 通过 BangumiUID 得到对应的 BangumiInfo
// 5. 通过 BangumiInfo 对 Bangumi 表进行更新/创建

// 要是没有对应的 BangumiParser, 调用 raw_parser 解析, 得到 对应的 BangumiParser, 再通过 MikanParser(如果有 homepage) 解析得到 mikan_id, 通过 mikan_id map 去找 BangumiUID
// 要是没找到 再去调用 tmdb_parser 解析, 得到 tmdb_id, 通过 tmdb_id map 去找 BangumiUID
// 对于找到了 BangumiUID 的, 更新 BangumiParserMapping, 其中要是没找到 mikan_id 的更新 mikan_id,
// 没找到 tmdb_id 的就是新的番剧了, 更新 tmdb_id, mikan_id,  BangumiInfo和对应的 mapping
// 对于 tmdb ,我们要拿到所有的集数, 用以显示下载了多少集, 还差多少集
