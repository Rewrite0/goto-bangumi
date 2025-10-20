package main

import (
	"fmt"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
)

func main() {
	// ANSI 颜色码
	cyan := "\033[36m"
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	reset := "\033[0m"
	bold := "\033[1m"

	fmt.Printf("\n%s%s=== GORM 关联自动保存测试 ===%s\n\n", cyan, bold, reset)

	// 1. 初始化数据库
	fmt.Printf("%s📦 连接数据库...%s\n", cyan, reset)
	db, err := database.NewDB("./data/test_association.db")
	if err != nil {
		fmt.Printf("%s❌ 数据库连接失败: %v%s\n", red, err, reset)
		return
	}
	defer db.Close()
	fmt.Printf("%s✅ 数据库连接成功！%s\n\n", green, reset)

	// 2. 方式一：完整创建（包含 MikanItem 和 TmdbItem）
	fmt.Printf("%s%s方式一：在 Bangumi 中直接包含关联对象%s\n", yellow, bold, reset)
	fmt.Printf("%s💡 GORM 会自动保存 MikanItem 和 TmdbItem，并自动设置外键%s\n\n", cyan, reset)

	bangumi1 := &model.Bangumi{
		OfficialTitle: "葬送的芙莉莲",
		Year:          "2023",
		Season:        1,
		Parse:        "mikan",
		RRSSLink:       "https://mikanani.me/RSS/Bangumi?bangumiId=3751",
		Deleted:       false,

		// 直接包含 MikanItem 对象
		MikanItem: &model.MikanItem{
			ID:            3751,
			OfficialTitle: "葬送的芙莉莲",
			Season:        1,
			PosterLink:    "https://mikanani.me/images/Bangumi/202510/0d10efc3.jpg",
		},

		// 直接包含 TmdbItem 对象
		TmdbItem: &model.TmdbItem{
			ID:            138502,
			Title:         "葬送的芙莉莲",
			OriginalTitle: "Sousou no Frieren",
			Year:          "2023",
			Season:        1,
			EpisodeCount:  28,
			PosterLink:     "https://image.tmdb.org/t/p/w500/e8yLJjTLn963IMiHEPCosFjrfcz.jpg",
			VoteAverage:   9.0,
		},
	}

	// 一次性保存，GORM 自动处理所有关联
	if err := db.CreateBangumi(bangumi1); err != nil {
		fmt.Printf("%s❌ 保存失败: %v%s\n", red, err, reset)
		return
	}

	fmt.Printf("%s✅ Bangumi 保存成功！%s\n", green, reset)
	fmt.Printf("  ID: %s%d%s\n", cyan, bangumi1.ID, reset)
	fmt.Printf("  MikanID: %s%d%s (自动设置)\n", cyan, *bangumi1.MikanID, reset)
	fmt.Printf("  TmdbID: %s%d%s (自动设置)\n", cyan, *bangumi1.TmdbID, reset)

	// 3. 方式二：只包含 MikanItem（没有 TmdbItem）
	fmt.Printf("\n%s%s方式二：只包含 MikanItem，不包含 TmdbItem%s\n", yellow, bold, reset)

	bangumi2 := &model.Bangumi{
		OfficialTitle: "孤独摇滚",
		Year:          "2022",
		Season:        1,
		Parse:        "mikan",
		RRSSLink:       "https://mikanani.me/RSS/Bangumi?bangumiId=3140",

		// 只有 MikanItem，没有 TmdbItem
		MikanItem: &model.MikanItem{
			ID:            3140,
			OfficialTitle: "孤独摇滚",
			Season:        1,
			PosterLink:    "https://mikanani.me/images/Bangumi/202210/bocchi.jpg",
		},
		// TmdbItem 为 nil，不保存
	}

	if err := db.CreateBangumi(bangumi2); err != nil {
		fmt.Printf("%s❌ 保存失败: %v%s\n", red, err, reset)
		return
	}

	fmt.Printf("%s✅ Bangumi 保存成功！%s\n", green, reset)
	fmt.Printf("  ID: %s%d%s\n", cyan, bangumi2.ID, reset)
	fmt.Printf("  MikanID: %s%d%s (自动设置)\n", cyan, *bangumi2.MikanID, reset)
	fmt.Printf("  TmdbID: %snil%s (因为没有提供 TmdbItem)\n", cyan, reset)

	// 4. 验证数据
	fmt.Printf("\n%s%s=== 验证保存结果 ===%s\n\n", cyan, bold, reset)

	// 使用预加载查询
	savedBangumi, err := db.GetBangumiWithDetails(uint(bangumi1.ID))
	if err != nil {
		fmt.Printf("%s❌ 查询失败: %v%s\n", red, err, reset)
		return
	}

	fmt.Printf("%s📺 番剧信息:%s\n", green, reset)
	fmt.Printf("  ID: %d\n", savedBangumi.ID)
	fmt.Printf("  标题: %s\n", savedBangumi.OfficialTitle)

	if savedBangumi.MikanItem != nil {
		fmt.Printf("\n%s📚 关联的 Mikan 信息:%s\n", green, reset)
		fmt.Printf("  ID: %d\n", savedBangumi.MikanItem.ID)
		fmt.Printf("  标题: %s\n", savedBangumi.MikanItem.OfficialTitle)
		fmt.Printf("  海报: %s\n", savedBangumi.MikanItem.PosterLink)
	}

	if savedBangumi.TmdbItem != nil {
		fmt.Printf("\n%s🎬 关联的 TMDB 信息:%s\n", green, reset)
		fmt.Printf("  ID: %d\n", savedBangumi.TmdbItem.ID)
		fmt.Printf("  标题: %s\n", savedBangumi.TmdbItem.Title)
		fmt.Printf("  原始标题: %s\n", savedBangumi.TmdbItem.OriginalTitle)
		fmt.Printf("  评分: %.1f\n", savedBangumi.TmdbItem.VoteAverage)
		fmt.Printf("  集数: %d\n", savedBangumi.TmdbItem.EpisodeCount)
	}

	fmt.Printf("\n%s%s=== 测试完成！ ===%s\n", cyan, bold, reset)
	fmt.Printf("%s💡 总结：只需要在 Bangumi 对象中设置 MikanItem/TmdbItem 字段%s\n", cyan, reset)
	fmt.Printf("%s   GORM 会自动保存关联对象并设置外键，无需手动操作！%s\n", cyan, reset)
}
