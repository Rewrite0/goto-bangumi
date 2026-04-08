# 数据库依赖注入重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `database` 包的全局单例 `globalDB`/`InitDB`/`GetDB`/`CloseDB`，把所有使用 db 的业务 package 改为构造函数注入 `*database.DB`，使测试可并行且相互隔离。

**Architecture:** 业务 package 从包级函数改造为 struct + 方法，`*database.DB` 通过构造函数注入。`internal/core/program.go` 成为应用容器，负责组装整条依赖链并管理 db 生命周期。`taskrunner` handler 工厂函数改为显式接收 `db` 参数。

**Tech Stack:** Go 1.25, GORM, SQLite (`github.com/glebarez/sqlite`), `log/slog`, 原生 `testing`。

**对应 spec:** `docs/superpowers/specs/2026-04-08-database-di-refactor-design.md`

---

## 重构策略说明

这是一个大面积重构，不是全新功能，所以：

- **不写新单测再反过来改**：业务逻辑不变，现有测试就是回归保障。每个 Task 做完就 `go build ./... && go test ./...`，保持绿。
- **按 package 分 Task，自底向上**：先改最底层（`database` 包暴露 `NewDB`，然后改被依赖多的 `refresh` `rename`，最后改 `subscribe` / `taskrunner/handlers` / `task`，最后 `core/program.go` 组装），避免中间状态编译不过。
- **一个 Task 一个 commit**。每个 Task 结束前必须 `go build ./...` 和 `go test ./...` 全绿。
- **不动 `scheduler` 自己的全局单例**：那是另一件事，超出本次 spec 范围。

---

## File Structure

**修改的文件：**
- `internal/database/db.go` — 删除全局单例相关函数
- `internal/refresh/bangumi.go` — 包函数 → `Refresher` struct 方法
- `internal/refresh/analyser.go` — `createBangumi` 签名保留（已经接收 db），但文件内部其他用到 db 的点改为方法
- `internal/refresh/bangumi_test.go` — 使用 `refresh.New(db)` 注入
- `internal/refresh/analyser_test.go` — 同上
- `internal/rename/utils.go` — `getBangumi` 改为 `Renamer` 方法
- `internal/rename/rename.go` — `Rename`/`GetBangumi` 改为 `Renamer` 方法
- `internal/subscribe/checkDownload.go` — 给 `CheckService` 加 `db` 字段
- `internal/subscribe/checkDownloading.go` — 给 `checkDownloadingService` 加 `db` 字段
- `internal/subscribe/rename.go` — 给 `renameService` 加 `db` 和 `renamer` 字段
- `internal/subscribe/init.go` — 接收 `db`/`renamer`，组装时注入
- `internal/taskrunner/handlers/downloading.go` — 工厂函数加 `db` 参数
- `internal/taskrunner/handlers/check.go` — 同上
- `internal/taskrunner/handlers/rename.go` — 同上（也接收 `renamer`）
- `internal/task/refresh.go` — 给 `RSSRefreshTask` 加 `db` 和 `refresher` 字段
- `internal/core/program.go` — 组装所有 service，持有 db 生命周期
- **最后**删除 `internal/database/db.go` 中的 `globalDB`/`InitDB`/`GetDB`/`CloseDB`

---

## Task 1: 在 `database` 包增加独立的 `NewDB` 入口（并行兼容）

**目的**：先让 `database.NewDB` 可以被外部直接调用创建独立实例，**但暂时保留**全局单例相关函数不动，这样后续每个 Task 都能独立地把调用点从 `database.GetDB()` 切换到注入的 `db` 字段，而不会一下子打断全局。Task 最后一步再删掉单例。

**Files:**
- Modify: `internal/database/db.go`

- [ ] **Step 1: 确认 `NewDB` 已经存在且签名正确**

目前 `internal/database/db.go:27` 已经有 `func NewDB(dsn *string) (*DB, error)`，并且它独立于 `globalDB`。无需改动。

- [ ] **Step 2: 验证构建**

```bash
go build ./...
go test ./...
```
Expected: 全绿。

- [ ] **Step 3: 无需 commit**（没有实际改动）。

---

