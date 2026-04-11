# Download Client 依赖注入重构设计

**日期**: 2026-04-11
**状态**: 待实现

## 背景

当前 `internal/download/client.go` 采用全局单例模式（`var Client = &DownloadClient{}`）。所有消费者直接引用 `download.Client`，存在与 database 旧模式相同的问题：

1. **隐式依赖**：消费者（handlers、rename）直接引用包级全局变量，依赖不可见
2. **测试污染**：测试通过直接赋值 `download.Client.Downloader = mock` 修改全局状态，测试之间互相影响
3. **无法 `t.Parallel()`**：所有测试共享同一个全局实例

**特殊点**：download client 支持运行时热更新——用户可通过配置切换 downloader 类型（如 qbittorrent → 其他），此时调用 `Init()` 原地替换内部 `Downloader` 字段。这不影响 DI 可行性，因为所有持有者共享同一个 `*DownloadClient` 指针，`Init()` 修改的是 struct 内部字段。

## 目标

- 删除 `download` 包的全局单例 `Client`
- 通过构造函数显式注入 `*download.DownloadClient` 到消费者
- 测试可以各自构造独立实例，互相隔离

## 非目标

- **不**引入接口抽象层（消费者数量少，直接注入具体类型即可）
- **不**改造 `notification.NotificationClient`（思路不同，不在本次范围）
- **不**引入 DI 框架
- **不**做无关重构

## 设计

### 1. `download` 包改造

新增工厂函数，删除全局变量：

```go
// 新增
func NewDownloadClient() *DownloadClient {
    return &DownloadClient{}
}

// 删除
// var Client = &DownloadClient{}
```

`Init()` 方法保持不变，依然原地修改内部字段（`c.Downloader = dl`、`c.SavePath`、`c.limiter` 等）。

### 2. `core/Program` 持有实例

`Program` struct 新增 `downloader` 字段，负责创建和生命周期管理：

```go
type Program struct {
    db         *database.DB
    downloader *download.DownloadClient
    // ...
}
```

在 `InitProgram()` 中：

```go
p.downloader = download.NewDownloadClient()
p.downloader.Init(&cfg.Downloader)
p.downloader.Login()
```

热更新时直接调用 `p.downloader.Init(&newCfg.Downloader)`，所有已注入的消费者自动看到新的 downloader。

### 3. 注入到消费者

4 个消费者需要改造：

#### `handlers.NewAddHandler`

```go
type AddHandler struct {
    db         *database.DB
    downloader *download.DownloadClient
}

func NewAddHandler(db *database.DB, dl *download.DownloadClient) *AddHandler
```

内部 `download.Client.AddTorrent(...)` 改为 `h.downloader.AddTorrent(...)`。

#### `handlers.NewCheckHandler`

```go
type CheckHandler struct {
    db         *database.DB
    downloader *download.DownloadClient
}

func NewCheckHandler(db *database.DB, dl *download.DownloadClient) *CheckHandler
```

#### `handlers.NewDownloadingHandler`

```go
type DownloadingHandler struct {
    db         *database.DB
    downloader *download.DownloadClient
}

func NewDownloadingHandler(db *database.DB, dl *download.DownloadClient) *DownloadingHandler
```

#### `rename` 模块

`Renamer` struct 新增 `downloader` 字段：

```go
type Renamer struct {
    db         *database.DB
    downloader *download.DownloadClient
}

func New(db *database.DB, dl *download.DownloadClient) *Renamer
```

内部所有 `download.Client.XXX()` 改为 `r.downloader.XXX()`。

### 4. `core/Program` 组装

```go
// 在 InitProgram 或 NewProgram 中
renamer := rename.New(db, p.downloader)

tr.Register(handlers.NewAddHandler(db, p.downloader))
tr.Register(handlers.NewCheckHandler(db, p.downloader))
tr.Register(handlers.NewDownloadingHandler(db, p.downloader))
tr.Register(handlers.NewRenameHandler(db, p.downloader, renamer))
```

### 5. Login 与 EnsureLogin

无需改造。`Login()` 和 `EnsureLogin()` 都是 `*DownloadClient` 的方法，DI 后行为完全不变——所有消费者通过持有的同一个实例调用，状态（`logined`、`LoginError`、`singleflight`）自动共享。

### 6. 测试

每个测试构造独立实例，不再污染全局状态：

```go
func TestSomething(t *testing.T) {
    t.Parallel()

    client := download.NewDownloadClient()
    client.Downloader = mockDownloader
    client.SavePath = "/tmp/test"

    handler := handlers.NewAddHandler(db, client)
    // 测试...
}
```

## 影响面

需要改动的文件：

- `internal/download/client.go` — 新增 `NewDownloadClient()`，删除 `var Client`
- `internal/core/program.go` — 新增 `downloader` 字段，改造初始化和组装逻辑
- `internal/taskrunner/handlers/add.go` — 注入 downloader
- `internal/taskrunner/handlers/check.go` — 注入 downloader
- `internal/taskrunner/handlers/downloading.go` — 注入 downloader
- `internal/rename/rename.go` — `Renamer` 新增 downloader 字段
- `internal/rename/utils.go` — 如有直接引用 `download.Client` 则改为方法
- `internal/rename/utils_test.go` — 改为构造独立实例

## 验收标准

- [ ] `grep -r "download\.Client" internal/` 无结果
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过
- [ ] rename 相关测试可以加 `t.Parallel()` 并通过
