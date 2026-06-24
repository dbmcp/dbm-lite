/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 *
 * 说明：本文件为平台核心元数据库初始化模块，针对 SQLite 高并发读写进行生产级配置。
 * 核心策略：
 *   1. journal_mode=WAL：启用预写日志，实现读写互不阻塞，是高并发能力的核心基础。
 *   2. synchronous=NORMAL：WAL 模式官方推荐值，事务提交不强制刷盘，检查点阶段同步主库。
 *   3. busy_timeout=60000：锁等待超时 60 秒，高并发下请求排队等待而非直接报错。
 *   4. cache_size=-65536：64MB 内存缓存，缓存热点数据以减少磁盘 IO。
 *   5. temp_store=MEMORY：临时表、排序、分组等中间数据直接放入内存，规避磁盘临时文件竞争。
 *   6. 连接池：MaxOpen=5，MaxIdle=5，避免过多写连接造成锁等待争抢。
 *   7. 后台协程每 60 秒执行一次 PASSIVE 模式 WAL checkpoint，防止 WAL 文件无限膨胀。
 *   8. 服务关闭时主动执行 TRUNCATE 模式 checkpoint，并关闭连接池，杜绝文件句柄泄漏与 WAL 残留。
 */

package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/crypto"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// ====== 全局可运维常量（集中管理，避免硬编码） ======

// SQLitePragmaBusyTimeout 单个 SQL 语句锁等待超时（毫秒）
// SQLite 数据库繁忙时请求排队等待，而不是立即返回 database is locked。
// 预留可通过环境变量 BUSY_TIMEOUT_MS 覆盖，测试环境建议 5000-10000。
const SQLitePragmaBusyTimeoutMs = 60000

// SQLitePragmaCacheSizeKB 64MB 内存缓存（负数单位为 KB）
const SQLitePragmaCacheSizeKB = -65536

// SQLiteMaxOpenConns SQLite 最大活跃连接数：由于 SQLite 为单写者，过多写连接只会加剧锁竞争。
// 读多写少场景可适度上调至 10-15，此处为兼顾双场景的保守值。
const SQLiteMaxOpenConns = 5

// SQLiteMaxIdleConns SQLite 最大空闲连接数，与活跃值保持一致以减少连接创建开销。
const SQLiteMaxIdleConns = 5

// SQLiteConnMaxLifetime 连接生命周期：30 分钟
const SQLiteConnMaxLifetime = 30 * time.Minute

// WALCheckpointInterval 自动 WAL checkpoint 周期，默认 1 分钟。
// 采用 PASSIVE 非阻塞模式，不影响业务读写，仅合并可提交数据。
const WALCheckpointInterval = 60 * time.Second

// MySQLMaxOpenConns MySQL 最大活跃连接数（兼容原配置，与 SQLite 解耦管理）
const MySQLMaxOpenConns = 100

// MySQLMaxIdleConns MySQL 最大空闲连接数
const MySQLMaxIdleConns = 10

// MySQLConnMaxLifetime MySQL 连接生命周期
const MySQLConnMaxLifetime = time.Hour

// ====== 全局状态 ======

// DB GORM 全局句柄，业务层统一使用该句柄进行数据持久化。
var DB *gorm.DB

// rawDBPlatform 保留底层 *sql.DB，用于执行 PRAGMA / checkpoint / 关闭连接池等操作。
var rawDBPlatform *sql.DB

// autoCheckpointStop 通知后台 WAL checkpoint 协程退出的信号通道。
var autoCheckpointStop chan struct{}

// autoCheckpointDone checkpoint 协程退出完成通知。
var autoCheckpointDone chan struct{}

// onceClose 保证 Shutdown 仅可被执行一次，避免重复关闭导致 panic。
var onceClose sync.Once

// ====== 对外 API ======

// Init 对外主入口，由 main.go 调用。根据配置选择 SQLite 或 MySQL 作为平台元数据库。
func Init() error {
	if err := initPlatformDB(); err != nil {
		return fmt.Errorf("init platform db failed: %w", err)
	}

	// 启动 SQLite 专用的 WAL checkpoint 后台协程（MySQL 模式下不启动）
	dbTypeForCheckpoint := strings.ToLower(strings.TrimSpace(config.App.DBType))
	if dbTypeForCheckpoint == "sqlite" && rawDBPlatform != nil {
		startAutoCheckpoint(rawDBPlatform, WALCheckpointInterval)
	}
	return nil
}

