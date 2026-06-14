/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"errors"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/database"
	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/crypto"

	"github.com/google/uuid"
)

type AuthService struct{}

func NewAuthService() *AuthService { return &AuthService{} }

func (s *AuthService) Login(username, password string) (*model.User, string, error) {
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}
	if user.Status != model.StatusActive {
		return nil, "", errors.New("账号已被禁用")
	}
	if !crypto.VerifyPassword(user.PasswordHash, password) {
		return nil, "", errors.New("用户名或密码错误")
	}

	now := time.Now()
	user.LastLoginAt = &now
	database.DB.Model(&user).Update("last_login_at", now)

	// 使用中间件中的GenerateToken
	token, err := GenerateTokenLocal(user.UserID, user.Username, user.Role, user.DisplayName)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (s *AuthService) GetUser(userId string) (*model.User, error) {
	var user model.User
	if err := database.DB.Where("user_id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) ChangePassword(userId, oldPassword, newPassword string) error {
	var user model.User
	if err := database.DB.Where("user_id = ?", userId).First(&user).Error; err != nil {
		return err
	}
	if !crypto.VerifyPassword(user.PasswordHash, oldPassword) {
		return errors.New("原密码错误")
	}
	pwdHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = pwdHash
	user.UpdatedAt = time.Now()
	return database.DB.Save(&user).Error
}

type UserService struct{}

func NewUserService() *UserService { return &UserService{} }

func (s *UserService) Create(user *model.User, rawPassword string) error {
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	pwdHash, err := crypto.HashPassword(rawPassword)
	if err != nil {
		return err
	}
	user.UserID = uuid.New().String()
	user.PasswordHash = pwdHash
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	if user.Role == "" {
		user.Role = model.RoleMember
	}
	if user.Status == "" {
		user.Status = model.StatusActive
	}
	return database.DB.Create(user).Error
}

func (s *UserService) List(page, pageSize int, keyword, role string) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	q := database.DB.Model(&model.User{})
	if keyword != "" {
		q = q.Where("username LIKE ? OR display_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if role != "" {
		q = q.Where("role = ?", role)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Omit("password_hash").Find(&list).Error
	return list, total, err
}

func (s *UserService) Update(userId string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.User{}).Where("user_id = ?", userId).Updates(updates).Error
}

func (s *UserService) Delete(userId string) error {
	return database.DB.Where("user_id = ?", userId).Delete(&model.User{}).Error
}

func (s *UserService) ResetPassword(userId, newPassword string) error {
	pwdHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.User{}).Where("user_id = ?", userId).Update("password_hash", pwdHash).Error
}

// 解密密码的辅助函数
func DecryptDatasourcePassword(encPwd string) (string, error) {
	if encPwd == "" {
		return "", nil
	}
	plain, err := crypto.DecryptAES(encPwd, config.App.AESKey)
	if err != nil {
		return "", err
	}
	return plain, nil
}

var _ = dbtype.EncryptPassword // 保持引用防止未使用警告
