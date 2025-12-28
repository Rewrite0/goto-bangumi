package model

//Config 总配置结构体, 构成一个基础的配置, 但实际用的时候, 还是用的子配置
// type Config struct {
// 	Program       ProgramConfig       `json:"program" mapstructure:"program"`
// 	Downloader    DownloaderConfig    `json:"downloader" mapstructure:"downloader"`
// 	RssParse      RssParserConfig     `json:"rss_parser" mapstructure:"rss_parser"`
// 	BangumiManage BangumiRenameConfig `json:"bangumi_manage" mapstructure:"bangumi_manage"`
// 	Proxy         ProxyConfig         `json:"proxy" mapstructure:"proxy"`
// 	Notification  NotificationConfig  `json:"notification" mapstructure:"notification"`
// }
//
// func NewConfig() *Config {
// 	return &Config{
// 		Program:       *NewProgramConfig(),
// 		Downloader:    *NewDownloaderConfig(),
// 		RssParse:      *NewRssParserConfig(),
// 		BangumiManage: *NewBangumiRenameConfig(),
// 		Proxy:         *NewProxyConfig(),
// 		Notification:  *NewNotificationConfig(),
// 	}
// }

type ProgramConfig struct {
	RssTime     int    `json:"rss_time" mapstructure:"rss_time" validate:"gte=300"`
	WebuiPort   int    `json:"webui_port" mapstructure:"webui_port" validate:"gte=1,lte=65535"`
	PassWord    string `json:"password" mapstructure:"password"`
	DebugEnable bool   `json:"debug_enable" mapstructure:"debug_enable"`
}

func NewProgramConfig() *ProgramConfig {
	return &ProgramConfig{
		RssTime:   600,
		WebuiPort: 7892,
		PassWord:  "adminadmin",
		DebugEnable: false,
	}
}

type DownloaderConfig struct {
	Type     string `json:"type" mapstructure:"type" validate:"oneof=qbittorrent transmission aria2"`
	SavePath string `json:"path" mapstructure:"path" validate:"required"`
	Host     string `json:"host" mapstructure:"host" validate:"required"`
	Ssl      bool   `json:"ssl" mapstructure:"ssl"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

func NewDownloaderConfig() *DownloaderConfig {
	return &DownloaderConfig{
		Type:     "qbittorrent",
		SavePath: "/downloads/Bangumi",
		Host:     "127.0.0.1:8080",
		Username: "admin",
		Password: "adminadmin",
		Ssl:      false,
	}
}

type RssParserConfig struct {
	Enable         bool     `json:"enable" mapstructure:"enable"`
	Filter         []string `json:"filter" mapstructure:"filter"`
	Include        []string `json:"include" mapstructure:"include"`
	Language       string   `json:"language" mapstructure:"language" validate:"oneof=zh en jp"`
	MikanCustomURL string   `json:"mikan_custom_url" mapstructure:"mikan_custom_url"`
	TmdbAPIKey     string   `json:"tmdb_api_key" mapstructure:"tmdb_api_key"`
}

// NewRssParserConfig TODO: 加一个对字幕组的选项, 防止 include 污染到全局的排除
func NewRssParserConfig() *RssParserConfig {
	return &RssParserConfig{
		Enable:         true,
		Filter:         []string{"720", "\\d+-\\d+"},
		Include:        []string{},
		Language:       "zh",
		MikanCustomURL: "mikanani.me",
		TmdbAPIKey:     "",
	}
}

type BangumiRenameConfig struct {
	Enable       bool   `json:"enable" mapstructure:"enable"`
	EpsComplete  bool   `json:"eps_complete" mapstructure:"eps_complete"`
	RenameMethod string `json:"rename_method" mapstructure:"rename_method" validate:"oneof=advanced normal pn"`
	Year         bool   `json:"year" mapstructure:"year"`
	Group        bool   `json:"group" mapstructure:"group"`
}

func NewBangumiRenameConfig() *BangumiRenameConfig {
	return &BangumiRenameConfig{
		Enable:       true,
		EpsComplete:  false,
		RenameMethod: "advanced",
		Year:         false,
		Group:        false,
	}
}


// ProxyConfig is defined in network.go

type NotificationConfig struct {
	Enable bool   `json:"enable" mapstructure:"enable"`
	Type   string `json:"type" mapstructure:"type" validate:"oneof=telegram bark"`
	Token  string `json:"token" mapstructure:"token"`
	ChatID string `json:"chat_id" mapstructure:"chat_id"`
}

func NewNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Enable: false,
		Type:   "telegram",
		Token:  "",
		ChatID: "",
	}
}
