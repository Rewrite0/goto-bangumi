package refresh

import (
	// "os"
	"testing"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
)

func TestFilter_torrent(t *testing.T) {
	tests := []struct {
		name     string
		torrent  model.Torrent
		bangumi  model.Bangumi
		expected bool
	}{
		{
			name: "exclude false",
			torrent: model.Torrent{
				Name: "[喵萌奶茶屋&LoliHouse] 败犬女主角也太多了！ / 败犬女主太多了！ / 负けヒロインが多すぎる！ / Make Heroine ga Oosugiru! [01-12合集][WebRip 1080p HEVC-10bit AAC][简繁日内封字幕][Fin]",
			},
			bangumi: model.Bangumi{
				ExcludeFilter: "1080p,meow",
			},
			expected: false,
		},
		{
			name: "exclude true",
			torrent: model.Torrent{
				Name: "[喵萌奶茶屋&LoliHouse] 败犬女主角也太多了！ / 败犬女主太多了！ / 负けヒロインが多すぎる！ / Make Heroine ga Oosugiru! [01-12合集][WebRip 720p HEVC-10bit AAC][简繁日内封字幕][Fin]",
			},
			bangumi: model.Bangumi{
				ExcludeFilter: "1080p,meow",
			},
			expected: true,
		},
		{
			name: "include true",
			torrent: model.Torrent{
				Name: "[喵萌奶茶屋&LoliHouse] 败犬女主角也太多了！ / 败犬女主太多了！ / 负けヒロインが多すぎる！ / Make Heroine ga Oosugiru! [01-12合集][WebRip 1080p HEVC-10bit AAC][简繁日内封字幕][Fin]",
			},
			bangumi: model.Bangumi{
				IncludeFilter: "1080p,meow",
			},
			expected: true,
		},
		{
			name: "include false",
			torrent: model.Torrent{
				Name: "[喵萌奶茶屋&LoliHouse] 败犬女主角也太多了！ / 败犬女主太多了！ / 负けヒロインが多すぎる！ / Make Heroine ga Oosugiru! [01-12合集][WebRip 720p HEVC-10bit AAC][简繁日内封字幕][Fin]",
			},
			bangumi: model.Bangumi{
				IncludeFilter: "1080p,meow",
			},
			expected: true,
		},
		{
			name: "include empty",
			torrent: model.Torrent{
				Name: "[喵萌奶茶屋&LoliHouse] 败犬女主角也太多了！ / 败犬女主太多了！ / 负けヒロインが多すぎる！ / Make Heroine ga Oosugiru! [01-12合集][WebRip 720p HEVC-10bit AAC][简繁日内封字幕][Fin]",
			},
			bangumi:  model.Bangumi{},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterTorrent(&tt.torrent, &tt.bangumi)
			if result != tt.expected {
				t.Errorf("Filter_torrent() = %v, want %v,torrent name %s", result, tt.expected, tt.torrent.Name)
			}
		})
	}
}

