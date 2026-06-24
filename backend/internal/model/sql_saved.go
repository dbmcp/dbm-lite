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

// SavedQuery SQL IDE 的保存查询（对应 Navicat 的"查询"节点）
type SavedQuery struct {
	QueryID      string    `gorm:"column:query_id;primaryKey;size:64" json:"queryId"`
	UserID       string    `gorm:"column:user_id;size:64;index" json:"userId"`
	Username     string    `gorm:"column:username;size:128" json:"username"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	DatabaseName string    `gorm:"column:database_name;size:128" json:"databaseName"`
	Title        string    `gorm:"column:title;size:256" json:"title"`
	Description  string    `gorm:"column:description;type:text" json:"description"`
	SQL          string    `gorm:"column:sql;type:text" json:"sql"`
	CreatedAt    time.Time `gorm:"column:created_at;index" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (SavedQuery) TableName() string { return "sql_saved_query" }
