// Package conf 提供配置获取辅助函数
package conf

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// validate 全局验证器实例
var validate = validator.New()

// GetConfigOrDefault 获取配置，失败时返回默认值
// 当配置不存在时，将默认值设置到 viper 中
// 当配置存在但部分字段缺失时，用默认值补全缺失字段
// 当配置值不符合验证规则时，用默认值替换不合法的字段
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
		SetConfigValue(key, *defaultValue)
		return defaultValue
	}

	// 验证配置值
	if err := validate.Struct(config); err != nil {
		// 验证失败，修正不合法的字段
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fixInvalidFields(key, &config, defaultValue, validationErrors)
		}
	}

	// 存储配置值，用于按顺序保存
	SetConfigValue(key, config)

	return &config
}

// fixInvalidFields 修正不合法的配置字段，用默认值替换
func fixInvalidFields[T any](key string, config *T, defaultValue *T, errors validator.ValidationErrors) {
	configVal := reflect.ValueOf(config).Elem()
	defaultVal := reflect.ValueOf(defaultValue).Elem()
	configType := configVal.Type()

	for _, err := range errors {
		fieldName := err.StructField()

		// 找到对应的字段
		for i := 0; i < configType.NumField(); i++ {
			field := configType.Field(i)
			if field.Name == fieldName {
				// 获取 json tag 用于日志
				jsonTag := field.Tag.Get("json")
				tagParts := strings.Split(jsonTag, ",")
				jsonName := tagParts[0]

				// 用默认值替换不合法的值
				invalidValue := configVal.Field(i).Interface()
				defaultFieldValue := defaultVal.Field(i)
				configVal.Field(i).Set(defaultFieldValue)

				// 更新 viper 中的值
				fullKey := key + "." + jsonName
				viper.Set(fullKey, defaultFieldValue.Interface())
				NeedUpdate = true

				slog.Warn("配置验证失败，使用默认值",
					"key", fullKey,
					"invalid_value", invalidValue,
					"default_value", defaultFieldValue.Interface(),
					"rule", err.Tag(),
				)
				break
			}
		}
	}
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
			slog.Debug("[config] 配置缺失，设置默认值", "key", fullKey, "value", fieldValue.Interface())
		}

		// 设置默认值
		// viper.SetDefault(fullKey, fieldValue.Interface())
	}
}
