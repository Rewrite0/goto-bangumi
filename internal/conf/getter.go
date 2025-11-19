// Package conf 提供配置获取辅助函数
package conf

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// GetConfigByKey 泛型函数：根据key从viper获取对应类型的配置
// 这是插件化系统的核心配置获取方法，支持任意配置类型
//
// 使用示例:
//
//	// 内置模块使用
//	cfg, err := conf.GetConfigByKey[model.DownloaderConfig]("downloader")
//	cfg, err := conf.GetConfigByKey[model.RssParserConfig]("rss_parser")
//
//	// 插件使用（插件自定义配置结构体）
//	type MyPluginConfig struct {
//	    ApiKey string `json:"api_key"`
//	    Timeout int   `json:"timeout"`
//	}
//	pluginCfg, err := conf.GetConfigByKey[MyPluginConfig]("plugins.my_plugin")
func GetConfigByKey[T any](key string) (*T, error) {
	var config T
	if err := viper.UnmarshalKey(key, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetConfigOrDefault 获取配置，失败时返回默认值
// 当配置不存在时，将默认值设置到 viper 中
// 当配置存在但部分字段缺失时，用默认值补全缺失字段
// 最后通过 SaveConfig() 一次性写回配置文件
func GetConfigOrDefault[T any](key string, defaultValue *T) *T {
	// 为该 key 的所有字段设置默认值
	// 这样 viper 在解析时会自动用默认值填充缺失的字段
	setDefaultsForKey(key, defaultValue)

	// 尝试解析配置（会自动使用默认值填充缺失字段）
	config := *defaultValue // start with defaults so missing fields stay populated
	if err := viper.UnmarshalKey(key, &config); err != nil {
		// 解析失败，使用完整的默认值
		viper.Set(key, defaultValue)
		NeedUpdate = true
		slog.Error("配置解析失败，使用默认值", "key", key, "error", err)
		return defaultValue
	}

	// 将合并后的配置设置回 viper
	// viper.Set(key, config)

	return &config
}

// setDefaultsForKey 为指定 key 的所有字段设置默认值
// 使用反射遍历结构体字段，为每个字段在 viper 中设置默认值
func setDefaultsForKey[T any](key string, defaultValue *T) {
	v := reflect.ValueOf(defaultValue)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 获取 json tag 作为配置键名
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// 处理 json tag 中的选项（如 omitempty）
		tagParts := strings.Split(jsonTag, ",")
		fieldName := tagParts[0]

		// 构造完整的 key 路径
		fullKey := key + "." + fieldName
		// 判断 fullKey 是否在 viper 中已存在
		if !viper.IsSet(fullKey) {
			NeedUpdate = true
			viper.Set(fullKey, fieldValue.Interface())
			slog.Debug("配置缺失，设置默认值", "key", fullKey, "value", fieldValue.Interface())
		}

		// 设置默认值
		// viper.SetDefault(fullKey, fieldValue.Interface())
	}
}
