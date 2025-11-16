package network

import (
	_ "embed"
	"os"
	"testing"
)

//go:embed testdata/rss_3391_583.xml
var rss3391583XML []byte

// TestMain 在所有测试运行前设置缓存
func TestMain(m *testing.M) {
	// 设置测试缓存
	rssURL := "https://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=583"
	SetTestCache(rssURL, rss3391583XML)

	// 运行测试
	code := m.Run()

	// 退出
	os.Exit(code)
}

func TestGetRSS(t *testing.T) {
	rssURL := "https://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=583"

	netClient := GetRequestClient()
	rss, err := netClient.GetRSS(rssURL)
	if err != nil {
		t.Fatalf("Error fetching RSS: %v", err)
	}

	// 验证 RSS 标题
	expectedTitle := "Mikan Project - 败犬女主太多了！"
	if rss.Title != expectedTitle {
		t.Errorf("Title = %q, want %q", rss.Title, expectedTitle)
	}

	// 验证 RSS Link
	expectedLink := "http://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=583"
	if rss.Link != expectedLink {
		t.Errorf("Link = %q, want %q", rss.Link, expectedLink)
	}

	// 验证种子数量
	expectedTorrentCount := 12
	if len(rss.Torrents) != expectedTorrentCount {
		t.Errorf("Torrent count = %d, want %d", len(rss.Torrents), expectedTorrentCount)
	}

	// 验证第一个种子的信息
	if len(rss.Torrents) > 0 {
		firstTorrent := rss.Torrents[0]
		expectedName := "[ANi] Make Heroine ga Oosugiru /  败北女角太多了！ - 12 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]"
		if firstTorrent.Name != expectedName {
			t.Errorf("First torrent name = %q, want %q", firstTorrent.Name, expectedName)
		}

		expectedEnclosureURL := "https://mikanani.me/Download/20240929/33fbab8f53fe4bad12f07afa5abdb7c4afa5956c.torrent"
		if firstTorrent.Enclosure.URL != expectedEnclosureURL {
			t.Errorf("First torrent enclosure URL = %q, want %q", firstTorrent.Enclosure.URL, expectedEnclosureURL)
		}
	}
}
func TestGetRSSTitle(t *testing.T) {
	rssURL := "https://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=583"

	netClient := GetRequestClient()
	title, err := netClient.GetRSSTitle(rssURL)
	if err != nil {
		t.Errorf("Error fetching RSS title: %v", err)
	}
	expectedTitle := "Mikan Project - 败犬女主太多了！"
	if title != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, title)
	}
}

func TestGetTorrents(t *testing.T) {
	rssURL := "https://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=583"

	netClient := GetRequestClient()
	torrents, err := netClient.GetTorrents(rssURL)
	if err != nil {
		t.Fatalf("Error fetching torrents: %v", err)
	}

	// 验证种子数量
	expectedTorrentCount := 12
	if len(torrents) != expectedTorrentCount {
		t.Errorf("Torrent count = %d, want %d", len(torrents), expectedTorrentCount)
	}

	// 验证第一个种子
	if len(torrents) > 0 {
		firstTorrent := torrents[0]

		// 验证名称
		expectedName := "[ANi] Make Heroine ga Oosugiru /  败北女角太多了！ - 12 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]"
		if firstTorrent.Name != expectedName {
			t.Errorf("First torrent name = %q, want %q", firstTorrent.Name, expectedName)
		}

		// 验证 URL（应该是 Enclosure.URL）
		expectedURL := "https://mikanani.me/Download/20240929/33fbab8f53fe4bad12f07afa5abdb7c4afa5956c.torrent"
		if firstTorrent.URL != expectedURL {
			t.Errorf("First torrent URL = %q, want %q", firstTorrent.URL, expectedURL)
		}

		// 验证 Homepage（应该是 Link）
		expectedHomepage := "https://mikanani.me/Home/Episode/33fbab8f53fe4bad12f07afa5abdb7c4afa5956c"
		if firstTorrent.Homepage != expectedHomepage {
			t.Errorf("First torrent Homepage = %q, want %q", firstTorrent.Homepage, expectedHomepage)
		}
	}

	// 验证最后一个种子
	if len(torrents) == expectedTorrentCount {
		lastTorrent := torrents[expectedTorrentCount-1]

		expectedName := "[ANi] Make Heroine ga Oosugiru /  败北女角太多了！ - 01 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]"
		if lastTorrent.Name != expectedName {
			t.Errorf("Last torrent name = %q, want %q", lastTorrent.Name, expectedName)
		}

		expectedURL := "https://mikanani.me/Download/20240714/4a6f89e788f32e84e65f4b14d33cf0964ad68c48.torrent"
		if lastTorrent.URL != expectedURL {
			t.Errorf("Last torrent URL = %q, want %q", lastTorrent.URL, expectedURL)
		}
	}
}
