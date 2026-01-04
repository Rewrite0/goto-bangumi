package refresh

import (
	_ "embed"
	"os"
	"testing"

	"goto-bangumi/internal/network"
	"goto-bangumi/internal/parser"
)

// RSS 测试数据
//
//go:embed testdata/rss_mybangumi.xml
var rssMybangumiXML []byte

// Mikan 页面测试数据
//
//go:embed testdata/mikan_3774.html
var mikan3774HTML []byte

//go:embed testdata/mikan_3749.html
var mikan3749HTML []byte

//go:embed testdata/mikan_3676.html
var mikan3676HTML []byte

//go:embed testdata/mikan_3784.html
var mikan3784HTML []byte

//go:embed testdata/mikan_3774_ep02.html
var mikan3774Ep02HTML []byte

// TMDB 搜索测试数据
//
//go:embed testdata/tmdb_search_chitose.json
var tmdbSearchChitose []byte

//go:embed testdata/tmdb_search_koikoi.json
var tmdbSearchKoikoi []byte

//go:embed testdata/tmdb_search_tougen.json
var tmdbSearchTougen []byte

//go:embed testdata/tmdb_search_isekai.json
var tmdbSearchIsekai []byte

// TMDB 详情测试数据
//
//go:embed testdata/tmdb_info_261343.json
var tmdbInfo261343 []byte

//go:embed testdata/tmdb_info_282662.json
var tmdbInfo282662 []byte

//go:embed testdata/tmdb_info_253811.json
var tmdbInfo253811 []byte

//go:embed testdata/tmdb_info_87478.json
var tmdbInfo87478 []byte

// TestMain 设置所有测试缓存
func TestMain(m *testing.M) {
	// 设置 RSS 缓存
	rssURL := "https://mikanani.me/RSS/MyBangumi?token=test"
	network.SetTestCache(rssURL, rssMybangumiXML)

	// 设置 Mikan 页面缓存 - FindNewBangumi 测试用
	network.SetTestCache("https://mikanani.me/Home/Episode/46a4d69be33f6923c3eab31fe70e27b42b57a643", mikan3774HTML) // 弹珠汽水瓶里的千岁同学
	network.SetTestCache("https://mikanani.me/Home/Episode/123fc0383afd2ccd36f49da3d31f1348c2a029b7", mikan3749HTML) // 跨越种族与你相恋
	network.SetTestCache("https://mikanani.me/Home/Episode/00e085199ec6226948d9851c7a789ce67ba94c61", mikan3749HTML) // 跨越种族与你相恋 另一集
	network.SetTestCache("https://mikanani.me/Home/Episode/fa57b5211750399db0c02feac09f8888f4180c3d", mikan3676HTML) // 桃源暗鬼
	network.SetTestCache("https://mikanani.me/Home/Episode/826e7bd3625020312f68d724c025b2a646ac044d", mikan3784HTML) // 异世界四重奏
	network.SetTestCache("https://mikanani.me/Home/Episode/8104af2208a83747aab89198f66e2b1e8acfcb05", mikan3784HTML) // 异世界四重奏 合集
	network.SetTestCache("https://mikanani.me/Home/Episode/d2de7ee4aeb90901df425b2f2b1dd67cf1ad0f5b", mikan3774HTML) // 弹珠汽水瓶里的千岁同学 另一集

	// 设置 Mikan 页面缓存 - TorrentToBangumi 和 CreateBangumi 测试用
	network.SetTestCache("https://mikanani.me/Home/Episode/7c8c41e409922d9f2c34a726c92e77daf05558ff", mikan3774Ep02HTML) // 弹珠汽水瓶里的千岁同学 EP02

	// 设置 TMDB 搜索缓存
	network.SetTestCache(parser.SearchURL("弹珠汽水瓶里的千岁同学"), tmdbSearchChitose)
	network.SetTestCache(parser.SearchURL("跨越种族与你相恋"), tmdbSearchKoikoi)
	network.SetTestCache(parser.SearchURL("桃源暗鬼"), tmdbSearchTougen)
	network.SetTestCache(parser.SearchURL("异世界四重奏"), tmdbSearchIsekai)
	// 也缓存去空格版本
	network.SetTestCache(parser.SearchURL("异世界四重奏3"), tmdbSearchIsekai)

	// 设置 TMDB 详情缓存
	network.SetTestCache(parser.InfoURL(261343, "zh"), tmdbInfo261343) // 弹珠汽水瓶里的千岁同学
	network.SetTestCache(parser.InfoURL(282662, "zh"), tmdbInfo282662) // 跨越种族与你相恋
	network.SetTestCache(parser.InfoURL(253811, "zh"), tmdbInfo253811) // 桃源暗鬼
	network.SetTestCache(parser.InfoURL(87478, "zh"), tmdbInfo87478)   // 异世界四重奏

	code := m.Run()
	os.Exit(code)
}
