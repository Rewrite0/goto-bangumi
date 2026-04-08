# TaskRunner Semaphore Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the taskrunner semaphore so it represents a "pipeline slot" — acquired at PhaseAdding, released when the task exits the pipeline (PhaseEnd or PhaseFailed).

**Architecture:** Remove per-phase `needsLimit` flag and `Config` struct. Runner acquires `addingSem` when a task enters PhaseAdding and releases it via `releaseSlot()` on task exit. Task carries `HoldingSlot bool` to track whether it holds a slot.

**Tech Stack:** Go, standard library only (`sync`, `context`, `chan struct{}`)

---

### Task 1: Add `HoldingSlot` field to `model.Task`

**Files:**
- Modify: `internal/model/task.go:48-60`

- [ ] **Step 1: Add `HoldingSlot` field**

In `internal/model/task.go`, add `HoldingSlot bool` to the Task struct after `Phase`:

```go
// Task 下载任务
type Task struct {
	Mu    sync.Mutex
	Phase TaskPhase

	HoldingSlot bool // 是否持有流水线槽位

	// 业务数据
	Guids     []string  // 可能的 hash 列表
	StartTime time.Time // 开始下载时间（用于超时判断）
	ErrorMsg  string

	// 关联对象（内存引用）
	Torrent *Torrent
	Bangumi *Bangumi
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/model/...`
Expected: success (no code references the new field yet)

- [ ] **Step 3: Commit**

```bash
git add internal/model/task.go
git commit -m "feat: add HoldingSlot field to Task"
```

---

### Task 2: Simplify Runner — remove Config, needsLimit, rename semaphore

**Files:**
- Modify: `internal/taskrunner/runner.go:1-217`

- [ ] **Step 1: Write failing test for new `New()` signature**

In `internal/taskrunner/runner_test.go`, add a test at the bottom that uses the new `New(queueSize, maxConcurrency int)` signature:

```go
func TestNew_DirectParams(t *testing.T) {
	runner := New(16, 2)
	if runner == nil {
		t.Fatal("New should return non-nil runner")
	}
	if cap(runner.queue) != 16 {
		t.Errorf("queue cap = %d, want 16", cap(runner.queue))
	}
	if cap(runner.addingSem) != 2 {
		t.Errorf("addingSem cap = %d, want 2", cap(runner.addingSem))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/taskrunner/ -run TestNew_DirectParams -v`
Expected: FAIL — `New` still takes `Config`, and field is `downloadSem` not `addingSem`

- [ ] **Step 3: Rewrite runner.go**

Replace the entire `internal/taskrunner/runner.go` with:

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
	store     *TaskStore
	phases    []phaseEntry
	queue     chan *model.Task
	addingSem chan struct{} // 流水线槽位信号量
	wg        sync.WaitGroup
	cancel    context.CancelFunc
}

// New 创建任务执行器
func New(queueSize, maxConcurrency int) *TaskRunner {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	if queueSize <= 0 {
		queueSize = 64
	}
	return &TaskRunner{
		store:     NewTaskStore(),
		queue:     make(chan *model.Task, queueSize),
		addingSem: make(chan struct{}, maxConcurrency),
	}
}

// Register 注册阶段处理器
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc) {
	r.phases = append(r.phases, phaseEntry{
		phase:   phase,
		handler: handler,
	})
}

// Store 返回任务存储
func (r *TaskRunner) Store() *TaskStore {
	return r.store
}

// Start 启动 dispatcher
func (r *TaskRunner) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	r.wg.Add(1)
	go r.dispatcher(ctx)
}

// Stop 优雅关闭
func (r *TaskRunner) Stop() {
	r.cancel()
	r.wg.Wait()
}

// Submit 提交任务
func (r *TaskRunner) Submit(task *model.Task) bool {
	if !r.store.Add(task) {
		return false // 重复任务
	}
	select {
	case r.queue <- task:
		return true
	default:
		r.store.Remove(task.Torrent.Link)
		return false // 队列满
	}
}

// Cancel 取消任务
func (r *TaskRunner) Cancel(link string) {
	r.store.Remove(link)
}

// dispatcher 从队列取任务，为每个任务启动 goroutine
func (r *TaskRunner) dispatcher(ctx context.Context) {
	defer r.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-r.queue:
			// 检查任务是否已被取消（从 store 移除）
			if r.store.Get(task.Torrent.Link) == nil {
				continue
			}
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				r.process(ctx, task)
			}()
		}
	}
}

