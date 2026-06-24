/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"strings"
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/crypto"

	"github.com/google/uuid"
)

// ====== 账号管理 ======

type AccountService struct{}

func NewAccountService() *AccountService { return &AccountService{} }

func (s *AccountService) List(page, pageSize int, keyword, status string) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.User
	var total int64
	query := database.DB.Model(&model.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *AccountService) Get(id string) (*model.User, error) {
	var u model.User
	err := database.DB.Where("user_id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *AccountService) Create(u *model.User, rawPassword string) error {
	if u.UserID == "" {
		u.UserID = uuid.New().String()
	}
	hashed, err := crypto.HashPassword(rawPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hashed
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	if u.Status == "" {
		u.Status = model.StatusActive
	}
	if u.Role == "" {
		u.Role = model.RoleMember
	}
	return database.DB.Create(u).Error
}

func (s *AccountService) Update(id string, updates map[string]interface{}) error {
	converted := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		converted[camelToSnakeHelper(k)] = v
	}
	converted["updated_at"] = time.Now()
	return database.DB.Model(&model.User{}).Where("user_id = ?", id).Updates(converted).Error
}

func (s *AccountService) Delete(id string) error {
	return database.DB.Model(&model.User{}).Where("user_id = ?", id).Update("status", model.StatusDeleted).Error
}

func (s *AccountService) ResetPassword(id, newPassword string) error {
	hashed, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.User{}).Where("user_id = ?", id).Update("password_hash", hashed).Error
}

func (s *AccountService) ToggleLock(id string) (*model.User, error) {
	var u model.User
	if err := database.DB.Where("user_id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	newStatus := model.StatusInactive
	if u.Status != model.StatusActive {
		newStatus = model.StatusActive
	}
	if err := database.DB.Model(&u).Update("status", newStatus).Error; err != nil {
		return nil, err
	}
	u.Status = newStatus
	return &u, nil
}

// GetUserRoles 获取用户绑定的角色ID列表
func (s *AccountService) GetUserRoles(userID string) ([]string, error) {
	var binds []model.UserRoleBind
	err := database.DB.Where("user_id = ?", userID).Find(&binds).Error
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(binds))
	for _, b := range binds {
		ids = append(ids, b.RoleID)
	}
	return ids, nil
}

// AssignRoles 为用户分配角色（先清空再绑定）
func (s *AccountService) AssignRoles(userID string, roleIDs []string) error {
	tx := database.DB.Begin()
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRoleBind{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, rid := range roleIDs {
		if rid == "" {
			continue
		}
		b := &model.UserRoleBind{UserID: userID, RoleID: rid, CreatedAt: time.Now()}
		if err := tx.Create(b).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// GetUserPermissionCodes 获取用户所有角色的权限码并集
func (s *AccountService) GetUserPermissionCodes(userID string) (map[string]bool, error) {
	roles, err := s.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}
	codes := make(map[string]bool)
	
	// 优先检查用户 role 字段：admin 用户直接获得所有超级管理员权限
	var u model.User
	if ferr := database.DB.Where("user_id = ?", userID).First(&u).Error; ferr == nil {
		if u.Role == model.RoleAdmin {
			for _, c := range model.DefaultSuperAdminCodes() {
				codes[c] = true
			}
			return codes, nil
		}
	}
	
	if len(roles) == 0 {
		// 普通用户默认权限
		for _, c := range model.DefaultMemberCodes() {
			codes[c] = true
		}
		return codes, nil
	}

	var roleList []model.Role
	if err := database.DB.Where("role_id IN ?", roles).Find(&roleList).Error; err != nil {
		return nil, err
	}
	for _, r := range roleList {
		// 兼容 role.Codes JSON字段
		if r.Codes != "" {
			for _, c := range splitCodes(r.Codes) {
				codes[c] = true
			}
		}
	}
	// 额外查询 role_permission_binds
	var binds []model.RolePermissionBind
	if err := database.DB.Where("role_id IN ?", roles).Find(&binds).Error; err == nil {
		for _, b := range binds {
			codes[b.PointCode] = true
		}
	}
	return codes, nil
}

// splitCodes 解析权限码字符串，支持 JSON数组 / 逗号分隔 / 空格分隔
func splitCodes(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), "\"")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
