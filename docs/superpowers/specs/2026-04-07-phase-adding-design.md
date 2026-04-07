# PhaseAdding：将"添加到下载"移入 TaskRunner

## Context

当前 `download.Client.Add()` 在 `DownloadTask.Run()` 中调用，位于 taskrunner 之外。这意味着添加下载的动作不受 taskrunner 的流水线管理，失败重试逻辑也与 taskrunner 的 phase 模型不一致。

**目标：** 将"添加到下载器"作为 taskrunner 的第一个阶段（`PhaseAdding`），并让 `downloadSem` 信号量仅限制这一个阶段。

## 设计

### TaskPhase 变更

在 iota 最前面插入 `PhaseAdding`，所有后续 phase 值顺移：

```go
const (
    PhaseAdding      TaskPhase = iota // 添加到下载器
    PhaseChecking                      // 检查下载是否成功添加
    PhaseDownloading                   // 下载中，等待完成
    PhaseRenaming                      // 重命名文件
    PhaseCompleted                     // 完成
    PhaseFailed                        // 失败
    PhaseEnd                           // 任务完成标志
)
```

纯内存不持久化，iota 值变化没有实际影响。

`NewTask` 初始 phase 从 `PhaseChecking` 改为 `PhaseAdding`。

### 新增 AddHandler

`internal/taskrunner/handlers/add.go`：

```go
func NewAddHandler() PhaseFunc {
    return func(ctx context.Context, task *model.Task) PhaseResult {
        guids, err := download.Client.Add(ctx, task.Torrent.Link, genSavePath(task.Bangumi))
        if err != nil {
            // handler 自己决定重试策略
            return PhaseResult{PollAfter: 5 * time.Second}
        }
        task.Guids = guids
        return PhaseResult{}
    }
}
```

重试逻辑由 handler 内部控制（次数上限、退避策略等）。

`genSavePath` 从 `internal/task/download.go` 迁移到此处。

### 信号量限制仅作用于 PhaseAdding

```go
runner.Register(model.PhaseAdding, handlers.NewAddHandler(), true)        // 唯一受限阶段
runner.Register(model.PhaseChecking, handlers.NewCheckHandler(), false)
runner.Register(model.PhaseDownloading, handlers.NewDownloadingHandler(), false)
runner.Register(model.PhaseRenaming, handlers.NewRenameHandler(), false)
```

理由：真正需要控制并发的瓶颈在"添加下载"这个动作上，check 和 downloading 只是轻量的状态查询。

### DownloadTask 简化

`DownloadTask.Run()` 不再调用 `download.Client.Add()`，简化为从 DQueue 取出后直接创建 Task 并 Submit：

```go
func (t *DownloadTask) Run(ctx context.Context) error {
    select {
    case tb := <-download.DQueue.Queue:
        download.DQueue.InQueue.Delete(tb.Torrent.Link)
        task := model.NewTask(tb.Torrent, tb.Bangumi)
        t.runner.Submit(task)
    default:
    }
    return nil
}
```

## 影响范围

| 文件 | 变更 |
|------|------|
| `internal/model/task.go` | 插入 `PhaseAdding`，`NewTask` 初始 phase 改为 `PhaseAdding`，`String()` 增加 case |
| `internal/taskrunner/handlers/add.go` | **新增**，AddHandler + genSavePath |
| `internal/task/download.go` | 删除 `download.Client.Add()` 调用和 `genSavePath`，简化 `Run()` |
| `internal/core/program.go` | 注册时加上 `PhaseAdding`，其余三个 phase 改 `needsLimit=false` |

## 验证

1. `go build ./...` 编译通过
2. 现有 runner 单元测试通过（phase 值变化不影响测试逻辑）
3. AddHandler 单元测试：成功添加、失败重试、重试超限
