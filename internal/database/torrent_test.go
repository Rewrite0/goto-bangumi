package database

import (
	"context"
	"testing"

	"goto-bangumi/internal/model"
)

// TestAddTorrent 测试添加种子
// 刷新的时候可以把 rss id 给加进去( 不一定有 rss id, collection 的是没有的)
// rename 的时候可以把 EpisodeMetadata 更新进去?
// 没想到有什么用呀, episode 可以是多个,感觉也不好统计更新了几集
// torrent 也不一定会有 BangumiID, 因为Collection
func TestAddTorrent(t *testing.T) {
	testdb := ":memory:"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// 测试数据来源于 mikanani.me 真实种子数据
	torrents := []model.Torrent{
		{
			Link:        "https://mikanani.me/Download/20250829/1b13cab156276b2d29db032e12ee5548afb3f847.torrent",
			Name:        "[ANi]  卡片战斗!! 先导者 Divinez 第四季「DELUXE 决胜篇」 - 07 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     false,
			DownloadUID: "1b13cab156276b2d29db032e12ee5548afb3f847",
			Homepage:    "https://mikanani.me/Home/Episode/1b13cab156276b2d29db032e12ee5548afb3f847",
		},
		{
			Link:        "https://mikanani.me/Download/20250829/19f198e968610981f93699397657cf5126cc8cdd.torrent",
			Name:        "[ANi] The Water Magician /  水属性的魔法师 - 08 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "19f198e968610981f93699397657cf5126cc8cdd",
			Homepage:    "https://mikanani.me/Home/Episode/19f198e968610981f93699397657cf5126cc8cdd",
		},
		{
			Link:        "https://mikanani.me/Download/20250829/a2c46c18f5ad6a7482dd124892017a662794e81f.torrent",
			Name:        "[ANi] Dan Da Dan S02 /  胆大党 第二季 - 21 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "a2c46c18f5ad6a7482dd124892017a662794e81f",
			Homepage:    "https://mikanani.me/Home/Episode/a2c46c18f5ad6a7482dd124892017a662794e81f",
		},
		{
			Link:        "https://mikanani.me/Download/20250828/515e32e6165f4fbfb77d8c4a84bde62422e70661.torrent",
			Name:        "[ANi] Dr STONE S04 /  Dr.STONE 新石纪 第四季 - 20 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "515e32e6165f4fbfb77d8c4a84bde62422e70661",
			Homepage:    "https://mikanani.me/Home/Episode/515e32e6165f4fbfb77d8c4a84bde62422e70661",
		},
		{
			Link:        "https://mikanani.me/Download/20250827/49c1f67d3c239ae061147b978999838483a65181.torrent",
			Name:        "[ANi] Tate no Yuusha no Nariagari /  盾之勇者成名录 Season 4 - 08 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "49c1f67d3c239ae061147b978999838483a65181",
			Homepage:    "https://mikanani.me/Home/Episode/49c1f67d3c239ae061147b978999838483a65181",
		},
		{
			Link:        "https://mikanani.me/Download/20250828/a59275b5fa5ede5124d8fc659225b2e1d3f4df07.torrent",
			Name:        "[ANi] Captivated by You /  为你著迷 - 02 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "a59275b5fa5ede5124d8fc659225b2e1d3f4df07",
			Homepage:    "https://mikanani.me/Home/Episode/a59275b5fa5ede5124d8fc659225b2e1d3f4df07",
		},
		{
			Link:        "https://mikanani.me/Download/20250828/c998cec3cbbba8d10eba2b310bbeb3f21ca58316.torrent",
			Name:        "[ANi]  阴阳回天 Re：Birth Verse - 09 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "c998cec3cbbba8d10eba2b310bbeb3f21ca58316",
			Homepage:    "https://mikanani.me/Home/Episode/c998cec3cbbba8d10eba2b310bbeb3f21ca58316",
		},
		{
			Link:        "https://mikanani.me/Download/20250828/e63306dfe6908739c03a4f318f75118fa5023de9.torrent",
			Name:        "[ANi] Uchūjin MūMū /  外星人姆姆 - 21 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "e63306dfe6908739c03a4f318f75118fa5023de9",
			Homepage:    "https://mikanani.me/Home/Episode/e63306dfe6908739c03a4f318f75118fa5023de9",
		},
		{
			Link:        "https://mikanani.me/Download/20250827/dbe13b6e2c4c8668527b5c9b93aac1ced2becb1c.torrent",
			Name:        "[ANi] Turkey Time to Strike /  保龄球少女！ - 08 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "dbe13b6e2c4c8668527b5c9b93aac1ced2becb1c",
			Homepage:    "https://mikanani.me/Home/Episode/dbe13b6e2c4c8668527b5c9b93aac1ced2becb1c",
		},
		{
			Link:        "https://mikanani.me/Download/20250826/a9247f1bbc55da3150a24d0f9bcffd29e97fefe5.torrent",
			Name:        "[ANi]  涅库罗诺美子的宇宙恐怖秀 - 03 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			Downloaded:  1,
			Renamed:     true,
			DownloadUID: "a9247f1bbc55da3150a24d0f9bcffd29e97fefe5",
			Homepage:    "https://mikanani.me/Home/Episode/a9247f1bbc55da3150a24d0f9bcffd29e97fefe5",
		},
	}

	ctx := context.Background()
	for i, torrent := range torrents {
		err := db.CreateTorrent(ctx, &torrent)
		if err != nil {
			t.Errorf("Failed to create torrent %d: %v", i, err)
		}
	}
	t.Logf("Successfully added %d torrents", len(torrents))
}