// Shutdown 服务关闭时需被调用，按顺序执行：
// 1. 通知后台 checkpoint 协程停止；2. 对 SQLite 执行 TRUNCATE checkpoint；3. 关闭底层连接池。
// 幂等，重复调用安全。
func Shutdown() {
	onceClose.Do(func() {
		// 1. 通知后台 checkpoint 协程退出
		if autoCheckpointStop != nil {
			close(autoCheckpointStop)
			if autoCheckpointDone != nil {
				select {
				case <-autoCheckpointDone:
				case <-time.After(5 * time.Second):
					fmt.Printf("[database] 等待 checkpoint 协程退出超时，继续关闭流程\n")
				}
			}
		}

		// 2. 对 SQLite 执行 TRUNCATE 模式 checkpoint，确保 WAL 文件被截断到 0 字节
		if rawDBPlatform != nil && DB != nil {
			// 判断当前是否为 SQLite：通过驱动名称或 DBType 判断，此处以配置为准
			dbType := strings.ToLower(strings.TrimSpace(config.App.DBType))
			if dbType == "sqlite" {
				if err := RunWALCheckpoint(rawDBPlatform, "TRUNCATE"); err != nil {
					fmt.Printf("[database] TRUNCATE checkpoint 失败: %v\n", err)
				} else {
					fmt.Println("[database] SQLite WAL 已清理，服务可安全退出")
				}
			}
		}

		// 3. 关闭底层 sql.DB 连接池，释放文件句柄
		if rawDBPlatform != nil {
			if err := rawDBPlatform.Close(); err != nil {
				fmt.Printf("[database] 关闭数据库连接池失败: %v\n", err)
			}
		}
	})
}

// ====== 内部实现 ======

// initPlatformDB 根据配置创建平台元数据库连接（SQLite 或 MySQL）。
func initPlatformDB() error {
	var (
		gdb    *gorm.DB
		rawDB  *sql.DB
		err    error
		dbType = strings.ToLower(strings.TrimSpace(config.App.DBType))
	)

	switch dbType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.App.MySQLUsername, config.App.MySQLPassword,
			config.App.MySQLHost, config.App.MySQLPort, config.App.MySQLDatabase)
		fmt.Printf("[database] Connecting to MySQL: %s:%s@%s:%s/%s\n", 
			config.App.MySQLUsername, "******", config.App.MySQLHost, config.App.MySQLPort, config.App.MySQLDatabase)
		
		gdb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Warn),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			fmt.Printf("[database] MySQL connection failed: %v\n", err)
			fmt.Printf("[database] Please check: 1) MySQL server is running 2) Host/Port is correct 3) Username/Password is correct 4) Database exists\n")
			return fmt.Errorf("open mysql failed: %w", err)
		}
		if rawDB, err = gdb.DB(); err != nil {
			return fmt.Errorf("get mysql raw db failed: %w", err)
		}
		
		// 验证MySQL连接
		if pingErr := rawDB.Ping(); pingErr != nil {
			fmt.Printf("[database] MySQL ping failed: %v\n", pingErr)
			return fmt.Errorf("mysql ping failed: %w", pingErr)
		}
		fmt.Println("[database] MySQL connection successful")
		
		rawDB.SetMaxOpenConns(MySQLMaxOpenConns)
		rawDB.SetMaxIdleConns(MySQLMaxIdleConns)
		rawDB.SetConnMaxLifetime(MySQLConnMaxLifetime)

	case "sqlite":
		// SQLite：统一拼接 DSN，启用进程共享缓存与读写创建模式
		dbPath := config.App.DBPath
		if dbPath == "" {
			dbPath = "dbm-lite.db"
		}
		// 确保上层目录存在，避免 Open 失败
		if dir := dirOf(dbPath); dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0o755)
		}
		dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbPath)

		gdb, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Warn),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			return fmt.Errorf("open sqlite failed: %w", err)
		}
		if rawDB, err = gdb.DB(); err != nil {
			return fmt.Errorf("get sqlite raw db failed: %w", err)
		}

		// 连接有效性验证：Ping
		if pingErr := rawDB.Ping(); pingErr != nil {
			_ = rawDB.Close()
			return fmt.Errorf("sqlite ping failed: %w", pingErr)
		}

		// 批量执行 PRAGMA：逐条捕获错误，失败则关闭连接并返回异常
		if err := applyConcurrencyPragmas(rawDB); err != nil {
			_ = rawDB.Close()
			return fmt.Errorf("apply sqlite pragmas failed: %w", err)
		}

		// 连接池：SQLite 单写者，不宜过大；此处设置为 5/5，兼顾读写并发。
		rawDB.SetMaxOpenConns(SQLiteMaxOpenConns)
		rawDB.SetMaxIdleConns(SQLiteMaxIdleConns)
		rawDB.SetConnMaxLifetime(SQLiteConnMaxLifetime)

	default:
		return fmt.Errorf("unsupported database type: %s (expected sqlite or mysql)", config.App.DBType)
	}

	DB = gdb
	rawDBPlatform = rawDB

	// SQLite schema 兼容性检查（与原逻辑保持一致，避免业务改动）
	if dbType == "sqlite" {
		if err := rebuildAuditLogTableIfNeeded(); err != nil {
			return fmt.Errorf("rebuild audit_logs failed: %w", err)
		}
		if err := rebuildSQLHistoryTableIfNeeded(); err != nil {
			return fmt.Errorf("rebuild sql_history failed: %w", err)
		}
	}

	if err := seedData(); err != nil {
		return fmt.Errorf("seed data failed: %w", err)
	}

	return nil
}

