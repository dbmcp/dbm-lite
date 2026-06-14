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

// SQL执行历史记录

const (
	SQLStatusSuccess = "success"
	SQLStatusFailed  = "failed"
	SQLStatusRunning = "running"
	SQLStatusStopped = "stopped"
)

type SQLHistory struct {
	HistoryID    string    `gorm:"column:history_id;primaryKey;size:64" json:"historyId"`
	UserID       string    `gorm:"column:user_id;size:64;index" json:"userId"`
	Username     string    `gorm:"column:username;size:128" json:"username"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	DatabaseName string    `gorm:"column:database_name;size:128" json:"databaseName"`
	SQL          string    `gorm:"column:sql;type:text" json:"sql"`
	RowsAffected int64     `gorm:"column:rows_affected" json:"rowsAffected"`
	DurationMs   int64     `gorm:"column:duration_ms" json:"durationMs"`
	IsHighRisk   bool      `gorm:"column:is_high_risk" json:"isHighRisk"`
	Status       string    `gorm:"column:status;size:32" json:"status"`
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"errorMessage,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at;index" json:"createdAt"`
}

func (SQLHistory) TableName() string { return "sql_history" }

