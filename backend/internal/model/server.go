/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import (
	"time"
)

const (
	ServerAuthPassword = "password"
	ServerAuthKey      = "key"

	ServerStatusActive   = "active"
	ServerStatusInactive = "inactive"
	ServerStatusDeleted  = "deleted"
)

type Server struct {
	ServerID   string    `gorm:"column:server_id;primaryKey;size:64" json:"serverId"`
	ProjectID  string    `gorm:"column:project_id;size:64;index" json:"projectId"`
	BusinessID string    `gorm:"column:business_id;size:64;index" json:"businessId"`
	Env        string    `gorm:"column:env;size:32;default:'dev'" json:"env"`
	Name       string    `gorm:"column:name;size:128;not null" json:"name"`
	Host       string    `gorm:"column:host;size:255;not null" json:"host"`
	Port       int       `gorm:"column:port;default:22" json:"port"`
	Username   string    `gorm:"column:username;size:128" json:"username"`
	AuthType   string    `gorm:"column:auth_type;size:32;default:'password'" json:"authType"`
	Password   string    `gorm:"column:password;size:512" json:"-"`
	PrivateKey string    `gorm:"column:private_key;type:text" json:"-"`
	OS         string    `gorm:"column:os;size:64" json:"os"`
	Status     string    `gorm:"column:status;size:32;default:'active'" json:"status"`
	Remark     string    `gorm:"column:remark;size:512" json:"remark"`
	CreatedBy  string    `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Server) TableName() string { return "servers" }
