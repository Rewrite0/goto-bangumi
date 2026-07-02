package rss

import (
	"context"
	_ "embed"
	"testing"
)

//go:embed testdata/rss_3391_583.xml
var rss3391583XML []byte

type fakeGetter struct {
	data []byte
	err  error
}

func (g fakeGetter) Get(ctx context.Context, url string) ([]byte, error) {
	return g.data, g.err
}

func TestParse(t *testing.T) {
	feed, err := Parse(rss3391583XML)
	if err != nil {
		t.Fatalf("Error parsing RSS: %v", err)
	}

	expectedTitle := "Mikan Project - 败犬女主太多了！"
	if feed.Title != expectedTitle {
		t.Errorf("Title = %q, want %q", feed.Title, expectedTitle)
	}

	expectedLink := "http://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=583"
	if feed.Link != expectedLink {
		t.Errorf("Link = %q, want %q", feed.Link, expectedLink)
	}

	expectedTorrentCount := 12
	if len(feed.Items) != expectedTorrentCount {
		t.Errorf("Torrent count = %d, want %d", len(feed.Items), expectedTorrentCount)
	}

	if len(feed.Items) > 0 {
		firstTorrent := feed.Items[0]
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

func TestGetTitle(t *testing.T) {
	rssURL := "https://mikanani.me/RSS/Bangumi?bangumiId=3391&subgroupid=583"

	title, err := GetTitle(context.Background(), fakeGetter{data: rss3391583XML}, rssURL)
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

	torrents, err := GetTorrents(context.Background(), fakeGetter{data: rss3391583XML}, rssURL)
	if err != nil {
		t.Fatalf("Error fetching torrents: %v", err)
	}

	expectedTorrentCount := 12
	if len(torrents) != expectedTorrentCount {
		t.Errorf("Torrent count = %d, want %d", len(torrents), expectedTorrentCount)
	}

	if len(torrents) > 0 {
		firstTorrent := torrents[0]

		expectedName := "[ANi] Make Heroine ga Oosugiru /  败北女角太多了！ - 12 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]"
		if firstTorrent.Name != expectedName {
			t.Errorf("First torrent name = %q, want %q", firstTorrent.Name, expectedName)
		}

		expectedURL := "https://mikanani.me/Download/20240929/33fbab8f53fe4bad12f07afa5abdb7c4afa5956c.torrent"
		if firstTorrent.Link != expectedURL {
			t.Errorf("First torrent URL = %q, want %q", firstTorrent.Link, expectedURL)
		}

		expectedHomepage := "https://mikanani.me/Home/Episode/33fbab8f53fe4bad12f07afa5abdb7c4afa5956c"
		if firstTorrent.Homepage != expectedHomepage {
			t.Errorf("First torrent Homepage = %q, want %q", firstTorrent.Homepage, expectedHomepage)
		}
	}

	if len(torrents) == expectedTorrentCount {
		lastTorrent := torrents[expectedTorrentCount-1]

		expectedName := "[ANi] Make Heroine ga Oosugiru /  败北女角太多了！ - 01 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]"
		if lastTorrent.Name != expectedName {
			t.Errorf("Last torrent name = %q, want %q", lastTorrent.Name, expectedName)
		}

		expectedURL := "https://mikanani.me/Download/20240714/4a6f89e788f32e84e65f4b14d33cf0964ad68c48.torrent"
		if lastTorrent.Link != expectedURL {
			t.Errorf("Last torrent URL = %q, want %q", lastTorrent.Link, expectedURL)
		}
	}
}
