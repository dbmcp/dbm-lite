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

// ====== 角色管理 ======

type RoleService struct{}

func NewRoleService() *RoleService { return &RoleService{} }

func (s *RoleService) List(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Role
	var total int64
	query := database.DB.Model(&model.Role{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *RoleService) All() ([]model.Role, error) {
	var list []model.Role
	err := database.DB.Model(&model.Role{}).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *RoleService) Get(id string) (*model.Role, error) {
	var r model.Role
	err := database.DB.Where("role_id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *RoleService) Create(r *model.Role, codes []string) error {
	if r.RoleID == "" {
		r.RoleID = uuid.New().String()
	}
	r.Codes = joinCodes(codes)
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	if r.Status == "" {
		r.Status = model.StatusActive
	}
	return database.DB.Create(r).Error
}

func (s *RoleService) Update(id string, updates map[string]interface{}) error {
	converted := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		converted[camelToSnakeHelper(k)] = v
	}
	if codes, ok := converted["codes"]; ok {
		if arr, ok2 := codes.([]string); ok2 {
			converted["codes"] = joinCodes(arr)
		}
	}
	converted["updated_at"] = time.Now()
	return database.DB.Model(&model.Role{}).Where("role_id = ?", id).Updates(converted).Error
}

func (s *RoleService) Delete(id string) error {
	tx := database.DB.Begin()
	if err := tx.Where("role_id = ?", id).Delete(&model.UserRoleBind{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("role_id = ?", id).Delete(&model.RolePermissionBind{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("role_id = ?", id).Delete(&model.Role{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// AssignPermissions 为角色分配权限点
func (s *RoleService) AssignPermissions(roleID string, codes []string) error {
	tx := database.DB.Begin()
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermissionBind{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, c := range codes {
		if c == "" {
			continue
		}
		b := &model.RolePermissionBind{RoleID: roleID, PointCode: c, CreatedAt: time.Now()}
		if err := tx.Create(b).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Model(&model.Role{}).Where("role_id = ?", roleID).Update("codes", joinCodes(codes)).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// GetRoleUsers 获取角色下的用户ID列表
func (s *RoleService) GetRoleUsers(roleID string) ([]string, error) {
	var binds []model.UserRoleBind
	err := database.DB.Where("role_id = ?", roleID).Find(&binds).Error
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(binds))
	for _, b := range binds {
		ids = append(ids, b.UserID)
	}
	return ids, nil
}

func joinCodes(codes []string) string {
	if len(codes) == 0 {
		return ""
	}
	result := ""
	for i, c := range codes {
		if i > 0 {
			result += ","
		}
		result += c
	}
	return result
}
