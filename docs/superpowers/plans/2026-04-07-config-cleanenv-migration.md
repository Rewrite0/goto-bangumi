# Config Migration (Viper → cleanenv) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Viper with cleanenv, use YAML config format, centralize config structs in model/, and inject configs into modules via dependency injection.

**Architecture:** Single top-level `model.Config` struct with nested sub-configs, loaded once by `conf.Init()`. Modules receive their sub-config as a parameter instead of pulling from `conf`. `conf` package exposes `Init()`, `Get()`, and `Update()`.

**Tech Stack:** `github.com/ilyakaznacheev/cleanenv`, `gopkg.in/yaml.v3`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/model/config.go` | Modify | Top-level `Config` struct + all sub-config structs with `yaml`/`env`/`env-default` tags |
| `internal/model/network.go` | Modify | `ProxyConfig` tag changes |
| `internal/conf/config.go` | Rewrite | `Init()`, `Get()`, `Update()` using cleanenv + yaml.v3 |
| `internal/conf/getter.go` | Delete | No longer needed |
| `internal/core/program.go` | Modify | Simplified init + dependency injection |
| `internal/task/refresh.go` | Modify | Accept config as parameter |
| `internal/download/client.go` | Modify | Remove `conf` import, remove `GetConfig()`, remove `InitModule()` |
| `internal/parser/raw.go` | Modify | Remove `conf` import, remove `InitModule()` |
| `internal/notification/notification_client.go` | Modify | Remove `conf` import, remove `InitModule()` |
| `internal/network/request_url.go` | Modify | Remove `conf` import, remove `GetConfig()` |
| `internal/rename/rename.go` | Modify | Remove `conf` import, remove `InitModule()` |
| `go.mod` | Modify | Add cleanenv, remove viper + validator |

---

### Task 1: Add cleanenv dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add cleanenv dependency**

Run:
```bash
cd C:/Users/19457/github/goto-bangumi && go get github.com/ilyakaznacheev/cleanenv
```

- [ ] **Step 2: Verify dependency was added**

Run:
```bash
grep cleanenv go.mod
```

Expected: `github.com/ilyakaznacheev/cleanenv vX.X.X`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add cleanenv dependency"
```

---

### Task 2: Update model config structs with YAML + env tags

**Files:**
- Modify: `internal/model/config.go`
- Modify: `internal/model/network.go`

- [ ] **Step 1: Rewrite `internal/model/config.go`**

Replace the entire file content with:

```go
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
```

Note: `Filter` and `Include` in `RssParserConfig` are `[]string` — cleanenv does not support `env-default` for slices easily, so they get their zero value (`nil`). The backfill step in `conf.Init()` will handle writing defaults to YAML.

- [ ] **Step 2: Update `internal/model/network.go` — change `ProxyConfig` tags**

Replace the `ProxyConfig` struct and delete `NewProxyConfig()`:

```go
// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	Enable   bool   `yaml:"enable" env:"ENABLE" env-default:"false"`
	Type     string `yaml:"type" env:"TYPE" env-default:"http"`
	Host     string `yaml:"host" env:"HOST"`
	Port     int    `yaml:"port" env:"PORT" env-default:"0"`
	Username string `yaml:"username" env:"USERNAME"`
	Password string `yaml:"password" env:"PASSWORD"`
}
```

Remove `NewProxyConfig()` function. Keep the `RSSXml`, `RSSTorrent`, and `Enclosure` structs unchanged.

