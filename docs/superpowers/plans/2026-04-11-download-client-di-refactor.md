# Download Client DI Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the global `download.Client` singleton with dependency injection via constructor parameters, matching the existing database DI pattern.

**Architecture:** Add `NewDownloadClient()` factory function, hold the instance in `Program`, inject `*download.DownloadClient` into all 4 consumer sites (3 handlers + rename). The `Init()` hot-reload mechanism stays unchanged — all holders share the same pointer.

**Tech Stack:** Go, existing project patterns (constructor injection, `*database.DB` as reference)

---

### Task 1: Add `NewDownloadClient()` factory and move `loginGroup` into struct

**Files:**
- Modify: `internal/download/client.go:21-36`

Note: `var Client` is kept temporarily so `program.go` still compiles. It gets deleted in Task 6 after all consumers are updated.

- [ ] **Step 1: Add `NewDownloadClient()` factory, move `loginGroup` into struct**

Change `internal/download/client.go`:

Remove only `loginGroup`:
```go
// loginGroup 确保同一时间只有一个登录协程
var loginGroup singleflight.Group
```

Keep `var Client` but change it to use the factory:
```go
// Client 为一个全局的下载客户端实例（将在 DI 改造完成后删除）
var Client = NewDownloadClient()
```

Add `loginGroup` field to `DownloadClient` struct:
```go
type DownloadClient struct {
	Downloader     downloader.BaseDownloader
	limiter        *rate.Limiter
	SavePath       string
	downloaderType string

	// 登录控制
	logined    bool // 是否已登录
	LoginError bool // 登录错误通道
	loginGroup singleflight.Group
}
```

Add factory function (after the struct definition):
```go
// NewDownloadClient 创建下载客户端实例
func NewDownloadClient() *DownloadClient {
	return &DownloadClient{}
}
```

Update `Login` method to use `c.loginGroup` instead of package-level `loginGroup`:
```go
func (c *DownloadClient) Login(ctx context.Context) error {
	_, err, _ := c.loginGroup.Do("login", func() (any, error) {
```

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: PASS (`var Client` still exists, all existing consumers unaffected)

- [ ] **Step 3: Commit**

```bash
git add internal/download/client.go
git commit -m "refactor(download): add NewDownloadClient factory, move loginGroup into struct"
```

---

### Task 2: Inject `*DownloadClient` into AddHandler

**Files:**
- Modify: `internal/taskrunner/handlers/add.go`

- [ ] **Step 1: Add `*download.DownloadClient` parameter to `NewAddHandler`**

Replace the full `NewAddHandler` function in `internal/taskrunner/handlers/add.go`:

```go
// NewAddHandler 创建添加下载处理器，将种子添加到下载器
func NewAddHandler(dl *download.DownloadClient) taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		savePath := genSavePath(task.Bangumi)
		guids, err := dl.Add(ctx, task.Torrent.Link, savePath)
		if err != nil {
			slog.Warn("[add handler] 添加下载失败，稍后重试",
				"torrent", task.Torrent.Name, "error", err)
			// TODO: 不应一直重试, 一是要有次数的限制, 二是要看是什么错误
			if apperrors.IsNetworkError(err) {
				return taskrunner.PhaseResult{PollAfter: 5 * time.Second}
			}
			return taskrunner.PhaseResult{Err: err}
		}

		task.Guids = guids
		slog.Debug("[add handler] 添加下载成功",
			"torrent", task.Torrent.Name, "guids", guids)
		return taskrunner.PhaseResult{}
	}
}
```

Remove the `"goto-bangumi/internal/download"` import (no longer needed directly — `dl` is passed in).

- [ ] **Step 2: Verify build of handlers package**

Run: `go build ./internal/taskrunner/handlers/...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/taskrunner/handlers/add.go
git commit -m "refactor(handlers): inject DownloadClient into AddHandler"
```

---

### Task 3: Inject `*DownloadClient` into CheckHandler

**Files:**
- Modify: `internal/taskrunner/handlers/check.go`

- [ ] **Step 1: Add `*download.DownloadClient` parameter to `NewCheckHandler`**

Replace the `NewCheckHandler` function in `internal/taskrunner/handlers/check.go`:

