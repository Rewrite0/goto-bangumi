# Stateful MockDownloader Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the stateless MockDownloader with a stateful in-memory mock that supports full workflow testing (add → query → rename → delete) with automatic download progress simulation.

**Architecture:** Single struct `MockDownloader` backed by a `map[string]*mockTorrent`. Progress simulation is query-count-driven (deterministic). Preloaded data from `mock_data.go` is injected at Init time.

**Tech Stack:** Go standard library (`sync`, `fmt`, `context`), existing `model` and `apperrors` packages.

---

### File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/download/downloader/mock.go` | Rewrite | Stateful MockDownloader implementation |
| `internal/download/downloader/mock_test.go` | Create | Unit tests for MockDownloader |
| `internal/download/downloader/mock_data.go` | No change | Preloaded test data |
| `internal/download/downloader/interface.go` | No change | BaseDownloader interface |

---

### Task 1: Write mockTorrent struct and MockDownloader skeleton

**Files:**
- Modify: `internal/download/downloader/mock.go`

- [ ] **Step 1: Write the test file with basic construction test**

Create `internal/download/downloader/mock_test.go`:

```go
package downloader

import (
	"goto-bangumi/internal/model"
	"testing"
)

func newTestMock(t *testing.T) *MockDownloader {
	t.Helper()
	d := NewMockDownloader()
	err := d.Init(&model.DownloaderConfig{
		Type:     "mock",
		SavePath: "/downloads/Bangumi",
		Host:     "127.0.0.1:8080",
		Username: "admin",
		Password: "adminadmin",
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	return d
}

func TestNewMockDownloader(t *testing.T) {
	d := newTestMock(t)

	if d.torrents == nil {
		t.Fatal("torrents map should be initialized")
	}
	if d.completionThreshold != 3 {
		t.Errorf("completionThreshold = %d, want 3", d.completionThreshold)
	}
	if d.APIInterval != 100 {
		t.Errorf("APIInterval = %d, want 100", d.APIInterval)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/download/downloader/ -run TestNewMockDownloader -v`
Expected: FAIL — `d.torrents` and `d.completionThreshold` are undefined fields.

- [ ] **Step 3: Rewrite mock.go with struct and Init**

Replace the entire content of `internal/download/downloader/mock.go` with:

