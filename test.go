package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	// "goto-bangumi/internal/parser"

	"goto-bangumi/internal/conf"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/parser/baseparser"
	"goto-bangumi/internal/refresh"

	// "goto-bangumi/internal/parser"

	// "goto-bangumi/internal/model"
	"github.com/spf13/viper"
)

func test_parser() {
	t := baseparser.NewTitleMetaParse()
	title := "【幻樱字幕组】【4月新番】【古见同学有交流障碍症 第二季 Komi-san wa, Komyushou Desu. S02】【22】【GB_MP4】【1920X1080】"
	// title :="[织梦字幕组][尼尔：机械纪元 NieR Automata Ver1.1a][02集][1080P][AVC][简日双语]"
	// title :="[梦蓝字幕组]New Doraemon 哆啦A梦新番[747][2023.02.25][AVC][1080P][GB_JP][MP4]"
	// title := "[ANi] Grand Blue Dreaming /  GRAND BLUE 碧蓝之海 2 - 04 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]"
	ans := t.Parse(title)
	fmt.Printf("%+v\n", ans)
}

func test_conf() {
	config, _ := conf.LoadConfig("./config")
	fmt.Println(config.BangumiManage)
	fmt.Printf("%+v\n", config.BangumiManage)

	// ANSI 颜色码
	cyan := "\033[36m"
	yellow := "\033[33m"
	green := "\033[32m"
	reset := "\033[0m"
	bold := "\033[1m"

	fmt.Printf("\n%s%s%s%s\n", cyan, bold, strings.Repeat("━", 70), reset)
	fmt.Printf("%s%s⚙️  配置信息%s\n", yellow, bold, reset)
	fmt.Printf("%s%s%s%s\n", cyan, bold, strings.Repeat("━", 70), reset)

	data, _ := json.MarshalIndent(viper.AllSettings(), "", "  ")
	fmt.Printf("%s%s%s\n", green, string(data), reset)

	fmt.Printf("%s%s%s%s\n\n", cyan, bold, strings.Repeat("━", 70), reset)
}

