package database

import (
	"log/slog"
	"sync"

	"goto-bangumi/internal/model"

	"gorm.io/gorm"
)

// ============ Bangumi 相关方法 ============

// 用于防止并发创建相同 Bangumi 的互斥锁
var bangumiCreateMutex sync.Mutex

// CreateBangumi 创建番剧
func (db *DB) CreateBangumi(bangumi *model.Bangumi) error {
	// 加锁防止并发创建重复的 Bangumi
	bangumiCreateMutex.Lock()
	defer bangumiCreateMutex.Unlock()

	// 对于 Bangumi 要进行一个查重, 主要是看其对应的 mikanid 和 tmdbid
	// 先是看 mikanid 有的话
	var oldBangumi model.Bangumi
	var tmdbID int
	if bangumi.TmdbID != nil {
		tmdbID = *bangumi.TmdbID
	} else if bangumi.TmdbItem != nil {
		tmdbID = bangumi.TmdbItem.ID
	}
	var mikanID int
	if bangumi.MikanID != nil {
		mikanID = *bangumi.MikanID
	} else if bangumi.MikanItem != nil {
		mikanID = bangumi.MikanItem.ID
	}
	// 通过 mikanID 和 tmdbID 来查找 Bangumi
	// err := db.Where("mikan_id = ? AND tmdb_id = ?", mikanID, tmdbID).First(&oldBangumi).Error
	err := db.Preload("MikanItem").
		Preload("TmdbItem").
		Preload("EpisodeMetadata").
		Where("mikan_id = ?", mikanID).
		Or("tmdb_id = ?", tmdbID).First(&oldBangumi).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if oldBangumi.ID != 0 {
		// 找到的话就更新一下 mikan, tmdb
		slog.Debug("[database] 番剧已存在，进行更新", "标题", oldBangumi.OfficialTitle)
		if oldBangumi.MikanID == nil && bangumi.MikanItem != nil {
			oldBangumi.MikanItem = bangumi.MikanItem
		}
		if oldBangumi.TmdbID == nil && bangumi.TmdbItem != nil {
			oldBangumi.TmdbItem = bangumi.TmdbItem
		}
		// FIXME: 这里会重复添加, 要改改
		oldBangumi.EpisodeMetadata = bangumi.EpisodeMetadata
		return db.Save(&oldBangumi).Error
	}
	return db.Save(bangumi).Error
}

// UpdateBangumi 更新番剧
func (db *DB) UpdateBangumi(bangumi *model.Bangumi) error {
	return db.Save(bangumi).Error
}

// DeleteBangumi 删除番剧
func (db *DB) DeleteBangumi(id int) error {
	return db.Delete(&model.Bangumi{}, id).Error
}

// GetBangumiByID 根据 ID 获取番剧
func (db *DB) GetBangumiByID(id int) (*model.Bangumi, error) {
	var bangumi model.Bangumi
	err := db.First(&bangumi, id).Error
	if err != nil {
		return nil, err
	}
	return &bangumi, nil
}

func (db *DB) GetBangumiByOfficialTitle(title string) (*model.Bangumi, error) {
	var bangumi model.Bangumi
	err := db.Where("official_title = ?", title).First(&bangumi).Error
	if err != nil {
		return nil, err
	}
	return &bangumi, nil
}

// ListBangumi 获取所有番剧
func (db *DB) ListBangumi() ([]*model.Bangumi, error) {
	var bangumis []*model.Bangumi
	err := db.Find(&bangumis).Error
	return bangumis, err
}
