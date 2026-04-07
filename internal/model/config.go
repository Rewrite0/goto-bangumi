package model

// Config is the top-level configuration struct containing all sub-configs.
type Config struct {
	Program      ProgramConfig       `yaml:"program" env-prefix:"PROGRAM_"`
	Downloader   DownloaderConfig    `yaml:"downloader" env-prefix:"DOWNLOADER_"`
	Parser       RssParserConfig     `yaml:"parser" env-prefix:"PARSER_"`
	Rename       BangumiRenameConfig `yaml:"rename" env-prefix:"RENAME_"`
	Notification NotificationConfig  `yaml:"notification" env-prefix:"NOTIFICATION_"`
	Proxy        ProxyConfig         `yaml:"proxy" env-prefix:"PROXY_"`
}

type ProgramConfig struct {
	RssTime     int    `yaml:"rss_time" env:"RSS_TIME" env-default:"600"`
	WebuiPort   int    `yaml:"webui_port" env:"WEBUI_PORT" env-default:"7892"`
	PassWord    string `yaml:"password" env:"PASSWORD" env-default:"adminadmin"`
	DebugEnable bool   `yaml:"debug_enable" env:"DEBUG_ENABLE" env-default:"false"`
}

type DownloaderConfig struct {
	Type     string `yaml:"type" env:"TYPE" env-default:"qbittorrent"`
	SavePath string `yaml:"path" env:"PATH" env-default:"/downloads/Bangumi"`
	Host     string `yaml:"host" env:"HOST" env-default:"127.0.0.1:8080"`
	Ssl      bool   `yaml:"ssl" env:"SSL" env-default:"false"`
	Username string `yaml:"username" env:"USERNAME" env-default:"admin"`
	Password string `yaml:"password" env:"PASSWORD" env-default:"adminadmin"`
}

type RssParserConfig struct {
	Enable         bool     `yaml:"enable" env:"ENABLE" env-default:"true"`
	Filter         []string `yaml:"filter"`
	Include        []string `yaml:"include"`
	Language       string   `yaml:"language" env:"LANGUAGE" env-default:"zh"`
	MikanCustomURL string   `yaml:"mikan_custom_url" env:"MIKAN_CUSTOM_URL" env-default:"mikanani.me"`
	TmdbAPIKey     string   `yaml:"tmdb_api_key" env:"TMDB_API_KEY"`
}

type BangumiRenameConfig struct {
	Enable       bool   `yaml:"enable" env:"ENABLE" env-default:"true"`
	EpsComplete  bool   `yaml:"eps_complete" env:"EPS_COMPLETE" env-default:"false"`
	RenameMethod string `yaml:"rename_method" env:"RENAME_METHOD" env-default:"advanced"`
	Year         bool   `yaml:"year" env:"YEAR" env-default:"false"`
	Group        bool   `yaml:"group" env:"GROUP" env-default:"false"`
}

type NotificationConfig struct {
	Enable bool   `yaml:"enable" env:"ENABLE" env-default:"false"`
	Type   string `yaml:"type" env:"TYPE" env-default:"telegram"`
	Token  string `yaml:"token" env:"TOKEN"`
	ChatID string `yaml:"chat_id" env:"CHAT_ID"`
}
