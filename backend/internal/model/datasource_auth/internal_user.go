package datasource_auth

import "time"

type DatasourceInternalUser struct {
	ID              string     `gorm:"column:id;primaryKey;size:64" json:"id"`
	DatasourceID    string     `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	Username        string     `gorm:"column:username;size:128" json:"username"`
	Host            string     `gorm:"column:host;size:128;default:'%'" json:"host"`
	IsBuiltIn       bool       `gorm:"column:is_built_in;default:false" json:"isBuiltIn"`
	Password        string     `gorm:"column:password;size:512" json:"-"`
	Status          string     `gorm:"column:status;size:32;default:'active'" json:"status"`
	Remark          string     `gorm:"column:remark;size:512" json:"remark"`
	PasswordExpire  *time.Time `gorm:"column:password_expire" json:"passwordExpire"`
	LockedUntil     *time.Time `gorm:"column:locked_until" json:"lockedUntil"`
	PasswordPolicy  string     `gorm:"column:password_policy;size:32" json:"passwordPolicy"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"createdAt"`
	CreatedBy       string     `gorm:"column:created_by;size:64" json:"createdBy"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (DatasourceInternalUser) TableName() string {
	return "datasource_internal_users"
}

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"

	PasswordPolicyNone     = "none"
	PasswordPolicyMedium   = "medium"
	PasswordPolicyStrong   = "strong"
)