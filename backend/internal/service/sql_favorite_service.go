/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"dbm-lite/internal/database"
	"dbm-lite/internal/model"

	"github.com/google/uuid"
)

type SQLFavoriteService struct{}

func NewSQLFavoriteService() *SQLFavoriteService { return &SQLFavoriteService{} }

func (s *SQLFavoriteService) List(userId, keyword string, page, pageSize int) ([]model.SQLFavorite, int64, error) {
	var list []model.SQLFavorite
	q := database.DB.Model(&model.SQLFavorite{}).Where("user_id = ?", userId)
	
	if keyword != "" {
		q = q.Where("title LIKE ?", "%"+keyword+"%")
	}
	
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	
	return list, total, nil
}

func (s *SQLFavoriteService) Get(userId, favoriteId string) (*model.SQLFavorite, error) {
	var fav model.SQLFavorite
	err := database.DB.Where("user_id = ? AND favorite_id = ?", userId, favoriteId).First(&fav).Error
	if err != nil {
		return nil, err
	}
	return &fav, nil
}

func (s *SQLFavoriteService) Create(userId, username, title, description, sql string) error {
	fav := &model.SQLFavorite{
		FavoriteID:  uuid.New().String(),
		UserID:      userId,
		Username:    username,
		Title:       title,
		Description: description,
		SqlText:     sql,
	}
	return database.DB.Create(fav).Error
}

func (s *SQLFavoriteService) Update(userId, favoriteId, title, description, sql string) error {
	return database.DB.Model(&model.SQLFavorite{}).
		Where("user_id = ? AND favorite_id = ?", userId, favoriteId).
		Updates(map[string]interface{}{
			"title":       title,
			"description": description,
			"sql":         sql,
		}).Error
}

func (s *SQLFavoriteService) Delete(userId, favoriteId string) error {
	return database.DB.Where("user_id = ? AND favorite_id = ?", userId, favoriteId).Delete(&model.SQLFavorite{}).Error
}