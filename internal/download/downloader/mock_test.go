package downloader

import (
	"context"
	"testing"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/model"
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

func TestMockDownloader_PreloadedData(t *testing.T) {
	d := newTestMock(t)

	if len(d.torrents) != len(MockTorrentInfos) {
		t.Errorf("torrents count = %d, want %d", len(d.torrents), len(MockTorrentInfos))
	}

	ctx := context.Background()

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

	files, err := d.GetTorrentFiles(ctx, "1317e47882474c771e29ed2271b282fbfb56e7d2")
	if err != nil {
		t.Fatalf("GetTorrentFiles error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("files count = %d, want 1", len(files))
	}
}

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

func TestMockDownloader_DeleteAndCheckHash(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	hash := "1317e47882474c771e29ed2271b282fbfb56e7d2"

	got, err := d.CheckHash(ctx, hash)
	if err != nil {
		t.Fatalf("CheckHash should succeed for preloaded torrent, got error: %v", err)
	}
	if got != hash {
		t.Errorf("CheckHash = %q, want %q", got, hash)
	}

	ok, err := d.Delete(ctx, []string{hash})
	if err != nil || !ok {
		t.Fatalf("Delete failed: ok=%v, err=%v", ok, err)
	}

	_, err = d.CheckHash(ctx, hash)
	if err == nil {
		t.Fatal("CheckHash should return error after delete")
	}
	if !apperrors.IsKeyError(err) {
		t.Errorf("expected DownloadKeyError, got %T: %v", err, err)
	}
}

func TestMockDownloader_RenameAndMove(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	torrentInfo := &model.TorrentInfo{
		Name:       "Rename Test",
		InfoHashV1: "rename111122223333444455556666777788889999",
	}
	hashes, _ := d.Add(ctx, torrentInfo, "/downloads/original")
	hash := hashes[0]

	files, _ := d.GetTorrentFiles(ctx, hash)
	if len(files) != 1 || files[0] != "[Mock] Rename Test.mp4" {
		t.Fatalf("initial files = %v, want [\"[Mock] Rename Test.mp4\"]", files)
	}

	ok, err := d.Rename(ctx, hash, "[Mock] Rename Test.mp4", "Rename Test - S01E01.mp4")
	if err != nil || !ok {
		t.Fatalf("Rename failed: ok=%v, err=%v", ok, err)
	}

	files, _ = d.GetTorrentFiles(ctx, hash)
	if files[0] != "Rename Test - S01E01.mp4" {
		t.Errorf("after rename: files[0] = %q, want %q", files[0], "Rename Test - S01E01.mp4")
	}

	ok, err = d.Move(ctx, []string{hash}, "/downloads/moved")
	if err != nil || !ok {
		t.Fatalf("Move failed: ok=%v, err=%v", ok, err)
	}

	for i := 0; i < 3; i++ {
		d.GetTorrentInfo(ctx, hash)
	}
	info, _ := d.GetTorrentInfo(ctx, hash)
	if info.SavePath != "/downloads/moved" {
		t.Errorf("after move: SavePath = %q, want %q", info.SavePath, "/downloads/moved")
	}
}

func TestMockDownloader_TorrentsInfo(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	all, err := d.TorrentsInfo(ctx, "", "", nil, 0)
	if err != nil {
		t.Fatalf("TorrentsInfo error: %v", err)
	}
	if len(all) != len(MockTorrentInfos) {
		t.Errorf("all count = %d, want %d", len(all), len(MockTorrentInfos))
	}

	d.Add(ctx, &model.TorrentInfo{
		Name:       "Filtering Test",
		InfoHashV1: "filter11112222333344445555666677778888aaaa",
	}, "/downloads/filter")

	downloading, _ := d.TorrentsInfo(ctx, "downloading", "", nil, 0)
	for _, t2 := range downloading {
		if t2["completed"].(int) != 0 {
			t.Errorf("downloading filter returned completed torrent: %v", t2["hash"])
		}
	}
	if len(downloading) != 1 {
		t.Errorf("downloading count = %d, want 1", len(downloading))
	}

	completed, _ := d.TorrentsInfo(ctx, "completed", "", nil, 0)
	if len(completed) != len(MockTorrentInfos) {
		t.Errorf("completed count = %d, want %d", len(completed), len(MockTorrentInfos))
	}

	limited, _ := d.TorrentsInfo(ctx, "", "", nil, 1)
	if len(limited) != 1 {
		t.Errorf("limited count = %d, want 1", len(limited))
	}
}

func TestMockDownloader_AuthLogoutInterval(t *testing.T) {
	d := newTestMock(t)
	ctx := context.Background()

	ok, err := d.Auth(ctx)
	if err != nil || !ok {
		t.Fatalf("Auth failed: ok=%v, err=%v", ok, err)
	}
	if !d.loggedIn {
		t.Error("loggedIn should be true after Auth")
	}

	ok, err = d.Logout(ctx)
	if err != nil || !ok {
		t.Fatalf("Logout failed: ok=%v, err=%v", ok, err)
	}
	if d.loggedIn {
		t.Error("loggedIn should be false after Logout")
	}

	if d.GetInterval() != 100 {
		t.Errorf("GetInterval = %d, want 100", d.GetInterval())
	}
}
