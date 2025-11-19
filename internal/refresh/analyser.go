package refresh

import (
	"log/slog"
	"strings"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser"
)

func OfficialTitleParse(torrent *model.Torrent) (*model.Bangumi, error) {
	bangumi := model.NewBangumi()
	if torrent.Homepage != "" {
		// 对于有 homepage 的, 默认进行一遍解析, 用以得到更准确的标题
		// 就算是 mikan 的, 也不一定有 homepage
		mikanParse := parser.NewMikanParser()
		mikanInfo, err := mikanParse.Parse(torrent.Homepage)
		// 这里要看看是网络问题还是解析问题, 不过感觉有 homepage ，那就一定是网络问题
		if err == nil {
			bangumi.OfficialTitle = mikanInfo.OfficialTitle
			bangumi.PosterLink = mikanInfo.PosterLink
			bangumi.MikanItem = mikanInfo
			bangumi.Season = mikanInfo.Season
		} else {
			slog.Debug("mikan 解析失败", slog.String("种子名称", torrent.Name), slog.String("错误信息", err.Error()))
			// 网络错误直接返回,不做后面的解析
			if apperrors.IsNetworkError(err) {
				return nil, err
			}
		}
	}
	if bangumi.Parse == "bangumi" {
		// Bangumi 解析, 没有做
	} else {
		tmdbParse := parser.NewTMDBParse()
		var title string
		if bangumi.OfficialTitle != "" {
			// 优先使用 mikan 解析到的标题
			title = bangumi.OfficialTitle
		} else {
			// 否则使用种子标题
			title = parser.NewTitleMetaParse().Parse(torrent.Name).Title
		}

		tmdbInfo, err := tmdbParse.TMDBParse(title, "zh")
		// 当 tmdb 也没有找到信息的时候，如果 mikan 也没有找到， 报错
		if err != nil {
			if bangumi.OfficialTitle == "" {
				return nil, err
			}
			return bangumi, err
		}
		// 只有在没有解析到标题的情况下才使用 tmdb 的结果
		if bangumi.OfficialTitle == "" {
			bangumi.OfficialTitle = tmdbInfo.Title
			bangumi.PosterLink = tmdbInfo.PosterLink
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
	slog.Debug("通过", slog.String("种子名称", torrent.Name))
	return true
}

// TorrentToBangumi 从 torrent 解析出 bangumi 信息,只会反回网络错误
func TorrentToBangumi(torrent *model.Torrent, rssLink string) (*model.Bangumi, error) {
	bangumi, err := OfficialTitleParse(torrent)
	metaInfo := parser.NewTitleMetaParse().Parse(torrent.Name)
	// 为空在两种可能
	// 1. torrent 的名字不太对, 当torrent 名字不对而没法解析的时候, 要显示bangumi
	// 2. 网络的问题 , 这会导致永远无法出来这个番剧,这是不对的
	// TODO: 后面会有一些合并， 现在先放着
	bangumi, err = OfficialTitleParse(torrent)
	// 对于网络错误, 不添加
	if err != nil && bangumi == nil {
		if apperrors.IsNetworkError(err) {
			return nil, err
		}
	}
	// 解析错误主要是 tmdb 没有找到对应的番剧
	// mikan 能解析成功的话,这里不会是 nil
	// 对于解析错误, 以 metaInfo 构造 bangumi
	if bangumi == nil {
		bangumi = model.NewBangumi()
		bangumi.OfficialTitle = metaInfo.Title
		bangumi.Season = metaInfo.Season
	}

	bangumi.IncludeFilter = strings.Join(parser.ParserConfig.Include, ",")
	bangumi.ExcludeFilter = strings.Join(parser.ParserConfig.Filter, ",")
	bangumi.RRSSLink = rssLink
	bangumi.EpisodeMetadata = append(bangumi.EpisodeMetadata, *metaInfo)
	return bangumi, nil
}

func createBangumi(torrent *model.Torrent, rssLink string) {
	bangumi, err := TorrentToBangumi(torrent, rssLink)
	if err != nil && apperrors.IsNetworkError(err) {
		slog.Warn("网络错误，跳过该番剧的添加", slog.String("种子名称", torrent.Name), slog.String("错误信息", err.Error()))
		return
	}
	// 对 mikan 部份错误进行处理

	slog.Debug("createBangumi", bangumi)
	if bangumi != nil {
		if torrent.Homepage != "" && bangumi.MikanItem == nil {
			// 这里对应 mikan 未添加的情况, 一般出现在季度初
			// TODO: 没想好怎么处理, 先放着
		}
		// 对 bangumi 进行处理，要看看有没有相同的 bangumi 项
		// 有相同的就只更新metadata
		db := database.GetDB()
		db.CreateBangumi(bangumi)
	}
}
