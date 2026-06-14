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

type BackupPolicy struct {
	PolicyID     string    `gorm:"column:policy_id;primaryKey;size:64" json:"policyId"`
	Name         string    `gorm:"column:name;size:128;not null" json:"name"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	BackupType   string    `gorm:"column:backup_type;size:32" json:"backupType"` // full | incremental
	Strategy     string    `gorm:"column:strategy;size:32" json:"strategy"`       // manual | cron
	CronExpr     string    `gorm:"column:cron_expr;size:128" json:"cronExpr"`
	KeepCount    int       `gorm:"column:keep_count;default:7" json:"keepCount"`
	StoragePath  string    `gorm:"column:storage_path;size:512" json:"storagePath"`
	Status       string    `gorm:"column:status;size:32;default:'enabled'" json:"status"`
	CreatedBy    string    `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (BackupPolicy) TableName() string { return "backup_policies" }

type BackupRecord struct {
	RecordID   string    `gorm:"column:record_id;primaryKey;size:64" json:"recordId"`
	PolicyID   string    `gorm:"column:policy_id;size:64;index" json:"policyId"`
	BackupType string    `gorm:"column:backup_type;size:32" json:"backupType"`
	FileName   string    `gorm:"column:file_name;size:256" json:"fileName"`
	SizeMB     float64   `gorm:"column:size_mb" json:"sizeMb"`
	DurationSec int      `gorm:"column:duration_sec" json:"durationSec"`
	Status     string    `gorm:"column:status;size:32" json:"status"`
	Remark     string    `gorm:"column:remark;size:512" json:"remark"`
	StartedAt  time.Time `gorm:"column:started_at" json:"startedAt"`
	FinishedAt *time.Time `gorm:"column:finished_at" json:"finishedAt,omitempty"`
}

func (BackupRecord) TableName() string { return "backup_records" }

const (
	BackupTypeFull        = "full"
	BackupTypeIncremental = "incremental"
	StrategyManual        = "manual"
	StrategyCron          = "cron"
	StatusEnabled         = "enabled"
	StatusDisabled        = "disabled"
	StatusSuccess         = "success"
	StatusFailed          = "failed"
	StatusRunning         = "running"
)

type InspectTask struct {
	TaskID       string    `gorm:"column:task_id;primaryKey;size:64" json:"taskId"`
	Name         string    `gorm:"column:name;size:128;not null" json:"name"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	Strategy     string    `gorm:"column:strategy;size:32" json:"strategy"`
	CronExpr     string    `gorm:"column:cron_expr;size:128" json:"cronExpr"`
	Status       string    `gorm:"column:status;size:32" json:"status"`
	CreatedBy    string    `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (InspectTask) TableName() string { return "inspect_tasks" }

type InspectReport struct {
	ReportID     string    `gorm:"column:report_id;primaryKey;size:64" json:"reportId"`
	TaskID       string    `gorm:"column:task_id;size:64;index" json:"taskId"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	CPUUsage     float64   `gorm:"column:cpu_usage" json:"cpuUsage"`
	MemUsage     float64   `gorm:"column:mem_usage" json:"memUsage"`
	DiskUsage    float64   `gorm:"column:disk_usage" json:"diskUsage"`
	Connections  int       `gorm:"column:connections" json:"connections"`
	SlowQueries  int       `gorm:"column:slow_queries" json:"slowQueries"`
	ReplDelay    int       `gorm:"column:repl_delay" json:"replDelay"`
	Score        int       `gorm:"column:score" json:"score"`
	Detail       string    `gorm:"column:detail;type:text" json:"detail"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (InspectReport) TableName() string { return "inspect_reports" }

type SlowLog struct {
	LogID        string    `gorm:"column:log_id;primaryKey;size:64" json:"logId"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	SQL          string    `gorm:"column:sql;type:text" json:"sql"`
	QueryTime    float64   `gorm:"column:query_time" json:"queryTime"`
	RowsExamined int64     `gorm:"column:rows_examined" json:"rowsExamined"`
	RowsSent     int64     `gorm:"column:rows_sent" json:"rowsSent"`
	LockTime     float64   `gorm:"column:lock_time" json:"lockTime"`
	DatabaseName string    `gorm:"column:database_name;size:128" json:"databaseName"`
	UserHost     string    `gorm:"column:user_host;size:256" json:"userHost"`
	CreatedAt    time.Time `gorm:"column:created_at;index" json:"createdAt"`
	Digest       string    `gorm:"column:digest;size:256;index" json:"digest"`
}

func (SlowLog) TableName() string { return "slow_logs" }

type HaCluster struct {
	ClusterID  string    `gorm:"column:cluster_id;primaryKey;size:64" json:"clusterId"`
	Name       string    `gorm:"column:name;size:128;not null" json:"name"`
	ClusterType string   `gorm:"column:cluster_type;size:32" json:"clusterType"` // mgr | replication
	BusinessID string    `gorm:"column:business_id;size:64;index" json:"businessId"`
	Status     string    `gorm:"column:status;size:32" json:"status"`
	PrimaryID  string    `gorm:"column:primary_id;size:64" json:"primaryId"`
	Detail     string    `gorm:"column:detail;type:text" json:"detail"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (HaCluster) TableName() string { return "ha_clusters" }

const (
	HaTypeMgr         = "mgr"
	HaTypeReplication = "replication"
)

type HaNode struct {
	NodeID       string `gorm:"column:node_id;primaryKey;size:64" json:"nodeId"`
	ClusterID    string `gorm:"column:cluster_id;size:64;index" json:"clusterId"`
	DatasourceID string `gorm:"column:datasource_id;size:64" json:"datasourceId"`
	Role         string `gorm:"column:role;size:32" json:"role"` // primary | secondary
	Status       string `gorm:"column:status;size:32" json:"status"`
	ReplDelay    int    `gorm:"column:repl_delay" json:"replDelay"`
}

func (HaNode) TableName() string { return "ha_nodes" }

type Plugin struct {
	PluginID    string    `gorm:"column:plugin_id;primaryKey;size:64" json:"pluginId"`
	Name        string    `gorm:"column:name;size:128;not null;uniqueIndex" json:"name"`
	Version     string    `gorm:"column:version;size:32" json:"version"`
	Description string    `gorm:"column:description;size:512" json:"description"`
	Config      string    `gorm:"column:config;type:text" json:"config"`
	DownloadURL string    `gorm:"column:download_url;size:512" json:"downloadUrl"`
	Params      string    `gorm:"column:params;type:text" json:"params"`
	Status      string    `gorm:"column:status;size:32;default:'active'" json:"status"`
	CreatedBy   string    `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Plugin) TableName() string { return "plugins" }

type CapacityStat struct {
	StatID       string    `gorm:"column:stat_id;primaryKey;size:64" json:"statId"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	DatabaseName string    `gorm:"column:database_name;size:128" json:"databaseName"`
	TableItem    string    `gorm:"column:table_name;size:128" json:"tableName"`
	SizeMB       float64   `gorm:"column:size_mb" json:"sizeMb"`
	RowCount     int64     `gorm:"column:row_count" json:"rowCount"`
	IndexSizeMB  float64   `gorm:"column:index_size_mb" json:"indexSizeMb"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (CapacityStat) TableName() string { return "capacity_stats" }

type DBUser struct {
	UserID       string    `gorm:"column:user_id;primaryKey;size:64" json:"userId"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	UserName     string    `gorm:"column:user_name;size:128" json:"userName"`
	Host         string    `gorm:"column:host;size:128" json:"host"`
	DatabaseName string    `gorm:"column:database_name;size:128" json:"databaseName"`
	Privileges   string    `gorm:"column:privileges;size:512" json:"privileges"`
	Status       string    `gorm:"column:status;size:32" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (DBUser) TableName() string { return "db_users" }

type SystemSetting struct {
	SettingKey string `gorm:"column:setting_key;primaryKey;size:128" json:"settingKey"`
	Value      string `gorm:"column:value;type:text" json:"value"`
	Remark     string `gorm:"column:remark;size:512" json:"remark"`
}

func (SystemSetting) TableName() string { return "system_settings" }