func TestDeleteTorrent(t *testing.T) {
	testdb := ":memory:"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	ctx := context.Background()

	// 先创建一个种子
	torrent := model.Torrent{
		Link:        "https://mikanani.me/Download/20250829/test_delete.torrent",
		Name:        "[ANi] Test Delete Torrent",
		Downloaded:  0,
		Renamed:     false,
		DownloadUID: "test_delete_uid",
		Homepage:    "https://mikanani.me/Home/Episode/test_delete",
	}

	err = db.CreateTorrent(ctx, &torrent)
	if err != nil {
		t.Fatalf("Failed to create torrent: %v", err)
	}

	// 删除种子
	err = db.DeleteTorrent(ctx, torrent.Link)
	if err != nil {
		t.Errorf("Failed to delete torrent: %v", err)
	}

	// 验证种子已被删除
	var count int64
	db.Model(&model.Torrent{}).Where("link = ?", torrent.Link).Count(&count)
	if count != 0 {
		t.Errorf("Torrent should be deleted, but found %d records", count)
	}

	t.Log("Successfully deleted torrent")
}

func TestAddTorrentDownload(t *testing.T) {
	testdb := ":memory:"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	ctx := context.Background()

	// 创建一个未下载的种子
	torrent := model.Torrent{
		Link:       "https://mikanani.me/Download/20250829/test_download.torrent",
		Name:       "[ANi] Test Download Torrent",
		Downloaded: 0,
		Renamed:    false,
		Homepage:   "https://mikanani.me/Home/Episode/test_download",
	}

	err = db.CreateTorrent(ctx, &torrent)
	if err != nil {
		t.Fatalf("Failed to create torrent: %v", err)
	}

	// 标记为已下载
	err = db.AddTorrentDownload(ctx, torrent.Link)
	if err != nil {
		t.Errorf("Failed to mark torrent as downloaded: %v", err)
	}

	// 验证已标记为下载
	var updated model.Torrent
	db.Where("link = ?", torrent.Link).First(&updated)
	if updated.Downloaded != model.DownloadDone {
		t.Errorf("Torrent should be marked as downloaded, got %d", updated.Downloaded)
	}

	t.Log("Successfully marked torrent as downloaded")
}

func TestTorrentRenamed(t *testing.T) {
	testdb := ":memory:"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	ctx := context.Background()

	// 创建一个未重命名的种子
	torrent := model.Torrent{
		Link:       "https://mikanani.me/Download/20250829/test_rename.torrent",
		Name:       "[ANi] Test Rename Torrent",
		Downloaded: 1,
		Renamed:    false,
		Homepage:   "https://mikanani.me/Home/Episode/test_rename",
	}

	err = db.CreateTorrent(ctx, &torrent)
	if err != nil {
		t.Fatalf("Failed to create torrent: %v", err)
	}

	// 标记为已重命名
	err = db.TorrentRenamed(ctx, torrent.Link)
	if err != nil {
		t.Errorf("Failed to mark torrent as renamed: %v", err)
	}

	// 验证已标记为重命名
	var updated model.Torrent
	db.Where("link = ?", torrent.Link).First(&updated)
	if !updated.Renamed {
		t.Errorf("Torrent should be marked as renamed")
	}

	t.Log("Successfully marked torrent as renamed")
}

