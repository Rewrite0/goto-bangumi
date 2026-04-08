# Remove DownloadQueue Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the DownloadQueue indirection layer and allow Tasks to be submitted directly to TaskRunner from any starting phase.

**Architecture:** Replace `DQueue.Add()` → `DownloadTask` scheduler → `TaskRunner.Submit()` with direct `TaskRunner.Submit()` calls. Add `NewAddTask`/`NewRenameTask` constructors to support different starting phases. Clean up dead code (`DownloadQueue`, `DownloadTask`, `TorrentBangumi`).

**Tech Stack:** Go, standard library only

---

### Task 1: Add `NewAddTask` and `NewRenameTask` constructors

**Files:**
- Modify: `internal/model/task.go:64-71`
- Test: `internal/taskrunner/runner_test.go`

- [ ] **Step 1: Write test for `NewRenameTask` skipping straight to rename phase**

Add this test at the end of `internal/taskrunner/runner_test.go`:

```go
func TestNewRenameTask_SkipsToRenamePhase(t *testing.T) {
	renameHandler, renameCount := successHandler()

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		t.Error("add handler should not be called for rename task")
		return PhaseResult{}
	})
	runner.Register(model.PhaseChecking, func(ctx context.Context, task *model.Task) PhaseResult {
		t.Error("check handler should not be called for rename task")
		return PhaseResult{}
	})
	runner.Register(model.PhaseDownloading, func(ctx context.Context, task *model.Task) PhaseResult {
		t.Error("download handler should not be called for rename task")
		return PhaseResult{}
	})
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := model.NewRenameTask(
		&model.Torrent{Link: "magnet:rename-only", Name: "test-rename"},
		&model.Bangumi{},
	)
	ok := runner.Submit(task)
	if !ok {
		t.Fatal("Submit should succeed")
	}

	deadline := time.After(3 * time.Second)
	for runner.Store().Get("magnet:rename-only") != nil {
		select {
		case <-deadline:
			t.Fatal("task did not complete within deadline")
		case <-time.After(10 * time.Millisecond):
		}
	}

	if v := renameCount.Load(); v != 1 {
		t.Errorf("rename handler called %d times, want 1", v)
	}
	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
	if task.HoldingSlot {
		t.Error("rename task should never hold a slot")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd C:/Users/19457/github/goto-bangumi && go test ./internal/taskrunner/ -run TestNewRenameTask -v`

Expected: compilation error — `model.NewRenameTask` does not exist yet.

- [ ] **Step 3: Rename `NewTask` to `NewAddTask` and add `NewRenameTask`**

In `internal/model/task.go`, replace the `NewTask` function (lines 64-71) with:

```go
// NewAddTask 创建下载任务（从 PhaseAdding 开始）
func NewAddTask(torrent *Torrent, bangumi *Bangumi) *Task {
	return &Task{
		Phase:     PhaseAdding,
		StartTime: time.Now(),
		Torrent:   torrent,
		Bangumi:   bangumi,
	}
}

// NewRenameTask 创建重命名任务（从 PhaseRenaming 开始）
func NewRenameTask(torrent *Torrent, bangumi *Bangumi) *Task {
	return &Task{
		Phase:   PhaseRenaming,
		Torrent: torrent,
		Bangumi: bangumi,
	}
}
```

- [ ] **Step 4: Update `newTestTask` helper in runner_test.go to use `NewAddTask` pattern**

The existing `newTestTask` helper constructs tasks manually (not via `NewTask`), so it does not need updating. Verify no other test file calls `model.NewTask`.

Run: `cd C:/Users/19457/github/goto-bangumi && grep -r "model\.NewTask(" internal/ --include="*.go" | grep -v "_test.go"`

This should show only `internal/task/download.go:54` — which we will delete in Task 3.

- [ ] **Step 5: Run the new test to verify it passes**

Run: `cd C:/Users/19457/github/goto-bangumi && go test ./internal/taskrunner/ -run TestNewRenameTask -v`

Expected: PASS — the rename task starts at PhaseRenaming, only the rename handler is called, task reaches PhaseEnd.

- [ ] **Step 6: Run all taskrunner tests**

Run: `cd C:/Users/19457/github/goto-bangumi && go test ./internal/taskrunner/ -v`