// applyConcurrencyPragmas 独立封装 SQLite 高并发 PRAGMA 执行逻辑，便于统一维护与扩展。
func applyConcurrencyPragmas(db *sql.DB) error {
	pragmas := []string{
		fmt.Sprintf("PRAGMA journal_mode=WAL;"),
		fmt.Sprintf("PRAGMA synchronous=NORMAL;"),
		fmt.Sprintf("PRAGMA busy_timeout=%d;", SQLitePragmaBusyTimeoutMs),
		fmt.Sprintf("PRAGMA cache_size=%d;", SQLitePragmaCacheSizeKB),
		fmt.Sprintf("PRAGMA temp_store=MEMORY;"),
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("exec %q failed: %w", p, err)
		}
	}
	return nil
}

// RunWALCheckpoint 通用 WAL checkpoint 函数，支持 PASSIVE/FULL/RESTART/TRUNCATE 四种模式。
// - PASSIVE：非阻塞，日常自动执行使用；
// - FULL：阻塞写操作，完成全量合并（期间读操作可正常执行）；
// - RESTART：等待所有读写完成，重置 WAL 文件；
// - TRUNCATE：在 RESTART 基础上进一步把 WAL 文件截断到 0 字节，用于服务停止、运维巡检。
func RunWALCheckpoint(db *sql.DB, mode string) error {
	mode = strings.ToUpper(strings.TrimSpace(mode))
	switch mode {
	case "PASSIVE", "FULL", "RESTART", "TRUNCATE":
	default:
		return fmt.Errorf("unsupported checkpoint mode: %s", mode)
	}
	stmt := fmt.Sprintf("PRAGMA wal_checkpoint(%s);", mode)
	_, err := db.Exec(stmt)
	return err
}

// startAutoCheckpoint 启动后台定时自动 checkpoint，独立协程运行，不阻塞主流程；仅使用 PASSIVE 模式。
// 通过 stop 通道可优雅退出，退出后 close done 以便外部等待。
func startAutoCheckpoint(db *sql.DB, interval time.Duration) {
	if db == nil {
		return
	}
	autoCheckpointStop = make(chan struct{})
	autoCheckpointDone = make(chan struct{})

	go func() {
		defer close(autoCheckpointDone)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-autoCheckpointStop:
				return
			case <-ticker.C:
				if err := RunWALCheckpoint(db, "PASSIVE"); err != nil {
					fmt.Printf("[database] 自动 PASSIVE checkpoint 失败: %v\n", err)
				}
			}
		}
	}()
}