```go
package downloader

import (
	"context"
	"fmt"
	"sync"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/model"
)

// mockTorrent 模拟种子的内部状态
type mockTorrent struct {
	hash       string
	info       *model.TorrentDownloadInfo
	files      []string
	queryCount int
	category   string
	tags       string
}

// MockDownloader 有状态的模拟下载器，用于测试
type MockDownloader struct {
	config              *model.DownloaderConfig
	APIInterval         int
	mu                  sync.RWMutex
	torrents            map[string]*mockTorrent
	loggedIn            bool
	completionThreshold int
}

// NewMockDownloader 创建新的模拟下载器
func NewMockDownloader() *MockDownloader {
	return &MockDownloader{
		APIInterval:         100,
		completionThreshold: 3,
	}
}

// Init 初始化下载器，加载预置数据
func (d *MockDownloader) Init(config *model.DownloaderConfig) error {
	d.config = config
	d.torrents = make(map[string]*mockTorrent)

	// 加载预置数据
	for hash, info := range MockTorrentInfos {
		files := MockFiles[hash]
		d.torrents[hash] = &mockTorrent{
			hash: hash,
			info: &model.TorrentDownloadInfo{
				ETA:       info.ETA,
				SavePath:  info.SavePath,
				Completed: info.Completed,
			},
			files:      files,
			queryCount: d.completionThreshold, // 预置数据初始即完成
		}
	}
	return nil
}

// Auth 认证登录
func (d *MockDownloader) Auth(ctx context.Context) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.loggedIn = true
	return true, nil
}

// Logout 登出
func (d *MockDownloader) Logout(ctx context.Context) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.loggedIn = false
	return true, nil
}

// Add 添加种子
func (d *MockDownloader) Add(ctx context.Context, torrentInfo *model.TorrentInfo, savePath string) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	mt := &mockTorrent{
		info: &model.TorrentDownloadInfo{
			ETA:       300,
			SavePath:  savePath,
			Completed: 0,
		},
		files:      []string{fmt.Sprintf("[Mock] %s.mp4", torrentInfo.Name)},
		queryCount: 0,
	}

	hashes := make([]string, 0, 2)
	if torrentInfo.InfoHashV1 != "" {
		mt.hash = torrentInfo.InfoHashV1
		d.torrents[torrentInfo.InfoHashV1] = mt
		hashes = append(hashes, torrentInfo.InfoHashV1)
	}
	if torrentInfo.InfoHashV2 != "" {
		v2Hash := torrentInfo.InfoHashV2
		if len(v2Hash) > 40 {
			v2Hash = v2Hash[:40]
		}
		if mt.hash == "" {
			mt.hash = v2Hash
		}
		d.torrents[v2Hash] = mt // 同一个对象
		hashes = append(hashes, v2Hash)
	}
	return hashes, nil
}

// Delete 删除种子
func (d *MockDownloader) Delete(ctx context.Context, hashes []string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, h := range hashes {
		delete(d.torrents, h)
	}
	return true, nil
}

// GetTorrentInfo 获取单个种子详细信息，自动推进下载进度
func (d *MockDownloader) GetTorrentInfo(ctx context.Context, hash string) (*model.TorrentDownloadInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	mt, ok := d.torrents[hash]
	if !ok {
		return nil, nil
	}

	mt.queryCount++
	if mt.queryCount >= d.completionThreshold {
		mt.info.Completed = 1
		mt.info.ETA = 0
	} else {
		eta := 300 - mt.queryCount*100
		if eta < 0 {
			eta = 0
		}
		mt.info.ETA = eta
	}

	// 返回副本
	return &model.TorrentDownloadInfo{
		ETA:       mt.info.ETA,
		SavePath:  mt.info.SavePath,
		Completed: mt.info.Completed,
	}, nil
}

// GetTorrentFiles 获取种子文件列表
func (d *MockDownloader) GetTorrentFiles(ctx context.Context, hash string) ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	mt, ok := d.torrents[hash]
	if !ok {
		return []string{}, nil
	}
	return mt.files, nil
}

// TorrentsInfo 获取种子信息列表
func (d *MockDownloader) TorrentsInfo(ctx context.Context, statusFilter, category string, tag *string, limit int) ([]map[string]any, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []map[string]any
	for hash, mt := range d.torrents {
		// 按 category 过滤
		if category != "" && mt.category != category {
			continue
		}
		// 按 tag 过滤
		if tag != nil && mt.tags != *tag {
			continue
		}
		// 按 statusFilter 过滤
		if statusFilter != "" {
			if statusFilter == "completed" && mt.info.Completed != 1 {
				continue
			}
			if statusFilter == "downloading" && mt.info.Completed != 0 {
				continue
			}
		}

		result = append(result, map[string]any{
			"hash":      hash,
			"name":      mt.hash,
			"category":  mt.category,
			"tags":      mt.tags,
			"save_path": mt.info.SavePath,
			"completed": mt.info.Completed,
			"eta":       mt.info.ETA,
		})

		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

// CheckHash 检查种子是否存在
func (d *MockDownloader) CheckHash(ctx context.Context, hash string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if _, ok := d.torrents[hash]; ok {
		return hash, nil
	}
	return "", &apperrors.DownloadKeyError{
		Err: fmt.Errorf("种子不存在"),
		Key: hash,
	}
}

// Rename 重命名种子文件
func (d *MockDownloader) Rename(ctx context.Context, torrentHash, oldPath, newPath string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	mt, ok := d.torrents[torrentHash]
	if !ok {
		return true, nil
	}
	for i, f := range mt.files {
		if f == oldPath {
			mt.files[i] = newPath
			break
		}
	}
	return true, nil
}

// Move 移动种子到新位置
func (d *MockDownloader) Move(ctx context.Context, hashes []string, newLocation string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, h := range hashes {
		if mt, ok := d.torrents[h]; ok {
			mt.info.SavePath = newLocation
		}
	}
	return true, nil
}

// GetInterval 获取 API 调用间隔
func (d *MockDownloader) GetInterval() int {
	return d.APIInterval
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/download/downloader/ -run TestNewMockDownloader -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/download/downloader/mock.go internal/download/downloader/mock_test.go
git commit -m "feat: rewrite MockDownloader as stateful in-memory mock"
```