func test_network() {
	config, _ := conf.LoadConfig("./config")
	network.Init(&config.Proxy)
	client := network.NewRequestClient()

	url := "https://mikanani.me/RSS/MyBangumi?token=rmT2qkfOxawQZBSwcM0%2ba2K0McpV%2fjZ6qiStU2Et73Q%3d"
	cnt, err := client.GetRSS(url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	fmt.Println(cnt.Title)
	fmt.Println(cnt.Link)
	for _, item := range cnt.Torrents {
		fmt.Println("Title:", item.Name)
		fmt.Println("Link:", item.Link)
		// fmt.Println("Homepage:", item.Homepage)
		fmt.Println("Homepage", item.Enclosure.URL)
		fmt.Println("-----")
	}
}

func test_torrent() {
	config, _ := conf.LoadConfig("./config")
	network.Init(&config.Proxy)
	client := network.NewRequestClient()

	url := "https://mikanani.me/RSS/MyBangumi?token=rmT2qkfOxawQZBSwcM0%2ba2K0McpV%2fjZ6qiStU2Et73Q%3d"
	cnt, err := client.GetTorrents(url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	for _, item := range cnt {
		fmt.Println("Title:", item.Name)
		fmt.Println("Link:", item.URL)
		// fmt.Println("Homepage:", item.Homepage)
		fmt.Println("Homepage", item.Homepage)
		fmt.Println("-----")
	}
}

// func test_database() {
// 	// 初始化数据库连接
// 	err := database.InitDB("./data/data.db")
// 	if err != nil {
// 		fmt.Println("Error initializing database:", err)
// 		return
// 	}
// 	defer database.CloseDB()
//
// 	// 通过ID获取番剧
// 	bangumi, err := database.GetBangumiByID(1)
// 	if err != nil {
// 		fmt.Println("Error getting bangumi:", err)
// 		return
// 	}
//
// 	// 打印结果
// 	fmt.Printf("\n=== 番剧信息 ===\n")
// 	fmt.Printf("ID: %d\n", bangumi.UID)
// 	fmt.Printf("中文名: %s\n", bangumi.OfficialTitle)
// 	// fmt.Printf("原名: %s\n", bangumi.TitleRaw)
// 	fmt.Printf("季度: %d\n", bangumi.Season)
// 	fmt.Printf("解析器: %s\n", bangumi.Parse)
// 	fmt.Printf("RSS链接: %s\n", bangumi.RssLink)
// }

func test_image_cache() {
	config, _ := conf.LoadConfig("./config")
	network.Init(&config.Proxy)
	client := network.NewRequestClient()

	imageCache, err := network.NewImageCache(client, "./data")
	if err != nil {
		fmt.Println("Error creating image cache:", err)
		return
	}

	url := "https://mikanani.me/images/Bangumi/202510/0d10efc3.jpg?width=460&height=640&format=webp"
	data, err := imageCache.LoadImage(url)
	if err != nil {
		fmt.Println("Error loading image:", err)
		return
	}
	fmt.Printf("Image data length: %d bytes\n", len(data))
}

func test_mikan_parser() {
	parser := baseparser.NewMikanParser()
	// url := "https://mikanani.me/Home/Bangumi/3751"
	// url := "https://mikanani.me/Home/Episode/8c94c1699735481c8b2b18dba38908042f53adcc"
	url := "https://mikanani.me/Home/Episode/7c8c41e409922d9f2c34a726c92e77daf05558ff"

	info, err := parser.Parse(url)
	if err != nil {
		fmt.Println("Error parsing Mikan page:", err)
		return
	}
	fmt.Printf("Mikan Info:\n")
	fmt.Printf("ID: %d\n", info.ID)
	fmt.Printf("Official Title: %s\n", info.OfficialTitle)
	fmt.Printf("Season: %d\n", info.Season)
	fmt.Printf("Poster Link: %s\n", info.PosterLink)
}

func test_split() {
	s := ""
	s1 := ",a,b"
	fmt.Println("SplitSeq result for s1:", strings.Split(s1, ","))
	for v := range strings.SplitSeq(s, ",") {
		fmt.Println("v=", v)
	}
}

func test_tmdb_parser() {
	// 测试 "狼与香辛料"
	// title := "囮物语"
	// title := "拥有超常技能的异世界流浪美食家 第二季"
	title := "狼与香辛料"
	language := "zh"

	fmt.Printf("\n%s正在解析 TMDB 信息...%s\n", "\033[36m", "\033[0m")
	fmt.Printf("标题: %s\n", title)
	fmt.Printf("语言: %s\n\n", language)

	tmdbInfo, err := baseparser.ParseTMDB(title, language)
	if err != nil {
		fmt.Printf("%s错误: %v%s\n", "\033[31m", err, "\033[0m")
		return
	}

	if tmdbInfo == nil {
		fmt.Printf("%s未找到 TMDB 信息%s\n", "\033[33m", "\033[0m")
		return
	}

	// 打印结果
	fmt.Printf("%s%s=== TMDB 信息 ===%s\n", "\033[32m", "\033[1m", "\033[0m")
	fmt.Printf("ID: %d\n", tmdbInfo.ID)
	fmt.Printf("标题: %s\n", tmdbInfo.Title)
	fmt.Printf("原始标题: %s\n", tmdbInfo.OriginalTitle)
	fmt.Printf("年份: %s\n", tmdbInfo.Year)
	fmt.Printf("最新季度: %d\n", tmdbInfo.Season)
	fmt.Printf("海报链接: %s\n", tmdbInfo.PosterLink)
	fmt.Printf("\n季度信息:\n")
}





func test_torrent_to_bangumi() {
	torrent := model.Torrent{
		// Name:     "[ANi] Chitose Is in the Ramune Bottle / 弹珠汽水瓶里的千岁同学 - 02 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
		Name:     "meowwwwwww",
		URL:      "magnet:?xt=urn:btih:EXAMPLE1",
		Homepage: "https://mikanani.me/Home/Episode/7c8c41e409922d9f2c34a726c92e77daf05558ff",
	}
	rss := model.RSSItem{
		Name: "Chitose Is in the Ramune Bottle / 弹珠汽水瓶里的千岁同学",
		URL:  "https://mikanani.me/RSS/Search?searchstr=ANI",
	}
	bangumi, err := refresh.TorrentToBangumi(torrent, rss)
	if err == nil {
		fmt.Printf("\n=== 番剧信息 ===\n")
		fmt.Printf("ID: %d\n", bangumi.ID)
		fmt.Printf("中文名: %s\n", bangumi.OfficialTitle)
		fmt.Printf("季度: %d\n", bangumi.Season)
		fmt.Printf("年份: %s\n", bangumi.Year)
		fmt.Printf("解析器: %s\n", bangumi.Parse)
		fmt.Printf("RSS链接: %s\n", bangumi.RRSSLink)
		fmt.Printf("封面链接: %s\n", bangumi.PosterLink)
		fmt.Printf("Tmdb信息:\n %+v\n", bangumi.TmdbItem)
		fmt.Printf("MikanID: \n%v\n", bangumi.MikanItem)
	}
	// database.InitDB("./data/data.db")
	// refresh.CreateBangumi(torrent, rss)
	// defer database.CloseDB()
	// if bangumi == nil {
	// 	fmt.Println("无法解析番剧标题")
	// 	return
	// }
	// fmt.Printf("\n=== 番剧信息 ===\n")
	// fmt.Printf("ID: %d\n", bangumi.ID)
	// fmt.Printf("中文名: %s\n", bangumi.OfficialTitle)
	// fmt.Printf("季度: %d\n", bangumi.Season)
	// fmt.Printf("年份: %s\n", bangumi.Year)
	// fmt.Printf("解析器: %s\n", bangumi.Parse)
	// fmt.Printf("RSS链接: %s\n", bangumi.RssLink)
	// fmt.Printf("封面链接: %s\n", bangumi.PosterLink)
	// fmt.Printf("Tmdb信息:\n %+v\n", bangumi.TmdbItem)
	// fmt.Printf("MikanID: \n%v\n", bangumi.MikanItem)
}

func test_add_bangumi() {
	database.InitDB("./data/data.db")
	defer database.CloseDB()

	// 创建 MikanItem 测试数据（参考数据库中的现有数据）
	mikanItem := model.MikanItem{
		ID:            3775,
		OfficialTitle: "弹珠汽水瓶里的千岁同学",
		Season:        1,
		PosterLink:    "https://mikanani.me/images/Bangumi/202510/37749647.jpg",
	}

	// 创建 TmdbItem 测试数据
	tmdbItem := model.TmdbItem{
		ID:            261344,
		Title:         "弹珠汽水瓶里的千岁同学",
		OriginalTitle: "千歳くんはラムネ瓶のなか",
		Year:          "2025",
		Season:        1,
		AirDate:       "2025-10-07",
		EpisodeCount:  13,
		PosterLink:    "https://image.tmdb.org/t/p/w780/ibr0vzPc7XImKJ5kqPOxgsvZEKB.jpg",
		VoteAverage:   6.667,
	}

	// 创建 EpisodeMetadata 测试数据
	metaInfo := model.EpisodeMetadata{
		Title:      "弹珠汽水瓶里的千岁同学",
		Group:      "ANi",
		Season:     1,
		SeasonRaw:  "",
		Resolution: "1080P",
		Sub:        "繁",
		SubType:    "",
		Source:     "Baha",
		AudioInfo:  "AAC",
		VideoInfo:  "AVC,MP4",
	}

	// 创建 Bangumi 测试数据
	// mikanID := 3774
	// tmdbID := 261343
	bangumi := model.Bangumi{
		OfficialTitle:   "弹珠汽水瓶里的千岁同学",
		Year:            "2025",
		Season:          1,
		MikanItem:       &mikanItem,
		TmdbItem:        &tmdbItem,
		EpisodeMetadata: []model.EpisodeMetadata{metaInfo},
		EpsCollect:      false,
		Offset:          0,
		IncludeFilter:   "",
		ExcludeFilter:   "",
		Parse:           "mikan",
		RRSSLink:        "https://mikanani.me/RSS/Search?searchstr=ANI",
		PosterLink:      "https://mikanani.me/images/Bangumi/202510/37749647.jpg",
		Deleted:         false,
	}
	fmt.Printf("添加番剧: %+v\n", bangumi.OfficialTitle)
	db := database.GetDB()
	_ = db.CreateBangumi(&bangumi)
	// fmt.Printf("查询到的番剧: %+v\n", dbBangumi)
	// err := db.CreateBangumi(&bangumi)
	// fmt.Printf("错误: %v\n", err)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // 设置为 Debug 级别
	}))
	slog.SetDefault(logger)

	// test_add_bangumi()
	test_torrent_to_bangumi()
	// test_parser()

	// test_get_bangumi_parser_by_title()
	// test_database()
	// test_image_cache()
	// test_mikan_parser()
	// test_tmdb_parser()
	// test_split()
	// test_tmdb_parser()
	// test_cache_performance()
	// test_concurrent_requests()
	// test_multi_client_singleflight()
}

// func main() {
// 	tr := http.Transport{}
// 	header := http.Header{}
// 	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
// 	client := http.Client{Transport: &tr}
// 	resp, err := client.Get("https://www.baidu.com")
// 	body, err := io.ReadAll(resp.Body)
// 	fmt.Println(resp, err)
// 	// 获取其中的文本
// 	fmt.Println(string(body))
// }
