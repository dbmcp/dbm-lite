/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import "time"

// QueryPrivApply 权限申请工单
type QueryPrivApply struct {
	ApplyID       string    `gorm:"column:apply_id;primaryKey;size:64" json:"applyId"`
	ApplicantID   string    `gorm:"column:applicant_id;size:64;index" json:"applicantId"`
	ApplicantName string   `gorm:"column:applicant_name;size:64" json:"applicantName"`
	ApproverID    string    `gorm:"column:approver_id;size:64" json:"approverId"`
	ApproverName string    `gorm:"column:approver_name;size:64" json:"approverName"`
	GroupID        string    `gorm:"column:group_id;size:64;index" json:"groupId"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	DatasourceName string `gorm:"column:datasource_name;size:128" json:"datasourceName"`
	DatabaseName string    `gorm:"column:database_name;size:128" json:"databaseName"`
	TblName      string    `gorm:"column:table_name;size:128" json:"tableName"`
	PrivType     string    `gorm:"column:priv_type;size:32;default:'table'" json:"privType"` // database / table / column
	OperationType string   `gorm:"column:operation_type;size:32;default:'select'" json:"operationType"` // select / dml / ddl / all
	Columns      string    `gorm:"column:columns;type:text" json:"columns"` // 授权列 JSON数组
	ValidDays    int       `gorm:"column:valid_days;default:7" json:"validDays"`
	RowLimit     int       `gorm:"column:row_limit;default:1000" json:"rowLimit"`
	Status       string    `gorm:"column:status;size:32;default:'pending';index" json:"status"` // pending / approved / rejected / expired
	ApplyRemark  string    `gorm:"column:apply_remark;type:text" json:"applyRemark"`
	ApprovalRemark string `gorm:"column:approval_remark;type:text" json:"approvalRemark"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (QueryPrivApply) TableName() string { return "query_priv_applies" }

// QueryPrivilege 审批通过后生成的有效权限（CheckSQLPriv 核心依据
type QueryPrivilege struct {
	PrivID       string    `gorm:"column:priv_id;primaryKey;size:64" json:"privId"`
	UserID       string    `gorm:"column:user_id;size:64;index" json:"userId"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	DatabaseName string    `gorm:"column:database_name;size:128" json:"databaseName"`
	TblName      string    `gorm:"column:table_name;size:128" json:"tableName"`
	PrivType     string    `gorm:"column:priv_type;size:32;default:'table'" json:"privType"`
	OperationType string   `gorm:"column:operation_type;size:32;default:'select'" json:"operationType"`
	Columns      string    `gorm:"column:columns;type:text" json:"columns"`
	RowLimit     int       `gorm:"column:row_limit;default:1000" json:"rowLimit"`
	ApplyID      string    `gorm:"column:apply_id;size:64" json:"applyId"`
	ExpireAt     time.Time `gorm:"column:expire_at;index" json:"expireAt"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
	IsExpired    bool      `gorm:"column:is_expired;default:false;index" json:"isExpired"`
}

func (QueryPrivilege) TableName() string { return "query_privileges" }