```go
// NewCheckHandler 创建检查处理器，验证下载是否成功添加到下载器
func NewCheckHandler(db *database.DB, dl *download.DownloadClient) taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		for _, guid := range task.Guids {
			trueID, err := dl.Check(ctx, guid)

			// GUID 没找到，试下一个
			if apperrors.IsKeyError(err) {
				continue
			}

			if err != nil {
				slog.Error("[check handler] 检查下载失败", "error", err)
				//TODO: 如果是网络问题,可以重试
				return taskrunner.PhaseResult{Err: err}
			}

			// 找到了真实 ID
			if trueID != "" {
				task.Torrent.DownloadUID = trueID

				if err := db.AddTorrentDUID(ctx, task.Torrent.Link, trueID); err != nil {
					slog.Error("[check handler] 更新 Torrent DUID 失败", "error", err)
					return taskrunner.PhaseResult{Err: err}
				}

				slog.Debug("[check handler] 获取到真实 DUID",
					"torrent", task.Torrent.Name, "duid", trueID)

				return taskrunner.PhaseResult{} // 成功
			}
		}

		return taskrunner.PhaseResult{Err: errors.New("no valid hash found")}
	}
}
```

Remove the `"goto-bangumi/internal/download"` import.

- [ ] **Step 2: Verify build**

Run: `go build ./internal/taskrunner/handlers/...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/taskrunner/handlers/check.go
git commit -m "refactor(handlers): inject DownloadClient into CheckHandler"
```

---

### Task 4: Inject `*DownloadClient` into DownloadingHandler

**Files:**
- Modify: `internal/taskrunner/handlers/downloading.go`

- [ ] **Step 1: Add `*download.DownloadClient` parameter to `NewDownloadingHandler`**

Replace the `NewDownloadingHandler` function in `internal/taskrunner/handlers/downloading.go`:

```go
// NewDownloadingHandler 创建下载监控处理器，合并进度检查和 ETA 计算
func NewDownloadingHandler(db *database.DB, dl *download.DownloadClient) taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		// 检查是否超时（4小时）
		if time.Since(task.StartTime) > 4*time.Hour {
			slog.Warn("[downloading handler] 下载超过4小时，标记为异常",
				"torrent", task.Torrent.Name,
				"duid", task.Torrent.DownloadUID,
				"elapsed", time.Since(task.StartTime))

			db.AddTorrentError(ctx, task.Torrent.Link)
			return taskrunner.PhaseResult{Err: fmt.Errorf("download timeout after 4 hours")}
		}

		// 获取种子信息
		info, err := dl.GetTorrentInfo(ctx, task.Torrent.DownloadUID)
		if err != nil {
			slog.Error("[downloading handler] 获取种子信息失败",
				"error", err, "duid", task.Torrent.DownloadUID)
			return taskrunner.PhaseResult{Err: err}
		}

		if info == nil {
			slog.Warn("[downloading handler] 种子不存在", "duid", task.Torrent.DownloadUID)
			return taskrunner.PhaseResult{Err: fmt.Errorf("torrent not found")}
		}

		// 检查是否下载完成（Completed > 0 表示已完成，为 Unix 时间戳）
		if info.Completed > 0 {
			task.Torrent.Downloaded = model.DownloadDone
			if err := db.AddTorrentDownload(ctx, task.Torrent.Link); err != nil {
				slog.Error("[downloading handler] 更新种子状态失败", "error", err)
				return taskrunner.PhaseResult{Err: err}
			}

			slog.Info("[downloading handler] 下载完成", "torrent", task.Torrent.Name)
			return taskrunner.PhaseResult{} // 成功，进入下一阶段
		}

		// 未完成，根据 ETA 自适应轮询
		interval := calculateEta(int64(info.ETA))
		slog.Debug("[downloading handler] 设置检查间隔",
			"torrent", task.Torrent.Name,
			"eta", info.ETA,
			"interval", interval)

		return taskrunner.PhaseResult{PollAfter: time.Duration(interval) * time.Second}
	}
}
```

Remove the `"goto-bangumi/internal/download"` import.

- [ ] **Step 2: Verify build**

Run: `go build ./internal/taskrunner/handlers/...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/taskrunner/handlers/downloading.go
git commit -m "refactor(handlers): inject DownloadClient into DownloadingHandler"
```

---

### Task 5: Inject `*DownloadClient` into Renamer

**Files:**
- Modify: `internal/rename/rename.go`
- Modify: `internal/rename/utils.go`

- [ ] **Step 1: Add `downloader` field to `Renamer` and update constructor**

In `internal/rename/rename.go`, update `Renamer` struct and `New`:

```go
// Renamer 封装重命名相关操作
type Renamer struct {
	db         *database.DB
	downloader *download.DownloadClient
}

// New 创建 Renamer 实例
func New(db *database.DB, dl *download.DownloadClient) *Renamer {
	return &Renamer{db: db, downloader: dl}
}
```

