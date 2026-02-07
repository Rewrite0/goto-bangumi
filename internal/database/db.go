// Package database 提供数据库连接和操作功能
package database

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"goto-bangumi/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 数据库连接包装
type DB struct {
	*gorm.DB
}

// 全局数据库实例（单例模式）
var globalDB *DB

// NewDB 创建数据库连接
// dsn 为 nil 时使用默认路径，传入 ":memory:" 可创建内存数据库
func NewDB(dsn *string) (*DB, error) {
	// 打开数据库连接，使用简单配置
	path := filepath.Join("./data/data.db")
	if dsn != nil {
		path = *dsn
	}
	gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	slog.Info("数据库连接成功", slog.String("path", path))
	// 自动迁移模型
	// 注意：迁移顺序很重要，基础表（无外键依赖）应该先迁移
	// 1. 首先迁移独立的基础表
	// 2. 然后迁移有外键关联的表
	// 3. GORM 会自动创建多对多关系的中间表（如 bangumi_parser_mappings）
	if err := gormDB.AutoMigrate(
		// 基础表（无外键依赖）
		&model.MikanItem{},
		&model.TmdbItem{},
		&model.EpisodeMetadata{},
		&model.RSSItem{},

		// 有外键依赖的表
		&model.Bangumi{}, // 依赖 MikanItem, TmdbItem，多对多关联 BangumiParse
		&model.Torrent{}, // 依赖 Bangumi, BangumiParse
	); err != nil {
		fmt.Println("Error migrating database:", err)
		return nil, err
	}

	return &DB{DB: gormDB}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// ============ 单例模式相关方法 ============

// InitDB 初始化全局数据库实例
// 不同的地址主要是为了测试方便,正常使用是不用的
func InitDB(dsn *string) error {
	if globalDB != nil {
		return nil // 已经初始化，直接返回
	}
	db, err := NewDB(dsn)
	if err != nil {
		return err
	}
	globalDB = db
	return nil
}

// GetDB 获取全局数据库实例
func GetDB() *DB {
	if globalDB == nil {
		InitDB(nil)
	}
	return globalDB
}

// CloseDB 关闭全局数据库连接
func CloseDB() error {
	if globalDB != nil {
		err := globalDB.Close()
		globalDB = nil
		return err
	}
	return nil
}


// ============ Torrent 相关方法 ============

// UpdateTorrent 更新种子
func (db *DB) UpdateTorrent(ctx context.Context, torrent *model.Torrent) error {
	return db.WithContext(ctx).Save(torrent).Error
}

// GetTorrentByID 根据 ID 获取种子
func (db *DB) GetTorrentByID(ctx context.Context, id uint) (*model.Torrent, error) {
	var torrent model.Torrent
	err := db.WithContext(ctx).First(&torrent, id).Error
	if err != nil {
		return nil, err
	}
	return &torrent, nil
}

// GetTorrentByURL 根据 URL 获取种子
func (db *DB) GetTorrentByURL(ctx context.Context, url string) (*model.Torrent, error) {
	var torrent model.Torrent
	err := db.WithContext(ctx).Where("url = ?", url).First(&torrent).Error
	if err != nil {
		return nil, err
	}
	return &torrent, nil
}

// GetTorrentByDownloadUID 根据下载 UID 获取种子
func (db *DB) GetTorrentByDownloadUID(ctx context.Context, duid string) (*model.Torrent, error) {
	var torrent model.Torrent
	err := db.WithContext(ctx).Where("download_uid = ?", duid).First(&torrent).Error
	if err != nil {
		return nil, err
	}
	return &torrent, nil
}

// ListTorrentByBangumi 根据番剧信息获取种子列表
func (db *DB) ListTorrentByBangumi(ctx context.Context, title string, season int, rssLink string) ([]*model.Torrent, error) {
	var torrents []*model.Torrent
	err := db.WithContext(ctx).Where("bangumi_official_title = ? AND bangumi_season = ? AND rss_link = ?",
		title, season, rssLink).Find(&torrents).Error
	return torrents, err
}

// FindUnrenamedTorrent 查询已下载但未重命名的种子
func (db *DB) FindUnrenamedTorrent(ctx context.Context) ([]*model.Torrent, error) {
	var torrents []*model.Torrent
	err := db.WithContext(ctx).Where("downloaded = ? AND renamed = ?", true, false).
		Find(&torrents).Error
	return torrents, err
}

// CheckNewTorrents 检查新种子（不存在的种子）
func (db *DB) CheckNewTorrents(ctx context.Context, torrents []*model.Torrent) ([]*model.Torrent, error) {
	var newTorrents []*model.Torrent

	for _, torrent := range torrents {
		existing, err := db.GetTorrentByURL(ctx, torrent.Link)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		// 不存在的种子
		if existing == nil {
			newTorrents = append(newTorrents, torrent)
		}
	}

	return newTorrents, nil
}

// DeleteTorrentByURL 根据 URL 删除种子
func (db *DB) DeleteTorrentByURL(ctx context.Context, url string) error {
	return db.WithContext(ctx).Where("url = ?", url).Delete(&model.Torrent{}).Error
}

// DeleteTorrentByDownloadUID 根据下载 UID 删除种子
func (db *DB) DeleteTorrentByDownloadUID(ctx context.Context, duid string) error {
	return db.WithContext(ctx).Where("download_uid = ?", duid).Delete(&model.Torrent{}).Error
}

// ============ Mikan 关联方法 ============

// CreateMikanItem 创建或更新 Mikan 项
func (db *DB) CreateMikanItem(ctx context.Context, item *model.MikanItem) error {
	return db.WithContext(ctx).Save(item).Error
}

// GetMikanItemByID 根据 MikanID 获取 Mikan 项
func (db *DB) GetMikanItemByID(ctx context.Context, mikanID int) (*model.MikanItem, error) {
	var item model.MikanItem
	err := db.WithContext(ctx).Where("mikan_id = ?", mikanID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetBangumisByMikanID 根据 MikanID 查找所有关联的 Bangumi
func (db *DB) GetBangumisByMikanID(ctx context.Context, mikanID int) ([]*model.Bangumi, error) {
	var bangumis []*model.Bangumi
	err := db.WithContext(ctx).Where("mikan_id = ?", mikanID).Find(&bangumis).Error
	return bangumis, err
}

// UpdateBangumiMikan 更新 Bangumi 的 Mikan 关联
func (db *DB) UpdateBangumiMikan(ctx context.Context, bangumiID uint, mikanID int) error {
	return db.WithContext(ctx).Model(&model.Bangumi{}).
		Where("id = ?", bangumiID).
		Update("mikan_id", mikanID).Error
}

// RemoveBangumiMikan 移除 Bangumi 的 Mikan 关联
func (db *DB) RemoveBangumiMikan(ctx context.Context, bangumiID uint) error {
	return db.WithContext(ctx).Model(&model.Bangumi{}).
		Where("id = ?", bangumiID).
		Update("mikan_id", nil).Error
}

// ============ TMDB 关联方法 ============

// CreateTmdbItem 创建或更新 TMDB 项
func (db *DB) CreateTmdbItem(ctx context.Context, item *model.TmdbItem) error {
	return db.WithContext(ctx).Save(item).Error
}

// GetTmdbItemByID 根据 TmdbID 获取 TMDB 项
func (db *DB) GetTmdbItemByID(ctx context.Context, tmdbID int) (*model.TmdbItem, error) {
	var item model.TmdbItem
	err := db.WithContext(ctx).Where("tmdb_id = ?", tmdbID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetBangumisByTmdbID 根据 TmdbID 查找所有关联的 Bangumi
func (db *DB) GetBangumisByTmdbID(ctx context.Context, tmdbID int) ([]*model.Bangumi, error) {
	var bangumis []*model.Bangumi
	err := db.WithContext(ctx).Where("tmdb_id = ?", tmdbID).Find(&bangumis).Error
	return bangumis, err
}

// UpdateBangumiTmdb 更新 Bangumi 的 TMDB 关联
func (db *DB) UpdateBangumiTmdb(ctx context.Context, bangumiID uint, tmdbID int) error {
	return db.WithContext(ctx).Model(&model.Bangumi{}).
		Where("id = ?", bangumiID).
		Update("tmdb_id", tmdbID).Error
}

// RemoveBangumiTmdb 移除 Bangumi 的 TMDB 关联
func (db *DB) RemoveBangumiTmdb(ctx context.Context, bangumiID uint) error {
	return db.WithContext(ctx).Model(&model.Bangumi{}).
		Where("id = ?", bangumiID).
		Update("tmdb_id", nil).Error
}

// ============ BangumiParse 关联方法 ============

// CreateBangumiParse 创建番剧解析器
func (db *DB) CreateBangumiParse(ctx context.Context, parser *model.EpisodeMetadata) error {
	return db.WithContext(ctx).Save(parser).Error
}

func (db *DB) GetBangumiParseByTitle(ctx context.Context, torrentName string) (*model.Bangumi, error) {
	// 要求 Title 和 Group 都在 torrentName 中出现
	// title 和 group 是 torrentName 的子串
	var metaData model.EpisodeMetadata
	err := db.WithContext(ctx).Where("instr(?, title) > 0 AND instr(?, `group`) > 0", torrentName, torrentName).First(&metaData).Error
	if err != nil {
		return nil, err
	}
	// 通过 id 获取 对应的bangumi
	var bangumi model.Bangumi
	err = db.WithContext(ctx).First(&bangumi, metaData.BangumiID).Error
	if err != nil {
		return nil, err
	}
	return &bangumi, nil
}

// GetBangumiParseByID 根据 ID 获取番剧解析器
func (db *DB) GetBangumiParseByID(ctx context.Context, id uint) (*model.EpisodeMetadata, error) {
	var parser model.EpisodeMetadata
	err := db.WithContext(ctx).First(&parser, id).Error
	if err != nil {
		return nil, err
	}
	return &parser, nil
}

// GetBangumisByParseID 根据 ParseID 查找所有关联的 Bangumi
func (db *DB) GetBangumisByParseID(ctx context.Context, parserID uint) ([]*model.Bangumi, error) {
	var bangumis []*model.Bangumi
	err := db.WithContext(ctx).Joins("JOIN bangumi_parser_mappings ON bangumi.id = bangumi_parser_mappings.bangumi_id").
		Where("bangumi_parser_mappings.bangumi_parser_id = ?", parserID).
		Find(&bangumis).Error
	return bangumis, err
}

// GetParsesByBangumiID 根据 BangumiID 查找所有关联的 Parse
func (db *DB) GetParsesByBangumiID(ctx context.Context, bangumiID uint) ([]*model.EpisodeMetadata, error) {
	var parsers []*model.EpisodeMetadata
	err := db.WithContext(ctx).Joins("JOIN bangumi_parser_mappings ON bangumi_parser.id = bangumi_parser_mappings.bangumi_parser_id").
		Where("bangumi_parser_mappings.bangumi_id = ?", bangumiID).
		Find(&parsers).Error
	return parsers, err
}

// ============ Bangumi 复合查询方法 ============

// GetBangumiWithDetails 获取 Bangumi 及其关联的 TMDB、Mikan、Parse 信息
func (db *DB) GetBangumiWithDetails(ctx context.Context, id uint) (*model.Bangumi, error) {
	var bangumi model.Bangumi
	err := db.WithContext(ctx).Preload("TmdbItem").
		Preload("MikanItem").
		Preload("EpisodeMetadata").
		First(&bangumi, id).Error
	if err != nil {
		return nil, err
	}
	return &bangumi, nil
}

// ListBangumiWithDetails 获取所有 Bangumi 及其关联信息
func (db *DB) ListBangumiWithDetails(ctx context.Context) ([]*model.Bangumi, error) {
	var bangumis []*model.Bangumi
	err := db.WithContext(ctx).Preload("TmdbItem").
		Preload("MikanItem").
		Preload("EpisodeMetadata").
		Find(&bangumis).Error
	return bangumis, err
}

// ============ Bangumi 和 BangumiParse 多对多关联方法 ============

// AddParsesToBangumi 为 Bangumi 添加多个 Parse（一对多关系）
func (db *DB) AddParsesToBangumi(ctx context.Context, bangumiID int, parsers []*model.EpisodeMetadata) error {
	var bangumi model.Bangumi
	if err := db.WithContext(ctx).First(&bangumi, bangumiID).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&bangumi).Association("EpisodeMetadata").Append(parsers)
}

// ReplaceParsesToBangumi 替换 Bangumi 的所有 Parse（一对多关系）
func (db *DB) ReplaceParsesToBangumi(ctx context.Context, bangumiID int, parsers []*model.EpisodeMetadata) error {
	var bangumi model.Bangumi
	if err := db.WithContext(ctx).First(&bangumi, bangumiID).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&bangumi).Association("EpisodeMetadata").Replace(parsers)
}

// RemoveParsesFromBangumi 从 Bangumi 中移除指定的 Parse（一对多关系）
func (db *DB) RemoveParsesFromBangumi(ctx context.Context, bangumiID int, parsers []*model.EpisodeMetadata) error {
	var bangumi model.Bangumi
	if err := db.WithContext(ctx).First(&bangumi, bangumiID).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&bangumi).Association("EpisodeMetadata").Delete(parsers)
}

// ClearParsesFromBangumi 清空 Bangumi 的所有 Parse（一对多关系）
func (db *DB) ClearParsesFromBangumi(ctx context.Context, bangumiID int) error {
	var bangumi model.Bangumi
	if err := db.WithContext(ctx).First(&bangumi, bangumiID).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&bangumi).Association("EpisodeMetadata").Clear()
}

// CountParsesOfBangumi 统计 Bangumi 关联的 Parse 数量
func (db *DB) CountParsesOfBangumi(ctx context.Context, bangumiID int) (int64, error) {
	var bangumi model.Bangumi
	if err := db.WithContext(ctx).First(&bangumi, bangumiID).Error; err != nil {
		return 0, err
	}
	return db.WithContext(ctx).Model(&bangumi).Association("EpisodeMetadata").Count(), nil
}

// AddBangumiToParse 为 Parse 添加 Bangumi（多对多关系反向操作）
func (db *DB) AddBangumiToParse(ctx context.Context, parserID int, bangumis []*model.Bangumi) error {
	var parser model.EpisodeMetadata
	if err := db.WithContext(ctx).First(&parser, parserID).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&parser).Association("Bangumis").Append(bangumis)
}

// ============ Torrent 关联查询优化方法 ============

// GetTorrentWithDetails 获取 Torrent 及其关联的 Bangumi 和 Parse 信息
func (db *DB) GetTorrentWithDetails(ctx context.Context, url string) (*model.Torrent, error) {
	var torrent model.Torrent
	err := db.WithContext(ctx).Preload("Bangumi").
		Preload("BangumiParse").
		Where("url = ?", url).
		First(&torrent).Error
	if err != nil {
		return nil, err
	}
	return &torrent, nil
}

// ListTorrentWithDetails 获取所有 Torrent 及其关联信息
func (db *DB) ListTorrentWithDetails(ctx context.Context) ([]*model.Torrent, error) {
	var torrents []*model.Torrent
	err := db.WithContext(ctx).Preload("Bangumi").
		Preload("BangumiParse").
		Find(&torrents).Error
	return torrents, err
}