---

### Task 2: Test preloaded data injection

**Files:**
- Modify: `internal/download/downloader/mock_test.go`

- [ ] **Step 1: Write the test**

Append to `mock_test.go`:

```go
func TestMockDownloader_PreloadedData(t *testing.T) {
	d := newTestMock(t)

	// 预置数据应该在 Init 后存在
	if len(d.torrents) != len(MockTorrentInfos) {
		t.Errorf("torrents count = %d, want %d", len(d.torrents), len(MockTorrentInfos))
	}

	ctx := context.Background()

	// 预置 torrent 应该已完成
	for hash := range MockTorrentInfos {
		info, err := d.GetTorrentInfo(ctx, hash)
		if err != nil {
			t.Fatalf("GetTorrentInfo(%s) error: %v", hash, err)
		}
		if info.Completed != 1 {
			t.Errorf("preloaded torrent %s Completed = %d, want 1", hash, info.Completed)
		}
		if info.ETA != 0 {
			t.Errorf("preloaded torrent %s ETA = %d, want 0", hash, info.ETA)
		}
	}

	// 预置 torrent 文件应该可查
	files, err := d.GetTorrentFiles(ctx, "1317e47882474c771e29ed2271b282fbfb56e7d2")
	if err != nil {
		t.Fatalf("GetTorrentFiles error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("files count = %d, want 1", len(files))
	}
}
```