// process 处理单个任务的当前阶段
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	entry := r.entryFor(task.Phase)
	if entry == nil {
		// 没有 handler 的阶段（如 PhaseCompleted），直接推进
		r.advance(ctx, task)
		return
	}

	// Adding 阶段 acquire 流水线槽位
	if task.Phase == model.PhaseAdding {
		select {
		case r.addingSem <- struct{}{}:
			task.Mu.Lock()
			task.HoldingSlot = true
			task.Mu.Unlock()
		case <-ctx.Done():
			return
		}
	}

	result := entry.handler(ctx, task)

	if result.Err != nil {
		task.Mu.Lock()
		task.Phase = model.PhaseFailed
		task.ErrorMsg = result.Err.Error()
		task.Mu.Unlock()
		r.releaseSlot(task)
		r.store.Remove(task.Torrent.Link)
		slog.Error("[taskrunner] 任务失败",
			"torrent", task.Torrent.Name,
			"phase", task.Phase,
			"error", result.Err)
		return
	}

	if result.PollAfter > 0 {
		// 延迟重入队列，goroutine 立即结束
		time.AfterFunc(result.PollAfter, func() {
			select {
			case r.queue <- task:
			case <-ctx.Done():
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
		r.store.Remove(task.Torrent.Link)
		slog.Info("[taskrunner] 任务完成", "torrent", task.Torrent.Name)
		return
	}

	slog.Debug("[taskrunner] 阶段变更",
		"torrent", task.Torrent.Name,
		"from", oldPhase,
		"to", nextPhase)

	// 立即入队处理下一阶段
	select {
	case r.queue <- task:
	case <-ctx.Done():
	}
}

// releaseSlot 释放流水线槽位
func (r *TaskRunner) releaseSlot(task *model.Task) {
	task.Mu.Lock()
	defer task.Mu.Unlock()
	if task.HoldingSlot {
		<-r.addingSem
		task.HoldingSlot = false
	}
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
```

- [ ] **Step 4: Run the new signature test**

Run: `go test ./internal/taskrunner/ -run TestNew_DirectParams -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/taskrunner/runner.go internal/taskrunner/runner_test.go
git commit -m "refactor: simplify runner - remove Config/needsLimit, add pipeline slot semaphore"
```

---

### Task 3: Update existing tests to match new API

**Files:**
- Modify: `internal/taskrunner/runner_test.go:1-317`

The existing tests use `New(Config{...})` and `Register(..., needsLimit)` — these need updating to the new signatures. The semaphore test also needs rewriting because the semaphore now represents a pipeline slot (acquired at PhaseAdding, released at PhaseEnd/PhaseFailed), not a per-phase gate.

- [ ] **Step 1: Update helper and all existing tests**

Replace the entire `internal/taskrunner/runner_test.go` with:

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

func TestNew_DirectParams(t *testing.T) {
	runner := New(16, 2)
	if runner == nil {
		t.Fatal("New should return non-nil runner")
	}
	if cap(runner.queue) != 16 {
		t.Errorf("queue cap = %d, want 16", cap(runner.queue))
	}
	if cap(runner.addingSem) != 2 {
		t.Errorf("addingSem cap = %d, want 2", cap(runner.addingSem))
	}
}

func TestHappyPath_AllPhasesComplete(t *testing.T) {
	addHandler, addCount := successHandler()
	checkHandler, checkCount := successHandler()
	dlHandler, dlCount := successHandler()
	renameHandler, renameCount := successHandler()

	runner := New(16, 2)
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

	deadline := time.After(3 * time.Second)
	for {
		if runner.Store().Get("magnet:happy") == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("task did not complete within deadline")
		case <-time.After(10 * time.Millisecond):
		}
	}

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

	// 槽位应已释放
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

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failHandler)
	runner.Register(model.PhaseDownloading, dlHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:fail")
	runner.Submit(task)

	deadline := time.After(3 * time.Second)
	for {
		if runner.Store().Get("magnet:fail") == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("failed task was not removed from store")
		case <-time.After(10 * time.Millisecond):
		}
	}

	if task.Phase != model.PhaseFailed {
		t.Errorf("task phase = %v, want PhaseFailed", task.Phase)
	}
	if task.ErrorMsg != "boom" {
		t.Errorf("task ErrorMsg = %q, want %q", task.ErrorMsg, "boom")
	}
	if v := dlCount.Load(); v != 0 {
		t.Errorf("download handler called %d times, want 0", v)
	}
	// 槽位应已释放（即使任务失败）
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

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseDownloading, pollHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:poll")

	runner.Submit(task)

	deadline := time.After(3 * time.Second)
	for {
		if runner.Store().Get("magnet:poll") == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("task did not complete within deadline")
		case <-time.After(10 * time.Millisecond):
		}
	}

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

	runner := New(16, 2)
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

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, blockHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	task := newTestTask("magnet:cancel")
	runner.Submit(task)

	deadline := time.After(3 * time.Second)
	for called.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler was never called")
		case <-time.After(1 * time.Millisecond):
		}
	}

	runner.Cancel("magnet:cancel")

	if runner.Store().Get("magnet:cancel") != nil {
		t.Error("task should be removed from store after Cancel")
	}
}

func TestPipelineSlot_LimitsConcurrency(t *testing.T) {
	const maxConcurrency = 2
	const totalTasks = 5

	var concurrent atomic.Int32
	var maxSeen atomic.Int32
	var wg sync.WaitGroup
	wg.Add(totalTasks)

	// Adding handler 立即成功
	addHandler, _ := successHandler()

	// Downloading handler 模拟耗时操作，用于测量并发
	slowHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		cur := concurrent.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		concurrent.Add(-1)
		wg.Done()
		return PhaseResult{}
	}

	runner := New(64, maxConcurrency)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseDownloading, slowHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < totalTasks; i++ {
		task := newTestTask("magnet:sem-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("tasks did not complete within deadline")
	}

	// 因为槽位从 Adding 持有到 PhaseEnd，同时在流水线中的任务不超过 maxConcurrency
	if v := maxSeen.Load(); v > int32(maxConcurrency) {
		t.Errorf("max concurrent = %d, want <= %d", v, maxConcurrency)
	}
	if v := maxSeen.Load(); v < int32(maxConcurrency) {
		t.Logf("warning: max concurrent = %d, expected %d (may be flaky on slow CI)", v, maxConcurrency)
	}
}

func TestReleaseSlot_OnFailure(t *testing.T) {
	// 验证任务失败时槽位被正确释放，不会死锁后续任务
	addHandler, _ := successHandler()

	var failCount atomic.Int32
	failThenSucceed := func(ctx context.Context, task *model.Task) PhaseResult {
		n := failCount.Add(1)
		if n <= 2 {
			return PhaseResult{Err: errors.New("fail")}
		}
		return PhaseResult{}
	}

	runner := New(16, 1) // 只有1个槽位
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failThenSucceed)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	// 提交3个任务，前2个会在 checking 阶段失败，第3个成功
	for i := 0; i < 3; i++ {
		task := newTestTask("magnet:release-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	// 如果槽位没有正确释放，只有1个槽位会导致后续任务永远拿不到槽位
	deadline := time.After(5 * time.Second)
	for failCount.Load() < 3 {
		select {
		case <-deadline:
			t.Fatalf("only %d tasks processed, expected 3 — slot may not be released on failure", failCount.Load())
		case <-time.After(10 * time.Millisecond):
		}
	}
}
```

- [ ] **Step 2: Run all tests**

Run: `go test ./internal/taskrunner/ -v`
Expected: all tests PASS

- [ ] **Step 3: Commit**

```bash
git add internal/taskrunner/runner_test.go
git commit -m "test: update tests for pipeline slot semaphore"
```

---

### Task 4: Update program.go to match new API

**Files:**
- Modify: `internal/core/program.go:56-69`

- [ ] **Step 1: Update `Start()` method**

In `internal/core/program.go`, replace lines 61-66 with:

```go
	// 创建并启动 taskrunner
	runner := taskrunner.New(64, 4)
	runner.Register(model.PhaseAdding, handlers.NewAddHandler())
	runner.Register(model.PhaseChecking, handlers.NewCheckHandler())
	runner.Register(model.PhaseDownloading, handlers.NewDownloadingHandler())
	runner.Register(model.PhaseRenaming, handlers.NewRenameHandler())
	runner.Start(p.ctx)
```

- [ ] **Step 2: Verify full project compiles**

Run: `go build ./...`
Expected: success

- [ ] **Step 3: Run all tests**

Run: `go test ./...`
Expected: all tests PASS

- [ ] **Step 4: Commit**

```bash
git add internal/core/program.go
git commit -m "refactor: adapt program.go to new taskrunner API"
```
