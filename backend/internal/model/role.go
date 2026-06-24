/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import "time"

// Role 角色（权限组）：一个用户可绑定多个角色，权限码取并集
type Role struct {
	RoleID      string    `gorm:"column:role_id;primaryKey;size:64" json:"roleId"`
	Name        string    `gorm:"column:name;size:128;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"column:description;size:512" json:"description"`
	Codes       string    `gorm:"column:codes;type:text" json:"codes"`
	Status      string    `gorm:"column:status;size:32;default:'active'" json:"status"`
	BuiltIn     bool      `gorm:"column:built_in;default:false" json:"builtIn"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Role) TableName() string { return "roles" }

// UserRoleBind 用户-角色多对多绑定
type UserRoleBind struct {
	UserID    string    `gorm:"column:user_id;primaryKey;size:64;index:idx_ur_user" json:"userId"`
	RoleID    string    `gorm:"column:role_id;primaryKey;size:64;index:idx_ur_role" json:"roleId"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (UserRoleBind) TableName() string { return "user_role_binds" }

// PermissionPoint 权限点定义：支持菜单/按钮/接口/库表操作
type PermissionPoint struct {
	PointID    string    `gorm:"column:point_id;primaryKey;size:64" json:"pointId"`
	Code       string    `gorm:"column:code;size:128;not null;uniqueIndex" json:"code"`
	Name       string    `gorm:"column:name;size:128" json:"name"`
	Type       string    `gorm:"column:type;size:32;default:'menu'" json:"type"` // menu/button/api/sql
	Module     string    `gorm:"column:module;size:64" json:"module"`
	SortOrder  int       `gorm:"column:sort_order;default:0" json:"sortOrder"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (PermissionPoint) TableName() string { return "permission_points" }

// RolePermissionBind 角色-权限点多对多
type RolePermissionBind struct {
	RoleID    string    `gorm:"column:role_id;primaryKey;size:64;index:idx_rp_role" json:"roleId"`
	PointCode string    `gorm:"column:point_code;primaryKey;size:128;index:idx_rp_code" json:"pointCode"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (RolePermissionBind) TableName() string { return "role_permission_binds" }

// ====== 权限码常量（分层分类） ======
const (
	// 账号相关菜单
	PermMenuAccount     = "menu_account"
	PermMenuRole        = "menu_role"
	PermMenuAuditLog    = "menu_audit_log"

	// 账号相关操作
	PermAccountEdit     = "account_edit"
	PermAccountResetPwd = "account_reset_pwd"
	PermAccountLock     = "account_lock"

	// 角色管理
	PermRoleEdit        = "role_edit"
	PermRoleAssignUser  = "role_assign_user"
	PermRoleAssignPerm  = "role_assign_permission"

	// 对象权限菜单
	PermMenuPrivGroup   = "menu_priv_group"
	PermMenuPrivApply   = "menu_priv_apply"
	PermMenuPrivMy      = "menu_priv_my"
	PermMenuPrivAudit   = "menu_priv_audit"

	// 对象权限操作
	PermPrivGroupEdit   = "priv_group_edit"
	PermPrivApplyOper   = "priv_apply_operate"
	PermPrivAuditOper   = "priv_audit_operate"

	// SQL 执行超级豁免权限
	PermSqlQueryAll     = "sql_query_all"
	PermSqlQueryGroup   = "sql_query_group"
)

// DefaultSuperAdminCodes 超级管理员角色默认拥有的权限码
func DefaultSuperAdminCodes() []string {
	return []string{
		PermMenuAccount, PermAccountEdit, PermAccountResetPwd, PermAccountLock,
		PermMenuRole, PermRoleEdit, PermRoleAssignUser, PermRoleAssignPerm,
		PermMenuAuditLog,
		PermMenuPrivGroup, PermPrivGroupEdit,
		PermMenuPrivApply, PermPrivApplyOper,
		PermMenuPrivMy,
		PermMenuPrivAudit, PermPrivAuditOper,
		PermSqlQueryAll, PermSqlQueryGroup,
	}
}

// DefaultMemberCodes 普通用户角色默认拥有的权限码
func DefaultMemberCodes() []string {
	return []string{
		PermMenuPrivMy,
		PermMenuPrivApply,
	}
}