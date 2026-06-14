/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package database

import (
	"fmt"
	"strings"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/crypto"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(config.App.DBPath), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	})
	if err != nil {
		return fmt.Errorf("open database failed: %w", err)
	}

	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	// Fix schema compatibility: old SQLite tables may have columns with
	// incompatible types (e.g., user_id as INTEGER instead of TEXT). SQLite
	// does not support ALTER COLUMN, so we detect and rebuild such tables.
	if err := rebuildAuditLogTableIfNeeded(); err != nil {
		return fmt.Errorf("rebuild audit_logs failed: %w", err)
	}
	if err := rebuildSQLHistoryTableIfNeeded(); err != nil {
		return fmt.Errorf("rebuild sql_history failed: %w", err)
	}

	// AutoMigrate all models
	models := []interface{}{
		&model.User{},
		&model.Project{},
		&model.ProjectMember{},
		&model.Business{},
		&model.BusinessMember{},
		&model.Server{},
		&model.Datasource{},
		&model.AuditLog{},
		&model.SQLHistory{},
		&model.BackupPolicy{},
		&model.BackupRecord{},
		&model.InspectTask{},
		&model.InspectReport{},
		&model.SlowLog{},
		&model.HaCluster{},
		&model.HaNode{},
		&model.Plugin{},
		&model.CapacityStat{},
		&model.DBUser{},
		&model.SystemSetting{},
	}

	if err := DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	if err := seedData(); err != nil {
		return fmt.Errorf("seed data failed: %w", err)
	}

	return nil
}

type sqliteColumn struct {
	Name string `gorm:"column:name"`
	Type string `gorm:"column:type"`
}

func rebuildAuditLogTableIfNeeded() error {
	// Check if audit_logs table exists
	var tableExists int
	err := DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='audit_logs'").Scan(&tableExists).Error
	if err != nil {
		// Table doesn't exist yet, nothing to fix — AutoMigrate will create it
		return nil
	}
	if tableExists == 0 {
		return nil
	}

	// Inspect column types
	var cols []sqliteColumn
	if err := DB.Raw("PRAGMA table_info(audit_logs)").Scan(&cols).Error; err != nil {
		return nil
	}

	needsRebuild := false
	for _, c := range cols {
		colName := strings.ToLower(c.Name)
		colType := strings.ToUpper(c.Type)
		// TEXT columns that may have been created as INTEGER in older schemas
		if (colName == "user_id" || colName == "target_id" || colName == "ip_address" ||
			colName == "username" || colName == "action" || colName == "module" ||
			colName == "status" || colName == "user_agent") &&
			(strings.Contains(colType, "INT") || strings.Contains(colType, "BOOL")) {
			needsRebuild = true
			break
		}
	}

	if !needsRebuild {
		return nil
	}

	// Drop the old table so AutoMigrate can recreate it with correct TEXT types
	if err := DB.Exec("DROP TABLE IF EXISTS audit_logs").Error; err != nil {
		return err
	}
	return nil
}

func rebuildSQLHistoryTableIfNeeded() error {
	var tableExists int
	err := DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sql_history'").Scan(&tableExists).Error
	if err != nil {
		return nil
	}
	if tableExists == 0 {
		return nil
	}

	var cols []sqliteColumn
	if err := DB.Raw("PRAGMA table_info(sql_history)").Scan(&cols).Error; err != nil {
		return nil
	}

	needsRebuild := false
	for _, c := range cols {
		colName := strings.ToLower(c.Name)
		colType := strings.ToUpper(c.Type)
		if (colName == "history_id" || colName == "user_id" || colName == "datasource_id" ||
			colName == "username" || colName == "database_name" || colName == "sql" ||
			colName == "status" || colName == "error_message") &&
			(strings.Contains(colType, "INT") || strings.Contains(colType, "BOOL")) {
			needsRebuild = true
			break
		}
	}

	if !needsRebuild {
		return nil
	}

	if err := DB.Exec("DROP TABLE IF EXISTS sql_history").Error; err != nil {
		return err
	}
	return nil
}

func seedData() error {
	var admin model.User
	err := DB.Model(&model.User{}).Where("username = ?", config.App.AdminUsername).First(&admin).Error
	if err != nil {
		// admin用户不存在，创建新的
		pwdHash, err := crypto.HashPassword(config.App.AdminPassword)
		if err != nil {
			return err
		}
		admin = model.User{
			UserID:       uuid.New().String(),
			Username:     config.App.AdminUsername,
			PasswordHash: pwdHash,
			DisplayName:  "系统管理员",
			Email:        "admin@dbm-lite.local",
			Role:         model.RoleAdmin,
			Status:       model.StatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := DB.Create(&admin).Error; err != nil {
			return err
		}
	} else {
		// admin用户存在，验证密码哈希是否与当前算法兼容
		if !crypto.VerifyPassword(admin.PasswordHash, config.App.AdminPassword) {
			// 旧数据库中的哈希可能不兼容当前算法，更新为新哈希
			pwdHash, err := crypto.HashPassword(config.App.AdminPassword)
			if err != nil {
				return err
			}
			DB.Model(&model.User{}).Where("username = ?", config.App.AdminUsername).Update("password_hash", pwdHash)
		}
		// 确保状态为活跃
		if admin.Status != model.StatusActive {
			DB.Model(&model.User{}).Where("username = ?", config.App.AdminUsername).Update("status", model.StatusActive)
		}
	}
	return nil
}
