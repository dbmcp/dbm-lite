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

	ServerConnOK    = "ok"
	ServerConnFail  = "fail"
	ServerConnNone  = "none"
	ServerConnUnset = ""
)

type Server struct {
	ServerID        string    `gorm:"column:server_id;primaryKey;size:64" json:"serverId"`
	ProjectID       string    `gorm:"column:project_id;size:64;index" json:"projectId"`
	BusinessID      string    `gorm:"column:business_id;size:64;index" json:"businessId"`
	Env             string    `gorm:"column:env;size:32;default:'dev'" json:"env"`
	Name            string    `gorm:"column:name;size:128;not null;index" json:"name"`
	Host            string    `gorm:"column:host;size:255;not null" json:"host"`
	Port            int       `gorm:"column:port;default:22" json:"port"`
	Username        string    `gorm:"column:username;size:128" json:"username"`
	AuthType        string    `gorm:"column:auth_type;size:32;default:'password'" json:"authType"`
	Password        string    `gorm:"column:password;size:1024" json:"-"`
	PrivateKey      string    `gorm:"column:private_key;type:text" json:"-"`
	KeyPassphrase   string    `gorm:"column:key_passphrase;size:1024" json:"-"`
	OS              string    `gorm:"column:os;size:64" json:"os"`
	Arch            string    `gorm:"column:arch;size:32" json:"arch"`
	Version         string    `gorm:"column:version;size:128" json:"version"`
	CPUCores        int       `gorm:"column:cpu_cores;default:0" json:"cpuCores"`
	MemoryGB        float64   `gorm:"column:memory_gb;default:0" json:"memoryGB"`
	DiskGB          float64   `gorm:"column:disk_gb;default:0" json:"diskGB"`
	Status          string    `gorm:"column:status;size:32;default:'active'" json:"status"`
	ConnStatus      string    `gorm:"column:conn_status;size:32" json:"connStatus"`
	ConnLatencyMs   int64     `gorm:"column:conn_latency_ms;default:0" json:"connLatencyMs"`
	LastCheckTime   time.Time `gorm:"column:last_check_time" json:"lastCheckTime"`
	Remark          string    `gorm:"column:remark;size:1024" json:"remark"`
	Tags            string    `gorm:"column:tags;size:255" json:"tags"`
	Timeout         int       `gorm:"column:timeout;default:30" json:"timeout"`
	CreatedBy       string    `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Server) TableName() string { return "servers" }
