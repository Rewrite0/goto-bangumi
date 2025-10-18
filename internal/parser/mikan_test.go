package parser

import (
	_ "embed"
	"testing"

	"goto-bangumi/internal/network"
)

//go:embed testdata/mikan_3599.html
var mikan3599HTML []byte

//go:embed testdata/mikan_3751.html
var mikan3751HTML []byte

//go:embed testdata/mikan_3790.html
var mikan3790HTML []byte

func TestMikanParse(t *testing.T) {
	parser := NewMikanParse()
	tests := []struct {
		name       string
		homepage   string
		mockHTML   []byte // 如果不为空，则使用缓存模拟数据
		wantTitle  string
		wantID     int
		wantSeason int
		wantPoster string
	}{
		{
			name:       "拥有超常技能的异世界流浪美食家 第二季（使用缓存）",
			homepage:   "https://mikanani.me/Home/Episode/8c94c1699735481c8b2b18dba38908042f53adcc",
			mockHTML:   mikan3751HTML,
			wantTitle:  "拥有超常技能的异世界流浪美食家",
			wantID:     3751,
			wantSeason: 2,
			wantPoster: "https://mikanani.me/images/Bangumi/202510/0710007f.jpg",
		},
		{
			name:       "妖怪旅馆营业中（使用缓存）",
			homepage:   "https://mikanani.me/Home/Episode/f2340bae48a4c7eae1421190d603d4c889d490b7",
			mockHTML:   mikan3790HTML,
			wantTitle:  "妖怪旅馆营业中",
			wantID:     3790,
			wantSeason: 2,
			wantPoster: "https://mikanani.me/images/Bangumi/202510/0d10efc3.jpg",
		},
		{
			name:       "夏日口袋（使用缓存）",
			homepage:   "https://mikanani.me/Home/Episode/8c2e3e9f7b71419a513d2647f5004f3a0f08a7f0",
			mockHTML:   mikan3599HTML,
			wantTitle:  "夏日口袋",
			wantID:     3599,
			wantSeason: 1,
			wantPoster: "https://mikanani.me/images/Bangumi/202504/076c1094.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 如果提供了 mockHTML，则将其加入缓存
			if tt.mockHTML != nil {
				network.SetTestCache(tt.homepage, tt.mockHTML)
			}

			mikanInfo := parser.Parse(tt.homepage)
			if mikanInfo == nil {
				t.Fatalf("Parse(%q) returned nil", tt.homepage)
			}
			if mikanInfo.OfficialTitle != tt.wantTitle {
				t.Errorf("OfficialTitle = %q, want %q", mikanInfo.OfficialTitle, tt.wantTitle)
			}
			if mikanInfo.Season != tt.wantSeason {
				t.Errorf("Season = %d, want %d", mikanInfo.Season, tt.wantSeason)
			}
			if mikanInfo.PosterLink != tt.wantPoster {
				t.Errorf("PosterLink = %q, want %q", mikanInfo.PosterLink, tt.wantPoster)
			}
		})
	}
}
