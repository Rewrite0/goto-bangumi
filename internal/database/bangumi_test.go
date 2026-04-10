package database

import (
	"context"
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

func TestBangumiLifecycle(t *testing.T) {
	testdb := ":memory:"
	// testdb := "./test.db"

	db, err := NewDB(&testdb)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	mikanID := 3599
	tmdbID := 131631
	mikanItem := model.MikanItem{
		ID:            mikanID,
		OfficialTitle: "夏日口袋",
		Season:        1,
		PosterLink:    "https://www.mikanani.me/attachment/202202/xxjXH8e6.jpg",
	}
	tmdbItem := model.TmdbItem{
		ID:            tmdbID,
		Year:          "2022",
		OriginalTitle: "Summer Pocket",
		AirDate:       "2022-07-01",
		EpisodeCount:  12,
		Title:         "Summer Pocket",
		Season:        1,
		PosterLink:    "https://www.themoviedb.org/t/p/w600_and_h900_bestv2/8m8n5Yq4x0dT3cR7W4a6F2kH0kP.jpg",
		VoteAverage:   7.5,
	}
	episodeMetadata := model.EpisodeMetadata{
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
		EpisodeMetadata: []model.EpisodeMetadata{episodeMetadata},
		RSSLink:         "https://example.com/rss",
	}

	t.Run("Create", func(t *testing.T) {
		if err := db.CreateBangumi(&bangumi); err != nil {
			t.Fatalf("CreateBangumi failed: %v", err)
		}
		if bangumi.ID == 0 {
			t.Fatal("Expected bangumi.ID to be set after creation")
		}
		// 验证 MikanItem 和 TmdbItem 被正确关联
		if bangumi.MikanID == nil || *bangumi.MikanID != mikanID {
			t.Fatalf("Expected MikanID=%d, got %v", mikanID, bangumi.MikanID)
		}
		if bangumi.TmdbID == nil || *bangumi.TmdbID != tmdbID {
			t.Fatalf("Expected TmdbID=%d, got %v", tmdbID, bangumi.TmdbID)
		}
	})


	t.Run("CreateDuplicate_ByMikanID", func(t *testing.T) {
		// 用完全相同的数据再次创建，应该幂等（更新而非重复插入）
		dup := model.Bangumi{
			OfficialTitle:   bangumi.OfficialTitle,
			Year:            bangumi.Year,
			Season:          bangumi.Season,
			MikanItem:       &mikanItem,
			TmdbItem:        &tmdbItem,
			EpisodeMetadata: []model.EpisodeMetadata{episodeMetadata},
			RSSLink:         bangumi.RSSLink,
		}
		if err := db.CreateBangumi(&dup); err != nil {
			t.Fatalf("CreateBangumi duplicate should not error, got: %v", err)
		}
		// 总数仍然只有一条
		var count int64
		db.Model(&model.Bangumi{}).Count(&count)
		if count != 1 {
			t.Fatalf("Expected 1 bangumi after duplicate insert, got %d", count)
		}
		// EpisodeMetadata 不应被清空，仍然是 1 条
		var emCount int64
		db.Model(&model.EpisodeMetadata{}).Where("bangumi_id = ?", bangumi.ID).Count(&emCount)
		if emCount == 0 {
			t.Fatal("EpisodeMetadata should not be cleared after duplicate insert")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		got, err := db.GetBangumiByID(bangumi.ID)
		if err != nil {
			t.Fatalf("GetBangumiByID failed: %v", err)
		}
		if got.OfficialTitle != "夏日口袋" {
			t.Fatalf("Expected OfficialTitle '夏日口袋', got %q", got.OfficialTitle)
		}
	})

	t.Run("GetByOfficialTitle", func(t *testing.T) {
		got, err := db.GetBangumiByOfficialTitle("夏日口袋")
		if err != nil {
			t.Fatalf("GetBangumiByOfficialTitle failed: %v", err)
		}
		if got.ID != bangumi.ID {
			t.Fatalf("Expected ID=%d, got %d", bangumi.ID, got.ID)
		}
	})

	t.Run("GetWithDetails", func(t *testing.T) {
		ctx := context.Background()
		got, err := db.GetBangumiWithDetails(ctx, uint(bangumi.ID))
		if err != nil {
			t.Fatalf("GetBangumiWithDetails failed: %v", err)
		}
		if got.MikanItem == nil {
			t.Fatal("Expected MikanItem to be preloaded")
		}
		if got.MikanItem.ID != mikanID {
			t.Fatalf("Expected MikanItem.ID=%d, got %d", mikanID, got.MikanItem.ID)
		}
		if got.TmdbItem == nil {
			t.Fatal("Expected TmdbItem to be preloaded")
		}
		if got.TmdbItem.ID != tmdbID {
			t.Fatalf("Expected TmdbItem.ID=%d, got %d", tmdbID, got.TmdbItem.ID)
		}
		if len(got.EpisodeMetadata) != 1 {
			t.Fatalf("Expected 1 EpisodeMetadata, got %d", len(got.EpisodeMetadata))
		}
		em := got.EpisodeMetadata[0]
		if em.Title != episodeMetadata.Title {
			t.Fatalf("Expected EpisodeMetadata.Title=%q, got %q", episodeMetadata.Title, em.Title)
		}
		if em.Season != episodeMetadata.Season {
			t.Fatalf("Expected EpisodeMetadata.Season=%d, got %d", episodeMetadata.Season, em.Season)
		}
		if em.BangumiID != bangumi.ID {
			t.Fatalf("Expected EpisodeMetadata.BangumiID=%d, got %d", bangumi.ID, em.BangumiID)
		}
	})

	t.Run("List", func(t *testing.T) {
		bangumis, err := db.ListBangumi()
		if err != nil {
			t.Fatalf("ListBangumi failed: %v", err)
		}
		if len(bangumis) != 1 {
			t.Fatalf("Expected 1 bangumi, got %d", len(bangumis))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := db.DeleteBangumi(bangumi.ID); err != nil {
			t.Fatalf("DeleteBangumi failed: %v", err)
		}
		var count int64
		db.Model(&model.Bangumi{}).Count(&count)
		if count != 0 {
			t.Fatalf("Expected 0 bangumis after delete, got %d", count)
		}
	})
}
