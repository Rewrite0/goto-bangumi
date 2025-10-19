package database

import (
	"testing"
	"goto-bangumi/internal/model"
)

func TestNewDB(t *testing.T) {
	_, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
}

func TestAddBangumi(t *testing.T) {
	// 什么时候会加 Bangumi
	// 1. 主要是调用 FindNewBangumi, 聚合以及日常刷新
	// 2. 其次就是非聚合的时候, 前端会点一个让我们去找新的番剧
	// 3. 通过rss_link 来连接吧, rss不一定会加进去
	db, err := NewDB("./test.db")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	mikanItem := model.MikanItem{
		ID: 3599,
		OfficialTitle: "夏日口袋",
		Season: 1,
		PosterLink: "https://www.mikanani.me/attachment/202202/xxjXH8e6.jpg",
	}
	tmdbItem := model.TmdbItem{
		ID: 131631,
		Year: "2022",
		OriginalTitle: "Summer Pocket",
		AirDate: "2022-07-01",
		EpisodeCount: 12,
		Title: "Summer Pocket",
		Season: 1,
		PosterLink: "https://www.themoviedb.org/t/p/w600_and_h900_bestv2/8m8n5Yq4x0dT3cR7W4a6F2kH0kP.jpg",
		VoteAverage: 7.5,
	}
	bangumi := model.Bangumi{
		OfficialTitle: "夏日口袋",
		Year: "2022",
		Season: 1,
		MikanItem: &mikanItem,
		TmdbItem: &tmdbItem,
		RRSSLink: "https://example.com/rss",
	}
	db.CreateBangumi(&bangumi)
}


// TestAddTorrent 测试添加种子
// 刷新的时候可以把 rss id 给加进去( 不一定有 rss id, collection 的是没有的)
// rename 的时候可以把 EpisodeMetadata 更新进去?
// 没想到有什么用呀, episode 可以是多个,感觉也不好统计更新了几集
// torrent 也不一定会有 BangumiID, 因为Collection
func TestAddTorrent(t *testing.T) {
}
