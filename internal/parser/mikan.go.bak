package parser

import (
	"strings"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser/baseparser"
)

type MikanParse struct{}

func NewMikanParse() *MikanParse {
	return &MikanParse{}
}

func (p *MikanParse) Parse(homepage string) (*model.Bangumi, error) {
	// mikan 解析

	mikanParse := baseparser.NewMikanParser()
	mikanInfo, err := mikanParse.Parse(homepage)
	if err != nil {
		// 直接传递错误，上层可以用 apperrors.IsNetworkError() 或 apperrors.IsParseError() 判断
		return nil, err
	}
	if mikanInfo == nil {
		return nil, nil
	}
	return &model.Bangumi{
		OfficialTitle: mikanInfo.OfficialTitle,
		Season:        mikanInfo.Season,
		PosterLink:    mikanInfo.PosterLink,
	}, nil
}

func (p *MikanParse) PosterParse(bangumi *model.Bangumi) (bool, error) {
	// if bangumi.MikanID == "" {
	// 	return false, nil
	// }
	// FIXME: 这里后面要修复
	homepage := ParserConfig.MikanCustomURL + "/Home/Bangumi/"
	// + bangumi.MikanID
	if !strings.HasPrefix(homepage, "http") {
		homepage = "https://" + homepage
	}
	mikanParse := baseparser.NewMikanParser()
	posterLink, err := mikanParse.PosterParse(homepage)
	if err != nil {
		// 直接传递错误，上层可以判断错误类型
		return false, err
	}
	if posterLink != "" {
		bangumi.PosterLink = posterLink
		return true, nil
	}

	return false, nil
}
