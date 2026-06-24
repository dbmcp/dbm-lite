/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import "time"

// PrivAuditLog 权限审计日志：记录所有账号/角色/权限变更操作
type PrivAuditLog struct {
	LogID      string    `gorm:"column:log_id;primaryKey;size:64" json:"logId"`
	OperatorID string    `gorm:"column:operator_id;size:64;index" json:"operatorId"`
	Operator   string    `gorm:"column:operator;size:64" json:"operator"`
	OperType   string    `gorm:"column:oper_type;size:64;index" json:"operType"`
	ApplyID    string    `gorm:"column:apply_id;size:64;index" json:"applyId"`
	TargetID   string    `gorm:"column:target_id;size:64" json:"targetId"`
	Before     string    `gorm:"column:before;type:text" json:"before"`
	After      string    `gorm:"column:after;type:text" json:"after"`
	Detail     string    `gorm:"column:detail;size:512" json:"detail"`
	CreatedAt  time.Time `gorm:"column:created_at;index" json:"createdAt"`
}

func (PrivAuditLog) TableName() string { return "priv_audit_logs" }

// 权限审计操作类型
const (
	OperTypeAccountCreate   = "account.create"
	OperTypeAccountUpdate   = "account.update"
	OperTypeAccountDelete   = "account.delete"
	OperTypeAccountResetPwd = "account.reset_password"
	OperTypeAccountLock     = "account.lock"
	OperTypeAccountUnlock   = "account.unlock"

	OperTypeRoleCreate    = "role.create"
	OperTypeRoleUpdate    = "role.update"
	OperTypeRoleDelete    = "role.delete"
	OperTypeUserBindRole  = "role.bind_user"
	OperTypeUserUnbindRole = "role.unbind_user"

	OperTypeGroupCreate     = "group.create"
	OperTypeGroupUpdate     = "group.update"
	OperTypeGroupDelete     = "group.delete"
	OperTypeGroupBindUser   = "group.bind_user"
	OperTypeGroupUnbindUser = "group.unbind_user"
	OperTypeGroupBindDs     = "group.bind_datasource"
	OperTypeGroupUnbindDs   = "group.unbind_datasource"

	OperTypeApplySubmit  = "apply.submit"
	OperTypeApplyApprove = "apply.approve"
	OperTypeApplyReject  = "apply.reject"
	OperTypeApplyExpire  = "apply.expire"
	OperTypePrivRevoke   = "priv.revoke"

	OperTypeSqlBlocked = "sql.blocked"
)
