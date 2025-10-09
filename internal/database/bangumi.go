package database

import (
	"goto-bangumi/internal/model"
)

// GetBangumiByID 根据 ID 获取番剧
func GetBangumiByID(id uint) (*model.Bangumi, error) {
	var bangumi model.Bangumi
	err := DB.First(&bangumi, id).Error
	if err != nil {
		return nil, err
	}
	return &bangumi, nil
}
