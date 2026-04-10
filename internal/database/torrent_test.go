package database

import (
	"context"
	"testing"

	"goto-bangumi/internal/model"
)

// TestAddTorrent 测试 Torrent 相关数据库操作，每个子测试验证一个函数的功能
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
	const testDUID = "abc123def456"

	t.Run("BatchCreate", func(t *testing.T) {
		for i := range torrents {
			if err := db.CreateTorrent(ctx, &torrents[i]); err != nil {
				t.Fatalf("Failed to create torrent %d: %v", i, err)
			}
		}
		var count int64
		db.Model(&model.Torrent{}).Count(&count)
		if count != int64(len(torrents)) {
			t.Fatalf("Expected %d torrents in DB, got %d", len(torrents), count)
		}
	})

	t.Run("CreateDuplicate", func(t *testing.T) {
		dup := torrents[0]
		dup.Name = "This name should NOT overwrite"
		if err := db.CreateTorrent(ctx, &dup); err != nil {
			t.Fatalf("Duplicate CreateTorrent should not error, got: %v", err)
		}
		var count int64
		db.Model(&model.Torrent{}).Count(&count)
		if count != int64(len(torrents)) {
			t.Fatalf("After duplicate insert: expected %d torrents, got %d", len(torrents), count)
		}
	})

	t.Run("GetByURL", func(t *testing.T) {
		got, err := db.GetTorrentByURL(ctx, torrents[0].Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.Name != torrents[0].Name {
			t.Fatalf("Expected name %q, got %q", torrents[0].Name, got.Name)
		}
	})

	// AddDUID 同时将 Downloaded 置为 DownloadSending
	t.Run("AddDUID", func(t *testing.T) {
		if err := db.AddTorrentDUID(ctx, torrents[1].Link, testDUID); err != nil {
			t.Fatalf("Failed to add torrent DUID: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrents[1].Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.DownloadUID != testDUID {
			t.Fatalf("Expected DownloadUID=%s, got %s", testDUID, got.DownloadUID)
		}
		if got.Downloaded != model.DownloadSending {
			t.Fatalf("Expected Downloaded=%d (DownloadSending), got %d", model.DownloadSending, got.Downloaded)
		}
	})

	t.Run("GetByDUID", func(t *testing.T) {
		got, err := db.GetTorrentByDownloadUID(ctx, testDUID)
		if err != nil {
			t.Fatalf("GetTorrentByDownloadUID failed: %v", err)
		}
		if got.Link != torrents[1].Link {
			t.Fatalf("Expected link %q, got %q", torrents[1].Link, got.Link)
		}
	})

	// torrents[0].Renamed=false，标记下载完成后 FindUnrenamed 应能找到它
	t.Run("MarkDownloaded", func(t *testing.T) {
		if err := db.AddTorrentDownload(ctx, torrents[0].Link); err != nil {
			t.Fatalf("Failed to mark torrent as downloaded: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrents[0].Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.Downloaded != model.DownloadDone {
			t.Fatalf("Expected Downloaded=%d, got %d", model.DownloadDone, got.Downloaded)
		}
	})

	t.Run("FindUnrenamed", func(t *testing.T) {
		results, err := db.FindUnrenamedTorrent(ctx)
		if err != nil {
			t.Fatalf("FindUnrenamedTorrent failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 unrenamed torrent, got %d", len(results))
		}
		if results[0].Link != torrents[0].Link {
			t.Fatalf("Expected link %q, got %q", torrents[0].Link, results[0].Link)
		}
	})

	t.Run("MarkRenamed", func(t *testing.T) {
		if err := db.TorrentRenamed(ctx, torrents[0].Link); err != nil {
			t.Fatalf("Failed to mark torrent as renamed: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrents[0].Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if !got.Renamed {
			t.Fatal("Expected Renamed=true")
		}
	})

	t.Run("MarkError", func(t *testing.T) {
		if err := db.AddTorrentError(ctx, torrents[2].Link); err != nil {
			t.Fatalf("AddTorrentError failed: %v", err)
		}
		got, err := db.GetTorrentByURL(ctx, torrents[2].Link)
		if err != nil {
			t.Fatalf("GetTorrentByURL failed: %v", err)
		}
		if got.Downloaded != model.DownloadError {
			t.Fatalf("Expected Downloaded=%d (DownloadError), got %d", model.DownloadError, got.Downloaded)
		}
	})

	t.Run("CheckNew", func(t *testing.T) {
		candidates := []*model.Torrent{
			{Link: torrents[0].Link},                                         // 已存在
			{Link: "https://mikanani.me/Download/new/new1.torrent"},           // 新的
			{Link: torrents[1].Link},                                         // 已存在
			{Link: "https://mikanani.me/Download/new/new2.torrent"},           // 新的
		}
		newOnes, err := db.CheckNewTorrents(ctx, candidates)
		if err != nil {
			t.Fatalf("CheckNewTorrents failed: %v", err)
		}
		if len(newOnes) != 2 {
			t.Fatalf("Expected 2 new torrents, got %d", len(newOnes))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := db.DeleteTorrent(ctx, torrents[9].Link); err != nil {
			t.Fatalf("Failed to delete torrent: %v", err)
		}
		_, err := db.GetTorrentByURL(ctx, torrents[9].Link)
		if err == nil {
			t.Fatal("Expected error when getting deleted torrent, got nil")
		}
	})

	t.Run("DeleteByURL", func(t *testing.T) {
		if err := db.DeleteTorrentByURL(ctx, torrents[8].Link); err != nil {
			t.Fatalf("DeleteTorrentByURL failed: %v", err)
		}
		_, err := db.GetTorrentByURL(ctx, torrents[8].Link)
		if err == nil {
			t.Fatal("Expected error when getting deleted torrent, got nil")
		}
	})

	t.Run("DeleteByDUID", func(t *testing.T) {
		if err := db.DeleteTorrentByDownloadUID(ctx, testDUID); err != nil {
			t.Fatalf("DeleteTorrentByDownloadUID failed: %v", err)
		}
		_, err := db.GetTorrentByDownloadUID(ctx, testDUID)
		if err == nil {
			t.Fatal("Expected error when getting deleted torrent, got nil")
		}
	})
}
