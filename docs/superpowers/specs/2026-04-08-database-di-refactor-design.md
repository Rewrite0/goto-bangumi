# 数据库依赖注入重构设计

**日期**: 2026-04-08
**状态**: 待实现

## 背景

当前 `internal/database/db.go` 采用全局单例模式（`globalDB` + `InitDB` / `GetDB` / `CloseDB`）。这导致：

1. **测试之间会互相污染**：`InitDB` 内部有 `if globalDB != nil { return nil }` 短路，第二个测试拿到的仍是第一个测试的库
2. **无法 `t.Parallel()`**：所有测试共享同一个全局变量
3. **隐式依赖**：调用方（`refresh` / `subscribe` / `taskrunner/handlers` / `rename` / `task`）直接 `database.GetDB()`，依赖不可见
4. **未来加新依赖（metrics、tracing…）时**，每个函数签名都得改一轮

## 目标

- 删除 `database` 包里的全局单例
- 业务 package struct 化，通过构造函数显式注入 `*database.DB`
- 测试可以各自建内存库、相互隔离、支持 `t.Parallel()`

## 非目标

- **不**引入 Repository 接口抽象层（YAGNI，CRUD 密集项目不需要）
- **不**引入 wire / fx 等 DI 框架（当前依赖图小，手写组装一目了然）
- **不**动 `eventbus`、`config` 的全局性（本次只聚焦 db）
- **不**做无关重构

## 设计

### 1. `database` 包瘦身

删除：
- `globalDB` 包级变量
- `InitDB(dsn *string) error`
- `GetDB() *DB`
- `CloseDB() error`

保留：
- `NewDB(dsn *string) (*DB, error)`
- `(*DB).Close() error`
- 其他所有 CRUD 方法保持不变

### 2. 业务 package struct 化

把每个 package 的包级函数改成某个 struct 的方法，struct 在构造函数里接收 `*database.DB` 和其他依赖。

#### `internal/refresh`

```go
type Refresher struct {
    db *database.DB
    // 其他依赖（eventbus 等）按需注入
}

func New(db *database.DB, ...) *Refresher

func (r *Refresher) FindNewBangumi(ctx context.Context) error
// 其他现有 refresh 函数改为方法
```

#### `internal/subscribe`

```go
type Subscriber struct {
    db *database.DB
    // ...
}

func New(db *database.DB, ...) *Subscriber

func (s *Subscriber) Rename(ctx context.Context) error
func (s *Subscriber) CheckDownload(ctx context.Context) error
func (s *Subscriber) CheckDownloading(ctx context.Context) error
```

#### `internal/rename`

```go
type Renamer struct {
    db *database.DB
}

func New(db *database.DB) *Renamer

// 现有 utils.go 中用到 db 的函数改为方法
```

#### `internal/task`

```go
type Task struct {
    db *database.DB
    // ...
}

func New(db *database.DB, ...) *Task

func (t *Task) Refresh(ctx context.Context) error
```

### 3. `taskrunner/handlers` struct 化

每个 handler 改为 struct，持有它需要的依赖（当前都是 `*database.DB`），实现统一的 handler 接口（沿用现有接口，只是从函数改成方法接收者）。

```go
// internal/taskrunner/handlers/downloading.go
type DownloadingHandler struct {
    db *database.DB
}

func NewDownloadingHandler(db *database.DB) *DownloadingHandler

func (h *DownloadingHandler) Handle(ctx context.Context, task *Task) error {
    // 原来的 database.GetDB().AddTorrentError(...) 改为 h.db.AddTorrentError(...)
}
```

同样处理 `CheckHandler`、`RenameHandler`。

### 4. `core/Program` 作为应用容器

`internal/core/program.go` 承担依赖组装和生命周期管理：

```go
type Program struct {
    db         *database.DB
    refresher  *refresh.Refresher
    subscriber *subscribe.Subscriber
    renamer    *rename.Renamer
    task       *task.Task
    scheduler  *scheduler.Scheduler
    taskrunner *taskrunner.Runner
    // ...
}

func NewProgram(...) (*Program, error) {
    db, err := database.NewDB(nil)
    if err != nil { return nil, err }

    refresher := refresh.New(db, ...)
    subscriber := subscribe.New(db, ...)
    renamer := rename.New(db)
    tsk := task.New(db, ...)

    tr := taskrunner.New()
    tr.Register(handlers.NewDownloadingHandler(db))
    tr.Register(handlers.NewCheckHandler(db))
    tr.Register(handlers.NewRenameHandler(db))

    sch := scheduler.New(tsk, ...) // scheduler 调用 tsk.Refresh() 而非 task.Refresh()

    return &Program{
        db: db, refresher: refresher, subscriber: subscriber,
        renamer: renamer, task: tsk, scheduler: sch, taskrunner: tr,
    }, nil
}

func (p *Program) Close() error {
    // 关闭 scheduler、taskrunner 等，最后关 db
    return p.db.Close()
}
```

### 5. `scheduler` 调用方式

`scheduler` 目前直接调用 `task.Refresh()` 之类的包级函数。改为持有 `*task.Task` 引用：

```go
type Scheduler struct {
    task *task.Task
    // ...
}

func New(t *task.Task, ...) *Scheduler
// 内部 cron 注册回调时调用 s.task.Refresh(ctx)
```

### 6. 测试

每个测试自己建内存库，完全隔离：

```go
func TestFindNewBangumi(t *testing.T) {
    t.Parallel()
    dsn := ":memory:"
    db, err := database.NewDB(&dsn)
    if err != nil { t.Fatal(err) }
    defer db.Close()

    r := refresh.New(db /*, deps...*/)
    err = r.FindNewBangumi(context.Background())
    // 断言...
}
```

无 test helper、无全局状态，`t.Parallel()` 安全。

## 影响面

需要改动的文件：

- `internal/database/db.go` — 删除 globalDB / InitDB / GetDB / CloseDB
- `internal/core/program.go` — 组装所有 service
- `internal/refresh/bangumi.go` — struct 化
- `internal/refresh/analyser.go` — 如用到 db 则迁移为方法
- `internal/refresh/bangumi_test.go` — 改为 `NewDB(":memory:")` + `refresh.New(db)`
- `internal/refresh/analyser_test.go` — 同上
- `internal/subscribe/rename.go`
- `internal/subscribe/checkDownload.go`
- `internal/subscribe/checkDownloading.go`
- `internal/taskrunner/handlers/downloading.go`
- `internal/taskrunner/handlers/check.go`
- `internal/taskrunner/handlers/rename.go`
- `internal/rename/utils.go`
- `internal/task/refresh.go`
- `internal/scheduler/scheduler.go`

## 风险 / 注意事项

1. **改动面广但机械**：涉及约 15 个文件，但大部分是"把 `database.GetDB()` 换成 `x.db`，把函数签名改成方法"。建议按 package 分批改，每改完一个 package 立刻跑 `go build ./... && go test ./...`。
2. **`scheduler` 与 `task` 的耦合方向变化**：现在是 `scheduler` import `task` 调用包函数，改后 `scheduler` 仍 import `task` 但持有 `*task.Task`，方向不变，不会引入循环依赖。
3. **`taskrunner` handler 接口**：如果当前是函数类型（`type HandlerFunc func(...)`), 需要改为 interface 或让 struct 提供方法值（`tr.Register(h.Handle)`），实现细节在 plan 里细化。

## 验收标准

- [ ] `grep -r "database.GetDB\|database.InitDB\|database.CloseDB\|globalDB" internal/` 无结果
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过
- [ ] `internal/refresh` 相关测试可以加 `t.Parallel()` 并通过
