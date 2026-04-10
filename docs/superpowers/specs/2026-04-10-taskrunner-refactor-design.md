# TaskRunner 重构设计：双队列 + 事件驱动调度

## Context

当前 `internal/taskrunner` 基于 channel 队列 + dispatcher goroutine 模型，存在以下问题：
- channel 队列无法跳过不可调度的任务（下载槽位满时，Renaming 任务也被阻塞）
- `addingSem` 信号量只限制 Adding 阶段，无法覆盖整个下载流水线（Adding→Checking→Downloading）
- Submit 中 channel 发送被注释掉，releaseSlot 也被注释掉，当前代码处于半成品状态
- TaskStore 作为独立类型引入了额外的锁，但功能可以合并到 runner 中

目标：用**双队列 + 事件驱动调度器**模型重写 taskrunner，实现两个独立的并发控制维度。

## 设计约束

- 纯内存，不需要持久化
- 线性流水线：Adding → Checking → Downloading → Renaming → Completed → End
- 两个独立的并发维度：
  - 总并发上限 `maxConcurrency`：同时执行 handler 的任务数
  - 下载槽位上限 `maxDownload`（默认 5）：处于 Adding→Downloading 流水线中的任务数（包括 sleep 中的）
- 任务 PollAfter sleep 后回到等待队列，保持下载槽位
- 失败不重试，直接标记失败
- 事件驱动调度，非固定间隔轮询

---

## 核心结构

### TaskRunner

```go
type TaskRunner struct {
    phases []phaseEntry

    // 状态（mu 保护）
    mu            sync.Mutex
    tasks         map[string]*model.Task  // 去重 + 查找（替代 TaskStore）
    downloadQueue []*model.Task           // Adding/Checking/Downloading 阶段
    generalQueue  []*model.Task           // Renaming 等阶段
    running       int                     // 当前正在执行 handler 的任务数
    downloadSlots int                     // 当前持有下载槽位的任务数

    // 配置
    maxConcurrency int  // 总并发上限
    maxDownload    int  // 下载槽位上限

    // 控制
    signal chan struct{}  // buffer 1，唤醒 scheduler
    wg     sync.WaitGroup
    cancel context.CancelFunc
}
```

关键变化：
- 去掉 `chan *model.Task` 队列和 `addingSem` 信号量
- 去掉 `TaskStore` 类型，功能合并为 `tasks map[string]*model.Task`
- 用两个 `[]*model.Task` slice 作为等待队列
- `running` 和 `downloadSlots` 计数器在 `mu` 保护下操作

### PhaseResult & PhaseFunc（不变）

```go
type PhaseResult struct {
    Err       error
    PollAfter time.Duration
}

type PhaseFunc func(ctx context.Context, task *model.Task) PhaseResult
```

---

## 调度机制

### Scheduler 循环

```go
func (r *TaskRunner) scheduler(ctx context.Context) {
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
```

### Schedule 调度逻辑

```go
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
```

### findRunnable 逻辑

在 downloadQueue 中查找可调度的任务：
1. 优先找 `HoldingSlot == true` 的任务（已持有槽位，只需 running < n）
2. 如果没有，且 `downloadSlots < maxDownload`，取第一个 `HoldingSlot == false` 的任务

### dispatch

```go
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
```

### 触发 signal 的时机

- `Submit()` 入队后
- 任务执行完毕（`finish` 中 running--）
- PollAfter 延迟后 `enqueue` 回来
- `releaseSlot` 释放下载槽位后

---

## 入队与资源释放

### enqueue

```go
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
```

### finish（任务执行完毕）

```go
func (r *TaskRunner) finish(task *model.Task) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.running--
    r.notify()
}
```

### releaseSlot（释放下载槽位）

```go
func (r *TaskRunner) releaseSlot(task *model.Task) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if task.HoldingSlot {
        r.downloadSlots--
        task.HoldingSlot = false
    }
    r.notify()
}
```

### needsDownloadSlot

```go
func needsDownloadSlot(phase model.TaskPhase) bool {
    return phase <= model.PhaseDownloading
}
```

---

## Submit、Cancel、Process、Advance

### Submit

```go
func (r *TaskRunner) Submit(task *model.Task) bool {
    r.mu.Lock()
    link := task.Torrent.Link
    if _, exists := r.tasks[link]; exists {
        r.mu.Unlock()
        return false
    }
    r.tasks[link] = task
    r.mu.Unlock()

    r.enqueue(task)
    return true
}
```

### Cancel

```go
func (r *TaskRunner) Cancel(link string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.tasks, link)
}
```

Cancel 只从 tasks map 移除。队列中可能残留"幽灵"任务，`process` 开头检查 `tasks[link]` 为 nil 时跳过并释放资源。

### Process

```go
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
        return
    }

    if result.PollAfter > 0 {
        time.AfterFunc(result.PollAfter, func() {
            if ctx.Err() == nil {
                r.enqueue(task)
            }
        })
        return
    }

    r.advance(ctx, task)
}
```

### Advance

```go
func (r *TaskRunner) advance(ctx context.Context, task *model.Task) {
    task.Mu.Lock()
    oldPhase := task.Phase
    nextPhase := r.nextPhase(task.Phase)
    task.Phase = nextPhase
    task.Mu.Unlock()

    if nextPhase == model.PhaseEnd {
        r.releaseSlot(task)
        r.remove(task.Torrent.Link)
        return
    }

    // 跨越 Downloading→Renaming 边界时释放下载槽位
    if needsDownloadSlot(oldPhase) && !needsDownloadSlot(nextPhase) {
        r.releaseSlot(task)
    }

    r.enqueue(task)
}
```

---

## 文件结构与变更

```
internal/taskrunner/
    runner.go       — TaskRunner 完整实现
    handler.go      — PhaseFunc、PhaseResult（不变）
    handlers/
        add.go          — 不变
        check.go        — 不变
        downloading.go  — 不变
        rename.go       — 不变
```

### 删除

- `store.go` — 功能合并到 runner

### 修改

| 文件 | 变更 |
|------|------|
| `runner.go` | 重写：双 slice 队列 + 事件驱动 scheduler；合并 store 功能 |
| `runner_test.go` | 重写：适配新 API，去掉 `runner.Store()` 引用 |
| `model/task.go` | 小改：`HoldingSlot` 和 `Mu` 保留 |
| `core/program.go` | 小改：`New(64, 4)` → `New(maxConcurrency, maxDownload)` |
| `refresh/bangumi.go` | 调整 `runner.Submit` 调用（如果引用了 Store） |

### 对外 API

```go
func New(maxConcurrency, maxDownload int) *TaskRunner
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc)
func (r *TaskRunner) Start(ctx context.Context)
func (r *TaskRunner) Stop()
func (r *TaskRunner) Submit(task *model.Task) bool
func (r *TaskRunner) Cancel(link string)
func (r *TaskRunner) Get(link string) *model.Task  // 替代 Store().Get()
```

## 验证计划

1. `go build ./...` 编译通过
2. 单元测试覆盖：
   - Happy path：任务走完全流程 Adding→End
   - PollAfter：sleep 后重新调度，保持下载槽位
   - 并发上限：n 个任务同时执行不超限
   - 下载槽位：超过 maxDownload 的任务排队等待
   - 失败释放：任务失败时正确释放 running 和 downloadSlots
   - 去重：重复 Submit 被拒绝
   - Cancel：取消后任务被跳过并释放资源
   - Renaming 不被下载槽位阻塞
3. 集成测试：完整 RSS → 下载 → 重命名链路
