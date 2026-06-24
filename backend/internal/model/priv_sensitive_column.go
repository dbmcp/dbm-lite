/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import "time"

// SensitiveColumn 敏感列配置：用于 SQL 查询结果自动脱敏
type SensitiveColumn struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DatasourceID string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	DatabaseName string    `gorm:"column:database_name;size:128;index" json:"databaseName"`
	TblName      string    `gorm:"column:table_name;size:128;index" json:"tableName"`
	ColumnName   string    `gorm:"column:column_name;size:128;index" json:"columnName"`
	Rule         string    `gorm:"column:rule;size:32;default:'mask'" json:"rule"` // mask / email / phone / hide
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (s SensitiveColumn) TableName() string { return "priv_sensitive_columns" }

// 脱敏规则常量
const (
	SensitiveRuleMask  = "mask"
	SensitiveRuleEmail = "email"
	SensitiveRulePhone = "phone"
	SensitiveRuleHide  = "hide"
)
