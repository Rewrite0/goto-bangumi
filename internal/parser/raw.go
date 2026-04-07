package parser

import (
	"strings"

	"goto-bangumi/internal/model"
)

var ParserConfig *model.RssParserConfig

func init() {
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
	metaParser := NewTitleMetaParse()
	episode := metaParser.Parse(title)
	if episode.Episode == -1 {
		return nil
	}
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
		Parse:         "raw",
		RSSLink:       "",
		PosterLink:    "",
		Deleted:       false,
	}
}