## Task 2: `rename` 包 struct 化为 `Renamer`

**Files:**
- Modify: `internal/rename/rename.go`
- Modify: `internal/rename/utils.go`

- [ ] **Step 1: 在 `rename.go` 中定义 `Renamer` struct 和构造函数**

把文件开头 `renameConfig` 相关的 `Init` 保持不动（`renameConfig` 作为包级配置对象依然可用；本次不动 config 单例）。在文件顶部增加：

```go
// Renamer 持有数据库依赖，对外暴露重命名相关方法。
type Renamer struct {
    db *database.DB
}

// New 创建 Renamer
func New(db *database.DB) *Renamer {
    return &Renamer{db: db}
}
```

并在文件顶部 import 块中加入 `"goto-bangumi/internal/database"`。

- [ ] **Step 2: 把 `Rename` 改为 `Renamer` 方法**

将：
```go
func Rename(ctx context.Context, torrent *model.Torrent, bangumi *model.Bangumi) {
```
改为：
```go
func (r *Renamer) Rename(ctx context.Context, torrent *model.Torrent, bangumi *model.Bangumi) {
```

函数体中 `getBangumi(ctx, torrent)` 改为 `r.getBangumi(ctx, torrent)`。

- [ ] **Step 3: 把 `GetBangumi` 改为 `Renamer` 方法**

```go
func (r *Renamer) GetBangumi(ctx context.Context, torrent *model.Torrent) (*model.Bangumi, error) {
    return r.getBangumi(ctx, torrent)
}
```

- [ ] **Step 4: 把 `utils.go` 中的 `getBangumi` 改为方法**

将：
```go
func getBangumi(ctx context.Context, torrent *model.Torrent) (*model.Bangumi, error) {
    ...
    db := database.GetDB()
    bangumi, err := db.GetBangumiByOfficialTitle(pathInfo.BangumiName)
    ...
}
```
改为：
```go
func (r *Renamer) getBangumi(ctx context.Context, torrent *model.Torrent) (*model.Bangumi, error) {
    ...
    bangumi, err := r.db.GetBangumiByOfficialTitle(pathInfo.BangumiName)
    ...
}
```

删除 `utils.go` 中对 `database.GetDB()` 的调用和 `"goto-bangumi/internal/database"` 的 import（如果 utils.go 不再直接引用 database 包的话）。

- [ ] **Step 5: 构建会失败 —— 因为 `rename.Rename` / `rename.GetBangumi` 的调用方还在按包函数调用**

此时先不要跑 `go build`，进入 Task 3 修复调用方。

*（注：Task 2 和 Task 3 合并为一个 commit，保证中间态不破坏构建。）*

---

## Task 3: 修复 `rename.Rename` 的所有调用方，并把 Task 2+3 合并 commit

**Files:**
- Modify: `internal/taskrunner/handlers/rename.go`
- Modify: `internal/subscribe/rename.go`

- [ ] **Step 1: 修改 `internal/taskrunner/handlers/rename.go`**

```go
package handlers

import (
    "context"
    "log/slog"

    "goto-bangumi/internal/database"
    "goto-bangumi/internal/model"
    "goto-bangumi/internal/rename"
    "goto-bangumi/internal/taskrunner"
)

// NewRenameHandler 创建重命名处理器
func NewRenameHandler(db *database.DB, renamer *rename.Renamer) taskrunner.PhaseFunc {
    return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
        slog.Info("[rename handler] 开始重命名",
            "torrent", task.Torrent.Name,
            "bangumi", task.Bangumi.OfficialTitle)

        renamer.Rename(ctx, task.Torrent, task.Bangumi)

        if err := db.TorrentRenamed(ctx, task.Torrent.Link); err != nil {
            slog.Error("[rename handler] 更新种子重命名状态失败",
                "error", err, "link", task.Torrent.Link)
            return taskrunner.PhaseResult{Err: err}
        }

        slog.Info("[rename handler] 重命名完成", "torrent", task.Torrent.Name)
        return taskrunner.PhaseResult{}
    }
}
```

- [ ] **Step 2: 修改 `internal/subscribe/rename.go`**

