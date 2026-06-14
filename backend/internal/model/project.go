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

// 环境常量
const (
	EnvDev      = "dev"
	EnvTest     = "test"
	EnvPreprod  = "preprod"
	EnvProd     = "prod"
	EnvDisaster = "disaster"
)

// 项目角色常量
const (
	ProjectRoleOwner     = "owner"
	ProjectRoleDeveloper = "developer"
	ProjectRoleViewer    = "viewer"
)

type Project struct {
	ProjectID   string    `gorm:"column:project_id;primaryKey;size:64" json:"projectId"`
	Name        string    `gorm:"column:name;size:128;not null;index" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	CreatedBy   string    `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Project) TableName() string { return "projects" }

type ProjectMember struct {
	ProjectID string    `gorm:"column:project_id;primaryKey;size:64" json:"projectId"`
	UserID    string    `gorm:"column:user_id;primaryKey;size:64" json:"userId"`
	Role      string    `gorm:"column:role;size:32;default:'developer'" json:"role"`
	JoinedAt  time.Time `gorm:"column:joined_at" json:"joinedAt"`
}

func (ProjectMember) TableName() string { return "project_members" }

