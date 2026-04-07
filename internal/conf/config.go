// Package conf provides configuration loading and persistence.
package conf

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"

	"goto-bangumi/internal/model"
)

var (
	cfg        *model.Config
	configDir  = "./config"
	configFile = "config.yaml"
)

// Init loads config from YAML file, fills defaults from env-default tags
// and environment variables, then writes the complete config back to YAML.
func Init() error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, configFile)

	// Create empty YAML file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte("{}"), 0o644); err != nil {
			return err
		}
		slog.Info("[conf] 配置文件创建成功", "path", configPath)
	}

	cfg = &model.Config{}
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		return err
	}

	// Write complete config back (backfill defaults)
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

// save marshals the config to YAML and writes it to the config file.
func save(c *model.Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, configFile)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return err
	}
	slog.Info("[conf] 配置文件已更新", "path", configPath)
	return nil
}
