package migrate

import "time"

const CurrentDBVersion = "v1.0.2"

type MigrateVersion struct {
	ID            uint      `gorm:"primaryKey"`
	VersionCode   string    `gorm:"column:version_code;unique;not null;size:64"`
	MigrateContent string   `gorm:"column:migrate_content;type:text"`
	ExecAt        time.Time `gorm:"column:exec_at;not null;default:CURRENT_TIMESTAMP"`
}

func (MigrateVersion) TableName() string {
	return "db_migrate_version"
}