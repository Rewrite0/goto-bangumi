package refresh

import (
	"fmt"
	"log/slog"
	"strings"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser"
	"goto-bangumi/internal/parser/baseparser"
)

func OfficialTitleParse(torrent model.Torrent) (*model.Bangumi, error) {
	bangumi := &model.Bangumi{}
	if torrent.Homepage != "" {
		// 对于有 homepage 的, 默认进行一遍解析, 用以得到更准确的标题
		// 就算是 mikan 的, 也不一定有 homepage
		mikanParse := baseparser.NewMikanParser()
		mikanInfo, err := mikanParse.Parse(torrent.Homepage)
		// 这里要看看是网络问题还是解析问题, 不过感觉有 homepage ，那就一定是网络问题
		if err == nil {
			bangumi.OfficialTitle = mikanInfo.OfficialTitle
			bangumi.PosterLink = mikanInfo.PosterLink
			bangumi.MikanItem = mikanInfo
			bangumi.Season = mikanInfo.Season
		} else {
			// TODO 对网络错误和解析错误进行区分
			slog.Debug("mikan 解析失败", slog.String("种子名称", torrent.Name), slog.String("错误信息", err.Error()))
		}
	}
	if bangumi.Parse == "bangumi" {
		// Bangumi 解析, 没有做
	} else {
		tmdbParse := baseparser.NewTMDBParse()
		var title string
		if bangumi.OfficialTitle != "" {
			// 优先使用 mikan 解析到的标题
			title = bangumi.OfficialTitle
		}
		if title == "" {
			// 否则使用种子标题
			title = baseparser.NewTitleMetaParse().Parse(torrent.Name).Title
		}

		tmdbInfo, err := tmdbParse.TMDBParse(title, "zh")
		// if err != nil {
		// 	if apperrors.IsNetworkError(err) {
		// 	}
		// }

		// 当 tmdb 也没有找到信息的时候，如果 mikan 也没有找到， 报错
		if err != nil && bangumi.OfficialTitle == "" {
			return nil, fmt.Errorf("无法解析番剧标题: %s, 错误信息: %s", torrent.Name, err.Error())
		}
		// 只有在没有解析到标题的情况下才使用 tmdb 的结果
		if bangumi.OfficialTitle == "" {
			bangumi.OfficialTitle = tmdbInfo.Title
			bangumi.PosterLink = tmdbInfo.PosterLink
			bangumi.TmdbItem = tmdbInfo
			bangumi.Parse = "tmdb"
		}
		// 总是以 tmdb 的季度为准
		bangumi.Season = tmdbInfo.Season
		bangumi.Year = tmdbInfo.Year
		bangumi.TmdbItem = tmdbInfo
	}
	return bangumi, nil
}

// FilterTorrent 通过bangumi信息判断torrent是否符合要求
func FilterTorrent(torrent *model.Torrent, bangumi *model.Bangumi) bool {
	// 排除过滤
	var exclude, include string
	if bangumi == nil {
		exclude = strings.Join(parser.ParserConfig.Filter, ",")
		include = strings.Join(parser.ParserConfig.Include, ",")

	} else {
		exclude = bangumi.ExcludeFilter
		include = bangumi.IncludeFilter
	}
	for v := range strings.SplitSeq(exclude, ",") {
		if v != "" && strings.Contains(torrent.Name, v) {
			slog.Debug("过滤种子", slog.String("种子名称", torrent.Name), slog.String("过滤关键词", v))
			return false
		}
	}
	// 包含过滤
	for v := range strings.SplitSeq(include, ",") {
		if v != "" && strings.Contains(torrent.Name, v) {
			slog.Debug("通过包含过滤", slog.String("种子名称", torrent.Name), slog.String("包含关键词", v))
			return true
		}
	}
	slog.Debug("通过", slog.String("种子名称", torrent.Name), slog.String("包含关键词", bangumi.IncludeFilter))
	return true
}

func TorrentToBangumi(torrent model.Torrent, rss model.RSSItem) *model.Bangumi {
	bangumi, err := OfficialTitleParse(torrent)
	metaInfo := baseparser.NewTitleMetaParse().Parse(torrent.Name)
	// 为空在两种可能
	// 1. torrent 的名字不太对, 当torrent 名字不对而没法解析的时候, 要显示bangumi
	// 2. 网络的问题 , 这会导致永远无法出来这个番剧,这是不对的
	// TODO: 后面会有一些合并， 现在先放着
	bangumi, err = OfficialTitleParse(torrent)
	if err != nil && apperrors.IsNetworkError(err) {
		return nil
	}
	bangumi.RssLink = rss.URL
	bangumi.EpisodeMetadata = append(bangumi.EpisodeMetadata, *metaInfo)
	return bangumi
}

func CreateBangumi(torrent model.Torrent, rss model.RSSItem) {
	bangumi := TorrentToBangumi(torrent, rss)
	if bangumi != nil {
		// 对 bangumi 进行处理，要看看有没有相同的 bangumi 项
		// 有相同的就只更新metadata
		db := database.GetDB()
		db.CreateBangumi(bangumi)
	}
}
