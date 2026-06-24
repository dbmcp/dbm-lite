package migrate

import (
	"time"

	"gorm.io/gorm"
)

func GetLatestVersion(db *gorm.DB) (string, error) {
	var v MigrateVersion
	err := db.Order("id DESC").First(&v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return v.VersionCode, nil
}

func CheckForeignKeyExists(db *gorm.DB, tableName, fkName string) (bool, error) {
	var count int
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.KEY_COLUMN_USAGE 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = ? 
		  AND CONSTRAINT_NAME = ? 
		  AND REFERENCED_TABLE_NAME IS NOT NULL
	`, tableName, fkName).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func CheckIndexExists(db *gorm.DB, tableName, indexName string) (bool, error) {
	var count int
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.STATISTICS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = ? 
		  AND INDEX_NAME = ?
	`, tableName, indexName).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func MarkMigrateCompleted(db *gorm.DB, versionCode, content string) error {
	v := MigrateVersion{
		VersionCode:    versionCode,
		MigrateContent: content,
		ExecAt:         time.Now(),
	}
	return db.Create(&v).Error
}

func CreateVersionTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS db_migrate_version (
			id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增主键',
			version_code VARCHAR(64) NOT NULL COMMENT '数据库结构语义化版本',
			migrate_content TEXT COMMENT '本次变更说明',
			exec_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '迁移执行时间',
			UNIQUE KEY uk_version (version_code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库迁移版本控制表'
	`).Error
}