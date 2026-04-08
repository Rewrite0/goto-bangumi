# TaskRunner 信号量重构设计

## 背景

当前 runner 使用 `needsLimit` per-phase 标记 + 共享 `downloadSem` 控制并发，但实际只有 PhaseAdding 需要限制。且信号量在 handler 执行完就释放，不符合实际需求——应该从 Adding 获取到任务退出流水线（Renaming 完成或失败）才释放，限制的是同时在流水线中的任务数量。

## 设计

### 核心改动

信号量代表"流水线槽位"：Adding 阶段 acquire，任务退出流水线时 release（PhaseFailed 或 PhaseEnd）。

### Runner 变更

```go
type TaskRunner struct {
    store     *TaskStore
    phases    []phaseEntry
    queue     chan *model.Task
    addingSem chan struct{}      // 流水线槽位信号量
    wg        sync.WaitGroup
    cancel    context.CancelFunc
}

// 删除 Config struct / DefaultConfig()
func New(queueSize, maxConcurrency int) *TaskRunner

type phaseEntry struct {
    phase   model.TaskPhase
    handler PhaseFunc
    // 删除 needsLimit
}

// Register 简化为两参数
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc)
```

### process() 逻辑

```go
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
    entry := r.entryFor(task.Phase)
    if entry == nil {
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
        slog.Error(...)
        return
    }

    if result.PollAfter > 0 {
        time.AfterFunc(result.PollAfter, func() { ... })
        return
    }

    r.advance(ctx, task)
}
```

### advance() 释放逻辑

```go
func (r *TaskRunner) advance(ctx context.Context, task *model.Task) {
    task.Mu.Lock()
    nextPhase := r.nextPhase(task.Phase)
    task.Phase = nextPhase
    task.Mu.Unlock()

    if nextPhase == model.PhaseEnd {
        r.releaseSlot(task)
        r.store.Remove(task.Torrent.Link)
        slog.Info(...)
        return
    }

    select {
    case r.queue <- task:
    case <-ctx.Done():
    }
}
```

### releaseSlot 辅助方法

```go
func (r *TaskRunner) releaseSlot(task *model.Task) {
    task.Mu.Lock()
    defer task.Mu.Unlock()
    if task.HoldingSlot {
        <-r.addingSem
        task.HoldingSlot = false
    }
}
```

### Task 变更

```go
type Task struct {
    Mu    sync.Mutex
    Phase TaskPhase

    HoldingSlot bool // 是否持有流水线槽位

    Guids     []string
    StartTime time.Time
    ErrorMsg  string

    Torrent *Torrent
    Bangumi *Bangumi
}
```

### program.go 变更

```go
runner := taskrunner.New(64, 4)
runner.Register(model.PhaseAdding, handlers.NewAddHandler())
runner.Register(model.PhaseChecking, handlers.NewCheckHandler())
runner.Register(model.PhaseDownloading, handlers.NewDownloadingHandler())
runner.Register(model.PhaseRenaming, handlers.NewRenameHandler())
runner.Start(p.ctx)
```

## 涉及文件

| 文件 | 变更 |
|------|------|
| `internal/taskrunner/runner.go` | 删除 Config/DefaultConfig，New 改签名，删除 needsLimit，重写 process() 信号量逻辑，新增 releaseSlot() |
| `internal/model/task.go` | 新增 `HoldingSlot bool` |
| `internal/core/program.go` | 适配 New() 和 Register() 新签名 |