func TestTorrentToBangumi(t *testing.T) {
	// ID: 0
	// 中文名: 弹珠汽水瓶里的千岁同学
	// 季度: 0
	// 解析器:
	// RSS链接: https://mikanani.me/RSS/Search?searchstr=ANI
	// 封面链接: https://mikanani.me/images/Bangumi/202510/37749647.jpg
	// Tmdb信息:
	//  TmdbID: 261343,
	//  Title: Chitose Is in the Ramune Bottle,
	//  Year: 2025,
	//  OriginalTitle: 千歳くんはラムネ瓶のなか,
	//  AirDate: 2025-10-07,
	//  EpisodeCount: 13,
	//  Season: 1,
	//  PosterLink: https://image.tmdb.org/t/p/w780/7tpcFkOpLcWkJU6mV5ooJyHA3DR.jpg,
	//  VoteAverage: 5.50
	// MikanID:
	// MikanID: 3774,
	//  OfficialTitle: 弹珠汽水瓶里的千岁同学,
	//  Season: 1,
	//  PosterLink: https://mikanani.me/images/Bangumi/202510/37749647.jpg
	tests := []struct {
		name        string
		torrent     model.Torrent
		rss         model.RSSItem
		wantBangumi *model.Bangumi
	}{
		{
			name: "test1",
			torrent: model.Torrent{
				Name:     "[ANi] Chitose Is in the Ramune Bottle / 弹珠汽水瓶里的千岁同学 - 02 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
				URL:      "magnet:?xt=urn:btih:EXAMPLE1",
				Homepage: "https://mikanani.me/Home/Episode/7c8c41e409922d9f2c34a726c92e77daf05558ff",
			},
			rss: model.RSSItem{
				Name: "Chitose Is in the Ramune Bottle / 弹珠汽水瓶里的千岁同学",
				URL:  "https://mikanani.me/RSS/Search?searchstr=ANI",
			},
			wantBangumi: &model.Bangumi{
				OfficialTitle: "弹珠汽水瓶里的千岁同学",
				RRSSLink:       "https://mikanani.me/RSS/Search?searchstr=ANI",
				Year:          "2025",
				Season:        1,
				PosterLink:    "https://mikanani.me/images/Bangumi/202510/37749647.jpg",
				MikanItem: &model.MikanItem{
					ID:            3774,
					OfficialTitle: "弹珠汽水瓶里的千岁同学",
					Season:        1,
					PosterLink:    "https://mikanani.me/images/Bangumi/202510/37749647.jpg",
				},
				TmdbItem: &model.TmdbItem{
					ID:            261343,
					Title:         "弹珠汽水瓶里的千岁同学",
					Year:          "2025",
					OriginalTitle: "千歳くんはラムネ瓶のなか",
					AirDate:       "2025-10-07",
					EpisodeCount:  13,
					Season:        1,
					PosterLink:    "https://image.tmdb.org/t/p/w780/7tpcFkOpLcWkJU6mV5ooJyHA3DR.jpg",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBangumi,err := TorrentToBangumi(&tt.torrent, tt.rss.URL)
			if err != nil {
				t.Errorf("TorrentToBangumi() error = %v, want nil", err)
				return
			}
			if gotBangumi == nil {
				t.Errorf("TorrentToBangumi() = nil, want non-nil")
				return
			}
			if gotBangumi.OfficialTitle != tt.wantBangumi.OfficialTitle {
				t.Errorf("OfficialTitle = %v, want %v", gotBangumi.OfficialTitle, tt.wantBangumi.OfficialTitle)
			}
			if gotBangumi.RRSSLink != tt.wantBangumi.RRSSLink {
				t.Errorf("RssLink = %v, want %v", gotBangumi.RRSSLink, tt.wantBangumi.RRSSLink)
			}
			if gotBangumi.Year != tt.wantBangumi.Year {
				t.Errorf("Year = %v, want %v", gotBangumi.Year, tt.wantBangumi.Year)
			}
			if gotBangumi.Season != tt.wantBangumi.Season {
				t.Errorf("Season = %v, want %v", gotBangumi.Season, tt.wantBangumi.Season)
			}
			if gotBangumi.PosterLink != tt.wantBangumi.PosterLink {
				t.Errorf("PosterLink = %v, want %v", gotBangumi.PosterLink, tt.wantBangumi.PosterLink)
			}
			if gotBangumi.MikanItem == nil {
				t.Errorf("MikanItem = nil, want non-nil")
			} else {
				if gotBangumi.MikanItem.ID != tt.wantBangumi.MikanItem.ID {
					t.Errorf("MikanItem.ID = %v, want %v", gotBangumi.MikanItem.ID, tt.wantBangumi.MikanItem.ID)
				}
				if gotBangumi.MikanItem.OfficialTitle != tt.wantBangumi.MikanItem.OfficialTitle {
					t.Errorf("MikanItem.OfficialTitle = %v, want %v", gotBangumi.MikanItem.OfficialTitle, tt.wantBangumi.MikanItem.OfficialTitle)
				}
			}
			if gotBangumi.TmdbItem == nil {
				t.Errorf("TmdbItem = nil, want non-nil")
			} else {
				if gotBangumi.TmdbItem.ID != tt.wantBangumi.TmdbItem.ID {
					t.Errorf("TmdbItem.ID = %v, want %v", gotBangumi.TmdbItem.ID, tt.wantBangumi.TmdbItem.ID)
				}
				if gotBangumi.TmdbItem.Title != tt.wantBangumi.TmdbItem.Title {
					t.Errorf("TmdbItem.Title = %v, want %v", gotBangumi.TmdbItem.Title, tt.wantBangumi.TmdbItem.Title)
				}
			}
		})
	}
}

func TestCreateBangumi(t *testing.T) {
	// 创建内存数据库，测试完成后自动释放
	memoryDB := ":memory:"
	db, err := database.NewDB(&memoryDB)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	defer db.Close()

	torrent := &model.Torrent{
		Name:     "[ANi] Chitose Is in the Ramune Bottle / 弹珠汽水瓶里的千岁同学 - 02 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
		URL:      "magnet:?xt=urn:btih:EXAMPLE1",
		Homepage: "https://mikanani.me/Home/Episode/7c8c41e409922d9f2c34a726c92e77daf05558ff",
	}
	rssLink := "https://mikanani.me/RSS/Search?searchstr=ANI"

	// 调用被测函数
	createBangumi(db, torrent, rssLink)

	// 验证数据库中是否创建了番剧
	bangumi, err := db.GetBangumiByOfficialTitle("弹珠汽水瓶里的千岁同学")
	if err != nil {
		t.Fatalf("查询番剧失败: %v", err)
	}
	if bangumi == nil {
		t.Fatal("未找到期望的番剧: 弹珠汽水瓶里的千岁同学")
	}

	// 验证基础信息
	if bangumi.OfficialTitle != "弹珠汽水瓶里的千岁同学" {
		t.Errorf("OfficialTitle = %v, want 弹珠汽水瓶里的千岁同学", bangumi.OfficialTitle)
	}
	if bangumi.Year != "2025" {
		t.Errorf("Year = %v, want 2025", bangumi.Year)
	}
	if bangumi.Season != 1 {
		t.Errorf("Season = %v, want 1", bangumi.Season)
	}
	if bangumi.RRSSLink != rssLink {
		t.Errorf("RRSSLink = %v, want %v", bangumi.RRSSLink, rssLink)
	}

	// 验证 MikanItem
	if bangumi.MikanID == nil || *bangumi.MikanID != 3774 {
		t.Errorf("MikanID = %v, want 3774", bangumi.MikanID)
	}

	// 验证 TmdbItem
	if bangumi.TmdbID == nil || *bangumi.TmdbID != 261343 {
		t.Errorf("TmdbID = %v, want 261343", bangumi.TmdbID)
	}
}
