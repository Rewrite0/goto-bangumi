package parser

import (
	"strings"

	"goto-bangumi/internal/model"
)

var ParserConfig *model.RssParserConfig

func init() {
	// 避免没有调用Init时报错
	ParserConfig = &model.RssParserConfig{}
}

func Init(config *model.RssParserConfig) {
	if config != nil {
		ParserConfig = config
	}
	InitTmdb(config.TmdbAPIKey)
}

type RawParse struct{}

func (p *RawParse) Parse(title string) *model.Bangumi {
	// language := "zh"
	metaParser := NewTitleMetaParse()
	episode := metaParser.Parse(title)
	if episode.Episode == -1 {
		return nil
	}
	// 依据 language 选择标题
	var officialTitle string
	season := episode.Season
	return &model.Bangumi{
		OfficialTitle: officialTitle,
		Year:          episode.Year,
		Season:        season,
		EpsCollect:    false,
		Offset:        0,
		IncludeFilter: strings.Join(ParserConfig.Include, ","),
		ExcludeFilter: strings.Join(ParserConfig.Filter, ","),
		Parse:        "raw",
		RRSSLink:       "",
		PosterLink:    "",
		Deleted:       false,
	}
}
