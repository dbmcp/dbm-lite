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

// 统一审计日志模型

const (
	AuditResultSuccess = "success"
	AuditResultFailed  = "failed"
)

// 模块常量
const (
	ModuleAuth       = "auth"
	ModuleUser       = "user"
	ModuleAccount    = "account"
	ModuleBusiness   = "business"
	ModuleServer     = "server"
	ModuleDatasource = "datasource"
	ModuleSQL        = "sql"
	ModuleBackup     = "backup"
	ModuleInspect    = "inspect"
	ModuleSlowLog    = "slowlog"
	ModuleHa         = "ha"
	ModulePlugin     = "plugin"
	ModuleProject    = "project"
	ModuleOps        = "ops"
	ModulePlatform   = "platform"
	ModuleImportExport = "import_export"
	ModuleSQLAudit     = "sql_audit"
	ModuleSensitiveData = "sensitive_data"
	ModuleMigration     = "migration"
)

// 动作类型常量
const (
	ActionLogin          = "auth.login"
	ActionLogout         = "auth.logout"
	ActionChangePassword = "auth.changePassword"

	ActionAccountCreate   = "account.create"
	ActionAccountUpdate   = "account.update"
	ActionAccountDelete   = "account.delete"
	ActionAccountResetPwd = "account.resetPassword"

	ActionDsCreate   = "datasource.create"
	ActionDsUpdate   = "datasource.update"
	ActionDsDelete   = "datasource.delete"
	ActionDsTestConn = "datasource.testConnection"

	ActionSqlExecute = "sql.execute"

	ActionProjectCreate = "project.create"
	ActionProjectUpdate = "project.update"
	ActionProjectDelete = "project.delete"

	ActionServerCreate   = "server.create"
	ActionServerUpdate   = "server.update"
	ActionServerDelete   = "server.delete"
	ActionServerTestConn = "server.testConnection"

	ActionBackupCreate  = "backup.create"
	ActionBackupDelete  = "backup.delete"
	ActionBackupTrigger = "backup.trigger"

	ActionBusinessCreate = "business.create"
	ActionBusinessUpdate = "business.update"
	ActionBusinessDelete = "business.delete"

	ActionPluginExecute = "plugin.execute"
)

type AuditLog struct {
	LogID     string    `gorm:"column:log_id;primaryKey;size:64" json:"logId"`
	UserID    string    `gorm:"column:user_id;size:64;index" json:"userId"`
	Username  string    `gorm:"column:username;size:128" json:"username"`
	Action    string    `gorm:"column:action;size:64;index" json:"action"`
	Module    string    `gorm:"column:module;size:64;index" json:"module"`
	TargetID  string    `gorm:"column:target_id;size:128" json:"targetId"`
	Target    string    `gorm:"column:target;size:128" json:"target"`
	IPAddress string    `gorm:"column:ip_address;size:64" json:"ipAddress"`
	UserAgent string    `gorm:"column:user_agent;size:512" json:"userAgent"`
	Status    string    `gorm:"column:status;size:32;index" json:"status"`
	Detail    string    `gorm:"column:detail;type:text" json:"detail"`
	CreatedAt time.Time `gorm:"column:created_at;index" json:"createdAt"`
}

func (AuditLog) TableName() string { return "audit_logs" }
