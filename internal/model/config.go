package model

type Config struct {
	Program       ProgramConfig       `json:"program" mapstructure:"program"`
	Downloader    DownloaderConfig    `json:"downloader" mapstructure:"downloader"`
	RssParse      RssParserConfig     `json:"rss_parser" mapstructure:"rss_parser"`
	BangumiManage BangumiRenameConfig `json:"bangumi_manage" mapstructure:"bangumi_manage"`
	Log           LogConfig           `json:"log" mapstructure:"log"`
	Proxy         ProxyConfig         `json:"proxy" mapstructure:"proxy"`
	Notification  NotificationConfig  `json:"notification" mapstructure:"notification"`
	Password      string              `json:"password" mapstructure:"password"`
}

type ProgramConfig struct {
	RssTime   int `json:"rss_time" mapstructure:"rss_time"`
	WebuiPort int `json:"webui_port" mapstructure:"webui_port"`
}

type DownloaderConfig struct {
	Type     string `json:"type" mapstructure:"type"`
	SavePath string `json:"path" mapstructure:"path"`
	Host     string `json:"host" mapstructure:"host"`
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
	Language       string   `json:"language" mapstructure:"language"`
	MikanCustomURL string   `json:"mikan_custom_url" mapstructure:"mikan_custom_url"`
	TmdbAPIKey     string   `json:"tmdb_api_key" mapstructure:"tmdb_api_key"`
}

type BangumiRenameConfig struct {
	Enable       bool   `json:"enable" mapstructure:"enable"`
	EpsComplete  bool   `json:"eps_complete" mapstructure:"eps_complete"`
	RenameMethod string `json:"rename_method" mapstructure:"rename_method"`
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

type LogConfig struct {
	DebugEnable bool `json:"debug_enable" mapstructure:"debug_enable"`
}

// ProxyConfig is defined in network.go

type NotificationConfig struct {
	Enable bool   `json:"enable" mapstructure:"enable"`
	Type   string `json:"type" mapstructure:"type"`
	Token  string `json:"token" mapstructure:"token"`
	ChatID string `json:"chat_id" mapstructure:"chat_id"`
}
