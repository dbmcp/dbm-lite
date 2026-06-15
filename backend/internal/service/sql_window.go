/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/model"

	"github.com/google/uuid"
)

type SQLWindowService struct{}

func NewSQLWindowService() *SQLWindowService { return &SQLWindowService{} }

// ListByUser 返回用户的窗口列表，按序号与创建时间排序
func (s *SQLWindowService) ListByUser(userID string) ([]model.SQLWindow, int64, error) {
	var list []model.SQLWindow
	var total int64
	db := database.DB.Model(&model.SQLWindow{}).Where("user_id = ?", userID)
	db.Count(&total)
	err := db.Order("sort_order ASC, created_at DESC").Find(&list).Error
	return list, total, err
}

// GetById 获取单个窗口
func (s *SQLWindowService) GetById(id, userID string) (*model.SQLWindow, error) {
	var w model.SQLWindow
	err := database.DB.Where("window_id = ? AND user_id = ?", id, userID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// Create 创建新窗口
func (s *SQLWindowService) Create(userID, username, title, sqlContent, datasourceID, datasourceName, databaseName string, order int) (*model.SQLWindow, error) {
	// 创建前，若新窗口标记为激活，则先把原所有窗口的激活状态清除
	if err := s.deactivateAll(userID, ""); err != nil {
		return nil, err
	}
	now := time.Now()
	if title == "" {
		title = "未命名窗口"
	}
	window := &model.SQLWindow{
		WindowID:       uuid.New().String(),
		UserID:         userID,
		Username:       username,
		Title:          title,
		SQL:            sqlContent,
		DatasourceID:   datasourceID,
		DatasourceName: datasourceName,
		DatabaseName:   databaseName,
		SortOrder:      order,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := database.DB.Create(window).Error; err != nil {
		return nil, err
	}
	return window, nil
}

// Update 更新窗口内容
func (s *SQLWindowService) Update(id, userID, title, sqlContent, datasourceID, datasourceName, databaseName string) (*model.SQLWindow, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"title":          title,
		"sql":            sqlContent,
		"datasource_id":  datasourceID,
		"datasource_name": datasourceName,
		"database_name":  databaseName,
		"updated_at":     now,
	}
	if err := database.DB.Model(&model.SQLWindow{}).Where("window_id = ? AND user_id = ?", id, userID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetById(id, userID)
}

// Delete 删除窗口
func (s *SQLWindowService) Delete(id, userID string) error {
	return database.DB.Where("window_id = ? AND user_id = ?", id, userID).Delete(&model.SQLWindow{}).Error
}

// BatchDelete 批量删除窗口
func (s *SQLWindowService) BatchDelete(ids []string, userID string) error {
	if len(ids) == 0 {
		return nil
	}
	return database.DB.Where("window_id IN ? AND user_id = ?", ids, userID).Delete(&model.SQLWindow{}).Error
}

// SetActive 设置某个窗口为激活状态
func (s *SQLWindowService) SetActive(id, userID string) error {
	if err := s.deactivateAll(userID, id); err != nil {
		return err
	}
	return database.DB.Model(&model.SQLWindow{}).
		Where("window_id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"is_active": true, "updated_at": time.Now()}).Error
}

// Recent 最近访问的窗口
func (s *SQLWindowService) Recent(userID string, limit int) ([]model.SQLWindow, error) {
	var list []model.SQLWindow
	err := database.DB.Model(&model.SQLWindow{}).Where("user_id = ?", userID).
		Order("updated_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

// deactivateAll 把用户的所有窗口（除了 exclude）置为非激活
func (s *SQLWindowService) deactivateAll(userID, exclude string) error {
	q := database.DB.Model(&model.SQLWindow{}).Where("user_id = ? AND is_active = ?", userID, true)
	if exclude != "" {
		q = q.Where("window_id != ?", exclude)
	}
	return q.UpdateColumn("is_active", false).Error
}
