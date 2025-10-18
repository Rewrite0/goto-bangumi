package baseparser

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed testdata/tmdb_search_wolf.json
var tmdbSearchWolf []byte

//go:embed testdata/tmdb_info_229676.json
var tmdbInfo229676 []byte

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