- [ ] **Step 3: Verify it compiles (it won't fully compile yet, but check for syntax)**

Run:
```bash
cd C:/Users/19457/github/goto-bangumi && go vet ./internal/model/...
```

Expected: PASS (model package has no dependency on conf)

- [ ] **Step 4: Commit**

```bash
git add internal/model/config.go internal/model/network.go
git commit -m "refactor: update config structs with yaml/env tags, add top-level Config"
```

---

### Task 3: Rewrite conf package

**Files:**
- Rewrite: `internal/conf/config.go`
- Delete: `internal/conf/getter.go`

- [ ] **Step 1: Rewrite `internal/conf/config.go`**

Replace the entire file content with:

```go
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
	cfg       *model.Config
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
```

- [ ] **Step 2: Delete `internal/conf/getter.go`**

```bash
rm C:/Users/19457/github/goto-bangumi/internal/conf/getter.go
```

- [ ] **Step 3: Verify conf package compiles**

Run:
```bash
cd C:/Users/19457/github/goto-bangumi && go vet ./internal/conf/...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/conf/config.go
git rm internal/conf/getter.go
git commit -m "refactor: rewrite conf package to use cleanenv"
```

---

### Task 4: Update modules to remove conf dependency

**Files:**
- Modify: `internal/download/client.go`
- Modify: `internal/parser/raw.go`
- Modify: `internal/notification/notification_client.go`
- Modify: `internal/network/request_url.go`
- Modify: `internal/rename/rename.go`
- Modify: `internal/task/refresh.go`

- [ ] **Step 1: Update `internal/download/client.go`**

Remove the `conf` import. Delete `GetConfig()` method and `InitModule()` function. The `Init` method on `DownloadClient` already accepts `*model.DownloaderConfig` — no change needed there.

Remove these lines:
```go
// In import block:
"goto-bangumi/internal/conf"

// Delete GetConfig method (lines 253-255):
func (c *DownloadClient) GetConfig() *model.DownloaderConfig {
	return conf.GetConfigOrDefault("downloader", model.NewDownloaderConfig())
}

// Delete InitModule function (lines 257-261):
func InitModule() {
	c := conf.GetConfigOrDefault("downloader", model.NewDownloaderConfig())
	Client.Init(c)
	slog.Debug("[download] 下载客户端初始化完成", "类型", c.Type, "保存路径", c.SavePath)
}
```

- [ ] **Step 2: Update `internal/parser/raw.go`**

Remove `conf` import. Delete `InitModule()`. Keep `Init()` and `init()` as-is. Update `init()` to not use `NewRssParserConfig()` — use a zero-value struct pointer instead:

Replace the file content with:
```go
package parser

import (
	"strings"

	"goto-bangumi/internal/model"
)

var ParserConfig *model.RssParserConfig

func init() {
	ParserConfig = &model.RssParserConfig{}
}

func Init(config *model.RssParserConfig) {
	if config != nil {
		ParserConfig = config
	}
	InitTmdb(config.TmdbAPIKey)
}

type RawParse struct{}

func (p *RawParse) Parse(title string) *model.Bangumi {
	metaParser := NewTitleMetaParse()
	episode := metaParser.Parse(title)
	if episode.Episode == -1 {
		return nil
	}
	var officialTitle string
	season := episode.Season
	return &model.Bangumi{
		OfficialTitle: officialTitle,
		Year:          episode.Year,
		Season:        season,
		EpsCollect:    false,
		Offset:        0,
		IncludeFilter: strings.Join(ParserConfig.Include, ","),
		ExcludeFilter: strings.Join(ParserConfig.Filter, ","),
		Parse:         "raw",
		RSSLink:       "",
		PosterLink:    "",
		Deleted:       false,
	}
}
```

- [ ] **Step 3: Update `internal/notification/notification_client.go`**

Remove `conf` import. Delete `InitModule()`:

Remove these lines:
```go
// In import block:
"goto-bangumi/internal/conf"

// Delete InitModule (lines 64-68):
func InitModule() {
	c :=conf.GetConfigOrDefault("notification", model.NewNotificationConfig())
	NotificationClient.Init(c)
}
```

- [ ] **Step 4: Update `internal/network/request_url.go`**

Remove `conf` import. Delete `GetConfig()`. Update `init()` to not use `NewProxyConfig()`:

Remove:
```go
// In import block:
"goto-bangumi/internal/conf"

// Delete GetConfig (lines 48-50):
func GetConfig() *model.ProxyConfig {
	return conf.GetConfigOrDefault("proxy", model.NewProxyConfig())
}
```

Update `init()`:
```go
func init() {
	globalCache = NewMemoryCacheManager(500, 60*time.Second)
	defaultProxyConfig = &model.ProxyConfig{}
	defaultClient = newRequestClient()
}
```

- [ ] **Step 5: Update `internal/rename/rename.go`**

Remove `conf` import. Delete `InitModule()`. Update the package-level default:

Replace:
```go
var renameConfig = model.NewBangumiRenameConfig()
```
With:
```go
var renameConfig = &model.BangumiRenameConfig{}
```

Remove:
```go
// In import block:
"goto-bangumi/internal/conf"

// Delete InitModule (lines 83-87):
func InitModule() {
	c := conf.GetConfigOrDefault("rename", model.NewBangumiRenameConfig())
	Init(c)
	slog.Debug("[rename] 重命名模块初始化完成", "配置", c)
}
```

- [ ] **Step 6: Update `internal/task/refresh.go`**

Remove `conf` import. Change `NewRSSRefreshTask()` to accept `model.ProgramConfig` as a parameter:

Replace:
```go
func NewRSSRefreshTask() *RSSRefreshTask {
	programConfig := conf.GetConfigOrDefault("program", model.NewProgramConfig())
	interval := programConfig.RssTime
```
With:
```go
func NewRSSRefreshTask(programConfig model.ProgramConfig) *RSSRefreshTask {
	interval := programConfig.RssTime
```

Remove `conf` from imports. Remove `model.NewProgramConfig` usage (no longer needed since there's no `NewProgramConfig` function).

Updated imports:
```go
import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/refresh"
)
```

- [ ] **Step 7: Verify all modified modules compile**

Run:
```bash
cd C:/Users/19457/github/goto-bangumi && go vet ./internal/download/... ./internal/parser/... ./internal/notification/... ./internal/network/... ./internal/rename/... ./internal/task/...
```

This will likely fail because `core/program.go` still references old APIs. That's expected — we fix it in the next task.

- [ ] **Step 8: Commit**

```bash
git add internal/download/client.go internal/parser/raw.go internal/notification/notification_client.go internal/network/request_url.go internal/rename/rename.go internal/task/refresh.go
git commit -m "refactor: remove conf dependency from modules, use dependency injection"
```

---

### Task 5: Update core/program.go to inject configs

**Files:**
- Modify: `internal/core/program.go`

- [ ] **Step 1: Rewrite `InitProgram` in `internal/core/program.go`**

Replace the `InitProgram` function with:

```go
func InitProgram(ctx context.Context) {
	// Load config
	if err := conf.Init(); err != nil {
		slog.Error("[program] 加载配置文件失败", "error", err)
		panic(err)
	}

	cfg := conf.Get()

	// Initialize logger
	logger.Init(cfg.Program.DebugEnable)

	// Initialize database
	if err := database.InitDB(nil); err != nil {
		slog.Error("[program] 初始化数据库失败", "error", err)
		panic(err)
	}

	// Initialize modules with injected config
	network.Init(&cfg.Proxy)
	parser.Init(&cfg.Parser)
	notification.NotificationClient.Init(&cfg.Notification)
	download.Client.Init(&cfg.Downloader)
	rename.Init(&cfg.Rename)
}
```

Update imports — remove unused imports if any. The import list should be:

```go
import (
	"context"
	"log/slog"

	"goto-bangumi/internal/conf"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/logger"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/notification"
	"goto-bangumi/internal/parser"
	"goto-bangumi/internal/rename"
	"goto-bangumi/internal/scheduler"
	"goto-bangumi/internal/task"
	"goto-bangumi/internal/taskrunner"
	"goto-bangumi/internal/taskrunner/handlers"
)
```

Remove unused imports: `"fmt"`, `"goto-bangumi/internal/model"`.

- [ ] **Step 2: Update `InitScheduler` to pass config**

In `InitScheduler`, update the call to `NewRSSRefreshTask`:

Replace:
```go
s.AddTask(task.NewRSSRefreshTask())
```
With:
```go
s.AddTask(task.NewRSSRefreshTask(conf.Get().Program))
```

- [ ] **Step 3: Verify full project compiles**

Run:
```bash
cd C:/Users/19457/github/goto-bangumi && go build ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/core/program.go
git commit -m "refactor: update program init to inject configs into modules"
```

---

### Task 6: Remove viper and validator dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Remove viper from go.mod**

Run:
```bash
cd C:/Users/19457/github/goto-bangumi && go mod tidy
```

This will automatically remove `github.com/spf13/viper` and `github.com/go-playground/validator/v10` (and their indirect deps) since nothing imports them anymore.

- [ ] **Step 2: Verify viper is removed**

Run:
```bash
grep -E "spf13/viper|go-playground/validator" go.mod
```

Expected: No output (both removed).

- [ ] **Step 3: Verify cleanenv is present**

Run:
```bash
grep cleanenv go.mod
```

Expected: `github.com/ilyakaznacheev/cleanenv vX.X.X`

- [ ] **Step 4: Final build check**

Run:
```bash
cd C:/Users/19457/github/goto-bangumi && go build ./...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: remove viper and validator deps, keep cleanenv"
```

---

### Task 7: Manual smoke test

- [ ] **Step 1: Delete old config file if it exists**

```bash
rm -f C:/Users/19457/github/goto-bangumi/config/config.json
rm -f C:/Users/19457/github/goto-bangumi/config/config.yaml
```

- [ ] **Step 2: Run the application briefly**

```bash
cd C:/Users/19457/github/goto-bangumi && go run . &
sleep 3
kill %1 2>/dev/null
```

- [ ] **Step 3: Verify config.yaml was created with defaults**

```bash
cat C:/Users/19457/github/goto-bangumi/config/config.yaml
```

Expected: A complete YAML file with all default values populated:
```yaml
program:
  rss_time: 600
  webui_port: 7892
  password: adminadmin
  debug_enable: false
downloader:
  type: qbittorrent
  path: /downloads/Bangumi
  host: 127.0.0.1:8080
  ssl: false
  username: admin
  password: adminadmin
parser:
  enable: true
  filter: []
  include: []
  language: zh
  mikan_custom_url: mikanani.me
  tmdb_api_key: ""
rename:
  enable: true
  eps_complete: false
  rename_method: advanced
  year: false
  group: false
notification:
  enable: false
  type: telegram
  token: ""
  chat_id: ""
proxy:
  enable: false
  type: http
  host: ""
  port: 0
  username: ""
  password: ""
```

- [ ] **Step 4: Verify environment variable override works**

```bash
PROGRAM_RSS_TIME=1200 go run . &
sleep 3
kill %1 2>/dev/null
cat C:/Users/19457/github/goto-bangumi/config/config.yaml | grep rss_time
```

Expected: `rss_time: 1200` (if the YAML file was empty or had no `rss_time` field; if it already had `rss_time: 600`, the YAML value takes precedence).
