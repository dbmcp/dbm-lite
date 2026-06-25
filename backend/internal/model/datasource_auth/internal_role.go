package datasource_auth

import "time"

type DatasourceInternalRole struct {
	ID            string    `gorm:"column:id;primaryKey;size:64" json:"id"`
	DatasourceID  string    `gorm:"column:datasource_id;size:64;index" json:"datasourceId"`
	Name          string    `gorm:"column:name;size:128" json:"name"`
	Description   string    `gorm:"column:description;size:512" json:"description"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (DatasourceInternalRole) TableName() string {
	return "datasource_internal_roles"
}