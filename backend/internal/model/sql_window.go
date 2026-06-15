/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import "time"

type SQLWindow struct {
	WindowID     string     `gorm:"column:window_id;primaryKey;size:64" json:"windowId"`
	UserID       string     `gorm:"column:user_id;size:64;index" json:"userId"`
	Username     string     `gorm:"column:username;size:128" json:"username"`
	Title        string     `gorm:"column:title;size:128" json:"title"`
	SQL          string     `gorm:"column:sql;type:text" json:"sql"`
	DatasourceID string     `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	DatasourceName string  `gorm:"column:datasource_name;size:128" json:"datasourceName"`
	DatabaseName string     `gorm:"column:database_name;size:128" json:"databaseName"`
	SortOrder    int        `gorm:"column:sort_order;default:0" json:"sortOrder"`
	IsActive     bool       `gorm:"column:is_active;default:false" json:"isActive"`
	CreatedAt    time.Time  `gorm:"column:created_at;index" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (SQLWindow) TableName() string { return "sql_windows" }
