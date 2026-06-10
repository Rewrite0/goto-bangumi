// Package conf provides configuration loading and persistence.
package conf

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ilyakaznacheev/cleanenv"

	"goto-bangumi/internal/model"
)

var (
	cfg        *model.Config
	configDir  = "./config"
	configFile = "config.toml"
	configPath = filepath.Join(configDir, configFile)
)

// Init loads config from TOML file, fills defaults from env-default tags
// and environment variables, then writes the complete config back to TOML.
func Init() error {
	loaded := &model.Config{}
	if err := cleanenv.ReadConfig(configPath, loaded); err != nil {
		// 判断是否是文件不存在错误，如果是则创建一个空的配置文件
		if !os.IsNotExist(err) {
			return err
		}
		// Ensure config directory exists
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(configPath, nil, 0o644); err != nil {
			return err
		}
		slog.Info("[conf] 配置文件创建成功", "path", configPath)
		if err := cleanenv.ReadConfig(configPath, loaded); err != nil {
			slog.Warn("[conf] 配置文件读取失败", "error", err)
			return err
		}
	}

	// Write complete config back (backfill defaults)
	cfg = loaded
	return save(cfg)
}

// Get returns the global config.
func Get() *model.Config {
	return cfg
}

// Update applies a mutation to the config and persists it.
func Update(fn func(*model.Config)) error {
	fn(cfg)
	return save(cfg)
}

// save marshals the config to TOML and writes it to the config file.
func save(c *model.Config) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, buf.Bytes(), 0o644); err != nil {
		return err
	}
	slog.Info("[conf] 配置文件已更新", "path", configPath)
	return nil
}
