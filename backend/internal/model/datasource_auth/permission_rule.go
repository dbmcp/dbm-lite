package datasource_auth

import "time"

type DatasourcePermissionRule struct {
	ID              string    `gorm:"column:id;primaryKey;size:64" json:"id"`
	DatasourceID    string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	PrincipalType   string    `gorm:"column:principal_type;size:32" json:"principalType"`
	PrincipalID     string    `gorm:"column:principal_id;size:64;index" json:"principalId"`
	PrivilegeType   string    `gorm:"column:privilege_type;size:32" json:"privilegeType"`
	ObjectLevel     string    `gorm:"column:object_level;size:32" json:"objectLevel"`
	DatabaseName    string    `gorm:"column:database_name;size:128" json:"databaseName"`
	Table           string    `gorm:"column:table_name;size:128" json:"tableName"`
	Columns         string    `gorm:"column:columns;type:text" json:"columns"`
	Enabled         bool      `gorm:"column:enabled;default:true" json:"enabled"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (DatasourcePermissionRule) TableName() string {
	return "datasource_permission_rules"
}

const (
	PrincipalTypeUser = "user"
	PrincipalTypeRole = "role"

	PrivilegeTypeReadonly = "readonly"
	PrivilegeTypeDML      = "dml"
	PrivilegeTypeDDL      = "ddl"

	ObjectLevelDatabase = "database"
	ObjectLevelTable    = "table"
	ObjectLevelColumn   = "column"
	ObjectLevelView     = "view"
	ObjectLevelTrigger  = "trigger"
)