Replace `download.Client.GetTorrentFiles` and `download.Client.Rename` in `Rename` method:

```go
	fileList, err := r.downloader.GetTorrentFiles(ctx, torrent.DownloadUID)
```

```go
		if err := r.downloader.Rename(ctx, torrent.DownloadUID, filePath, newPath); err != nil {
```

Remove `"goto-bangumi/internal/download"` from `rename.go` imports.

- [ ] **Step 2: Update `utils.go` to use `r.downloader`**

In `internal/rename/utils.go`, replace the two `download.Client` references in `getBangumi`:

```go
	downloadInfo, err := r.downloader.GetTorrentInfo(ctx, torrent.DownloadUID)
```

```go
	relativePath, err := filepath.Rel(r.downloader.SavePath, savePath)
```

Remove `"goto-bangumi/internal/download"` from `utils.go` imports.

- [ ] **Step 3: Verify build**

Run: `go build ./internal/rename/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rename/rename.go internal/rename/utils.go
git commit -m "refactor(rename): inject DownloadClient into Renamer"
```

---

### Task 6: Delete global singleton, update `Program` to create and inject `DownloadClient`

**Files:**
- Modify: `internal/download/client.go` (delete `var Client`)
- Modify: `internal/core/program.go`

- [ ] **Step 1: Delete `var Client` from `internal/download/client.go`**

Remove:
```go
// Client 为一个全局的下载客户端实例（将在 DI 改造完成后删除）
var Client = NewDownloadClient()
```

- [ ] **Step 2: Add `downloader` field and wire everything together**

In `internal/core/program.go`, update `Program` struct:

```go
type Program struct {
	// 这里可以添加程序的全局状态和配置
	ctx        context.Context
	cancel     context.CancelFunc
	db         *database.DB
	downloader *download.DownloadClient
}
```

In `InitProgram`, replace `download.Client.Init(...)` with:

```go
	downloader := download.NewDownloadClient()
	downloader.Init(&cfg.Downloader)

	return &Program{db: db, downloader: downloader}
```

In `Start`, replace `download.Client.Login(...)` and update handler construction:

```go
func (p *Program) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	go p.downloader.Login(p.ctx)

	// 创建并启动 taskrunner
	renamer := rename.New(p.db, p.downloader)
	refresher := refresh.New(p.db)
	runner := taskrunner.New(4, 5)
	runner.Register(model.PhaseAdding, handlers.NewAddHandler(p.downloader))
	runner.Register(model.PhaseChecking, handlers.NewCheckHandler(p.db, p.downloader))
	runner.Register(model.PhaseDownloading, handlers.NewDownloadingHandler(p.db, p.downloader))
	runner.Register(model.PhaseRenaming, handlers.NewRenameHandler(p.db, renamer))
	runner.Start(p.ctx)

	// 启动调度器
	InitScheduler(p.ctx, runner, p.db, refresher)
}
```

- [ ] **Step 3: Verify full project builds**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/download/client.go internal/core/program.go
git commit -m "refactor(core): wire DownloadClient via DI, delete global singleton"
```

---

### Task 7: Update tests to use injected instances

**Files:**
- Modify: `internal/rename/utils_test.go`

- [ ] **Step 1: Replace global `download.Client` usage with constructed instance**

In `internal/rename/utils_test.go`, update `TestGetBangumi`:

Replace:
```go
	// 设置 download.Client
	download.Client.Downloader = mockDownloader
	download.Client.SavePath = mockConfig.SavePath
```

With:
```go
	// 创建独立的 download client 实例
	dlClient := download.NewDownloadClient()
	dlClient.Downloader = mockDownloader
	dlClient.SavePath = mockConfig.SavePath
```

Update the `Renamer` construction inside the test loop:
```go
			r := New(nil, dlClient)
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/rename/... -v`
Expected: All tests PASS

- [ ] **Step 3: Commit**

```bash
git add internal/rename/utils_test.go
git commit -m "test(rename): use injected DownloadClient instead of global singleton"
```

---

### Task 8: Final verification

- [ ] **Step 1: Verify no remaining references to `download.Client`**

Run: `grep -r "download\.Client" internal/`
Expected: No results

- [ ] **Step 2: Full build and test**

Run: `go build ./... && go test ./...`
Expected: All PASS

- [ ] **Step 3: Verify `go vet`**

Run: `go vet ./...`
Expected: No issues
