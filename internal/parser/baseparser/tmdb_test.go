package baseparser

import (
	"fmt"
	"testing"
)

func TestTmdbParser(t *testing.T){
	tests := []struct {
		name     string
		title    string
		wantID   int
		wantTitle string
		wantOriginalTitle string
		wantYear string
		wantSeason string
		wantPosterLink string
	}{
		{

			name:     "狼与香辛料",
			title:    "狼与香辛料",
			wantID:   229676,
			wantTitle: "狼与香辛料 行商邂逅贤狼",
			wantOriginalTitle: "狼と香辛料 MERCHANT MEETS THE WISE WOLF",
			wantYear: "2024",
			wantSeason: "1",
			wantPosterLink: "https://image.tmdb.org/t/p/w780/vgfhyqA6n8WWiDhHXdVRBMHAqQw.jpg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewTMDBParser()
			if err != nil {
				t.Fatalf("Failed to create TMDB parser: %v", err)
			}
			info, err := parser.TMDBParser(tt.title, "zh")
			fmt.Printf("TMDB Info: %+v\n", info)
			if err != nil {
				t.Fatalf("TMDBParser() error = %v", err)
			}
			if info == nil {
				t.Fatalf("TMDBParser() returned nil for title: %s", tt.title)
			}
			if tt.wantID != 0 && info.ID != tt.wantID {
				t.Errorf("TMDBParser() ID = %v, want %v", info.ID, tt.wantID)
			}
			if tt.wantTitle != "" && info.Title != tt.wantTitle {
				t.Errorf("TMDBParser() Title = %v, want %v", info.Title, tt.wantTitle)
			}
			if tt.wantOriginalTitle != "" && info.OriginalTitle != tt.wantOriginalTitle {
				t.Errorf("TMDBParser() OriginalTitle = %v, want %v", info.OriginalTitle, tt.wantOriginalTitle)
			}
			if tt.wantYear != "" && info.Year != tt.wantYear {
				t.Errorf("TMDBParser() Year = %v, want %v", info.Year, tt.wantYear)
			}
			if tt.wantSeason != "" && fmt.Sprintf("%d", info.LastSeason) != tt.wantSeason {
				t.Errorf("TMDBParser() LastSeason = %v, want %v", info.LastSeason, tt.wantSeason)
			}
			t.Logf("TMDBParser() returned: %+v", info)
		})
	}

}