Add `"context"` to the test file imports.

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/download/downloader/ -run TestMockDownloader_PreloadedData -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/download/downloader/mock_test.go
git commit -m "test: add preloaded data test for MockDownloader"
```

---

### Task 3: Test Add → GetTorrentInfo progress simulation

**Files:**
- Modify: `internal/download/downloader/mock_test.go`

- [ ] **Step 1: Write the test**

Append to `mock_test.go`:

```go
func TestMockDownloader_AddAndProgress(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	torrentInfo := &model.TorrentInfo{
		Name:       "Test Anime - 01",
		InfoHashV1: "aaaa1111bbbb2222cccc3333dddd4444eeee5555",
		InfoHashV2: "ffff6666777788889999000011112222333344445555666677778888",
	}
	hashes, err := d.Add(ctx, torrentInfo, "/downloads/test")
	if err != nil {
		t.Fatalf("Add error: %v", err)
	}
	if len(hashes) != 2 {
		t.Fatalf("hashes count = %d, want 2", len(hashes))
	}

	// 第 1 次查询：未完成，ETA=200
	info, err := d.GetTorrentInfo(ctx, hashes[0])
	if err != nil {
		t.Fatalf("GetTorrentInfo error: %v", err)
	}
	if info.Completed != 0 {
		t.Errorf("query 1: Completed = %d, want 0", info.Completed)
	}
	if info.ETA != 200 {
		t.Errorf("query 1: ETA = %d, want 200", info.ETA)
	}

	// 第 2 次查询：未完成，ETA=100
	info, _ = d.GetTorrentInfo(ctx, hashes[0])
	if info.Completed != 0 {
		t.Errorf("query 2: Completed = %d, want 0", info.Completed)
	}
	if info.ETA != 100 {
		t.Errorf("query 2: ETA = %d, want 100", info.ETA)
	}

	// 第 3 次查询：已完成，ETA=0
	info, _ = d.GetTorrentInfo(ctx, hashes[0])
	if info.Completed != 1 {
		t.Errorf("query 3: Completed = %d, want 1", info.Completed)
	}
	if info.ETA != 0 {
		t.Errorf("query 3: ETA = %d, want 0", info.ETA)
	}

	// V2 hash 指向同一个对象，进度共享
	info, _ = d.GetTorrentInfo(ctx, hashes[1])
	if info.Completed != 1 {
		t.Errorf("v2 hash should also be completed, got Completed = %d", info.Completed)
	}
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/download/downloader/ -run TestMockDownloader_AddAndProgress -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/download/downloader/mock_test.go
git commit -m "test: add progress simulation test for MockDownloader"
```

---

### Task 4: Test Delete and CheckHash

**Files:**
- Modify: `internal/download/downloader/mock_test.go`

- [ ] **Step 1: Write the test**

Append to `mock_test.go`:

```go
func TestMockDownloader_DeleteAndCheckHash(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	hash := "1317e47882474c771e29ed2271b282fbfb56e7d2"

	// 预置数据存在
	got, err := d.CheckHash(ctx, hash)
	if err != nil {
		t.Fatalf("CheckHash should succeed for preloaded torrent, got error: %v", err)
	}
	if got != hash {
		t.Errorf("CheckHash = %q, want %q", got, hash)
	}

	// 删除
	ok, err := d.Delete(ctx, []string{hash})
	if err != nil || !ok {
		t.Fatalf("Delete failed: ok=%v, err=%v", ok, err)
	}

	// 删除后 CheckHash 应返回 DownloadKeyError
	_, err = d.CheckHash(ctx, hash)
	if err == nil {
		t.Fatal("CheckHash should return error after delete")
	}
	if !apperrors.IsKeyError(err) {
		t.Errorf("expected DownloadKeyError, got %T: %v", err, err)
	}
}
```

Add `"goto-bangumi/internal/apperrors"` to the test file imports.

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/download/downloader/ -run TestMockDownloader_DeleteAndCheckHash -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/download/downloader/mock_test.go
git commit -m "test: add delete and CheckHash test for MockDownloader"
```

---

### Task 5: Test Rename and Move

**Files:**
- Modify: `internal/download/downloader/mock_test.go`

- [ ] **Step 1: Write the test**

Append to `mock_test.go`:

```go
func TestMockDownloader_RenameAndMove(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	// 添加一个 torrent
	torrentInfo := &model.TorrentInfo{
		Name:       "Rename Test",
		InfoHashV1: "rename111122223333444455556666777788889999",
	}
	hashes, _ := d.Add(ctx, torrentInfo, "/downloads/original")
	hash := hashes[0]

	// 验证初始文件名
	files, _ := d.GetTorrentFiles(ctx, hash)
	if len(files) != 1 || files[0] != "[Mock] Rename Test.mp4" {
		t.Fatalf("initial files = %v, want [\"[Mock] Rename Test.mp4\"]", files)
	}

	// Rename
	ok, err := d.Rename(ctx, hash, "[Mock] Rename Test.mp4", "Rename Test - S01E01.mp4")
	if err != nil || !ok {
		t.Fatalf("Rename failed: ok=%v, err=%v", ok, err)
	}

	files, _ = d.GetTorrentFiles(ctx, hash)
	if files[0] != "Rename Test - S01E01.mp4" {
		t.Errorf("after rename: files[0] = %q, want %q", files[0], "Rename Test - S01E01.mp4")
	}

	// Move
	ok, err = d.Move(ctx, []string{hash}, "/downloads/moved")
	if err != nil || !ok {
		t.Fatalf("Move failed: ok=%v, err=%v", ok, err)
	}

	// 查询 3 次使其完成，然后验证 SavePath
	for i := 0; i < 3; i++ {
		d.GetTorrentInfo(ctx, hash)
	}
	info, _ := d.GetTorrentInfo(ctx, hash)
	if info.SavePath != "/downloads/moved" {
		t.Errorf("after move: SavePath = %q, want %q", info.SavePath, "/downloads/moved")
	}
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/download/downloader/ -run TestMockDownloader_RenameAndMove -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/download/downloader/mock_test.go
git commit -m "test: add rename and move test for MockDownloader"
```

---

### Task 6: Test TorrentsInfo filtering

**Files:**
- Modify: `internal/download/downloader/mock_test.go`

- [ ] **Step 1: Write the test**

Append to `mock_test.go`:

```go
func TestMockDownloader_TorrentsInfo(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	// 无过滤 — 应返回所有预置数据
	all, err := d.TorrentsInfo(ctx, "", "", nil, 0)
	if err != nil {
		t.Fatalf("TorrentsInfo error: %v", err)
	}
	if len(all) != len(MockTorrentInfos) {
		t.Errorf("all count = %d, want %d", len(all), len(MockTorrentInfos))
	}

	// 添加一个未完成的 torrent
	d.Add(ctx, &model.TorrentInfo{
		Name:       "Filtering Test",
		InfoHashV1: "filter11112222333344445555666677778888aaaa",
	}, "/downloads/filter")

	// statusFilter = "downloading" 应只返回未完成的
	downloading, _ := d.TorrentsInfo(ctx, "downloading", "", nil, 0)
	for _, t2 := range downloading {
		if t2["completed"].(int) != 0 {
			t.Errorf("downloading filter returned completed torrent: %v", t2["hash"])
		}
	}
	if len(downloading) != 1 {
		t.Errorf("downloading count = %d, want 1", len(downloading))
	}

	// statusFilter = "completed" 应只返回已完成的
	completed, _ := d.TorrentsInfo(ctx, "completed", "", nil, 0)
	if len(completed) != len(MockTorrentInfos) {
		t.Errorf("completed count = %d, want %d", len(completed), len(MockTorrentInfos))
	}

	// limit
	limited, _ := d.TorrentsInfo(ctx, "", "", nil, 1)
	if len(limited) != 1 {
		t.Errorf("limited count = %d, want 1", len(limited))
	}
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/download/downloader/ -run TestMockDownloader_TorrentsInfo -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/download/downloader/mock_test.go
git commit -m "test: add TorrentsInfo filtering test for MockDownloader"
```

---

### Task 7: Test Auth/Logout and GetInterval

**Files:**
- Modify: `internal/download/downloader/mock_test.go`

- [ ] **Step 1: Write the test**

Append to `mock_test.go`:

```go
func TestMockDownloader_AuthLogoutInterval(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	// Auth
	ok, err := d.Auth(ctx)
	if err != nil || !ok {
		t.Fatalf("Auth failed: ok=%v, err=%v", ok, err)
	}
	if !d.loggedIn {
		t.Error("loggedIn should be true after Auth")
	}

	// Logout
	ok, err = d.Logout(ctx)
	if err != nil || !ok {
		t.Fatalf("Logout failed: ok=%v, err=%v", ok, err)
	}
	if d.loggedIn {
		t.Error("loggedIn should be false after Logout")
	}

	// GetInterval
	if d.GetInterval() != 100 {
		t.Errorf("GetInterval = %d, want 100", d.GetInterval())
	}
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/download/downloader/ -run TestMockDownloader_AuthLogoutInterval -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/download/downloader/mock_test.go
git commit -m "test: add auth, logout, and interval test for MockDownloader"
```

---

### Task 8: Run all tests and verify interface compliance

**Files:**
- No new files

- [ ] **Step 1: Verify interface compliance at compile time**

The factory in `interface.go` already assigns `NewMockDownloader()` to `BaseDownloader`, so any missing method will be a compile error. Run:

Run: `go build ./internal/download/downloader/`
Expected: Success (no errors)

- [ ] **Step 2: Run all downloader tests**

Run: `go test ./internal/download/downloader/ -v`
Expected: All tests PASS

- [ ] **Step 3: Run full project tests to check for regressions**

Run: `go test ./...`
Expected: No failures in existing tests

- [ ] **Step 4: Commit (if any fixes were needed)**

Only if previous steps required changes.
