# Stateful MockDownloader Design

## 目标

替换现有的无状态 `MockDownloader`，实现一个有状态的内存模拟器，满足完整工作流测试（添加→查询→重命名→删除）。

## 需求总结

- 实现 `BaseDownloader` 接口，直接替换现有 mock
- 有状态：Add 存入内存，Delete 删除，Rename/Move 修改状态
- 预置数据 + 动态数据结合
- 自动模拟下载进度（基于查询次数）
- 只覆盖正常路径，不模拟错误场景

## 内部数据结构

```go
type mockTorrent struct {
    hash       string
    info       *model.TorrentDownloadInfo  // ETA, SavePath, Completed
    files      []string                     // 文件列表
    queryCount int                          // GetTorrentInfo 被调用的次数
    category   string
    tags       string
}

type MockDownloader struct {
    config              *model.DownloaderConfig
    APIInterval         int
    mu                  sync.RWMutex
    torrents            map[string]*mockTorrent  // hash → torrent
    loggedIn            bool
    completionThreshold int  // 查询多少次后自动完成，默认 3
}
```

## 各方法实现

### Init

初始化 `torrents` map，将 `MockTorrentInfos` 和 `MockFiles` 中的预置数据注入，预置 torrent 的 `queryCount` 设为 `completionThreshold`（初始即完成状态）。`completionThreshold` 默认为 3。

### Auth / Logout

- Auth：设 `loggedIn = true`，返回 `(true, nil)`
- Logout：设 `loggedIn = false`，返回 `(true, nil)`

### Add

接收 `*model.TorrentInfo` 和 `savePath`，创建 `mockTorrent`：
- `Completed = 0`，`ETA = 300`
- `files`：生成一个假文件名 `[Mock] <torrentInfo.Name>.mp4`
- 使用 `InfoHashV1` 和 `InfoHashV2` 作为 key 存入 map（同一个 mockTorrent 对象）
- 返回 `[InfoHashV1, InfoHashV2]`（与现有 mock 行为一致）

### Delete

遍历 hashes，从 map 中删除。返回 `(true, nil)`。

### GetTorrentInfo

按 hash 查 map：
- 不存在：返回空 info（与现有行为一致）
- 存在：`queryCount++`，若 `queryCount >= completionThreshold` 则设 `Completed = 1, ETA = 0`，否则 `ETA = max(0, 300 - queryCount * 100)`
- 返回 info 的副本

### GetTorrentFiles

按 hash 查 map，返回 files 列表。不存在则返回空列表。

### TorrentsInfo

遍历 `torrents` map，按参数过滤：
- `category`：非空时只返回匹配的
- `tag`：非 nil 时只返回匹配的
- `statusFilter`：非空时根据 Completed 状态过滤（`completed` 只返回已完成的，`downloading` 只返回未完成的）
- `limit`：限制返回数量

返回 `[]map[string]any`，每个 map 包含 `hash`、`name`、`category`、`tags`、`save_path`、`completed`、`eta` 等字段。

### CheckHash

按 hash 查 map，存在返回 `(hash, nil)`，不存在返回 `("", DownloadKeyError)`。

### Rename

按 hash 查 map，在 files 中找到 oldPath 并替换为 newPath。返回 `(true, nil)`。

### Move

遍历 hashes，修改对应 torrent 的 `SavePath`。返回 `(true, nil)`。

### GetInterval

返回 `APIInterval`（默认 100ms）。

## 进度模拟

Add 产生的新 torrent 初始状态为未完成（`Completed=0, ETA=300`），每次调用 `GetTorrentInfo` 时 `queryCount` 自增，ETA 按 `max(0, 300 - queryCount * 100)` 递减。当 `queryCount >= completionThreshold`（默认 3）时，设为完成状态（`Completed=1, ETA=0`）。

预置数据的 `queryCount` 初始就等于 `completionThreshold`，所以它们从一开始就是完成状态。

这种基于查询次数而非时间的模拟方式保证了测试的确定性。

## 预置数据

保留 `mock_data.go` 中现有的 `MockTorrentInfos` 和 `MockFiles` 定义不变。Init 时读取这些数据注入到 torrents map。

## 文件变更

| 文件 | 动作 |
|------|------|
| `internal/download/downloader/mock.go` | 完全重写为有状态实现 |
| `internal/download/downloader/mock_data.go` | 保留，不变 |
| `internal/download/downloader/interface.go` | 不变 |

## 并发安全

所有读写 `torrents` map 的操作使用 `sync.RWMutex` 保护。读操作用 `RLock`，写操作用 `Lock`。