func TestAddTorrentDUID(t *testing.T) {
	testdb := ":memory:"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	ctx := context.Background()

	// 创建一个没有 DownloadUID 的种子
	torrent := model.Torrent{
		Link:        "https://mikanani.me/Download/20250829/test_duid.torrent",
		Name:        "[ANi] Test DUID Torrent",
		Downloaded:  0,
		Renamed:     false,
		DownloadUID: "",
		Homepage:    "https://mikanani.me/Home/Episode/test_duid",
	}

	err = db.CreateTorrent(ctx, &torrent)
	if err != nil {
		t.Fatalf("Failed to create torrent: %v", err)
	}

	// 添加 DownloadUID
	expectedUID := "abc123def456"
	err = db.AddTorrentDUID(ctx, torrent.Link, expectedUID)
	if err != nil {
		t.Errorf("Failed to add torrent DUID: %v", err)
	}

	// 验证 DownloadUID 已添加
	var updated model.Torrent
	db.Where("link = ?", torrent.Link).First(&updated)
	if updated.DownloadUID != expectedUID {
		t.Errorf("Torrent DownloadUID should be %s, got %s", expectedUID, updated.DownloadUID)
	}

	t.Log("Successfully added torrent DUID")
}

// TestTorrentLifecycle 测试种子的完整生命周期流程
// 共享一个内存数据库，子测试按顺序执行，模拟真实的业务流程
func TestTorrentLifecycle(t *testing.T) {
	testdb := ":memory:"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	ctx := context.Background()

	// 3 条种子，各自走不同的生命周期路径
	torrentA := model.Torrent{
		Link:     "https://mikanani.me/Download/lifecycle/torrent_a.torrent",
		Name:     "[ANi] Lifecycle Test A - 01 [1080P]",
		Homepage: "https://mikanani.me/Home/Episode/torrent_a",
	}
	torrentB := model.Torrent{
		Link:     "https://mikanani.me/Download/lifecycle/torrent_b.torrent",
		Name:     "[ANi] Lifecycle Test B - 02 [1080P]",
		Homepage: "https://mikanani.me/Home/Episode/torrent_b",
	}
	torrentC := model.Torrent{
		Link:     "https://mikanani.me/Download/lifecycle/torrent_c.torrent",
		Name:     "[ANi] Lifecycle Test C - 03 [1080P]",
		Homepage: "https://mikanani.me/Home/Episode/torrent_c",
	}

	t.Run("Create", func(t *testing.T) {
		for _, tor := range []*model.Torrent{&torrentA, &torrentB, &torrentC} {
			if err := db.CreateTorrent(ctx, tor); err != nil {
				t.Fatalf("Failed to create torrent %s: %v", tor.Link, err)
			}
		}
		// 验证 3 条都在
		var count int64
		db.Model(&model.Torrent{}).Count(&count)
		if count != 3 {
			t.Fatalf("Expected 3 torrents, got %d", count)
		}
	})

	t.Run("CreateDuplicate", func(t *testing.T) {
		// 重复插入 torrentA，应该幂等（不报错，不覆盖）
		dup := model.Torrent{
			Link:     torrentA.Link,
			Name:     "This name should NOT overwrite",
			Homepage: "https://should.not.overwrite",
		}
		if err := db.CreateTorrent(ctx, &dup); err != nil {
			t.Fatalf("CreateTorrent duplicate should not error, got: %v", err)
		}
		// 验证原数据没被覆盖
		got, err := db.GetTorrentByURL(ctx, torrentA.Link)
		if err != nil {
			t.Fatalf("Failed to get torrent: %v", err)
		}
		if got.Name != torrentA.Name {
			t.Fatalf("Duplicate insert overwrote name: expected %q, got %q", torrentA.Name, got.Name)
		}
		// 总数仍然是 3
		var count int64
		db.Model(&model.Torrent{}).Count(&count)
		if count != 3 {
			t.Fatalf("Expected 3 torrents after duplicate insert, got %d", count)
		}
	})

	t.Run("GetByURL", func(t *testing.T) {
		got, err := db.GetTorrentByURL(ctx, torrentB.Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.Name != torrentB.Name {
			t.Fatalf("Expected name %q, got %q", torrentB.Name, got.Name)
		}
	})

	t.Run("AddDUID", func(t *testing.T) {
		// 给 torrentA 设置 download_uid
		uid := "uid_torrent_a_001"
		if err := db.AddTorrentDUID(ctx, torrentA.Link, uid); err != nil {
			t.Fatalf("AddTorrentDUID failed: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrentA.Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.DownloadUID != uid {
			t.Fatalf("Expected DownloadUID %q, got %q", uid, got.DownloadUID)
		}
		if got.Downloaded != model.DownloadSending {
			t.Fatalf("Expected Downloaded=%d (DownloadSending), got %d", model.DownloadSending, got.Downloaded)
		}
	})

	t.Run("GetByDUID", func(t *testing.T) {
		got, err := db.GetTorrentByDownloadUID(ctx, "uid_torrent_a_001")
		if err != nil {
			t.Fatalf("GetTorrentByDownloadUID failed: %v", err)
		}
		if got.Link != torrentA.Link {
			t.Fatalf("Expected link %q, got %q", torrentA.Link, got.Link)
		}
	})

	t.Run("MarkDownloaded", func(t *testing.T) {
		if err := db.AddTorrentDownload(ctx, torrentA.Link); err != nil {
			t.Fatalf("AddTorrentDownload failed: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrentA.Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.Downloaded != model.DownloadDone {
			t.Fatalf("Expected Downloaded=%d (DownloadDone), got %d", model.DownloadDone, got.Downloaded)
		}
	})

	t.Run("FindUnrenamed", func(t *testing.T) {
		// torrentA: Downloaded=Done, Renamed=false → 应该被找到
		torrents, err := db.FindUnrenamedTorrent(ctx)
		if err != nil {
			t.Fatalf("FindUnrenamedTorrent failed: %v", err)
		}
		if len(torrents) != 1 {
			t.Fatalf("Expected 1 unrenamed torrent, got %d", len(torrents))
		}
		if torrents[0].Link != torrentA.Link {
			t.Fatalf("Expected unrenamed torrent link %q, got %q", torrentA.Link, torrents[0].Link)
		}
	})

	t.Run("MarkRenamed", func(t *testing.T) {
		if err := db.TorrentRenamed(ctx, torrentA.Link); err != nil {
			t.Fatalf("TorrentRenamed failed: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrentA.Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if !got.Renamed {
			t.Fatal("Expected Renamed=true")
		}
	})

	t.Run("FindUnrenamed2", func(t *testing.T) {
		// torrentA 已重命名，其他两条没下载完成 → 应该找不到
		torrents, err := db.FindUnrenamedTorrent(ctx)
		if err != nil {
			t.Fatalf("FindUnrenamedTorrent failed: %v", err)
		}
		if len(torrents) != 0 {
			t.Fatalf("Expected 0 unrenamed torrents, got %d", len(torrents))
		}
	})

	t.Run("MarkError", func(t *testing.T) {
		// 先给 torrentB 设置 DUID 并发送到下载器
		if err := db.AddTorrentDUID(ctx, torrentB.Link, "uid_torrent_b_001"); err != nil {
			t.Fatalf("AddTorrentDUID failed: %v", err)
		}
		// 标记下载出错
		if err := db.AddTorrentError(ctx, torrentB.Link); err != nil {
			t.Fatalf("AddTorrentError failed: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrentB.Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.Downloaded != model.DownloadError {
			t.Fatalf("Expected Downloaded=%d (DownloadError), got %d", model.DownloadError, got.Downloaded)
		}
	})

	t.Run("CheckNew", func(t *testing.T) {
		candidates := []*model.Torrent{
			{Link: torrentA.Link},                                        // 已存在
			{Link: "https://mikanani.me/Download/lifecycle/new1.torrent"}, // 新的
			{Link: torrentC.Link},                                        // 已存在
			{Link: "https://mikanani.me/Download/lifecycle/new2.torrent"}, // 新的
		}
		newOnes, err := db.CheckNewTorrents(ctx, candidates)
		if err != nil {
			t.Fatalf("CheckNewTorrents failed: %v", err)
		}
		if len(newOnes) != 2 {
			t.Fatalf("Expected 2 new torrents, got %d", len(newOnes))
		}
	})

	t.Run("DeleteByURL", func(t *testing.T) {
		if err := db.DeleteTorrentByURL(ctx, torrentC.Link); err != nil {
			t.Fatalf("DeleteTorrentByURL failed: %v", err)
		}
		_, err := db.GetTorrentByURL(ctx, torrentC.Link)
		if err == nil {
			t.Fatal("Expected error when getting deleted torrent, got nil")
		}
	})

	t.Run("DeleteByDUID", func(t *testing.T) {
		if err := db.DeleteTorrentByDownloadUID(ctx, "uid_torrent_b_001"); err != nil {
			t.Fatalf("DeleteTorrentByDownloadUID failed: %v", err)
		}
		_, err := db.GetTorrentByDownloadUID(ctx, "uid_torrent_b_001")
		if err == nil {
			t.Fatal("Expected error when getting deleted torrent, got nil")
		}
		// 最终只剩 torrentA
		var count int64
		db.Model(&model.Torrent{}).Count(&count)
		if count != 1 {
			t.Fatalf("Expected 1 remaining torrent, got %d", count)
		}
	})
}
