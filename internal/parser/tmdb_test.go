package parser

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed testdata/tmdb_search_wolf.json
var tmdbSearchWolf []byte

//go:embed testdata/tmdb_info_229676.json
var tmdbInfo229676 []byte

//go:embed testdata/tmdb_search_hyakusho.json
var tmdbSearchHyakusho []byte

//go:embed testdata/tmdb_search_hyakusho_star.json
var tmdbSearchHyakushoStar []byte

func TestTMDBSearch(t *testing.T) {
	tests := []struct {
		name          string
		keyword       string
		wantCount     int
		wantFirstID   int
		wantFirstName string
	}{
		{
			name:          "百姓贵族 - 正常搜索",
			keyword:       "百姓贵族",
			wantCount:     1,
			wantFirstID:   221165,
			wantFirstName: "Hyakusho Kizoku-the farmer's days",
		},
		{
			name:      "★百姓贵族 - 带特殊字符返回空结果",
			keyword:   "★百姓贵族",
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTMDBParse()
			results, err := parser.TMDBSearch(tt.keyword)
			if err != nil {
				t.Fatalf("TMDBSearch() error = %v", err)
			}
			if len(results) != tt.wantCount {
				t.Errorf("TMDBSearch() results count = %d, want %d", len(results), tt.wantCount)
			}
			if tt.wantCount > 0 {
				if results[0].ID != tt.wantFirstID {
					t.Errorf("TMDBSearch() first result ID = %d, want %d", results[0].ID, tt.wantFirstID)
				}
				if results[0].Name != tt.wantFirstName {
					t.Errorf("TMDBSearch() first result Name = %s, want %s", results[0].Name, tt.wantFirstName)
				}
			}
			t.Logf("TMDBSearch(%s) returned %d results", tt.keyword, len(results))
		})
	}
}

func TestTmdbParse(t *testing.T){
	tests := []struct {
		name              string
		title             string
		mockSearchData    []byte // 模拟搜索 API 响应
		mockInfoData      []byte // 模拟详情 API 响应
		wantID            int
		wantTitle         string
		wantOriginalTitle string
		wantYear          string
		wantSeason        string
		wantPosterLink    string
	}{
		{
			name:              "狼与香辛料（使用缓存）",
			title:             "狼与香辛料",
			mockSearchData:    tmdbSearchWolf,
			mockInfoData:      tmdbInfo229676,
			wantID:            229676,
			wantTitle:         "狼与香辛料 行商邂逅贤狼",
			wantOriginalTitle: "狼と香辛料 MERCHANT MEETS THE WISE WOLF",
			wantYear:          "2024",
			wantSeason:        "1",
			wantPosterLink:    "https://image.tmdb.org/t/p/w780/vgfhyqA6n8WWiDhHXdVRBMHAqQw.jpg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 缓存已在 TestMain 中集中设置
			parser := NewTMDBParse()
			info, err := parser.TMDBParse(tt.title, "zh")
			if err != nil {
				t.Fatalf("TMDBParse() error = %v", err)
			}
			if info == nil {
				t.Fatalf("TMDBParse() returned nil for title: %s", tt.title)
			}
			if tt.wantID != 0 && info.ID != tt.wantID {
				t.Errorf("TMDBParse() ID = %v, want %v", info.ID, tt.wantID)
			}
			if tt.wantTitle != "" && info.Title != tt.wantTitle {
				t.Errorf("TMDBParse() Title = %v, want %v", info.Title, tt.wantTitle)
			}
			if tt.wantOriginalTitle != "" && info.OriginalTitle != tt.wantOriginalTitle {
				t.Errorf("TMDBParse() OriginalTitle = %v, want %v", info.OriginalTitle, tt.wantOriginalTitle)
			}
			if tt.wantYear != "" && info.Year != tt.wantYear {
				t.Errorf("TMDBParse() Year = %v, want %v", info.Year, tt.wantYear)
			}
			if tt.wantSeason != "" && fmt.Sprintf("%d", info.Season) != tt.wantSeason {
				t.Errorf("TMDBParse() Season = %v, want %v", info.Season, tt.wantSeason)
			}
			t.Logf("TMDBParse() returned: %+v", info)
		})
	}

}
