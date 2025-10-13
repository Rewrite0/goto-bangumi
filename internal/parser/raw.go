package parser

import (
	"strings"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser/baseparser"
)

var parserConfig *model.RssParserConfig

func init() {
	// 避免没有调用Init时报错
	parserConfig = &model.RssParserConfig{}
}

func Init(config *model.RssParserConfig) {
	if config != nil {
		parserConfig = config
	}
}

type RawParser struct{}

func (p *RawParser) Parse(title string) *model.Bangumi {
	// language := "zh"
	meta_parser := baseparser.NewTitleMetaParser()
	episode := meta_parser.Parser(title)
	if episode.Episode == -1 {
		return nil
	}
	title_raw := episode.TitleRaw
	// if title_raw == "" {
	// 	title_raw = episode.TitleZh
	// }
	// if title_raw == "" {
	// 	title_raw = episode.TitleJp
	// }
	// 依据 language 选择标题
	var official_title string
	// if language == "zh" {
	// 	official_title = episode.TitleZh
	// } else if language == "en" {
	// 	official_title = episode.TitleEn
	// } else {
	// 	official_title = episode.TitleJp
	// }
	season := episode.Season
	return &model.Bangumi{
		OfficialTitle: official_title,
		TitleRaw:      title_raw,
		Year:          episode.Year,
		Season:        season,
		SeasonRaw:     episode.SeasonRaw,
		GroupName:     episode.Group,
		DPI:           episode.Resolution,
		Source:        episode.Source,
		Subtitle:      episode.Sub,
		EpsCollect:    false,
		Offset:        0,
		IncludeFilter: strings.Join(parserConfig.Include, ","),
		ExcludeFilter: strings.Join(parserConfig.Filter, ","),
		Parser:        "raw",
		RssLink:       "",
		PosterLink:    "",
		RuleName:      "default",
		Added:         false,
		Deleted:       false,
	}
}
