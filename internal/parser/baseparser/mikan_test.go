package baseparser

import (
	_ "embed"
	"testing"

	"goto-bangumi/internal/apperrors"
)

//go:embed testdata/mikan_3599.html
var mikan3599HTML []byte

//go:embed testdata/mikan_3751.html
var mikan3751HTML []byte

//go:embed testdata/mikan_3790.html
var mikan3790HTML []byte

//go:embed testdata/mikan_edge_case.html
var mikanEdgeCaseHTML []byte

func TestMikanParse(t *testing.T) {
	parser := NewMikanParser()
	tests := []struct {
		name        string
		homepage    string
		mockHTML    []byte // 如果不为空，则使用缓存模拟数据
		wantMikanID int
		wantTitle   string
		wantSeason  int
		wantPoster  string
	}{
		{
			name:        "拥有超常技能的异世界流浪美食家 第二季（使用缓存）",
			homepage:    "https://mikanani.me/Home/Episode/8c94c1699735481c8b2b18dba38908042f53adcc",
			mockHTML:    mikan3751HTML,
			wantMikanID: 3751,
			wantTitle:   "拥有超常技能的异世界流浪美食家",
			wantSeason:  2,
			wantPoster:  "https://mikanani.me/images/Bangumi/202510/0710007f.jpg",
		},
		{
			name:        "妖怪旅馆营业中（使用缓存）",
			homepage:    "https://mikanani.me/Home/Episode/f2340bae48a4c7eae1421190d603d4c889d490b7",
			mockHTML:    mikan3790HTML,
			wantMikanID: 3790,
			wantTitle:   "妖怪旅馆营业中",
			wantSeason:  2,
			wantPoster:  "https://mikanani.me/images/Bangumi/202510/0d10efc3.jpg",
		},
		{
			name:        "夏日口袋（使用缓存）",
			homepage:    "https://mikanani.me/Home/Episode/8c2e3e9f7b71419a513d2647f5004f3a0f08a7f0",
			mockHTML:    mikan3599HTML,
			wantMikanID: 3599,
			wantTitle:   "夏日口袋",
			wantSeason:  1,
			wantPoster:  "https://mikanani.me/images/Bangumi/202504/076c1094.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 缓存已在 TestMain 中集中设置
			mikanInfo, _ := parser.Parse(tt.homepage)
			if mikanInfo == nil {
				t.Fatalf("Parse(%q) returned nil", tt.homepage)
			}
			if mikanInfo.ID != tt.wantMikanID {
				t.Errorf("MikanID = %d, want %d", mikanInfo.ID, tt.wantMikanID)
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
		mockHTML   []byte // 如果不为空，则使用缓存模拟数据
		wantID     string
		wantTitle  string
		wantSeason int
		wantPoster string
	}{
		{
			name:       "拥有超常技能的异世界流浪美食家 第二季（使用缓存）",
			homepage:   "https://mikanani.me/Home/Episode/8c94c1699735481c8b2b18dba38908042f53adcc",
			mockHTML:   mikan3751HTML,
			wantPoster: "https://mikanani.me/images/Bangumi/202510/0710007f.jpg",
		},
		{
			name:       "夏日口袋（使用缓存）",
			homepage:   "https://mikanani.me/Home/Episode/8c2e3e9f7b71419a513d2647f5004f3a0f08a7f0",
			mockHTML:   mikan3599HTML,
			wantPoster: "https://mikanani.me/images/Bangumi/202504/076c1094.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 缓存已在 TestMain 中集中设置
			posterLink, _ := parser.PosterParse(tt.homepage)
			if posterLink == "" {
				t.Fatalf("Parse(%q) returned nil", tt.homepage)
			}
			if posterLink != tt.wantPoster {
				t.Errorf("PosterLink = %q, want %q", posterLink, tt.wantPoster)
			}
		})
	}
}

// TestMikanParseEdgeCase 测试边缘情况：没有mikanID、没有官方标题、没有poster链接
func TestMikanParseEdgeCase(t *testing.T) {
	parser := NewMikanParser()

	t.Run("没有RSS链接（无法获取mikanID）", func(t *testing.T) {
		homepage := "https://mikanani.me/Home/Episode/699000310671bae565c37abb20d119824efeb6f0"
		// 缓存已在 TestMain 中集中设置
		mikanInfo, err := parser.Parse(homepage)

		// 应该返回错误，因为页面没有RSS链接
		if err == nil {
			t.Errorf("Parse(%q) expected error, got nil", homepage)
		}

		// 验证是否返回了 ParseError
		if err != nil {
			if !apperrors.IsParseError(err) {
				t.Errorf("Parse(%q) expected ParseError, got %T: %v", homepage, err, err)
			}
		}

		// mikanInfo 应该为 nil
		if mikanInfo != nil {
			t.Errorf("Parse(%q) expected nil, got %+v", homepage, mikanInfo)
		}
	})

	// mikan 是有默认图片的，所以不会报错
	t.Run("没有poster链接时使用默认图片", func(t *testing.T) {
		homepage := "https://mikanani.me/Home/Episode/699000310671bae565c37abb20d119824efeb6f0"
		// 缓存已在 TestMain 中集中设置
		posterLink, err := parser.PosterParse(homepage)
		// 应该能够解析出poster链接（即使是默认图片）
		if err != nil {
			t.Logf("PosterParse error (expected): %v", err)
		}

		// 验证是否返回了默认图片链接
		if posterLink == "" {
			t.Errorf("PosterParse(%q) expected default image link, got empty string", homepage)
		}
	})
}
