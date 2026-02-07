package database

import (
	"context"

	"goto-bangumi/internal/model"
)

// ============ RSS 相关方法 ============

// CreateRSS 创建 RSS 项
func (db *DB) CreateRSS(ctx context.Context, item *model.RSSItem) error {
	return db.WithContext(ctx).Save(item).Error
}

// UpdateRSS 更新 RSS 项
func (db *DB) UpdateRSS(ctx context.Context, item *model.RSSItem) error {
	return db.WithContext(ctx).Save(item).Error
}

// DeleteRSS 删除 RSS 项
func (db *DB) DeleteRSS(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&model.RSSItem{}, id).Error
}

// GetRSSByID 根据 ID 获取 RSS 项
func (db *DB) GetRSSByID(ctx context.Context, id uint) (*model.RSSItem, error) {
	var item model.RSSItem
	err := db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetRSSByURL 根据 URL 获取 RSS 项
func (db *DB) GetRSSByURL(ctx context.Context, url string) (*model.RSSItem, error) {
	var item model.RSSItem
	err := db.WithContext(ctx).Where("url = ?", url).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ListRSS 获取所有 RSS 项
func (db *DB) ListRSS(ctx context.Context) ([]*model.RSSItem, error) {
	var items []*model.RSSItem
	err := db.WithContext(ctx).Find(&items).Error
	return items, err
}

// ListActiveRSS 获取所有激活的 RSS 项
func (db *DB) ListActiveRSS(ctx context.Context) ([]*model.RSSItem, error) {
	var items []*model.RSSItem
	err := db.WithContext(ctx).Where("enabled = ?", true).Find(&items).Error
	return items, err
}

// SetRSSEnabled 设置 RSS 项的启用状态
func (db *DB) SetRSSEnabled(ctx context.Context, id uint, enabled bool) error {
	return db.WithContext(ctx).Model(&model.RSSItem{}).
		Where("id = ?", id).
		Update("enabled", enabled).Error
}
