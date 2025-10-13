package parser

import (
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser/baseparser"
	// "strconv"
)

type TmdbParser struct{}

func NewTmdbParser() *TmdbParser {
	return &TmdbParser{}
}

func (p *TmdbParser) Parse(title string, language string) *model.Bangumi {
	tmdb, _ := baseparser.NewTMDBParser()
	tmdbInfo, _ := tmdb.TMDBParser(title, language)
	if tmdbInfo == nil{
		return nil
	}
	//TODO: original title 是一个更好的标准
	return &model.Bangumi{
		OfficialTitle: tmdbInfo.Title,
		TitleRaw:      title,
		Year:          tmdbInfo.Year,
		Season:        tmdbInfo.LastSeason,
		PosterLink:    tmdbInfo.PosterLink,
	}
}

func (p *TmdbParser) PosterParser(bangumi *model.Bangumi) bool {
tmdb, err := baseparser.NewTMDBParser()
	if err != nil {
		return false
	}
tmdbInfo,_:= tmdb.TMDBParser(bangumi.OfficialTitle, parserConfig.Language)
	if tmdbInfo == nil{
		return false
	}
	if tmdbInfo.PosterLink != "" {
		bangumi.PosterLink = tmdbInfo.PosterLink
		// bangumi.TmdbID = strconv.Itoa(tmdbInfo.ID)
		return true
	}
	return false
}