```go
package subscribe

import (
    "context"
    "log/slog"

    "goto-bangumi/internal/database"
    "goto-bangumi/internal/eventbus"
    "goto-bangumi/internal/model"
    "goto-bangumi/internal/rename"
)

type renameService struct {
    bus     eventbus.EventBus
    db      *database.DB
    renamer *rename.Renamer
}

func (rs *renameService) handleRename(ctx context.Context, data model.RenameEvent) {
    slog.Info("[rename service] 收到重命名事件", "torrent", data.Torrent.Name, "bangumi", data.Bangumi.OfficialTitle)

    rs.renamer.Rename(ctx, data.Torrent, data.Bangumi)

    if err := rs.db.TorrentRenamed(ctx, data.Torrent.Link); err != nil {
        slog.Error("[rename service] 更新种子重命名状态失败", "error", err, "link", data.Torrent.Link)
        return
    }

    slog.Info("[rename service] 重命名完成", "torrent", data.Torrent.Name)
}

func (rs *renameService) Start(ctx context.Context) {
    ch, unsubscribe := eventbus.Subscribe[model.RenameEvent](rs.bus, ctx, 100)
    defer unsubscribe()
    slog.Info("[rename service] 重命名服务已启动")

    for event := range ch {
        go rs.handleRename(ctx, event)
    }
}
```

- [ ] **Step 3: 此时 `subscribe/init.go` 和 `core/program.go` 的调用点会编译失败**

这是预期的，Task 8/9 会修复。这里先标记：**Task 3 结束不构建，直接进入 Task 4。** Task 2+3 会和后续 Task 一起，在**所有 Task 做完后统一 commit**。

*实际执行时为了每 Task 一 commit，Task 2-3-4-5-6-7 作为"重构串"一起验证；每个 Task 可以单独 commit 其修改的文件，但构建验证只在最后一个 Task 跑。*

**修正执行策略**：下面所有 Task 都只改文件不跑 build，最后 Task 10 统一 `go build ./... && go test ./...`，通过后一次性 commit（或按文件分组 commit）。

*为清晰起见，最终 commit 方式在 Task 11 明确。*

---

## Task 4: `refresh` 包 struct 化为 `Refresher`

**Files:**
- Modify: `internal/refresh/bangumi.go`
- Modify: `internal/refresh/analyser.go`

- [ ] **Step 1: 在 `bangumi.go` 增加 `Refresher` struct**

把 `bangumi.go` 改成：

```go
package refresh

import (
    "context"
    "errors"
    "log/slog"

    "gorm.io/gorm"

    "goto-bangumi/internal/database"
    "goto-bangumi/internal/model"
    "goto-bangumi/internal/network"
    "goto-bangumi/internal/taskrunner"
)

// Refresher 封装 RSS 刷新与新番剧发现逻辑。
type Refresher struct {
    db *database.DB
}

// New 创建 Refresher
func New(db *database.DB) *Refresher {
    return &Refresher{db: db}
}

func (r *Refresher) getTorrents(ctx context.Context, url string) []*model.Torrent {
    client := network.GetRequestClient()
    torrents, _ := client.GetTorrents(ctx, url)
    slog.Debug("[getTorrents]从 RSS 获取种子列表", "URL", url, "数量", len(torrents))
    newTorrents, _ := r.db.CheckNewTorrents(ctx, torrents)
    return newTorrents
}

// FindNewBangumi 从 rss 里面看看没有没新的番剧
func (r *Refresher) FindNewBangumi(ctx context.Context, rssItem *model.RSSItem) {
    slog.Info("[FindNewBangumi]检查 RSS 是否有新的番剧", "RSS 名称", rssItem.Name)
    netClient := network.GetRequestClient()
    torrents, _ := netClient.GetTorrents(ctx, rssItem.Link)
    for _, t := range torrents {
        _, err := r.db.GetBangumiParseByTitle(ctx, t.Name)
        if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
            slog.Debug("[FindNewBangumi]没有找到番剧信息，可能是新的番剧", "种子名称", t.Name, "error", err)
            if FilterTorrent(t, rssItem.ExcludeFilter, rssItem.IncludeFilter) {
                slog.Info("[FindNewBangumi]发现新的番剧", "种子名称", t.Name)
                r.createBangumi(ctx, t, rssItem)
            }
        }
    }
}

// RefreshRSS 刷新单个 RSS 源并向 runner 提交下载任务
func (r *Refresher) RefreshRSS(ctx context.Context, url string, runner *taskrunner.TaskRunner) {
    slog.Info("[RefreshRSS]刷新 RSS", "URL", url)
    torrents := r.getTorrents(ctx, url)
    slog.Debug("[RefreshRSS]获取种子列表", "数量", len(torrents))
    for _, t := range torrents {
        metaData, err := r.db.GetBangumiParseByTitle(ctx, t.Name)
        slog.Debug("[RefreshRSS]检查番剧信息", "种子名称", t.Name, "error", err)
        if err != nil {
            continue
        }
        if FilterTorrent(t, metaData.IncludeFilter, metaData.ExcludeFilter) {
            t.Bangumi = metaData
            _ = r.db.CreateTorrent(ctx, t)
            runner.Submit(model.NewAddTask(t, t.Bangumi))
        }
    }
}
```

