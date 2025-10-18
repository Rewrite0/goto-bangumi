//Package conf 提供了基础的配置加载功能
package conf

import (
	"github.com/spf13/viper"
	"goto-bangumi/internal/model"
)

var GlobalConfig *model.Config

// setDefaultValues sets default values from Python's DEFAULT_SETTINGS
// TODO: 一些检查和初使化的工作应该放到 init 函数中完成
// WARN: 想了想还是不太行, 测试的时候我又不想 init
func setDefaultValues() {
	// Program defaults
	viper.SetDefault("program.rss_time", 7200)
	viper.SetDefault("program.webui_port", 7892)

	// Downloader defaults
	viper.SetDefault("downloader.type", "qbittorrent")
	viper.SetDefault("downloader.host", "127.0.0.1:8080")
	viper.SetDefault("downloader.username", "admin")
	viper.SetDefault("downloader.password", "adminadmin")
	viper.SetDefault("downloader.path", "/downloads/Bangumi")
	viper.SetDefault("downloader.ssl", false)

	// RSS Parse defaults
	viper.SetDefault("rss_parser.enable", true)
	viper.SetDefault("rss_parser.filter", []string{"720", "\\d+-\\d+"})
	viper.SetDefault("rss_parser.include", []string{})
	viper.SetDefault("rss_parser.language", "zh")
	viper.SetDefault("rss_parser.mikan_custom_url", "mikanani.me")
	viper.SetDefault("rss_parser.tmdb_api_key", "")

	// Bangumi Manage defaults
	viper.SetDefault("bangumi_manage.enable", true)
	viper.SetDefault("bangumi_manage.eps_complete", false)
	viper.SetDefault("bangumi_manage.rename_method", "pn")
	viper.SetDefault("bangumi_manage.group_tag", false)
	viper.SetDefault("bangumi_manage.remove_bad_torrent", false)

	// Log defaults
	viper.SetDefault("log.debug_enable", false)

	// Proxy defaults
	viper.SetDefault("proxy.enable", false)
	viper.SetDefault("proxy.type", "http")
	viper.SetDefault("proxy.host", "")
	viper.SetDefault("proxy.port", 1080)
	viper.SetDefault("proxy.username", "")
	viper.SetDefault("proxy.password", "")

	// Notification defaults
	viper.SetDefault("notification.enable", false)
	viper.SetDefault("notification.type", "telegram")
	viper.SetDefault("notification.token", "")
	viper.SetDefault("notification.chat_id", "")

	// password
	viper.SetDefault("password", "adminadmin")
}

// LoadConfig 读取配置文件
//TODO: 要在 conf 里面初始化 network 包,等其他的包
func LoadConfig(configPath string) (*model.Config, error) {
	// Set default values first
	setDefaultValues()

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config model.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	GlobalConfig = &config
	return &config, nil
}
