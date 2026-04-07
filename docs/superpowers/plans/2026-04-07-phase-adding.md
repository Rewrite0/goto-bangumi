# PhaseAdding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move "add to downloader" into the taskrunner as the first pipeline phase (`PhaseAdding`), with download semaphore only limiting this phase.

**Architecture:** Insert `PhaseAdding` before `PhaseChecking` in the iota enum. Create `AddHandler` that calls `download.Client.Add()` with handler-controlled retry. Simplify `DownloadTask.Run()` to only create and submit tasks. Change registration so only `PhaseAdding` has `needsLimit=true`.

**Tech Stack:** Go, existing taskrunner framework

---

### Task 1: Add `PhaseAdding` to TaskPhase enum

**Files:**
- Modify: `internal/model/task.go:11-17` (iota block)
- Modify: `internal/model/task.go:20-37` (String method)
- Modify: `internal/model/task.go:59-66` (NewTask)

- [ ] **Step 1: Insert `PhaseAdding` as first iota value and update `String()`**

In `internal/model/task.go`, change the const block to insert `PhaseAdding` at the top:

```go
const (
	PhaseAdding      TaskPhase = iota // 添加到下载器
	PhaseChecking                     // 检查下载是否成功添加
	PhaseDownloading                  // 下载中，等待完成
	PhaseRenaming                     // 重命名文件
	PhaseCompleted                    // 完成
	PhaseFailed                       // 失败
	PhaseEnd                          // 任务完成标志
)
```

Add the `"adding"` case to `String()`:

```go
func (p TaskPhase) String() string {
	switch p {
	case PhaseAdding:
		return "adding"
	case PhaseChecking:
		return "checking"
	case PhaseDownloading:
		return "downloading"
	case PhaseRenaming:
		return "renaming"
	case PhaseCompleted:
		return "completed"
	case PhaseFailed:
		return "failed"
	case PhaseEnd:
		return "end"
	default:
		return "unknown"
	}
}
```

- [ ] **Step 2: Change `NewTask` initial phase to `PhaseAdding`**

In `internal/model/task.go`, change `NewTask`:

```go
func NewTask(torrent *Torrent, bangumi *Bangumi) *Task {
	return &Task{
		Phase:     PhaseAdding,
		StartTime: time.Now(),
		Torrent:   torrent,
		Bangumi:   bangumi,
	}
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/model/...`
Expected: SUCCESS (other packages that reference PhaseChecking by name won't break — only iota values shift)

- [ ] **Step 4: Commit**

```bash
git add internal/model/task.go
git commit -m "feat: add PhaseAdding as first TaskPhase"
```

---

### Task 2: Update existing tests for new initial phase

**Files:**
- Modify: `internal/taskrunner/runner_test.go:17-24` (newTestTask helper)

The existing `newTestTask` helper creates tasks with `Phase: model.PhaseChecking`. Since `PhaseAdding` is now the first phase but tests don't register an AddHandler, tests that start from `PhaseChecking` need to explicitly set the phase.

- [ ] **Step 1: Update `newTestTask` to set `PhaseChecking` explicitly**

The helper currently sets `Phase: model.PhaseChecking`. After our change, `NewTask` defaults to `PhaseAdding`, but this test helper manually constructs the struct so it already explicitly sets `PhaseChecking`. Verify this still compiles — `model.PhaseChecking` is referenced by name, not by value, so it's fine.

Run: `go test ./internal/taskrunner/ -run . -count=1 -v`
Expected: All 6 tests PASS. The iota values shifted but tests reference phases by name (`model.PhaseChecking`, `model.PhaseEnd`, etc.), so behavior is unchanged.

- [ ] **Step 2: Commit (if any changes were needed)**

```bash
git add internal/taskrunner/runner_test.go
git commit -m "test: update runner tests for PhaseAdding enum shift"
```

---

### Task 3: Create AddHandler with genSavePath

**Files:**
- Create: `internal/taskrunner/handlers/add.go`
- Modify: `internal/task/download.go:76-84` (remove genSavePath)

- [ ] **Step 1: Write the test for AddHandler**

Create `internal/taskrunner/handlers/add_test.go`. Since `AddHandler` calls `download.Client.Add()` which is a global, we need to use the existing mock downloader. Check how `check.go` handler tests work — if there are none, write a focused unit test using the mock.

Actually, looking at the codebase, handlers call global `download.Client` directly. For now, write the handler first, then test via integration. The AddHandler is simple enough.

- [ ] **Step 2: Create `internal/taskrunner/handlers/add.go`**

```go
package handlers

import (
	"context"
	"log/slog"
	"path/filepath"
	"strconv"
	"time"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/taskrunner"
)

// NewAddHandler 创建添加下载处理器，将种子添加到下载器
func NewAddHandler() taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		savePath := genSavePath(task.Bangumi)
		guids, err := download.Client.Add(ctx, task.Torrent.Link, savePath)
		if err != nil {
			slog.Warn("[add handler] 添加下载失败，稍后重试",
				"torrent", task.Torrent.Name, "error", err)
			return taskrunner.PhaseResult{PollAfter: 5 * time.Second}
		}

		task.Guids = guids
		slog.Debug("[add handler] 添加下载成功",
			"torrent", task.Torrent.Name, "guids", guids)
		return taskrunner.PhaseResult{}
	}
}

// genSavePath 根据番剧信息生成保存路径
func genSavePath(bangumi *model.Bangumi) string {
	folder := bangumi.OfficialTitle
	if bangumi.Year != "" {
		folder += " (" + bangumi.Year + ")"
	}
	season := "Season " + strconv.Itoa(bangumi.Season)
	return filepath.Join(folder, season)
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/taskrunner/handlers/...`
Expected: SUCCESS

- [ ] **Step 4: Commit**

```bash
git add internal/taskrunner/handlers/add.go
git commit -m "feat: add AddHandler for PhaseAdding"
```

---

### Task 4: Simplify DownloadTask — remove Add call and genSavePath

**Files:**
- Modify: `internal/task/download.go`

- [ ] **Step 1: Simplify `DownloadTask.Run()` and remove `genSavePath`**

Replace `internal/task/download.go` contents with:

```go
package task

import (
	"context"
	"log/slog"
	"time"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/taskrunner"
)

// DownloadTask 下载任务
type DownloadTask struct {
	interval time.Duration
	enabled  bool
	runner   *taskrunner.TaskRunner
}

// NewDownloadTask 创建下载任务
func NewDownloadTask(runner *taskrunner.TaskRunner) *DownloadTask {
	return &DownloadTask{
		interval: 5 * time.Second,
		enabled:  true,
		runner:   runner,
	}
}

// Name 返回任务名称
func (t *DownloadTask) Name() string {
	return "下载任务"
}

// Interval 返回执行间隔
func (t *DownloadTask) Interval() time.Duration {
	return t.interval
}

// Enable 返回是否启用
func (t *DownloadTask) Enable() bool {
	return t.enabled
}

// Run 从队列取出种子，创建任务提交到 taskrunner
func (t *DownloadTask) Run(ctx context.Context) error {
	select {
	case tb := <-download.DQueue.Queue:
		torrent := tb.Torrent
		bangumi := tb.Bangumi
		download.DQueue.InQueue.Delete(torrent.Link)

		slog.Debug("[download task] 提交任务到 taskrunner", "Name", torrent.Name)
		task := model.NewTask(torrent, bangumi)
		t.runner.Submit(task)
	default:
	}
	return nil
}
```

Key changes:
- Removed `download.Client.Add()` call — now handled by `AddHandler`
- Removed `genSavePath()` — moved to `handlers/add.go`
- Removed imports: `"path/filepath"`, `"strconv"`
- `NewTask` now starts at `PhaseAdding` (from Task 1), so the runner's `AddHandler` will call `download.Client.Add()`

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/task/...`
Expected: SUCCESS

- [ ] **Step 3: Commit**

```bash
git add internal/task/download.go
git commit -m "refactor: simplify DownloadTask, move Add logic to AddHandler"
```

---

### Task 5: Update phase registration in program.go

**Files:**
- Modify: `internal/core/program.go:61-64`

- [ ] **Step 1: Register `PhaseAdding` and change needsLimit flags**

In `internal/core/program.go`, replace the runner registration block (lines 61-64):

```go
	// 创建并启动 taskrunner
	runner := taskrunner.New(taskrunner.DefaultConfig())
	runner.Register(model.PhaseAdding, handlers.NewAddHandler(), true)          // 唯一受限阶段
	runner.Register(model.PhaseChecking, handlers.NewCheckHandler(), false)     // 轻量查询
	runner.Register(model.PhaseDownloading, handlers.NewDownloadingHandler(), false) // 轻量轮询
	runner.Register(model.PhaseRenaming, handlers.NewRenameHandler(), false)    // 本地文件操作
	runner.Start(p.ctx)
```

- [ ] **Step 2: Verify full project compilation**

Run: `go build ./...`
Expected: SUCCESS

- [ ] **Step 3: Commit**

```bash
git add internal/core/program.go
git commit -m "feat: register PhaseAdding, limit semaphore to adding phase only"
```

---

### Task 6: Run all tests and verify

- [ ] **Step 1: Run all tests**

Run: `go test ./... -count=1`
Expected: All tests PASS

- [ ] **Step 2: Verify runner tests specifically**

Run: `go test ./internal/taskrunner/ -run . -count=1 -v`
Expected: All 6 tests PASS — they reference phases by name and don't depend on iota values

- [ ] **Step 3: Final commit (if any fixes needed)**

If any test adjustments were needed, commit them:

```bash
git add -A
git commit -m "fix: adjust tests for PhaseAdding changes"
```