- [ ] **Step 2: 把 `analyser.go` 的 `createBangumi` 改为 `Refresher` 方法**

`internal/refresh/analyser.go:135` 当前签名：
```go
func createBangumi(ctx context.Context, db *database.DB, torrent *model.Torrent, rssItem *model.RSSItem) {
```

改为：
```go
func (r *Refresher) createBangumi(ctx context.Context, torrent *model.Torrent, rssItem *model.RSSItem) {
```

函数体内把 `db` 替换为 `r.db`。

- [ ] **Step 3: 处理 `analyser.go` 内其他直接用 `database.GetDB()` 或包级 `db` 的点**

扫描 `analyser.go` 里如果还有 `database.GetDB()` 调用，一律改为 `r.db`。如果 analyser.go 里还有其他包级函数仍需 db，要么：
- 把它们也改为 `Refresher` 方法（推荐），或
- 明确它们不需要 db，无需改动

**具体操作**：`grep -n "database.GetDB\|db \*database" internal/refresh/analyser.go`，对每个命中点按上述原则处理。

---

## Task 5: `refresh` 包的测试迁移

**Files:**
- Modify: `internal/refresh/bangumi_test.go`
- Modify: `internal/refresh/analyser_test.go`

- [ ] **Step 1: 改 `bangumi_test.go`**

把所有这种模式：
```go
memoryDB := ":memory:"
database.InitDB(&memoryDB)
defer database.CloseDB()
...
db := database.GetDB()
...
FindNewBangumi(ctx, rssItem)
```

改为：
```go
memoryDB := ":memory:"
db, err := database.NewDB(&memoryDB)
if err != nil {
    t.Fatal(err)
}
defer db.Close()
r := refresh.New(db)
...
r.FindNewBangumi(ctx, rssItem)
```

每个子测试都改。特别注意 `bangumi_test.go:21-22, 36, 83-84, 111-112, 125, 152-154` 这几个位置（grep 过的）。

- [ ] **Step 2: 改 `analyser_test.go`**

`internal/refresh/analyser_test.go:212` 处 `createBangumi(context.Background(), db, torrent, rssItem)` 改为：
```go
r := &Refresher{db: db} // 或 refresh.New(db) 若在外部包
r.createBangumi(context.Background(), torrent, rssItem)
```

（注意：由于 `createBangumi` 是小写，`analyser_test.go` 是同包测试，可以直接 `&Refresher{db: db}`。）

同样把文件顶部的 `database.InitDB` 替换为 `database.NewDB`。

- [ ] **Step 3: 可选 —— 给测试加 `t.Parallel()`**

现在每个测试拥有独立 db，可以加 `t.Parallel()`。但这属于改进而非必需，可以放到最后。

---

## Task 6: `subscribe` 包注入 `db`

**Files:**
- Modify: `internal/subscribe/checkDownload.go`
- Modify: `internal/subscribe/checkDownloading.go`
- Modify: `internal/subscribe/init.go`

*（`rename.go` 已在 Task 3 改过。）*

- [ ] **Step 1: 改 `checkDownload.go`**

