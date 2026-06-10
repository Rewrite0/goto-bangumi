package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goto-bangumi/internal/model"
)

func useTempConfig(t *testing.T) string {
	t.Helper()

	originalCfg := cfg
	originalDir := configDir
	originalFile := configFile
	originalPath := configPath

	configDir = t.TempDir()
	configFile = "config.toml"
	configPath = filepath.Join(configDir, configFile)
	cfg = nil

	t.Cleanup(func() {
		cfg = originalCfg
		configDir = originalDir
		configFile = originalFile
		configPath = originalPath
	})

	return configPath
}

func TestInitCreatesTomlConfig(t *testing.T) {
	path := useTempConfig(t)

	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if Get() == nil {
		t.Fatal("Get() returned nil after Init()")
	}

	content := readConfigFile(t, path)
	if !strings.Contains(content, "[program]") {
		t.Fatalf("generated config is not TOML: %s", content)
	}
	if !strings.Contains(content, "rss_time = ") {
		t.Fatalf("generated config did not backfill defaults: %s", content)
	}
}

func TestInitReadsTomlConfig(t *testing.T) {
	path := useTempConfig(t)

	data := []byte("[parser]\nfilter = [\"720p\"]\ninclude = [\"1080p\"]\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	got := Get()
	if got == nil {
		t.Fatal("Get() returned nil after Init()")
	}
	if len(got.Parser.Filter) != 1 || got.Parser.Filter[0] != "720p" {
		t.Fatalf("Parser.Filter = %#v, want [720p]", got.Parser.Filter)
	}
	if len(got.Parser.Include) != 1 || got.Parser.Include[0] != "1080p" {
		t.Fatalf("Parser.Include = %#v, want [1080p]", got.Parser.Include)
	}
}

func TestUpdateWritesTomlConfig(t *testing.T) {
	path := useTempConfig(t)

	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if err := Update(func(c *model.Config) {
		c.Program.DebugEnable = true
		c.Notification.ChatID = "12345"
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	content := readConfigFile(t, path)
	if !strings.Contains(content, "debug_enable = true") {
		t.Fatalf("updated config missing debug_enable: %s", content)
	}
	if !strings.Contains(content, "chat_id = \"12345\"") {
		t.Fatalf("updated config missing chat_id: %s", content)
	}
}

func readConfigFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(data)
}
