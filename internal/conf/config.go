// Package conf 提供了基础的配置加载功能
package conf

import (
	"encoding/json"
	// "goto-bangumi/internal/model"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Module interface {
	Init(config any) error
	DefaultConfig() any
	ConfigKey() string
}

// var GlobalConfig *model.Config

var (
	configDir  = "./config"
	configName = "config.json"
	NeedUpdate = false
)

func init() {
	// 检测配置文件夹是否存在, 不存在则创建
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			slog.Error("[conf] 创建配置文件夹失败", "error", err)
			return
		}
		slog.Info("[conf] 配置文件夹创建成功", "path", configDir)
	}
	configPath := filepath.Join(configDir, configName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建空的 JSON 配置文件
		if err := os.WriteFile(configPath, []byte("{}"), 0o644); err != nil {
			slog.Error("[conf] 创建配置文件失败", "error", err)
			return
		}
		slog.Info("[conf] 配置文件创建成功", "path", configPath)
	}
}

func Init() {
	// 加载配置文件
	// 要是有一些项没有的话, 要把文件反写回去
}

// LoadConfig 读取配置文件
// 只负责读取配置，不设置默认值
// 默认值由各模块通过 GetConfigOrDefault 自行处理
func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// SaveConfig 将 viper 中的配置写回到配置文件
// 在所有模块通过 GetConfigOrDefault 加载配置后调用
// 自动补全缺失的配置字段
func SaveConfig() error {
	configPath := filepath.Join(configDir, configName)

	// 获取 viper 中的所有配置
	allSettings := viper.AllSettings()

	// 序列化为 JSON，使用缩进格式便于阅读
	data, err := json.MarshalIndent(allSettings, "", "  ")
	if err != nil {
		return err
	}

	// 写回文件
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return err
	}
	NeedUpdate = false
	slog.Info("[conf] 配置文件已更新", "path", configPath)
	return nil
}