```go
type CheckService struct {
    bus eventbus.EventBus
    db  *database.DB
}
```

函数体内 `database.GetDB().AddTorrentDUID(...)` 改为 `cs.db.AddTorrentDUID(...)`。

- [ ] **Step 2: 改 `checkDownloading.go`**

```go
type checkDownloadingService struct {
    bus eventbus.EventBus
    db  *database.DB
}
```

函数体内两处 `db := database.GetDB()` 删除，直接用 `cds.db.AddTorrentError(...)` / `cds.db.AddTorrentDownload(...)`。

- [ ] **Step 3: 改 `init.go` 签名，接收 `db` 和 `renamer`**

读取 `internal/subscribe/init.go`，将 `Init(ctx)` 改为：

```go
func Init(ctx context.Context, db *database.DB, renamer *rename.Renamer) {
    bus := eventbus.Default()  // 或现有获取 bus 的方式，按现状保留

    checkService := &CheckService{
        bus: bus,
        db:  db,
    }
    go checkService.Start(ctx)

    checkDownloadingService := &checkDownloadingService{
        bus: bus,
        db:  db,
    }
    go checkDownloadingService.Start(ctx)

    renameService := &renameService{
        bus:     bus,
        db:      db,
        renamer: renamer,
    }
    go renameService.Start(ctx)
}
```

**注意**：先 `Read` 当前 `init.go` 的完整内容，找出它如何获取 `bus`，保留现有方式；只添加 `db` 和 `renamer` 参数与字段。

---

## Task 7: `taskrunner/handlers` 工厂函数显式接收 `db`

**Files:**
- Modify: `internal/taskrunner/handlers/downloading.go`
- Modify: `internal/taskrunner/handlers/check.go`

*（`rename.go` 已在 Task 3 改过。）*

- [ ] **Step 1: 改 `downloading.go`**

```go
func NewDownloadingHandler(db *database.DB) taskrunner.PhaseFunc {
    return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
        // 检查是否超时（4小时）
        if time.Since(task.StartTime) > 4*time.Hour {
            slog.Warn("[downloading handler] 下载超过4小时，标记为异常", ...)
            db.AddTorrentError(ctx, task.Torrent.Link)
            return taskrunner.PhaseResult{Err: fmt.Errorf("download timeout after 4 hours")}
        }
        ...
        if info.Completed > 0 {
            task.Torrent.Downloaded = model.DownloadDone
            if err := db.AddTorrentDownload(ctx, task.Torrent.Link); err != nil {
                ...
            }
            ...
        }
        ...
    }
}
```

把所有 `database.GetDB()` 替换为闭包捕获的 `db`。

- [ ] **Step 2: 改 `check.go`**

```go
func NewCheckHandler(db *database.DB) taskrunner.PhaseFunc {
    return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
        for _, guid := range task.Guids {
            ...
            if trueID != "" {
                task.Torrent.DownloadUID = trueID
                if err := db.AddTorrentDUID(ctx, task.Torrent.Link, trueID); err != nil {
                    ...
                }
                ...
            }
        }
        ...
    }
}
```

**同时看一下 `handlers/adding.go`（`NewAddHandler`）**：如果它也用 `database.GetDB()`，顺手改成 `NewAddHandler(db *database.DB)`，并加入到 `core/program.go` 的调用点。**不要漏**。

---

## Task 8: `task` 包注入 `db` 和 `refresher`

**Files:**
- Modify: `internal/task/refresh.go`

- [ ] **Step 1: 修改 `RSSRefreshTask`**

