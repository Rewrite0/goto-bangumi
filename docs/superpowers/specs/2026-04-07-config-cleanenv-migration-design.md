# Config Migration: Viper → cleanenv + YAML

## Goal

Replace Viper with cleanenv, centralize all config structs in `model/`, use YAML format, and inject configs into modules instead of modules pulling from `conf`.

## Config Structure

### Top-level struct (`internal/model/config.go`)

```go
type Config struct {
    Program      ProgramConfig       `yaml:"program" env-prefix:"PROGRAM_"`
    Downloader   DownloaderConfig    `yaml:"downloader" env-prefix:"DOWNLOADER_"`
    Parser       RssParserConfig     `yaml:"parser" env-prefix:"PARSER_"`
    Rename       BangumiRenameConfig `yaml:"rename" env-prefix:"RENAME_"`
    Notification NotificationConfig  `yaml:"notification" env-prefix:"NOTIFICATION_"`
    Proxy        ProxyConfig         `yaml:"proxy" env-prefix:"PROXY_"`
}
```

### Sub-struct tag format

All sub-structs use `yaml` + `env` + `env-default` tags. Example:

```go
type ProgramConfig struct {
    RssTime     int    `yaml:"rss_time" env:"RSS_TIME" env-default:"600"`
    WebuiPort   int    `yaml:"webui_port" env:"WEBUI_PORT" env-default:"7892"`
    Password    string `yaml:"password" env:"PASSWORD" env-default:"adminadmin"`
    DebugEnable bool   `yaml:"debug_enable" env:"DEBUG_ENABLE" env-default:"false"`
}
```

Default values are declared in `env-default` tags. `NewXxxConfig()` factory functions are deleted.

### Config priority

YAML file value > environment variable > `env-default` tag default.

If a field is missing from the YAML file, cleanenv fills it from environment variable. If the environment variable is also unset, the `env-default` value is used.

## conf Package (`internal/conf/config.go`)

Simplified to three functions:

```go
var cfg *model.Config

func Init() error          // Load config.yaml via cleanenv, backfill defaults, write back
func Get() *model.Config   // Return global config (read-only)
func Update(fn func(*model.Config)) error  // Mutate config and persist to YAML
```

### Init behavior

1. Ensure `./config/` directory exists.
2. If `./config/config.yaml` does not exist, create it with `{}`.
3. Call `cleanenv.ReadConfig("./config/config.yaml", cfg)` — this fills defaults from `env-default` tags and environment variables.
4. Write the complete config back to `config.yaml` using `yaml.v3.Marshal` (backfill/auto-complete).

### Update behavior

Accept a mutation function, apply it to `cfg`, then marshal and write back to `config.yaml`.

### Deleted

- `internal/conf/getter.go` — entire file
- All global state: `configOrder`, `configValues`, `NeedUpdate`
- All reflection-based default/validation logic

## Dependency Injection

Modules do NOT import `conf`. The startup entry point (`internal/core/program.go`) loads config once and passes sub-configs to each module:

```go
cfg := conf.Get()
download.Init(cfg.Downloader)
parser.Init(cfg.Parser)
rename.Init(cfg.Rename)
notification.Init(cfg.Notification)
network.Init(cfg.Proxy)
```

Each module's `Init` function (or equivalent) accepts its own config struct type directly.

## Files Changed

### Deleted
- `internal/conf/getter.go`

### Rewritten
- `internal/conf/config.go` — viper → cleanenv, three-function API

### Modified (tag changes)
- `internal/model/config.go` — add `Config` struct, change sub-struct tags to `yaml`+`env`+`env-default`, delete `NewXxxConfig()` functions
- `internal/model/network.go` — `ProxyConfig` tag changes

### Modified (injection pattern)
- `internal/core/program.go` — simplify to `conf.Init()` + inject sub-configs
- `internal/task/refresh.go` — receive config as parameter
- `internal/download/client.go` — receive config as parameter
- `internal/rename/rename.go` — receive config as parameter
- `internal/parser/raw.go` — receive config as parameter
- `internal/notification/notification_client.go` — receive config as parameter
- `internal/network/request_url.go` — receive config as parameter

### Dependency changes (`go.mod`)
- Add: `github.com/ilyakaznacheev/cleanenv`
- Remove: `github.com/spf13/viper`, `github.com/go-playground/validator/v10`

## Config File

Format: YAML at `./config/config.yaml`.

No migration from old JSON file. Clean start.

## API routes

`api/routes/config.go` remains TODO. Future implementation will use `conf.Get()` and `conf.Update()`.
