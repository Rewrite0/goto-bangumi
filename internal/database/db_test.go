package database

import (
	"testing"

	"goto-bangumi/internal/model"
)

func TestNewDB(t *testing.T) {
	dbPath := ":memory:"
	_, err := NewDB(&dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
}

func TestAddBangumi(t *testing.T) {
	// 什么时候会加 Bangumi
	// 1. 主要是调用 FindNewBangumi, 聚合以及日常刷新
	// 2. 其次就是非聚合的时候, 前端会点一个让我们去找新的番剧
	// 3. 通过rss_link 来连接吧, rss不一定会加进去
	testdb := "./test.db"
	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	mikanItem := model.MikanItem{
		ID:            3599,
		OfficialTitle: "夏日口袋",
		Season:        1,
		PosterLink:    "https://www.mikanani.me/attachment/202202/xxjXH8e6.jpg",
	}
	tmdbItem := model.TmdbItem{
		ID:            131631,
		Year:          "2022",
		OriginalTitle: "Summer Pocket",
		AirDate:       "2022-07-01",
		EpisodeCount:  12,
		Title:         "Summer Pocket",
		Season:        1,
		PosterLink:    "https://www.themoviedb.org/t/p/w600_and_h900_bestv2/8m8n5Yq4x0dT3cR7W4a6F2kH0kP.jpg",
		VoteAverage:   7.5,
	}
	EpisodeMetadata := model.EpisodeMetadata{
		Title:     "第1话 夏日口袋",
		Season:    1,
		SeasonRaw: "",
		Episode:   1,
	}
	bangumi := model.Bangumi{
		OfficialTitle:   "夏日口袋",
		Year:            "2022",
		Season:          1,
		MikanItem:       &mikanItem,
		TmdbItem:        &tmdbItem,
		EpisodeMetadata: []model.EpisodeMetadata{EpisodeMetadata},
		RSSLink:         "https://example.com/rss",
	}
	db.CreateBangumi(&bangumi)
}
