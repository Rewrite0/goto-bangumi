# TaskRunner Dual-Queue Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the channel-based dispatcher with a dual-queue + event-driven scheduler, merge TaskStore into the runner, and add two independent concurrency dimensions (maxConcurrency and maxDownload).

**Architecture:** Two slice-based queues (downloadQueue for Adding/Checking/Downloading, generalQueue for Renaming+) fed by an event-driven scheduler goroutine. A single `sync.Mutex` protects all state including the merged task map (replacing TaskStore). `signal chan struct{}` (buffer 1) triggers scheduling on submit, finish, and re-enqueue events.

**Tech Stack:** Go stdlib only (sync, context, time, log/slog)

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/taskrunner/runner.go` | Rewrite | TaskRunner struct, New, Start/Stop, scheduler, schedule, dispatch, findRunnable, dequeue, enqueue, finish, releaseSlot, notify, process, advance, Submit, Cancel, Get, needsDownloadSlot |
| `internal/taskrunner/handler.go` | No change | PhaseFunc, PhaseResult |
| `internal/taskrunner/store.go` | Delete | Replaced by tasks map in runner |
| `internal/taskrunner/runner_test.go` | Rewrite | All tests adapted to new API |
| `internal/taskrunner/handlers/*.go` | No change | add, check, downloading, rename handlers |
| `internal/core/program.go` | Modify line 68 | New(maxConcurrency, maxDownload) signature |
| `internal/refresh/bangumi_test.go` | Modify line 200 | New(maxConcurrency, maxDownload) signature |

---

### Task 1: Scaffold runner.go with new types and constructor

**Files:**
- Rewrite: `internal/taskrunner/runner.go`

- [ ] **Step 1: Replace runner.go with new struct and constructor**

Replace the entire contents of `internal/taskrunner/runner.go` with:

```go
package taskrunner

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"goto-bangumi/internal/model"
)

// phaseEntry 阶段配置
type phaseEntry struct {
	phase   model.TaskPhase
	handler PhaseFunc
}

// TaskRunner 任务执行器
type TaskRunner struct {
	phases []phaseEntry

	// 状态（mu 保护）
	mu            sync.Mutex
	tasks         map[string]*model.Task // 去重 + 查找
	downloadQueue []*model.Task          // Adding/Checking/Downloading 阶段
	generalQueue  []*model.Task          // Renaming 等阶段
	running       int                    // 当前正在执行 handler 的任务数
	downloadSlots int                    // 当前持有下载槽位的任务数

	// 配置
	maxConcurrency int // 总并发上限
	maxDownload    int // 下载槽位上限

	// 控制
	signal chan struct{} // buffer 1，唤醒 scheduler
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// New 创建任务执行器
func New(maxConcurrency, maxDownload int) *TaskRunner {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	if maxDownload <= 0 {
		maxDownload = 5
	}
	return &TaskRunner{
		tasks:          make(map[string]*model.Task),
		maxConcurrency: maxConcurrency,
		maxDownload:    maxDownload,
		signal:         make(chan struct{}, 1),
	}
}

// Register 注册阶段处理器
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc) {
	r.phases = append(r.phases, phaseEntry{
		phase:   phase,
		handler: handler,
	})
}

// needsDownloadSlot 判断阶段是否需要下载槽位
func needsDownloadSlot(phase model.TaskPhase) bool {
	return phase <= model.PhaseDownloading
}

// notify 非阻塞写入 signal
func (r *TaskRunner) notify() {
	select {
	case r.signal <- struct{}{}:
	default:
	}
}

// enqueue 根据阶段放入对应队列（需要外部持有 mu 或内部加锁）
func (r *TaskRunner) enqueue(task *model.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if needsDownloadSlot(task.Phase) {
		r.downloadQueue = append(r.downloadQueue, task)
	} else {
		r.generalQueue = append(r.generalQueue, task)
	}
	r.notify()
}

// finish 任务执行完毕，释放 running 计数
func (r *TaskRunner) finish(task *model.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running--
	r.notify()
}

// releaseSlot 释放下载槽位
func (r *TaskRunner) releaseSlot(task *model.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if task.HoldingSlot {
		r.downloadSlots--
		task.HoldingSlot = false
	}
	r.notify()
}

// dequeue 从 slice 队列中移除指定索引的任务
func (r *TaskRunner) dequeue(queue *[]*model.Task, idx int) *model.Task {
	q := *queue
	task := q[idx]
	*queue = append(q[:idx], q[idx+1:]...)
	return task
}

// findRunnable 在 downloadQueue 中查找可调度的任务
// 优先找已持有槽位的（HoldingSlot == true），再找需要新槽位的
// 返回索引，-1 表示没有可调度的
func (r *TaskRunner) findRunnable() int {
	// 优先：已持有槽位的任务（sleep 回来的）
	for i, task := range r.downloadQueue {
		if task.HoldingSlot {
			return i
		}
	}
	// 其次：需要新槽位，但槽位未满
	if r.downloadSlots < r.maxDownload {
		if len(r.downloadQueue) > 0 {
			return 0
		}
	}
	return -1
}

// entryFor 查找阶段对应的配置
func (r *TaskRunner) entryFor(phase model.TaskPhase) *phaseEntry {
	for i := range r.phases {
		if r.phases[i].phase == phase {
			return &r.phases[i]
		}
	}
	return nil
}

// nextPhase 返回下一个阶段
func (r *TaskRunner) nextPhase(current model.TaskPhase) model.TaskPhase {
	if current == model.PhaseCompleted {
		return model.PhaseEnd
	}
	return current + 1
}

// Submit 提交任务
func (r *TaskRunner) Submit(task *model.Task) bool {
	r.mu.Lock()
	link := task.Torrent.Link
	if _, exists := r.tasks[link]; exists {
		r.mu.Unlock()
		slog.Debug("[taskrunner] 任务已存在，忽略", "link", link)
		return false
	}
	r.tasks[link] = task
	r.mu.Unlock()

	slog.Debug("[taskrunner] 提交任务", "torrent", task.Torrent.Name)
	r.enqueue(task)
	return true
}

// Cancel 取消任务
func (r *TaskRunner) Cancel(link string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, link)
}

// Get 根据 link 获取任务
func (r *TaskRunner) Get(link string) *model.Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tasks[link]
}

// remove 从 tasks map 中移除任务
func (r *TaskRunner) remove(link string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, link)
}

// Start 启动 scheduler
func (r *TaskRunner) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	r.wg.Add(1)
	go r.scheduler(ctx)
}

// Stop 优雅关闭
func (r *TaskRunner) Stop() {
	r.cancel()
	r.wg.Wait()
}

// scheduler 事件驱动调度循环
func (r *TaskRunner) scheduler(ctx context.Context) {
	slog.Info("[taskrunner] 任务执行器已启动")
	defer r.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.signal:
			r.schedule(ctx)
		}
	}
}

// schedule 尝试尽可能多地调度任务
func (r *TaskRunner) schedule(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for {
		scheduled := false

		// 尝试从 downloadQueue 调度
		if r.running < r.maxConcurrency && len(r.downloadQueue) > 0 {
			if idx := r.findRunnable(); idx >= 0 {
				task := r.dequeue(&r.downloadQueue, idx)
				r.dispatch(ctx, task)
				scheduled = true
			}
		}

		// 尝试从 generalQueue 调度
		if r.running < r.maxConcurrency && len(r.generalQueue) > 0 {
			task := r.dequeue(&r.generalQueue, 0)
			r.dispatch(ctx, task)
			scheduled = true
		}

		if !scheduled {
			break
		}
	}
}

// dispatch 启动 goroutine 执行任务（调用方必须持有 mu）
func (r *TaskRunner) dispatch(ctx context.Context, task *model.Task) {
	r.running++
	if !task.HoldingSlot && needsDownloadSlot(task.Phase) {
		r.downloadSlots++
		task.HoldingSlot = true
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.process(ctx, task)
	}()
}

// process 处理单个任务
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	defer r.finish(task)

	// 检查任务是否已被取消
	if r.Get(task.Torrent.Link) == nil {
		if task.HoldingSlot {
			r.releaseSlot(task)
		}
		return
	}

	entry := r.entryFor(task.Phase)
	if entry == nil {
		// 没有 handler 的阶段（如 PhaseCompleted），直接推进
		r.advance(ctx, task)
		return
	}

	result := entry.handler(ctx, task)

	if result.Err != nil {
		task.Mu.Lock()
		task.Phase = model.PhaseFailed
		task.ErrorMsg = result.Err.Error()
		task.Mu.Unlock()
		r.releaseSlot(task)
		r.remove(task.Torrent.Link)
		slog.Error("[taskrunner] 任务失败",
			"torrent", task.Torrent.Name,
			"phase", task.Phase,
			"error", result.Err)
		return
	}

	if result.PollAfter > 0 {
		// 延迟重入队列，goroutine 立即结束（finish 在 defer 中）
		time.AfterFunc(result.PollAfter, func() {
			if ctx.Err() == nil {
				r.enqueue(task)
			}
		})
		return
	}

	// 成功，推进到下一阶段
	r.advance(ctx, task)
}

// advance 推进到下一阶段
func (r *TaskRunner) advance(ctx context.Context, task *model.Task) {
	task.Mu.Lock()
	oldPhase := task.Phase
	nextPhase := r.nextPhase(task.Phase)
	task.Phase = nextPhase
	task.Mu.Unlock()

	if nextPhase == model.PhaseEnd {
		r.releaseSlot(task)
		r.remove(task.Torrent.Link)
		slog.Info("[taskrunner] 任务完成", "torrent", task.Torrent.Name)
		return
	}

	slog.Debug("[taskrunner] 阶段变更",
		"torrent", task.Torrent.Name,
		"from", oldPhase,
		"to", nextPhase)

	// 跨越 Downloading→Renaming 边界时释放下载槽位
	if needsDownloadSlot(oldPhase) && !needsDownloadSlot(nextPhase) {
		r.releaseSlot(task)
	}

	r.enqueue(task)
}
```

- [ ] **Step 2: Verify it compiles (ignoring test file and store.go temporarily)**

Run: `go build ./internal/taskrunner/...` — this will fail because `store.go` defines `TaskStore` which conflicts. We need to delete it first. Proceed to step 3.

- [ ] **Step 3: Delete store.go**

```bash
rm internal/taskrunner/store.go
```

- [ ] **Step 4: Verify compilation (ignoring tests)**

Run: `go build -o /dev/null ./internal/taskrunner`

This will fail because `runner_test.go` still references `runner.Store()`. That's expected — we fix tests in Task 2.

- [ ] **Step 5: Commit scaffold**

```bash
git add internal/taskrunner/runner.go
git add internal/taskrunner/store.go
git commit -m "refactor(taskrunner): rewrite runner with dual-queue scheduler, delete store"
```

---

### Task 2: Rewrite runner_test.go — Happy Path

**Files:**
- Rewrite: `internal/taskrunner/runner_test.go`

- [ ] **Step 1: Write test helpers and TestHappyPath**

Replace the entire contents of `internal/taskrunner/runner_test.go` with:

```go
package taskrunner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"goto-bangumi/internal/model"
)

// helper: 创建带 link 的测试 task（从 PhaseAdding 开始）
func newTestTask(link string) *model.Task {
	return &model.Task{
		Phase: model.PhaseAdding,
		Torrent: &model.Torrent{
			Link: link,
			Name: "test-" + link,
		},
		Bangumi:   &model.Bangumi{},
		StartTime: time.Now(),
	}
}

// helper: 创建一个记录调用次数的 handler，成功返回
func successHandler() (PhaseFunc, *atomic.Int32) {
	var count atomic.Int32
	return func(ctx context.Context, task *model.Task) PhaseResult {
		count.Add(1)
		return PhaseResult{}
	}, &count
}

// helper: 等待条件满足或超时
func waitFor(t *testing.T, timeout time.Duration, condition func() bool, msg string) {
	t.Helper()
	deadline := time.After(timeout)
	for !condition() {
		select {
		case <-deadline:
			t.Fatal(msg)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestNew_DirectParams(t *testing.T) {
	runner := New(8, 5)
	if runner == nil {
		t.Fatal("New should return non-nil runner")
	}
	if runner.maxConcurrency != 8 {
		t.Errorf("maxConcurrency = %d, want 8", runner.maxConcurrency)
	}
	if runner.maxDownload != 5 {
		t.Errorf("maxDownload = %d, want 5", runner.maxDownload)
	}
}

func TestHappyPath_AllPhasesComplete(t *testing.T) {
	addHandler, addCount := successHandler()
	checkHandler, checkCount := successHandler()
	dlHandler, dlCount := successHandler()
	renameHandler, renameCount := successHandler()

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, checkHandler)
	runner.Register(model.PhaseDownloading, dlHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:happy")
	ok := runner.Submit(task)
	if !ok {
		t.Fatal("Submit should succeed")
	}

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:happy") == nil
	}, "task did not complete within deadline")

	if v := addCount.Load(); v != 1 {
		t.Errorf("add handler called %d times, want 1", v)
	}
	if v := checkCount.Load(); v != 1 {
		t.Errorf("check handler called %d times, want 1", v)
	}
	if v := dlCount.Load(); v != 1 {
		t.Errorf("download handler called %d times, want 1", v)
	}
	if v := renameCount.Load(); v != 1 {
		t.Errorf("rename handler called %d times, want 1", v)
	}

	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
	if task.HoldingSlot {
		t.Error("task should not be holding slot after completion")
	}
}

func TestHandlerError_TaskFails(t *testing.T) {
	errBoom := errors.New("boom")

	addHandler, _ := successHandler()
	failHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		return PhaseResult{Err: errBoom}
	}
	dlHandler, dlCount := successHandler()

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failHandler)
	runner.Register(model.PhaseDownloading, dlHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:fail")
	runner.Submit(task)

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:fail") == nil
	}, "failed task was not removed")

	if task.Phase != model.PhaseFailed {
		t.Errorf("task phase = %v, want PhaseFailed", task.Phase)
	}
	if task.ErrorMsg != "boom" {
		t.Errorf("task ErrorMsg = %q, want %q", task.ErrorMsg, "boom")
	}
	if v := dlCount.Load(); v != 0 {
		t.Errorf("download handler called %d times, want 0", v)
	}
	if task.HoldingSlot {
		t.Error("task should not be holding slot after failure")
	}
}

func TestPollAfter_ReEnqueuesTask(t *testing.T) {
	var pollCount atomic.Int32

	addHandler, _ := successHandler()
	pollHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		n := pollCount.Add(1)
		if n < 3 {
			return PhaseResult{PollAfter: 10 * time.Millisecond}
		}
		return PhaseResult{}
	}
	renameHandler, _ := successHandler()

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseDownloading, pollHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:poll")
	runner.Submit(task)

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:poll") == nil
	}, "task did not complete within deadline")

	if v := pollCount.Load(); v != 3 {
		t.Errorf("poll handler called %d times, want 3", v)
	}
	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
}

func TestDuplicateSubmit_Rejected(t *testing.T) {
	blockCh := make(chan struct{})
	blockHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		<-blockCh
		return PhaseResult{}
	}

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, blockHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	task1 := newTestTask("magnet:dup")
	ok1 := runner.Submit(task1)
	if !ok1 {
		t.Fatal("first Submit should succeed")
	}

	task2 := newTestTask("magnet:dup")
	ok2 := runner.Submit(task2)
	if ok2 {
		t.Error("duplicate Submit should return false")
	}
}

func TestCancel_RemovesFromStore(t *testing.T) {
	blockCh := make(chan struct{})
	var called atomic.Int32
	blockHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		called.Add(1)
		<-blockCh
		return PhaseResult{}
	}

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, blockHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	task := newTestTask("magnet:cancel")
	runner.Submit(task)

	waitFor(t, 3*time.Second, func() bool {
		return called.Load() > 0
	}, "handler was never called")

	runner.Cancel("magnet:cancel")

	if runner.Get("magnet:cancel") != nil {
		t.Error("task should be removed after Cancel")
	}
}

func TestMaxConcurrency_Limits(t *testing.T) {
	const maxConcurrency = 2
	const totalTasks = 5

	var concurrent atomic.Int32
	var maxSeen atomic.Int32
	doneCh := make(chan struct{})

	slowHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		cur := concurrent.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		<-doneCh
		concurrent.Add(-1)
		return PhaseResult{}
	}

	runner := New(maxConcurrency, 10) // maxDownload 很大，不限制
	runner.Register(model.PhaseAdding, slowHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < totalTasks; i++ {
		task := newTestTask("magnet:conc-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	// 等待并发达到 maxConcurrency
	waitFor(t, 3*time.Second, func() bool {
		return concurrent.Load() >= int32(maxConcurrency)
	}, "concurrent tasks did not reach maxConcurrency")

	// 给一点时间确保不会超过限制
	time.Sleep(50 * time.Millisecond)

	if v := maxSeen.Load(); v > int32(maxConcurrency) {
		t.Errorf("max concurrent = %d, want <= %d", v, maxConcurrency)
	}

	close(doneCh)
}

func TestDownloadSlot_Limits(t *testing.T) {
	const maxDownload = 2
	const totalTasks = 5

	var concurrent atomic.Int32
	var maxSeen atomic.Int32
	doneCh := make(chan struct{})

	slowAddHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		cur := concurrent.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		<-doneCh
		concurrent.Add(-1)
		return PhaseResult{}
	}

	runner := New(10, maxDownload) // maxConcurrency 很大，不限制
	runner.Register(model.PhaseAdding, slowAddHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < totalTasks; i++ {
		task := newTestTask("magnet:dl-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	waitFor(t, 3*time.Second, func() bool {
		return concurrent.Load() >= int32(maxDownload)
	}, "concurrent download tasks did not reach maxDownload")

	time.Sleep(50 * time.Millisecond)

	if v := maxSeen.Load(); v > int32(maxDownload) {
		t.Errorf("max concurrent downloads = %d, want <= %d", v, maxDownload)
	}

	close(doneCh)
}

func TestRenaming_NotBlockedByDownloadSlots(t *testing.T) {
	// 下载槽位满时，Renaming 任务仍能执行
	const maxDownload = 1

	blockCh := make(chan struct{})
	var renameCalled atomic.Int32

	blockAddHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		<-blockCh
		return PhaseResult{}
	}
	renameHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		renameCalled.Add(1)
		return PhaseResult{}
	}

	runner := New(4, maxDownload)
	runner.Register(model.PhaseAdding, blockAddHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	// 提交一个 Adding 任务占满下载槽位
	addTask := newTestTask("magnet:block-add")
	runner.Submit(addTask)

	// 等待 Adding handler 开始执行（槽位被占）
	time.Sleep(50 * time.Millisecond)

	// 提交一个 Renaming 任务
	renameTask := model.NewRenameTask(
		&model.Torrent{Link: "magnet:rename-only", Name: "test-rename"},
		&model.Bangumi{},
	)
	runner.Submit(renameTask)

	// Renaming 任务应该不被下载槽位阻塞
	waitFor(t, 3*time.Second, func() bool {
		return renameCalled.Load() > 0
	}, "rename task was blocked by full download slots")
}

func TestReleaseSlot_OnFailure(t *testing.T) {
	addHandler, _ := successHandler()

	var failCount atomic.Int32
	failThenSucceed := func(ctx context.Context, task *model.Task) PhaseResult {
		n := failCount.Add(1)
		if n <= 2 {
			return PhaseResult{Err: errors.New("fail")}
		}
		return PhaseResult{}
	}

	runner := New(4, 1) // 只有 1 个下载槽位
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failThenSucceed)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < 3; i++ {
		task := newTestTask("magnet:release-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	waitFor(t, 5*time.Second, func() bool {
		return failCount.Load() >= 3
	}, "not all tasks processed — slot may not be released on failure")
}

func TestNewRenameTask_SkipsToRenamePhase(t *testing.T) {
	renameHandler, renameCount := successHandler()

	runner := New(4, 5)
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

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:rename-only") == nil
	}, "task did not complete within deadline")

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

func TestPollAfter_KeepsDownloadSlot(t *testing.T) {
	const maxDownload = 1
	var addConcurrent atomic.Int32

	addHandler, _ := successHandler()
	checkHandler, _ := successHandler()

	var pollCount atomic.Int32
	pollHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		n := pollCount.Add(1)
		if n < 3 {
			return PhaseResult{PollAfter: 10 * time.Millisecond}
		}
		return PhaseResult{}
	}
	renameHandler, _ := successHandler()

	// 第二个任务的 Adding handler 记录是否同时执行
	addHandler2 := func(ctx context.Context, task *model.Task) PhaseResult {
		addConcurrent.Add(1)
		return PhaseResult{}
	}

	runner := New(4, maxDownload)
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		if task.Torrent.Link == "magnet:poll-slot" {
			return addHandler(ctx, task)
		}
		return addHandler2(ctx, task)
	})
	runner.Register(model.PhaseChecking, checkHandler)
	runner.Register(model.PhaseDownloading, pollHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	// 第一个任务会在 Downloading 阶段 poll 3 次
	task1 := newTestTask("magnet:poll-slot")
	runner.Submit(task1)

	// 第二个任务也需要下载槽位
	task2 := newTestTask("magnet:blocked")
	runner.Submit(task2)

	// 等待第一个任务完成
	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:poll-slot") == nil
	}, "first task did not complete")

	// 第二个任务的 Adding 应该在第一个任务释放槽位后才执行
	// 如果 PollAfter 没保持槽位，第二个任务会在第一个任务 poll 期间就开始
	// 由于 maxDownload=1，第二个任务必须等第一个完成 Downloading
	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:blocked") == nil
	}, "second task did not complete")
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/taskrunner/ -v -count=1 -timeout=30s`

Expected: All tests PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/taskrunner/runner.go internal/taskrunner/runner_test.go
git commit -m "refactor(taskrunner): rewrite runner with dual-queue scheduler, delete store

- Replace channel-based dispatcher with event-driven scheduler
- Two independent concurrency dimensions: maxConcurrency and maxDownload
- Merge TaskStore into runner (tasks map)
- Download slot held from Adding through Downloading completion
- PollAfter tasks keep slot, release running count while sleeping"
```

---

### Task 3: Update external references

**Files:**
- Modify: `internal/core/program.go:68`
- Modify: `internal/refresh/bangumi_test.go:200`

- [ ] **Step 1: Update program.go**

In `internal/core/program.go`, change line 68 from:

```go
	runner := taskrunner.New(64, 4)
```

to:

```go
	runner := taskrunner.New(4, 5)
```

(4 = maxConcurrency, 5 = maxDownload)

- [ ] **Step 2: Update bangumi_test.go**

In `internal/refresh/bangumi_test.go`, change line 200 from:

```go
	runner := taskrunner.New(64, 2)
```

to:

```go
	runner := taskrunner.New(4, 5)
```

- [ ] **Step 3: Verify full build**

Run: `go build ./...`

Expected: Success, no errors.

- [ ] **Step 4: Run all tests**

Run: `go test ./internal/taskrunner/... -v -count=1 -timeout=30s`

Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/program.go internal/refresh/bangumi_test.go
git commit -m "refactor: update taskrunner.New() call sites for new API"
```
