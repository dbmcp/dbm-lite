package datasource_auth

import "time"

type DatasourceUserRoleRel struct {
	ID         string    `gorm:"column:id;primaryKey;size:64" json:"id"`
	UserID     string    `gorm:"column:user_id;size:64;index" json:"userId"`
	RoleID     string    `gorm:"column:role_id;size:64;index" json:"roleId"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (DatasourceUserRoleRel) TableName() string {
	return "datasource_user_role_rels"
}