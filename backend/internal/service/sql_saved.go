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
	"errors"
	"fmt"
	"strings"
	"time"
)

type SavedQueryService struct{}

func NewSavedQueryService() *SavedQueryService {
	return &SavedQueryService{}
}

func (s *SavedQueryService) List(datasourceID string, page, pageSize int, keyword string) ([]*model.SavedQuery, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 100
	}
	db := database.DB
	if datasourceID != "" {
		db = db.Where("datasource_id = ?", datasourceID)
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		db = db.Where("(title LIKE ? OR description LIKE ? OR sql LIKE ?)",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := db.Model(&model.SavedQuery{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count failed: %w", err)
	}
	var list []*model.SavedQuery
	err := db.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list failed: %w", err)
	}
	return list, total, nil
}

func (s *SavedQueryService) Get(queryID string) (*model.SavedQuery, error) {
	if queryID == "" {
		return nil, errors.New("queryID 不能为空")
	}
	var q model.SavedQuery
	if err := database.DB.Where("query_id = ?", queryID).First(&q).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

func (s *SavedQueryService) Save(userID, username, datasourceID, databaseName, title, description, sqlText string) (*model.SavedQuery, error) {
	if title = strings.TrimSpace(title); title == "" {
		title = "未命名查询"
	}
	now := time.Now()
	q := &model.SavedQuery{
		QueryID:      "sq_" + now.Format("20060102150405") + fmt.Sprintf("_%03d", time.Now().Nanosecond()/1e6),
		UserID:       userID,
		Username:     username,
		DatasourceID: datasourceID,
		DatabaseName: databaseName,
		Title:        title,
		Description: description,
		SQL:         sqlText,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := database.DB.Create(q).Error; err != nil {
		return nil, fmt.Errorf("create failed: %w", err)
	}
	return q, nil
}

func (s *SavedQueryService) Update(queryID, title, description, sqlText, databaseName string) error {
	if queryID == "" {
		return errors.New("queryID 不能为空")
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if title != "" {
		updates["title"] = title
	}
	if description != "" {
		updates["description"] = description
	}
	if sqlText != "" {
		updates["sql"] = sqlText
	}
	if databaseName != "" {
		updates["database_name"] = databaseName
	}
	return database.DB.Model(&model.SavedQuery{}).Where("query_id = ?", queryID).Updates(updates).Error
}

func (s *SavedQueryService) Delete(queryID string) error {
	if queryID == "" {
		return errors.New("queryID 不能为空")
	}
	return database.DB.Where("query_id = ?", queryID).Delete(&model.SavedQuery{}).Error
}
