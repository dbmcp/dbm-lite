/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import "time"

// Department 部门表，支持多级树形组织结构（整合 ferry 的 dept 能力）
type Department struct {
	DeptID    string     `gorm:"column:dept_id;primaryKey;size:64" json:"deptId"`
	Name      string     `gorm:"column:name;size:128;not null" json:"name"`
	ParentID  string     `gorm:"column:parent_id;size:64;index" json:"parentId"`
	Leader    string     `gorm:"column:leader;size:64" json:"leader"`
	Phone     string     `gorm:"column:phone;size:32" json:"phone"`
	Email     string     `gorm:"column:email;size:128" json:"email"`
	SortOrder int        `gorm:"column:sort_order;default:0" json:"sortOrder"`
	Status    string     `gorm:"column:status;size:32;default:'active'" json:"status"`
	Remark    string     `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (Department) TableName() string { return "departments" }

// SystemConfig 系统配置（键值对，整合 ferry 的 sys_dict / sys_config）
type SystemConfig struct {
	ConfigID  string    `gorm:"column:config_id;primaryKey;size:64" json:"configId"`
	ConfigKey string    `gorm:"column:config_key;size:128;uniqueIndex;not null" json:"configKey"`
	ConfigVal string    `gorm:"column:config_val;type:text" json:"configVal"`
	Category  string    `gorm:"column:category;size:64;index" json:"category"`
	Label     string    `gorm:"column:label;size:128" json:"label"`
	SortOrder int       `gorm:"column:sort_order;default:0" json:"sortOrder"`
	ValueType string    `gorm:"column:value_type;size:16;default:'string'" json:"valueType"` // string / bool / int / json
	Remark    string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (SystemConfig) TableName() string { return "system_configs" }
