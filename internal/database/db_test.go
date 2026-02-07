package database

import (
	"context"
	"testing"

	"goto-bangumi/internal/model"
)

func TestNewDB(t *testing.T) {
	dbPath := ":memory:"
	_, err := NewDB(&dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
}

func TestAddBangumi(t *testing.T) {
	// 什么时候会加 Bangumi
	// 1. 主要是调用 FindNewBangumi, 聚合以及日常刷新
	// 2. 其次就是非聚合的时候, 前端会点一个让我们去找新的番剧
	// 3. 通过rss_link 来连接吧, rss不一定会加进去
	testdb := "./test.db"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	mikanItem := model.MikanItem{
		ID:            3599,
		OfficialTitle: "夏日口袋",
		Season:        1,
		PosterLink:    "https://www.mikanani.me/attachment/202202/xxjXH8e6.jpg",
	}
	tmdbItem := model.TmdbItem{
		ID:            131631,
		Year:          "2022",
		OriginalTitle: "Summer Pocket",
		AirDate:       "2022-07-01",
		EpisodeCount:  12,
		Title:         "Summer Pocket",
		Season:        1,
		PosterLink:    "https://www.themoviedb.org/t/p/w600_and_h900_bestv2/8m8n5Yq4x0dT3cR7W4a6F2kH0kP.jpg",
		VoteAverage:   7.5,
	}
	EpisodeMetadata := model.EpisodeMetadata{
		Title:     "第1话 夏日口袋",
		Season:    1,
		SeasonRaw: "",
		Episode:   1,
	}
	bangumi := model.Bangumi{
		OfficialTitle:   "夏日口袋",
		Year:            "2022",
		Season:          1,
		MikanItem:       &mikanItem,
		TmdbItem:        &tmdbItem,
		EpisodeMetadata: []model.EpisodeMetadata{EpisodeMetadata},
		RSSLink:         "https://example.com/rss",
	}
	db.CreateBangumi(&bangumi)
}

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
	if updated.Downloaded != 1 {
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
