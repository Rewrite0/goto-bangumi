package baseparser

import (
	"testing"
)

func TestMikanParser(t *testing.T) {
	parser := NewMikanParser()
	tests := []struct {
		name       string
		homepage   string
		wantID     string
		wantTitle  string
		wantSeason int
		wantPoster string
	}{
		{
			name:       "拥有超常技能的异世界流浪美食家 第二季",
			homepage:   "https://mikanani.me/Home/Episode/8c94c1699735481c8b2b18dba38908042f53adcc",
			wantID:     "3751#1230",
			wantTitle:  "拥有超常技能的异世界流浪美食家",
			wantSeason: 2,
			wantPoster: "https://mikanani.me/images/Bangumi/202510/0710007f.jpg",
		},
		{
			name:       "妖怪旅馆营业中",
			homepage:   "https://mikanani.me/Home/Episode/f2340bae48a4c7eae1421190d603d4c889d490b7",
			wantID:     "3790#370",
			wantTitle:  "妖怪旅馆营业中",
			wantSeason: 2,
			wantPoster: "https://mikanani.me/images/Bangumi/202510/0d10efc3.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mikanInfo, _ := parser.Parse(tt.homepage)
			if mikanInfo == nil {
				t.Fatalf("Parse(%q) returned nil", tt.homepage)
			}
			if mikanInfo.ID != tt.wantID {
				t.Errorf("MikanID = %q, want %q", mikanInfo.ID, tt.wantID)
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

func TestMikanPoster(t *testing.T) {
	parser := NewMikanParser()
	tests := []struct {
		name       string
		homepage   string
		wantID     string
		wantTitle  string
		wantSeason int
		wantPoster string
	}{
		{
			name:       "拥有超常技能的异世界流浪美食家 第二季",
			homepage:   "https://mikanani.me/Home/Episode/8c94c1699735481c8b2b18dba38908042f53adcc",
			wantPoster: "https://mikanani.me/images/Bangumi/202510/0710007f.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posterLink, _ := parser.PosterParser(tt.homepage)
			if posterLink == "" {
				t.Fatalf("Parse(%q) returned nil", tt.homepage)
			}
			if posterLink != tt.wantPoster {
				t.Errorf("PosterLink = %q, want %q", posterLink, tt.wantPoster)
			}
		})
	}
}
