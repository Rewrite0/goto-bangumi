package baseparser

import (
	"os"
	"testing"
	"goto-bangumi/internal/network"
)

// TestMain 在所有测试运行前设置缓存
func TestMain(m *testing.M) {
	// 设置 TMDB 测试缓存 - 狼与香辛料
	searchURL := SearchURL("狼与香辛料")
	network.SetTestCache(searchURL, tmdbSearchWolf)

	infoURL := InfoURL(229676, "zh")
	network.SetTestCache(infoURL, tmdbInfo229676)

	// 设置 Mikan 测试缓存 - 拥有超常技能的异世界流浪美食家 第二季
	network.SetTestCache("https://mikanani.me/Home/Episode/8c94c1699735481c8b2b18dba38908042f53adcc", mikan3751HTML)

	// 设置 Mikan 测试缓存 - 妖怪旅馆营业中
	network.SetTestCache("https://mikanani.me/Home/Episode/f2340bae48a4c7eae1421190d603d4c889d490b7", mikan3790HTML)

	// 设置 Mikan 测试缓存 - 夏日口袋
	network.SetTestCache("https://mikanani.me/Home/Episode/8c2e3e9f7b71419a513d2647f5004f3a0f08a7f0", mikan3599HTML)

	// 设置 Mikan 测试缓存 - 边缘情况（无mikanID、无官方标题、无poster）
	network.SetTestCache("https://mikanani.me/Home/Episode/699000310671bae565c37abb20d119824efeb6f0", mikanEdgeCaseHTML)

	// 运行测试
	code := m.Run()

	// 退出
	os.Exit(code)
}
