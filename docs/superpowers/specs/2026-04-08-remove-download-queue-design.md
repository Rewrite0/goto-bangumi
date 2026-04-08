# 移除 DownloadQueue，支持任意起始阶段提交 Task

## 背景

当前架构中，RSS 刷新发现新种子后，先放入 `download.DQueue`（channel-based queue），再由 `DownloadTask` scheduler 每 5 秒从中取出并提交到 `TaskRunner`。这层间接跳转已无必要——`TaskRunner.Submit()` 本身就有 channel + `TaskStore` 去重，功能完全覆盖。

同时，`model.NewTask()` 总是从 `PhaseAdding` 开始，无法支持"只需重命名"等场景直接提交一个处于 `PhaseRenaming` 的 Task。

## 目标

1. 移除 `DownloadQueue` 和 `DownloadTask` scheduler，RSS 刷新直接提交 Task 到 `TaskRunner`
2. 提供多个构造函数（`NewAddTask`、`NewRenameTask`），支持从不同阶段开始
3. 清理不再使用的 `TorrentBangumi` 模型
4. 保持无循环引用

## 变更详情

### 1. 删除文件

| 文件 | 原因 |
|------|------|
| `internal/download/download_queue.go` | 整个 DownloadQueue 不再需要 |
| `internal/task/download.go` | DownloadTask scheduler 不再需要 |

### 2. 修改 `internal/model/task.go`

- `NewTask()` 改名为 `NewAddTask(torrent, bangumi)`，语义不变（Phase = `PhaseAdding`，设 StartTime）
- 新增 `NewRenameTask(torrent, bangumi)`：
  ```go
  func NewRenameTask(torrent *Torrent, bangumi *Bangumi) *Task {
      return &Task{
          Phase:   PhaseRenaming,
          Torrent: torrent,
          Bangumi: bangumi,
      }
  }
  ```
  - 不设 `HoldingSlot`（Rename 不经过 Adding 阶段，不占流水线槽位）
  - 不设 `Guids`（Rename 不需要 hash 候选列表）
  - 不设 `StartTime`（无下载超时判断需求）

### 3. 修改 `internal/model/torrent.go`

- 删除 `TorrentBangumi` struct（仅被 `DownloadQueue` 使用）

### 4. 修改 `internal/refresh/bangumi.go`

`RefreshRSS` 签名变更：

```go
// before
func RefreshRSS(ctx context.Context, url string)

// after
func RefreshRSS(ctx context.Context, url string, runner *taskrunner.TaskRunner)
```

函数体变更：
```go
// before
go download.DQueue.Add(ctx, t, t.Bangumi)

// after
runner.Submit(model.NewAddTask(t, t.Bangumi))
```

同时需要把 `DQueue.Add` 中的 `db.CreateTorrent(ctx, torrent)` 调用移到 `RefreshRSS` 中（在 Submit 之前调用），保持种子入库逻辑不丢失。

移除 `import "goto-bangumi/internal/download"`，新增 `import "goto-bangumi/internal/taskrunner"`。

### 5. 修改 `internal/task/refresh.go`

`RSSRefreshTask` 增加 `runner` 字段：

```go
type RSSRefreshTask struct {
    interval time.Duration
    enabled  bool
    runner   *taskrunner.TaskRunner
}

func NewRSSRefreshTask(programConfig model.ProgramConfig, runner *taskrunner.TaskRunner) *RSSRefreshTask {
    // ...
    return &RSSRefreshTask{
        interval: time.Duration(interval) * time.Second,
        enabled:  true,
        runner:   runner,
    }
}
```

`Run` 方法中将 runner 传递给 `RefreshRSS`：

```go
refresh.RefreshRSS(ctx, rss.Link, t.runner)
```

### 6. 修改 `internal/core/program.go`

`InitScheduler` 中：
- 删除 `s.AddTask(task.NewDownloadTask(runner))`
- `NewRSSRefreshTask` 改为 `task.NewRSSRefreshTask(conf.Get().Program, runner)`
- 移除 `import "goto-bangumi/internal/download"`（如果不再需要）

### 7. 去重逻辑

原来由 `DQueue.InQueue`（sync.Map）做去重。移除后由 `TaskRunner.Submit()` → `TaskStore.Add()` 承担，它已经按 `torrent.Link` 去重，逻辑等价。

## 依赖方向（变更后）

```
model (无内部依赖)
  ^
  |
taskrunner (仅依赖 model)
  ^
  |
refresh (依赖 taskrunner, model, database, network)
  ^
  |
task (依赖 refresh, model, database, taskrunner)
  ^
  |
core (组装一切)
```

无循环引用。`refresh` 新增对 `taskrunner` 的依赖，方向单一向上。

## 不变的部分

- `TaskRunner` 本身（runner.go、store.go、handler.go）不需修改
- 所有 phase handler（add.go、check.go、downloading.go、rename.go）不需修改
- `download.Client` 不受影响
- `rename` 包不受影响
- 阶段推进逻辑（`nextPhase`、`advance`）不变——`NewRenameTask` 从 `PhaseRenaming` 开始，完成后推进到 `PhaseCompleted` → `PhaseEnd`，自然走完生命周期
