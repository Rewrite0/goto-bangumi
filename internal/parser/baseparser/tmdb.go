package baseparser

import (
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
)

const (
	// tmdbURL is the base URL for TMDB API
	tmdbURL = "https://api.themoviedb.org"

	// tmdbImgURL is the base URL for TMDB images
	tmdbImgURL = "https://image.tmdb.org/t/p/w780"

	// Note: Using a public key is not recommended for production use.
	k = "291237f90b24267380d6176c98f7619f"
)

var tmdbKey string = k

var Language = map[string]string{
	"zh": "zh-CN",
	"jp": "ja-JP",
	"en": "en-US",
}

func Init(key string) {
	tmdbKey = key
}

// TMDBParser handles TMDB API interactions and parsing
type TMDBParser struct {
	client *network.RequestClient
}

// SearchURL 生成 TMDB 搜索 URL
func SearchURL(keyword string) string {
	// 对 keyword 进行 URL 编码
	keyword = strings.ReplaceAll(keyword, " ", "%20")
	keyword = url.QueryEscape(keyword)
	// 使用 tmdbKey 作为 API Key
	return fmt.Sprintf("%s/3/search/tv?api_key=%s&page=1&query=%s&include_adult=false",
		tmdbURL, tmdbKey, keyword)
}

// InfoURL 生成 TMDB 详情 URL
func InfoURL(showID int, language string) string {
	lang, ok := Language[language]
	if !ok {
		lang = Language["zh"] // Default to English if language not found
	}
	return fmt.Sprintf("%s/3/tv/%d?api_key=%s&language=%s",
		tmdbURL, showID, tmdbKey, lang)
}

// NewTMDBParse creates a new TMDB parser instance
func NewTMDBParse() *TMDBParser {
	client := network.NewRequestClient()

	return &TMDBParser{
		client: client,
	}
}

// TMDBSearch searches for TV shows on TMDB by keyword
func (p *TMDBParser) TMDBSearch(keyword string) ([]model.ShowInfo, error) {
	url := SearchURL(keyword)
	slog.Debug("[TMDB] Searching TV shows", "keyword", keyword, "url", url)

	var searchResult model.SearchResult
	if err := p.client.GetJSONTo(url, &searchResult); err != nil {
		return nil, err
	}

	slog.Debug("[TMDB] Search completed", "results_count", len(searchResult.Results))
	return searchResult.Results, nil
}

// TMDBInfo fetches detailed information for a specific TV show
func (p *TMDBParser) TMDBInfo(id int, language string) (*model.TVShow, error) {
	url := InfoURL(id, language)
	slog.Debug("[TMDB] Fetching TV show info", "id", id, "language", language)

	var tvShow model.TVShow
	if err := p.client.GetJSONTo(url, &tvShow); err != nil {
		return nil, err
	}

	slog.Debug("[TMDB] TV show info fetched", "name", tvShow.Name, "seasons", len(tvShow.Seasons))
	return &tvShow, nil
}

// IsAnimation checks if a show is an animation based on genre IDs
// Genre ID 16 represents Animation in TMDB
func IsAnimation(genreIDs []int) bool {
	return slices.Contains(genreIDs, 16)
}

// GetSeason 到最新的已播季度
// Returns season number and poster path
func GetSeason(seasons []model.SeasonTmdb) model.SeasonTmdb {
	validSeasons := make([]model.SeasonTmdb, 0)
	currentTime := time.Now()
	for _, s := range seasons {
		if s.AirDate != "" && s.SeasonNumber > 0 {
			airDate, err := time.Parse("2006-01-02", s.AirDate)
			if err != nil {
				continue
			}
			// 只考虑已经播出的季度 , 有一些确定播出的季度 AirDate 也会有时间,不能加入这种
			if airDate.Before(currentTime) {
				validSeasons = append(validSeasons, s)
			}
		}
	}
	// 不知道为什么会这样, 对没有有效的处理一下
	if len(validSeasons) == 0 {
		for _, s := range seasons {
			if s.SeasonNumber > 0 {
				return s
			}
		}
	}

	// 按 AirDate 降序排序
	slog.Debug("[TMDB] Valid seasons found", "count", len(validSeasons))
	sort.Slice(validSeasons, func(i, j int) bool {
		return validSeasons[i].AirDate > validSeasons[j].AirDate
	})

	// 时间不能早于现在时间
	lastSeason := validSeasons[len(validSeasons)-1]
	return lastSeason
}