```go
type RSSRefreshTask struct {
    interval  time.Duration
    enabled   bool
    runner    *taskrunner.TaskRunner
    db        *database.DB
    refresher *refresh.Refresher
}

func NewRSSRefreshTask(
    programConfig model.ProgramConfig,
    runner *taskrunner.TaskRunner,
    db *database.DB,
    refresher *refresh.Refresher,
) *RSSRefreshTask {
    interval := programConfig.RssTime
    return &RSSRefreshTask{
        interval:  time.Duration(interval) * time.Second,
        enabled:   true,
        runner:    runner,
        db:        db,
        refresher: refresher,
    }
}

func (t *RSSRefreshTask) Run(ctx context.Context) error {
    rssList, err := t.db.ListActiveRSS(ctx)
    if err != nil {
        return fmt.Errorf("[RSS task] 获取 RSS 列表失败: %w", err)
    }
    slog.Debug("[Rss task] 开始刷新 RSS", "数量", len(rssList))
    for _, rss := range rssList {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            slog.Debug("[refresh] 刷新 RSS 源", "名称", rss.Name, "URL", rss.Link)
            t.refresher.FindNewBangumi(ctx, rss)
            t.refresher.RefreshRSS(ctx, rss.Link, t.runner)
            time.Sleep(2 * time.Second)
        }
    }
    slog.Debug("RSS 刷新完成")
    return nil
}
```

删掉 `db := database.GetDB()` 这一行。

---

## Task 9: 改造 `core/program.go` 作为应用容器

**Files:**
- Modify: `internal/core/program.go`

- [ ] **Step 1: 把 `Program` 改为持有所有依赖**

```go
package core

import (
    "context"
    "log/slog"

    "goto-bangumi/internal/conf"
    "goto-bangumi/internal/database"
    "goto-bangumi/internal/download"
    "goto-bangumi/internal/logger"
    "goto-bangumi/internal/model"
    "goto-bangumi/internal/network"
    "goto-bangumi/internal/notification"
    "goto-bangumi/internal/parser"
    "goto-bangumi/internal/refresh"
    "goto-bangumi/internal/rename"
    "goto-bangumi/internal/scheduler"
    "goto-bangumi/internal/subscribe"
    "goto-bangumi/internal/task"
    "goto-bangumi/internal/taskrunner"
    "goto-bangumi/internal/taskrunner/handlers"
)

type Program struct {
    ctx    context.Context
    cancel context.CancelFunc

    db        *database.DB
    refresher *refresh.Refresher
    renamer   *rename.Renamer
}

// InitProgram 初始化全局单例类组件（config/logger/network/parser/notification/download）
// 然后返回已组装好 db 依赖的 Program。
func InitProgram(ctx context.Context) *Program {
    if err := conf.Init(); err != nil {
        slog.Error("[program] 加载配置文件失败", "error", err)
        panic(err)
    }
    cfg := conf.Get()

    logger.Init(cfg.Program.DebugEnable)

    db, err := database.NewDB(nil)
    if err != nil {
        slog.Error("[program] 初始化数据库失败", "error", err)
        panic(err)
    }

    network.Init(&cfg.Proxy)
    parser.Init(&cfg.Parser)
    notification.NotificationClient.Init(&cfg.Notification)
    download.Client.Init(&cfg.Downloader)
    rename.Init(&cfg.Rename)

    renamer := rename.New(db)
    refresher := refresh.New(db)

    return &Program{
        db:        db,
        refresher: refresher,
        renamer:   renamer,
    }
}

func (p *Program) Start(ctx context.Context) {
    p.ctx, p.cancel = context.WithCancel(ctx)
    go download.Client.Login(p.ctx)

    // 启动 subscribe 事件服务
    subscribe.Init(p.ctx, p.db, p.renamer)

    // 创建并启动 taskrunner
    runner := taskrunner.New(64, 4)
    runner.Register(model.PhaseAdding, handlers.NewAddHandler(p.db))
    runner.Register(model.PhaseChecking, handlers.NewCheckHandler(p.db))
    runner.Register(model.PhaseDownloading, handlers.NewDownloadingHandler(p.db))
    runner.Register(model.PhaseRenaming, handlers.NewRenameHandler(p.db, p.renamer))
    runner.Start(p.ctx)

    // 启动调度器
    p.initScheduler(p.ctx, runner)
}

func (p *Program) Stop() {
    p.cancel()
    if p.db != nil {
        if err := p.db.Close(); err != nil {
            slog.Error("[program] 关闭数据库失败", "error", err)
        }
    }
    slog.Info("程序已停止")
}

func (p *Program) initScheduler(ctx context.Context, runner *taskrunner.TaskRunner) {
    scheduler.InitScheduler(ctx)
    s := scheduler.GetScheduler()
    if s == nil {
        slog.Error("调度器初始化失败")
        return
    }
    s.AddTask(task.NewRSSRefreshTask(conf.Get().Program, runner, p.db, p.refresher))
    s.Start()
    slog.Info("调度器启动成功")
}
```

