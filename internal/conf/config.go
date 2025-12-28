// Package conf 提供了基础的配置加载功能
package conf

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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

	// configOrder 记录配置的注册顺序
	configOrder []string
	// configValues 存储配置值，用于按顺序保存
	configValues = make(map[string]any)
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

// SetConfigValue 存储配置值，用于 SaveConfig 时按顺序保存
// 由 GetConfigOrDefault 自动调用
func SetConfigValue(key string, value any) {
	// 记录顺序（首次注册时）
	if _, exists := configValues[key]; !exists {
		configOrder = append(configOrder, key)
	}
	configValues[key] = value
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

// SaveConfig 将配置写回到配置文件
// 按照注册顺序和结构体字段顺序保存
func SaveConfig() error {
	configPath := filepath.Join(configDir, configName)

	var parts []string
	processed := make(map[string]bool)

	// 1. 按注册顺序处理已注册的配置
	for _, key := range configOrder {
		value, ok := configValues[key]
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("  %q: %s", key, formatValue(value)))
		processed[key] = true
	}

	// 2. 处理未注册的配置（如 plugin）
	for key, value := range viper.AllSettings() {
		if processed[key] {
			continue
		}
		parts = append(parts, fmt.Sprintf("  %q: %s", key, formatValue(value)))
	}

	result := "{\n" + strings.Join(parts, ",\n") + "\n}"

	if err := os.WriteFile(configPath, []byte(result), 0o644); err != nil {
		return err
	}
	NeedUpdate = false
	slog.Info("[conf] 配置文件已更新", "path", configPath)
	return nil
}

// formatValue 格式化值为带正确缩进的 JSON
func formatValue(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	// 给非首行添加缩进
	lines := strings.Split(string(data), "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = "  " + lines[i]
	}
	return strings.Join(lines, "\n")
}
