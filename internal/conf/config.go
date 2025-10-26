// Package conf 提供了基础的配置加载功能
package conf

import (
	"goto-bangumi/internal/model"

	"github.com/spf13/viper"
)

var GlobalConfig *model.Config



// LoadConfig 读取配置文件
// TODO: 要在 conf 里面初始化 network 包,等其他的包
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
