/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import (
	"time"
)

const (
	RoleAdmin  = "admin"
	RoleMember = "member"

	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusDeleted  = "deleted"
)

type User struct {
	UserID       string    `gorm:"column:user_id;primaryKey;size:64" json:"userId"`
	Username     string    `gorm:"column:username;size:64;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255" json:"-"`
	Email        string    `gorm:"column:email;size:128" json:"email"`
	DisplayName  string    `gorm:"column:display_name;size:64" json:"displayName"`
	Role         string    `gorm:"column:role;size:32;default:'member'" json:"role"`
	Status       string    `gorm:"column:status;size:32;default:'active'" json:"status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (User) TableName() string { return "users" }

