package refresh

import (
	"context"
	"strings"
	"testing"
	"time"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser"
	"goto-bangumi/internal/taskrunner"
)

// TestFindNewBangumi_NormalFlow 测试 FindNewBangumi 的正常流程
func TestFindNewBangumi_NormalFlow(t *testing.T) {
	// 创建内存数据库
	memoryDB := ":memory:"
	// 调用 FindNewBangumi
	// 注意：FindNewBangumi 使用全局数据库，这里需要先初始化
	// 由于 FindNewBangumi 内部调用 database.GetDB()，我们需要设置全局数据库
	database.InitDB(&memoryDB)
	defer database.CloseDB()

	rssURL := "https://mikanani.me/RSS/MyBangumi?token=test"
	rssItem := &model.RSSItem{
		Name: "我的番组",
		Link: rssURL,
	}
	FindNewBangumi(context.Background(), rssItem)

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
		"异世界四重奏":      {year: "2019", season: 3},
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
	torrents := getTorrents(context.Background(), rssURL)

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
		if torrent.Link == "" {
			t.Errorf("第 %d 个种子 Link 为空", i)
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

	ctx := context.Background()

	// 第一次获取
	firstTorrents := getTorrents(ctx, rssURL)
	if len(firstTorrents) == 0 {
		t.Fatal("第一次获取种子失败")
	}

	// 将第一个种子标记为已下载
	db := database.GetDB()
	firstTorrents[0].Downloaded = model.DownloadSending
	db.CreateTorrent(ctx, firstTorrents[0])

	// 第二次获取，应该少一个
	secondTorrents := getTorrents(ctx, rssURL)
	if len(secondTorrents) != len(firstTorrents)-1 {
		t.Errorf("期望 %d 个种子，实际 %d 个", len(firstTorrents)-1, len(secondTorrents))
	}
}

// TestRefreshRSS 测试 RefreshRSS 的完整流程
// RSS 源: https://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=370
// 番剧: 败犬女主太多了！ (13 条种子: 1 条合集 + 12 集单集)
// 排除: 合集
func TestRefreshRSS(t *testing.T) {
	ctx := context.Background()

	// 设置 parser config，让 FindNewBangumi 创建的 bangumi 带上 exclude filter
	oldConfig := parser.ParserConfig
	parser.ParserConfig = &model.RssParserConfig{
		Filter: []string{"合集"},
	}
	defer func() { parser.ParserConfig = oldConfig }()

	// 1. 初始化内存数据库
	memoryDB := ":memory:"
	database.InitDB(&memoryDB)
	defer database.CloseDB()
	db := database.GetDB()

	// 2. 创建 RSS 订阅
	rssURL := "https://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=370"
	rssItem := &model.RSSItem{
		Name:          "败犬女主太多了！",
		Link:          rssURL,
		Enabled:       true,
		ExcludeFilter: "合集",
	}
	if err := db.CreateRSS(ctx, rssItem); err != nil {
		t.Fatalf("创建 RSS 失败: %v", err)
	}

	// 3. 调用 FindNewBangumi 发现新番并创建 Bangumi + EpisodeMetadata
	FindNewBangumi(ctx, rssItem)

	// 验证 bangumi 已创建
	bangumis, err := db.ListBangumi()
	if err != nil {
		t.Fatalf("查询番剧列表失败: %v", err)
	}
	if len(bangumis) == 0 {
		t.Fatal("FindNewBangumi 未创建任何番剧")
	}
	t.Logf("FindNewBangumi 创建了 %d 个番剧", len(bangumis))
	for _, b := range bangumis {
		t.Logf("  番剧: %s, Year: %s, Season: %d, ExcludeFilter: %q", b.OfficialTitle, b.Year, b.Season, b.ExcludeFilter)
	}

	// 4. 创建 TaskRunner 并调用 RefreshRSS
	runner := taskrunner.New(64, 2)
	// 注册一个空的 handler，让 submit 能成功
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		return taskrunner.PhaseResult{}
	})
	runner.Start(ctx)
	defer runner.Stop()

	RefreshRSS(ctx, rssURL, runner)

	// 5. 验证结果
	// 等待一小段时间让 runner 处理
	time.Sleep(500 * time.Millisecond)

	// 检查入库的种子数量（应该是 12，排除了 1 条合集）
	var torrents []*model.Torrent
	if err := db.Find(&torrents).Error; err != nil {
		t.Fatalf("查询种子列表失败: %v", err)
	}

	t.Logf("入库种子数量: %d", len(torrents))
	for _, torrent := range torrents {
		t.Logf("  种子: %s", torrent.Name)
		// 验证没有合集被入库
		if strings.Contains(torrent.Name, "合集") {
			t.Errorf("合集种子不应该被入库: %s", torrent.Name)
		}
	}

	// RSS 共 13 条，排除 1 条合集，应该入库 12 条
	if len(torrents) != 12 {
		t.Errorf("期望入库 12 个种子，实际 %d 个", len(torrents))
	}
}
