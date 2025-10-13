package parser

import (
	"strings"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser/baseparser"
)

type MikanParser struct{}

func NewMikanParser() *MikanParser {
	return &MikanParser{}
}

func (p *MikanParser) Parse(homepage string) *model.Bangumi {
	// mikan 解析

	mikanParser := baseparser.NewMikanParser()
	mikanInfo, err := mikanParser.Parse(homepage)
	if err != nil {
		// 解析失败
		return nil
	}
	if mikanInfo == nil || mikanInfo.PosterLink == "" || mikanInfo.ID == "" {
		return nil
	}
	return &model.Bangumi{
		OfficialTitle: mikanInfo.OfficialTitle,
		Season:        mikanInfo.Season,
		PosterLink:    mikanInfo.PosterLink,
	}
}

func (p *MikanParser) PosterParser(bangumi *model.Bangumi) bool {
	// if bangumi.MikanID == "" {
	// 	return false
	// }
	// FIXME: 这里后面要修复
	homepage := parserConfig.MikanCustomURL + "/Home/Bangumi/" 
	// + bangumi.MikanID
	if !strings.HasPrefix(homepage, "http") {
		homepage = "https://" + homepage
	}
	mikanParser := baseparser.NewMikanParser()
	posterLink, err := mikanParser.PosterParser(homepage)
	if err != nil {
		// 解析失败
		return false
	}
	if posterLink != "" {
		bangumi.PosterLink = posterLink
		return true
	}

	return false
}
