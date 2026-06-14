/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package dbtype

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"dbm-lite/config"
	"dbm-lite/pkg/crypto"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	TypeMySQL  = "mysql"
	TypeTiDB   = "tidb"
	TypeSQLite = "sqlite"
)

type ConnectionParams struct {
	Type       string
	Host       string
	Port       int
	Username   string
	Password   string
	Database   string
	FilePath   string
	OpenMode   string // rw, ro
	Charset    string // utf8mb4, utf8, latin1, etc.
	Timezone   string // UTC, Asia/Shanghai, Local
	SSLMode    string // true/false
	SSLCAFile  string // ca certificate file
	TimeoutSec int    // connection timeout seconds
}

type ConnectionInfo struct {
	DB       *sql.DB
	GormDB   *gorm.DB
	Database string
	Type     string
}

type TestResult struct {
	Success   bool
	Message   string
	LatencyMs int64
	Version   string
}

var (
	connMap  = make(map[string]*ConnectionInfo)
	connLock sync.RWMutex
)

func SupportedTypes() []string {
	return []string{TypeMySQL, TypeTiDB, TypeSQLite}
}

func IsSupported(t string) bool {
	t = strings.ToLower(t)
	return t == TypeMySQL || t == TypeTiDB || t == TypeSQLite
}

func IsSSL(mode string) bool {
	m := strings.ToLower(strings.TrimSpace(mode))
	if m == "true" || m == "1" || m == "on" || m == "yes" || m == "tls" || m == "require" {
		return true
	}
	return false
}

func buildDSN(p *ConnectionParams) (string, string, error) {
	switch strings.ToLower(p.Type) {
	case TypeMySQL, TypeTiDB:
		charset := p.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		timezone := p.Timezone
		if timezone == "" {
			timezone = "Local"
		}
		timeoutSec := p.TimeoutSec
		if timeoutSec <= 0 {
			timeoutSec = 10
		}
		port := p.Port
		if port <= 0 {
			port = 3306
		}
		// 组装参数列表
		params := []string{
			fmt.Sprintf("charset=%s", charset),
			"parseTime=True",
			fmt.Sprintf("loc=%s", timezone),
			fmt.Sprintf("timeout=%ds", timeoutSec),
		}
		if IsSSL(p.SSLMode) {
			params = append(params, "tls=custom")
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			p.Username, p.Password, p.Host, port, p.Database, strings.Join(params, "&"))
		return dsn, p.Type, nil
	case TypeSQLite:
		path := p.FilePath
		if path == "" {
			return ":memory:", TypeSQLite, nil
		}
		mode := strings.ToLower(p.OpenMode)
		if mode != "rw" && mode != "ro" {
			mode = "rw"
		}
		// modernc.org/sqlite 使用 DSN 格式: file:///path?mode=rw
		// gorm.io/driver/sqlite 直接使用 file path
		if mode == "ro" {
			return path + "?_pragma=journal_mode(WAL)&mode=ro", TypeSQLite, nil
		}
		return path + "?_pragma=journal_mode(WAL)", TypeSQLite, nil
	default:
		return "", "", fmt.Errorf("unsupported database type: %s", p.Type)
	}
}

// ValidateSQLiteFile 校验指定路径是否是合法的 SQLite 数据库文件
func ValidateSQLiteFile(path string) error {
	if path == "" {
		return errors.New("SQLite 文件路径不能为空")
	}
	if path == ":memory:" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", path)
		}
		return fmt.Errorf("无法访问文件: %s (%w)", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("路径指向目录而不是文件: %s", path)
	}
	if info.Size() > 0 && info.Size() < 512 {
		return fmt.Errorf("文件大小异常 (%d 字节)，可能不是合法的 SQLite 数据库文件", info.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("读取文件头失败: %w", err)
	}
	if n < 16 {
		return fmt.Errorf("文件内容不足，不是合法的 SQLite 数据库文件")
	}
	magic := "SQLite format 3\x00"
	if string(header) != magic {
		return fmt.Errorf("文件头不匹配 SQLite 格式（应为 'SQLite format 3'），文件可能不是合法的 SQLite 数据库文件")
	}
	return nil
}

func Connect(key string, p *ConnectionParams) (*ConnectionInfo, error) {
	connLock.RLock()
	if ci, ok := connMap[key]; ok {
		connLock.RUnlock()
		return ci, nil
	}
	connLock.RUnlock()

	dsn, dbType, err := buildDSN(p)
	if err != nil {
		return nil, err
	}

	var sqlDB *sql.DB
	var gormDB *gorm.DB

	switch dbType {
	case TypeMySQL, TypeTiDB:
		sqlDB, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("connect failed: %w", err)
		}
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(3)
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
		if err = sqlDB.Ping(); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("ping failed: %w", err)
		}
		gormDB, err = gorm.Open(mysql.New(mysql.Config{Conn: sqlDB}), &gorm.Config{})
	case TypeSQLite:
		sqlDB, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, fmt.Errorf("connect failed: %w", err)
		}
		sqlDB.SetMaxOpenConns(1)
		if err = sqlDB.Ping(); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("ping failed: %w", err)
		}
		gormDB, err = gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported type: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("gorm init failed: %w", err)
	}

	ci := &ConnectionInfo{DB: sqlDB, GormDB: gormDB, Database: p.Database, Type: dbType}

	connLock.Lock()
	connMap[key] = ci
	connLock.Unlock()

	return ci, nil
}

