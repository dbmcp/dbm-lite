package migrate

import (
	"dbm-lite/internal/model"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type DropTask struct {
	Table    string
	Type     string
	Name     string
}

var onceCleanTasks = []DropTask{
	{Table: "users", Type: "foreign key", Name: "uni_users_username"},
	{Table: "plugins", Type: "foreign key", Name: "uni_plugins_name"},
	{Table: "roles", Type: "foreign key", Name: "uni_roles_name"},
	{Table: "permission_points", Type: "foreign key", Name: "uni_permission_points_code"},
	{Table: "business_members", Type: "foreign key", Name: "uni_bu_user"},
	
	{Table: "business_members", Type: "index", Name: "idx_bu_user"},
	{Table: "permission_points", Type: "index", Name: "idx_permission_points_code"},
	{Table: "plugins", Type: "index", Name: "idx_plugins_name"},
	{Table: "roles", Type: "index", Name: "idx_roles_name"},
	{Table: "users", Type: "index", Name: "idx_users_username"},
	{Table: "audit_logs", Type: "index", Name: "idx_audit_logs_user_id"},
}

func ExecOnceCleanDDL(db *gorm.DB) {
	for _, task := range onceCleanTasks {
		var exists bool
		var err error
		
		if task.Type == "foreign key" {
			exists, err = CheckForeignKeyExists(db, task.Table, task.Name)
		} else {
			exists, err = CheckIndexExists(db, task.Table, task.Name)
		}
		
		if err != nil {
			fmt.Printf("[migrate] Check %s %s.%s failed: %v\n", task.Type, task.Table, task.Name, err)
			continue
		}
		
		if !exists {
			continue
		}
		
		var ddl string
		if task.Type == "foreign key" {
			ddl = fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY `%s`", task.Table, task.Name)
		} else {
			ddl = fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", task.Table, task.Name)
		}
		
		if execErr := db.Exec(ddl).Error; execErr != nil {
			fmt.Printf("[migrate] Execute DROP %s %s failed: %v\n", task.Type, task.Name, execErr)
		} else {
			fmt.Printf("[migrate] Cleaned %s `%s`.`%s`\n", task.Type, task.Table, task.Name)
		}
	}
}

func allModels() []interface{} {
	return []interface{}{
		&model.User{},
		&model.Project{},
		&model.ProjectMember{},
		&model.Business{},
		&model.BusinessMember{},
		&model.Server{},
		&model.Datasource{},
		&model.AuditLog{},
		&model.SQLHistory{},
		&model.SQLWindow{},
		&model.SavedQuery{},
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
		&model.QueryPrivApply{},
		&model.QueryPrivilege{},
		&model.PrivAuditLog{},
		&model.SensitiveColumn{},
		&model.Role{},
		&model.RolePermissionBind{},
		&model.PermissionPoint{},
		&model.UserRoleBind{},
		&model.SQLFavorite{},
	}
}

func RunAutoMigrate(db *gorm.DB) error {
	if err := CreateVersionTable(db); err != nil {
		return fmt.Errorf("create version table failed: %w", err)
	}
	
	latestVersion, err := GetLatestVersion(db)
	if err != nil {
		return fmt.Errorf("get latest version failed: %w", err)
	}
	
	if latestVersion == CurrentDBVersion {
		fmt.Println("[migrate] Database schema is up to date (version:", CurrentDBVersion, "), skipping migration")
		return nil
	}
	
	fmt.Printf("[migrate] Starting migration from version '%s' to '%s'\n", latestVersion, CurrentDBVersion)
	
	if latestVersion == "" {
		fmt.Println("[migrate] First time setup, performing one-time cleanup...")
		ExecOnceCleanDDL(db)
	}
	
	models := allModels()
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "1091") || strings.Contains(errStr, "1061") ||
				strings.Contains(errStr, "Duplicate key") || strings.Contains(errStr, "Duplicate entry") ||
				strings.Contains(errStr, "already exists") || strings.Contains(strings.ToLower(errStr), "duplicate") {
				continue
			}
			fmt.Printf("[migrate] Table migration warning: %T: %v (continuing)\n", m, err)
		}
	}
	
	if err := MarkMigrateCompleted(db, CurrentDBVersion, "Auto migration completed"); err != nil {
		return fmt.Errorf("mark migrate completed failed: %w", err)
	}
	
	fmt.Printf("[migrate] Migration completed successfully, current version: %s\n", CurrentDBVersion)
	return nil
}