// FindAnimation finds the first animation from a list of search results
// Results are sorted by first_air_date (newest first)
func FindAnimation(contents []model.ShowInfo) *model.ShowInfo {
	// Sort by first_air_date in descending order
	// 按 first_air_date 降序排序
	sortedContents := make([]model.ShowInfo, len(contents))
	copy(sortedContents, contents)
	sort.Slice(sortedContents, func(i, j int) bool {
		return sortedContents[i].FirstAirDate > sortedContents[j].FirstAirDate
	})

	// 返回第一个动画
	for i := range sortedContents {
		if IsAnimation(sortedContents[i].GenreIds) {
			return &sortedContents[i]
		}
	}

	return nil
}

// TMDBParse searches and parses TMDB information for a bangumi
// Returns TMDBInfo or nil if not found
func (p *TMDBParser) TMDBParse(title string, language string) (*model.TmdbItem, error) {
	slog.Debug("[TMDB] Starting TMDB parser", "title", title, "language", language)

	// First search attempt
	contents, err := p.TMDBSearch(title)
	if err != nil {
		return nil, fmt.Errorf("failed to search TMDB: %w", err)
	}

	// 尝试去掉空格再搜索一次
	if len(contents) == 0 {
		slog.Debug("[TMDB] No results found, retrying without spaces")
		titleNoSpaces := strings.ReplaceAll(title, " ", "")
		contents, err = p.TMDBSearch(titleNoSpaces)
		if err != nil {
			return nil, fmt.Errorf("failed to search TMDB (no spaces): %w", err)
		}
	}

	// Still no results, return nil
	if len(contents) == 0 {
		slog.Warn("[TMDB] No results found for title", "title", title)
		return nil, nil
	}

	// 只对搜索结果中的动画进行处理
	// 不用考虑新的还没发布的问题, tmdb 没有的不应该会有种子
	content := FindAnimation(contents)
	if content == nil {
		slog.Warn("[TMDB] No animation found in search results", "title", title)
		return nil, nil
	}

	slog.Debug("[TMDB] Animation found", "name", content.Name, "id", content.ID)

	// 对找到的动画获取详细信息
	tvShow, err := p.TMDBInfo(content.ID, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get TV show info: %w", err)
	}

	// 用最后一个季度作为当前季度
	lastSeason := GetSeason(tvShow.Seasons)

	// 不以季度的年份为准， 因为tmdb 是以最先播出的时间为准
	seasonTime, err := time.Parse("2006-01-02", content.FirstAirDate)
	var year string
	if err != nil {
		year = strconv.Itoa(time.Now().Year())
	} else {
		year = strconv.Itoa(seasonTime.Year())
	}

	// 构造海报链接
	posterLink := tmdbImgURL + lastSeason.PosterPath

	tmdbInfo := &model.TmdbItem{
		ID:            tvShow.ID,
		Year:          year,
		OriginalTitle: tvShow.OriginalName,
		AirDate:       lastSeason.AirDate,
		EpisodeCount:  lastSeason.EpisodeCount,
		Title:         tvShow.Name,
		Season:        lastSeason.SeasonNumber,
		PosterLink:    posterLink,
		VoteAverage:   tvShow.VoteAverage,
	}

	return tmdbInfo, nil
}

// // Close closes the TMDB parser and releases resources
// func (p *TMDBParser) Close() error {
// 	if p.client != nil {
// 		return p.client.Close()
// 	}
// 	return nil
// }

// ParseTMDB is a convenience function that creates a parser, parses, and closes
func ParseTMDB(title string, language string) (*model.TmdbItem, error) {
	parser := NewTMDBParse()

	return parser.TMDBParse(title, language)
}
