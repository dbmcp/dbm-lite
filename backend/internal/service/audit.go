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

type AuditService struct{}

func NewAuditService() *AuditService { return &AuditService{} }

func (s *AuditService) Log(userId, username, ip, module, action, targetId, detail, result, remark string) error {
	remarkPart := ""
	if remark != "" {
		remarkPart = " " + remark
	}
	log := &model.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    userId,
		Username:  username,
		Action:    action,
		Module:    module,
		TargetID:  targetId,
		Detail:    detail + remarkPart,
		IPAddress: ip,
		Status:    result,
		CreatedAt: time.Now(),
	}
	return database.DB.Create(log).Error
}

func (s *AuditService) QueryList(page, pageSize int, keyword, userId, module, action, startTime, endTime string) ([]model.AuditLog, int64, error) {
	var list []model.AuditLog
	var total int64

	q := database.DB.Model(&model.AuditLog{})
	if keyword != "" {
		q = q.Where("username LIKE ? OR detail LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if userId != "" {
		q = q.Where("user_id = ?", userId)
	}
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if startTime != "" {
		q = q.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		q = q.Where("created_at <= ?", endTime)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *AuditService) Stats(days int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var total int64
	database.DB.Model(&model.AuditLog{}).Count(&total)
	stats["total"] = total

	var todayCount int64
	start := time.Now().Truncate(24 * time.Hour)
	database.DB.Model(&model.AuditLog{}).Where("created_at >= ?", start).Count(&todayCount)
	stats["todayCount"] = todayCount

	var failCount int64
	database.DB.Model(&model.AuditLog{}).Where("status = ?", model.AuditResultFailed).Count(&failCount)
	stats["failCount"] = failCount

	type ModuleStat struct {
		Module string `json:"module"`
		Count  int64  `json:"count"`
	}
	var moduleStats []ModuleStat
	database.DB.Model(&model.AuditLog{}).Select("module, COUNT(*) as count").Group("module").Order("count DESC").Limit(10).Scan(&moduleStats)
	stats["moduleStats"] = moduleStats

	return stats, nil
}
