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

type Business struct {
	BusinessID  string    `gorm:"column:business_id;primaryKey;size:64" json:"businessId"`
	ProjectID   string    `gorm:"column:project_id;size:64;index" json:"projectId"`
	Code        string    `gorm:"column:code;size:128;index" json:"code"`
	Name        string    `gorm:"column:name;size:128;not null" json:"name"`
	Description string    `gorm:"column:description;size:512" json:"description"`
	Env         string    `gorm:"column:env;size:32;default:'dev'" json:"env"`
	CreatedBy   string    `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Business) TableName() string { return "businesses" }

type BusinessMember struct {
	BusinessID string    `gorm:"column:business_id;size:64;index:idx_bu_user,unique" json:"businessId"`
	UserID     string    `gorm:"column:user_id;size:64;index:idx_bu_user,unique" json:"userId"`
	Role       string    `gorm:"column:role;size:32;default:'developer'" json:"role"`
	JoinedAt   time.Time `gorm:"column:joined_at" json:"joinedAt"`
}

func (BusinessMember) TableName() string { return "business_members" }