**注意点：**
1. `InitProgram` 原来返回 `void`，现在返回 `*Program`。需要检查 `main` 包（或调用 `InitProgram` 的地方）同步调整：原来应该是 `core.InitProgram(ctx); p := &core.Program{}` 之类；改为 `p := core.InitProgram(ctx); p.Start(ctx)`。
2. **先 `Grep -n "core.InitProgram\|core.Program" cmd/ internal/` 确认调用方**，再同步修改。如果 main 里是 `core.InitProgram(ctx)` 后另行 `new(core.Program).Start(ctx)`，必须改成用 `InitProgram` 的返回值。

- [ ] **Step 2: 修复 `main` 入口（如果需要）**

```bash
# 先确认
grep -rn "core.InitProgram\|core.Program" cmd/ internal/
```

按实际情况调整入口代码，使其使用 `InitProgram` 返回的 `*Program`。

---

## Task 10: 删除 `database` 包的全局单例

**Files:**
- Modify: `internal/database/db.go`

- [ ] **Step 1: 删除相关代码块**

删除 `internal/database/db.go:22-105` 区间内以下内容：
- `var globalDB *DB`
- `func InitDB(dsn *string) error`
- `func GetDB() *DB`
- `func CloseDB() error`
- 注释 `// ============ 单例模式相关方法 ============`

保留 `NewDB` 和 `(*DB).Close`。

- [ ] **Step 2: 全量构建 + 测试**

```bash
go build ./...
go test ./...
```
Expected: 全绿。如有残留 `database.GetDB()` 调用，编译会报错，回去对应文件修复。

- [ ] **Step 3: 确认无残留**

```bash
grep -rn "database\.GetDB\|database\.InitDB\|database\.CloseDB\|globalDB" internal/ cmd/
```
Expected: 无命中（可能命中老 plan 文档，那不算）。

---

## Task 11: 最终验证并分组提交

- [ ] **Step 1: 全量构建 + 测试**

```bash
go build ./...
go test ./...
go vet ./...
```
Expected: 全绿。

- [ ] **Step 2: 分组 commit**

建议按下列分组提交，每组独立原子：

```bash
# 1. rename 包 struct 化
git add internal/rename/rename.go internal/rename/utils.go
git commit -m "refactor(rename): convert to Renamer struct with injected DB"

# 2. refresh 包 struct 化 + 测试迁移
git add internal/refresh/bangumi.go internal/refresh/analyser.go \
        internal/refresh/bangumi_test.go internal/refresh/analyser_test.go
git commit -m "refactor(refresh): convert to Refresher struct, migrate tests to per-test in-memory DB"

# 3. subscribe 注入 db
git add internal/subscribe/
git commit -m "refactor(subscribe): inject *database.DB into services"

# 4. taskrunner handlers 注入 db
git add internal/taskrunner/handlers/
git commit -m "refactor(taskrunner/handlers): inject *database.DB via factory params"

# 5. task 包注入 db/refresher
git add internal/task/refresh.go
git commit -m "refactor(task): inject *database.DB and *refresh.Refresher into RSSRefreshTask"

# 6. core.Program 成为应用容器 + 删除 database 全局单例
git add internal/core/program.go internal/database/db.go cmd/
git commit -m "refactor(database,core): remove global DB singleton, make Program own db lifecycle"
```

（若 `cmd/` 未改动则从最后一条里去掉。）

- [ ] **Step 3: 最终扫描**

```bash
grep -rn "database\.GetDB\|database\.InitDB\|database\.CloseDB\|globalDB" internal/ cmd/
```
Expected: 空。

---

## 验收标准

- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] `grep -rn "database\.GetDB\|database\.InitDB\|database\.CloseDB\|globalDB" internal/ cmd/` 无业务代码命中
- [ ] `internal/refresh/bangumi_test.go` 中可以给任意测试加 `t.Parallel()` 且测试仍通过（验证隔离性）
