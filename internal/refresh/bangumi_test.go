package refresh

import (
	"testing"
	"time"

	"goto-bangumi/internal/database"
)

// TestFindNewBangumi_NormalFlow 测试 FindNewBangumi 的正常流程
func TestFindNewBangumi_NormalFlow(t *testing.T) {
	// 创建内存数据库
	memoryDB := ":memory:"
	db, err := database.NewDB(&memoryDB)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	defer db.Close()

	// 验证初始状态：数据库为空
	initialBangumis, err := db.ListBangumi()
	if err != nil {
		t.Fatalf("查询番剧列表失败: %v", err)
	}
	if len(initialBangumis) != 0 {
		t.Errorf("期望初始番剧数量为 0，但得到 %d", len(initialBangumis))
	}

	// 调用 FindNewBangumi
	// 注意：FindNewBangumi 使用全局数据库，这里需要先初始化
	// 由于 FindNewBangumi 内部调用 database.GetDB()，我们需要设置全局数据库
	database.InitDB(&memoryDB)
	defer database.CloseDB()

	rssURL := "https://mikanani.me/RSS/MyBangumi?token=test"
	FindNewBangumi(rssURL)

	// 等待 goroutine 完成
	time.Sleep(1 * time.Second)

	// 验证创建的番剧
	globalDB := database.GetDB()
	finalBangumis, err := globalDB.ListBangumi()
	if err != nil {
		t.Fatalf("查询番剧列表失败: %v", err)
	}

	t.Logf("创建了 %d 个番剧", len(finalBangumis))

	// 验证创建了预期的番剧
	expectedBangumis := map[string]struct {
		year   string
		season int
	}{
		"弹珠汽水瓶里的千岁同学": {year: "2025", season: 1},
		"跨越种族与你相恋":    {year: "2025", season: 1},
		"桃源暗鬼":        {year: "2025", season: 1},
		"异世界四重奏":      {year: "2019", season: 1},
	}

	for _, bangumi := range finalBangumis {
		t.Logf("番剧: %s, Year: %s, Season: %d, MikanID: %v, TmdbID: %v",
			bangumi.OfficialTitle, bangumi.Year, bangumi.Season, bangumi.MikanID, bangumi.TmdbID)

		expected, ok := expectedBangumis[bangumi.OfficialTitle]
		if !ok {
			t.Logf("意外的番剧: %s", bangumi.OfficialTitle)
			continue
		}

		if bangumi.Year != expected.year {
			t.Errorf("番剧 '%s' 的 Year = %v, 期望 %v", bangumi.OfficialTitle, bangumi.Year, expected.year)
		}
		if bangumi.Season != expected.season {
			t.Errorf("番剧 '%s' 的 Season = %v, 期望 %v", bangumi.OfficialTitle, bangumi.Season, expected.season)
		}
	}

	// 验证创建了 4 个番剧
	if len(finalBangumis) < 4 {
		t.Errorf("期望至少创建 4 个番剧，但只创建了 %d 个", len(finalBangumis))
	}
}

// TestGetTorrents 测试 getTorrents 函数
func TestGetTorrents(t *testing.T) {
	// 初始化内存数据库
	memoryDB := ":memory:"
	database.InitDB(&memoryDB)
	defer database.CloseDB()

	rssURL := "https://mikanani.me/RSS/MyBangumi?token=test"
	torrents := getTorrents(rssURL)

	// 验证返回的种子数量
	if len(torrents) == 0 {
		t.Fatal("期望获取到种子，但返回了空列表")
	}

	t.Logf("获取到 %d 个种子", len(torrents))

	// 验证种子基本信息
	for i, torrent := range torrents {
		if torrent.Name == "" {
			t.Errorf("第 %d 个种子名称为空", i)
		}
		if torrent.URL == "" {
			t.Errorf("第 %d 个种子 URL 为空", i)
		}
		t.Logf("种子 %d: %s", i+1, torrent.Name)
	}
}

// TestGetTorrents_WithExisting 测试 getTorrents 过滤已存在种子
func TestGetTorrents_WithExisting(t *testing.T) {
	memoryDB := ":memory:"
	database.InitDB(&memoryDB)
	defer database.CloseDB()

	rssURL := "https://mikanani.me/RSS/MyBangumi?token=test"

	// 第一次获取
	firstTorrents := getTorrents(rssURL)
	if len(firstTorrents) == 0 {
		t.Fatal("第一次获取种子失败")
	}

	// 将第一个种子标记为已下载
	db := database.GetDB()
	firstTorrents[0].Downloaded = 1
	db.CreateTorrent(firstTorrents[0])

	// 第二次获取，应该少一个
	secondTorrents := getTorrents(rssURL)
	if len(secondTorrents) != len(firstTorrents)-1 {
		t.Errorf("期望 %d 个种子，实际 %d 个", len(firstTorrents)-1, len(secondTorrents))
	}
}