Expected: all tests PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/model/task.go internal/taskrunner/runner_test.go
git commit -m "feat: add NewAddTask and NewRenameTask constructors, rename NewTask to NewAddTask"
```

---

### Task 2: Wire `TaskRunner` into RSS refresh path

**Files:**
- Modify: `internal/refresh/bangumi.go:48-61`
- Modify: `internal/task/refresh.go:16-19,22-28,69`
- Modify: `internal/core/program.go:78-93`

- [ ] **Step 1: Update `RefreshRSS` to accept and use `TaskRunner`**

In `internal/refresh/bangumi.go`, change the imports and `RefreshRSS` function:

Replace the import block:
```go
import (
	"context"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
)
```

With:
```go
import (
	"context"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/taskrunner"
)
```

Replace the `RefreshRSS` function (lines 48-61):
```go
func RefreshRSS(ctx context.Context, url string, runner *taskrunner.TaskRunner) {
	torrents := getTorrents(ctx, url)
	db := database.GetDB()
	for _, t := range torrents {
		metaData, err := db.GetBangumiParseByTitle(ctx, t.Name)
		if err != nil {
			continue
		}
		if FilterTorrent(t, metaData.IncludeFilter, metaData.ExcludeFilter) {
			t.Bangumi = metaData
			_ = db.CreateTorrent(ctx, t)
			runner.Submit(model.NewAddTask(t, t.Bangumi))
		}
	}
}
```

Key changes:
- Added `runner *taskrunner.TaskRunner` parameter
- Replaced `go download.DQueue.Add(ctx, t, t.Bangumi)` with `db.CreateTorrent` + `runner.Submit`
- `db.CreateTorrent` is moved here from `DQueue.Add` to preserve the torrent-creation side effect
- Removed `download` import, added `taskrunner` import

- [ ] **Step 2: Update `RSSRefreshTask` to hold and pass `TaskRunner`**

In `internal/task/refresh.go`, update the struct and constructor:

Replace the struct definition and constructor (lines 16-32):
```go
// RSSRefreshTask RSS 刷新任务
type RSSRefreshTask struct {
	interval time.Duration
	enabled  bool
	runner   *taskrunner.TaskRunner
}

// NewRSSRefreshTask 创建 RSS 刷新任务
func NewRSSRefreshTask(programConfig model.ProgramConfig, runner *taskrunner.TaskRunner) *RSSRefreshTask {
	interval := programConfig.RssTime

	task := &RSSRefreshTask{
		interval: time.Duration(interval) * time.Second,
		enabled:  true,
		runner:   runner,
	}

	slog.Debug("[task rss]创建 RSS 刷新任务", "间隔", task.interval)
	return task
}
```

Update the `Run` method — replace line 69:
```go
refresh.RefreshRSS(ctx, rss.Link)
```
With:
```go
refresh.RefreshRSS(ctx, rss.Link, t.runner)
```

Add the `taskrunner` import to the import block:
```go
import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/refresh"
	"goto-bangumi/internal/taskrunner"
)
```

- [ ] **Step 3: Update `InitScheduler` in `core/program.go`**

Replace lines 87-88:
```go
	s.AddTask(task.NewRSSRefreshTask(conf.Get().Program))
	s.AddTask(task.NewDownloadTask(runner))
```
With:
```go
	s.AddTask(task.NewRSSRefreshTask(conf.Get().Program, runner))
```

Remove `"goto-bangumi/internal/download"` from the imports if it is no longer used (it is still used on line 58: `download.Client.Login`), so **keep the import**.

- [ ] **Step 4: Verify compilation**

Run: `cd C:/Users/19457/github/goto-bangumi && go build ./...`

Expected: compiles successfully. The deleted `DownloadTask` references are gone, `RefreshRSS` now takes 3 args, all callers updated.

- [ ] **Step 5: Run all tests**

Run: `cd C:/Users/19457/github/goto-bangumi && go test ./internal/taskrunner/ ./internal/refresh/ ./internal/task/ -v`

Expected: all tests PASS. Note: `bangumi_test.go` does not call `RefreshRSS`, so it is unaffected.

- [ ] **Step 6: Commit**

```bash
git add internal/refresh/bangumi.go internal/task/refresh.go internal/core/program.go
git commit -m "refactor: wire TaskRunner directly into RSS refresh, bypass DownloadQueue"
```

---

### Task 3: Delete `DownloadQueue`, `DownloadTask`, and `TorrentBangumi`

**Files:**
- Delete: `internal/download/download_queue.go`
- Delete: `internal/task/download.go`
- Modify: `internal/model/torrent.go:8-13`

- [ ] **Step 1: Delete `download_queue.go`**

```bash
cd C:/Users/19457/github/goto-bangumi && rm internal/download/download_queue.go
```

- [ ] **Step 2: Delete `download.go` (DownloadTask scheduler)**

```bash
cd C:/Users/19457/github/goto-bangumi && rm internal/task/download.go
```

- [ ] **Step 3: Remove `TorrentBangumi` from `torrent.go`**

In `internal/model/torrent.go`, delete lines 8-13:

```go
// TorrentBangumi 种子和番剧关联模型
// 用于下载时传递
type TorrentBangumi struct {
	Bangumi *Bangumi
	Torrent *Torrent
}
```

- [ ] **Step 4: Verify compilation**

Run: `cd C:/Users/19457/github/goto-bangumi && go build ./...`

Expected: compiles successfully — no remaining references to `DQueue`, `DownloadQueue`, `DownloadTask`, or `TorrentBangumi`.

- [ ] **Step 5: Run full test suite**

Run: `cd C:/Users/19457/github/goto-bangumi && go test ./... -count=1`

Expected: all tests PASS.

- [ ] **Step 6: Verify no dangling references**

Run: `cd C:/Users/19457/github/goto-bangumi && grep -r "DQueue\|DownloadQueue\|DownloadTask\|TorrentBangumi" internal/ --include="*.go"`

Expected: no output — all references are cleaned up.

- [ ] **Step 7: Commit**

```bash
git add -A internal/download/download_queue.go internal/task/download.go internal/model/torrent.go
git commit -m "refactor: remove DownloadQueue, DownloadTask scheduler, and TorrentBangumi model"
```
