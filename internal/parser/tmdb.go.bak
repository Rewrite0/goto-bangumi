package parser

import (
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser/baseparser"
	// "strconv"
)

type TmdbParse struct{}

func NewTmdbParse() *TmdbParse {
	return &TmdbParse{}
}

func (p *TmdbParse) Parse(title string, language string) *model.Bangumi {
	tmdb := baseparser.NewTMDBParse()
	tmdbInfo, _ := tmdb.TMDBParse(title, language)
	if tmdbInfo == nil {
		return nil
	}
	// TODO: original title 是一个更好的标准
	return &model.Bangumi{
		OfficialTitle: tmdbInfo.Title,
		Year:          tmdbInfo.Year,
		Season:        tmdbInfo.Season,
		PosterLink:    tmdbInfo.PosterLink,
	}
}

func (p *TmdbParse) PosterParse(bangumi *model.Bangumi) bool {
	tmdb := baseparser.NewTMDBParse()
	tmdbInfo, _ := tmdb.TMDBParse(bangumi.OfficialTitle, ParserConfig.Language)
	if tmdbInfo == nil {
		return false
	}
	if tmdbInfo.PosterLink != "" {
		bangumi.PosterLink = tmdbInfo.PosterLink
		// bangumi.TmdbID = strconv.Itoa(tmdbInfo.ID)
		return true
	}
	return false
}