// ====== 原有 schema 兼容逻辑（保持不变，确保不破坏业务层） ======

type sqliteColumn struct {
	Name string `gorm:"column:name"`
	Type string `gorm:"column:type"`
}

func rebuildAuditLogTableIfNeeded() error {
	var tableExists int
	err := DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='audit_logs'").Scan(&tableExists).Error
	if err != nil {
		return nil
	}
	if tableExists == 0 {
		return nil
	}

	var cols []sqliteColumn
	if err := DB.Raw("PRAGMA table_info(audit_logs)").Scan(&cols).Error; err != nil {
		return nil
	}

	needsRebuild := false
	for _, c := range cols {
		colName := strings.ToLower(c.Name)
		colType := strings.ToUpper(c.Type)
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

	hasDatasourceName := false
	hasSqlText := false
	for _, c := range cols {
		switch strings.ToLower(c.Name) {
		case "datasource_name":
			hasDatasourceName = true
		case "sql_text":
			hasSqlText = true
		}
	}

	needsRebuild := false
	if !hasDatasourceName || !hasSqlText {
		needsRebuild = true
	}

	if !needsRebuild {
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
	}

	if !needsRebuild {
		return nil
	}
	if err := DB.Exec("DROP TABLE IF EXISTS sql_history").Error; err != nil {
		return err
	}
	return nil
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
	}
}

// AutoMigrateExtra 用于在 database.Init() 之外注册额外表（由上层模块调用，避免循环依赖）
func AutoMigrateExtra(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("db not initialized")
	}
	for _, m := range models {
		if err := DB.AutoMigrate(m); err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "1091") || strings.Contains(errStr, "1061") ||
				strings.Contains(errStr, "Duplicate key") || strings.Contains(errStr, "Duplicate entry") ||
				strings.Contains(errStr, "already exists") ||
				strings.Contains(strings.ToLower(errStr), "duplicate") {
				continue
			}
			fmt.Printf("[database] extra table migrate warning: %T: %v（继续）\n", m, err)
		}
	}
	return nil
}

func seedData() error {
	var admin model.User
	err := DB.Model(&model.User{}).Where("username = ?", config.App.AdminUsername).First(&admin).Error
	if err != nil {
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
		return nil
	}

	if !crypto.VerifyPassword(admin.PasswordHash, config.App.AdminPassword) {
		pwdHash, err := crypto.HashPassword(config.App.AdminPassword)
		if err != nil {
			return err
		}
		DB.Model(&model.User{}).Where("username = ?", config.App.AdminUsername).Update("password_hash", pwdHash)
	}
	if admin.Status != model.StatusActive {
		DB.Model(&model.User{}).Where("username = ?", config.App.AdminUsername).Update("status", model.StatusActive)
	}
	return nil
}

// safeAutoMigrate 对每个表逐个执行 AutoMigrate，单个表迁移失败不影响其他表。
// 自动忽略常见的索引/外键/重复数据等非致命错误，保证服务能正常启动。
func safeAutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("db not initialized")
	}
	models := allModels()
	for _, m := range models {
		if migrateErr := DB.AutoMigrate(m); migrateErr != nil {
			errStr := migrateErr.Error()
			if strings.Contains(errStr, "1091") || strings.Contains(errStr, "1061") ||
				strings.Contains(errStr, "Duplicate key") || strings.Contains(errStr, "Duplicate entry") ||
				strings.Contains(errStr, "already exists") || strings.Contains(errStr, "already exists") ||
				strings.Contains(strings.ToLower(errStr), "duplicate") {
				continue
			}
			fmt.Printf("[database] 表迁移警告: %T: %v（继续）\n", m, migrateErr)
		}
	}
	return nil
}



// dirOf 返回文件路径所在目录，不依赖 path/filepath 以避免 Windows/Linux 路径差异。
func dirOf(fp string) string {
	fp = strings.TrimSpace(fp)
	if fp == "" {
		return "."
	}
	// 统一按最后一个斜杠或反斜杠切分
	idx := strings.LastIndexAny(fp, "/\\")
	if idx < 0 {
		return "."
	}
	return fp[:idx]
}