// TestConnect 先验证数据库实例连通性（不依赖具体数据库名），
// 如配置了 database 再验证该数据库是否存在。记录延迟和版本信息。
func TestConnect(p *ConnectionParams) *TestResult {
	result := &TestResult{Success: false, Message: "", LatencyMs: 0, Version: ""}
	start := time.Now()
	dbType := strings.ToLower(p.Type)

	// --- 第一步：连接到实例（不指定具体数据库） ---
	instanceParams := &ConnectionParams{
		Type:     p.Type,
		Host:     p.Host,
		Port:     p.Port,
		Username: p.Username,
		Password: p.Password,
		Database: "", // 关键：空数据库名以验证实例连通性
		FilePath: p.FilePath,
		OpenMode: p.OpenMode,
		Charset:  p.Charset,
		Timezone: p.Timezone,
	}
	instanceDSN, _, err := buildDSN(instanceParams)
	if err != nil {
		result.Message = err.Error()
		return result
	}

	var db *sql.DB
	switch dbType {
	case TypeMySQL, TypeTiDB:
		db, err = sql.Open("mysql", instanceDSN)
	case TypeSQLite:
		db, err = sql.Open("sqlite", instanceDSN)
	default:
		result.Message = fmt.Sprintf("unsupported type: %s", p.Type)
		return result
	}
	if err != nil {
		result.Message = fmt.Sprintf("打开连接失败: %v", err)
		return result
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(30 * time.Second)

	// 执行真实连接
	pingStart := time.Now()
	if err = db.Ping(); err != nil {
		lat := time.Since(pingStart).Milliseconds()
		result.LatencyMs = lat
		lowerErr := strings.ToLower(err.Error())
		switch {
		case strings.Contains(lowerErr, "timeout") || strings.Contains(lowerErr, "i/o timeout"):
			result.Message = fmt.Sprintf("无法连接到 %s:%d (耗时 %dms，超时): %v",
				p.Host, p.Port, lat, err)
		case strings.Contains(lowerErr, "access denied") || strings.Contains(lowerErr, "password"):
			result.Message = fmt.Sprintf("认证失败 (用户名或密码错误): %v", err)
		case strings.Contains(lowerErr, "unknown database"):
			result.Message = fmt.Sprintf("连接失败 (%dms): %v", lat, err)
		default:
			result.Message = fmt.Sprintf("连接失败 (%dms): %v", lat, err)
		}
		return result
	}

	// --- 第二步：查询版本信息 ---
	var version string
	switch dbType {
	case TypeMySQL, TypeTiDB:
		row := db.QueryRow("SELECT VERSION()")
		_ = row.Scan(&version)
	case TypeSQLite:
		row := db.QueryRow("SELECT sqlite_version()")
		_ = row.Scan(&version)
	}

	// --- 第三步（可选）：如果指定了 database，验证该数据库是否存在 ---
	if p.Database != "" {
		dbExists := false
		switch dbType {
		case TypeMySQL, TypeTiDB:
			var cnt int
			err = db.QueryRow("SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", p.Database).Scan(&cnt)
			if err == nil && cnt > 0 {
				dbExists = true
			}
		case TypeSQLite:
			// SQLite 通过 PRAGMA database_list 检查
			var seq int
			var name, f string
			rows, qErr := db.Query("PRAGMA database_list")
			if qErr == nil {
				for rows.Next() {
					if scanErr := rows.Scan(&seq, &name, &f); scanErr == nil && name == p.Database {
						dbExists = true
						break
					}
				}
				rows.Close()
			}
		}
		if !dbExists {
			latency := time.Since(start).Milliseconds()
			result.Success = false
			result.LatencyMs = latency
			result.Version = version
			result.Message = fmt.Sprintf("实例连通正常 (%dms)，但指定的数据库 %q 不存在", latency, p.Database)
			return result
		}
	}

	latency := time.Since(start).Milliseconds()
	result.Success = true
	result.LatencyMs = latency
	result.Version = version
	if p.Database != "" {
		result.Message = fmt.Sprintf("连接成功，数据库 %q 存在，延迟 %d ms", p.Database, latency)
	} else {
		result.Message = fmt.Sprintf("连接成功，延迟 %d ms", latency)
	}
	return result
}

func Get(key string) *ConnectionInfo {
	connLock.RLock()
	defer connLock.RUnlock()
	return connMap[key]
}

func Close(key string) {
	connLock.Lock()
	defer connLock.Unlock()
	if ci, ok := connMap[key]; ok {
		if ci.DB != nil {
			_ = ci.DB.Close()
		}
		delete(connMap, key)
	}
}

func DecryptPassword(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}
	plain, err := crypto.DecryptAES(encrypted, config.App.AESKey)
	if err != nil {
		return "", fmt.Errorf("decrypt failed: %w", err)
	}
	return plain, nil
}

func EncryptPassword(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	return crypto.EncryptAES(plain, config.App.AESKey)
}
