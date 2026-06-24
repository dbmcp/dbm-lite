/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package dbtype

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"dbm-lite/config"
	"dbm-lite/pkg/crypto"

	dmysql "github.com/go-sql-driver/mysql"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
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
	connMap   = make(map[string]*ConnectionInfo)
	connLock  sync.RWMutex
	txMap     = make(map[string]*sql.Tx)
	txLock    sync.RWMutex

	tlsConfigRegistry = make(map[string]*tls.Config)
	tlsRegistryLock   sync.Mutex
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

func registerTLSConfig(caInput string) (string, error) {
	caData := []byte(caInput)

	tlsRegistryLock.Lock()
	defer tlsRegistryLock.Unlock()

	sum := sha256.Sum256(caData)
	name := "dbmlite_" + hex.EncodeToString(sum[:12])

	if _, ok := tlsConfigRegistry[name]; ok {
		return name, nil
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if ok := rootCAs.AppendCertsFromPEM(caData); !ok {
		return "", fmt.Errorf("failed to parse CA certificate")
	}

	tlsCfg := &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}

	if err := dmysql.RegisterTLSConfig(name, tlsCfg); err != nil {
		if !strings.Contains(err.Error(), "already registered") {
			return "", fmt.Errorf("register tls config failed: %w", err)
		}
	}

	tlsConfigRegistry[name] = tlsCfg
	return name, nil
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
			timeoutSec = 60
		}
		port := p.Port
		if port <= 0 {
			port = 3306
		}
		params := []string{
			fmt.Sprintf("charset=%s", charset),
			"parseTime=True",
			fmt.Sprintf("loc=%s", url.QueryEscape(timezone)),
			fmt.Sprintf("timeout=%ds", timeoutSec),
		}
		if IsSSL(p.SSLMode) {
			trimmedCA := strings.TrimSpace(p.SSLCAFile)
			if trimmedCA == "" {
				params = append(params, "tls=true")
			} else {
				caData := ""
				if strings.Contains(trimmedCA, "-----BEGIN") {
					caData = trimmedCA
				} else {
					data, err := os.ReadFile(trimmedCA)
					if err != nil {
						return "", "", fmt.Errorf("read CA file failed: %w", err)
					}
					caData = string(data)
				}
				tlsName, err := registerTLSConfig(caData)
				if err != nil {
					return "", "", fmt.Errorf("register tls config failed: %w", err)
				}
				params = append(params, "tls="+tlsName)
			}
		}
		escapedUser := url.QueryEscape(p.Username)
		escapedPwd := url.QueryEscape(p.Password)
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			escapedUser, escapedPwd, p.Host, port, p.Database, strings.Join(params, "&"))
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
		if mode == "ro" {
			return path + "?_pragma=journal_mode(WAL)&mode=ro", TypeSQLite, nil
		}
		return path + "?_pragma=journal_mode(WAL)", TypeSQLite, nil
	default:
		return "", "", fmt.Errorf("unsupported database type: %s", p.Type)
	}
}

// BuildDSN 对外暴露的 DSN 构造函数，便于测试/复现连接字符串构造
func BuildDSN(p *ConnectionParams) (string, string, error) {
	return buildDSN(p)
}

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
		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
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

func TestConnect(p *ConnectionParams) *TestResult {
	result := &TestResult{Success: false, Message: "", LatencyMs: 0, Version: ""}
	start := time.Now()
	dbType := strings.ToLower(p.Type)

	instanceParams := &ConnectionParams{
		Type:      p.Type,
		Host:      p.Host,
		Port:      p.Port,
		Username:  p.Username,
		Password:  p.Password,
		Database:  "",
		FilePath:  p.FilePath,
		OpenMode:  p.OpenMode,
		Charset:   p.Charset,
		Timezone:  p.Timezone,
		SSLMode:   p.SSLMode,
		SSLCAFile: p.SSLCAFile,
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

	var version string
	switch dbType {
	case TypeMySQL, TypeTiDB:
		row := db.QueryRow("SELECT VERSION()")
		_ = row.Scan(&version)
	case TypeSQLite:
		row := db.QueryRow("SELECT sqlite_version()")
		_ = row.Scan(&version)
	}

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

func BeginTransaction(key string) error {
	connLock.RLock()
	ci, ok := connMap[key]
	connLock.RUnlock()
	if !ok {
		return fmt.Errorf("connection not found: %s", key)
	}

	tx, err := ci.DB.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	txLock.Lock()
	txMap[key] = tx
	txLock.Unlock()
	return nil
}

func CommitTransaction(key string) error {
	txLock.Lock()
	tx, ok := txMap[key]
	delete(txMap, key)
	txLock.Unlock()
	if !ok {
		return fmt.Errorf("transaction not found: %s", key)
	}
	return tx.Commit()
}

func RollbackTransaction(key string) error {
	txLock.Lock()
	tx, ok := txMap[key]
	delete(txMap, key)
	txLock.Unlock()
	if !ok {
		return fmt.Errorf("transaction not found: %s", key)
	}
	return tx.Rollback()
}

func GetTransaction(key string) *sql.Tx {
	txLock.RLock()
	defer txLock.RUnlock()
	return txMap[key]
}

func DefaultPort(dbType string) int {
	switch strings.ToLower(dbType) {
	case TypeMySQL:
		return 3306
	case TypeTiDB:
		return 4000
	default:
		return 0
	}
}

func SystemDatabases(dbType string) []string {
	switch strings.ToLower(dbType) {
	case TypeMySQL:
		return []string{"information_schema", "mysql", "performance_schema", "sys"}
	case TypeTiDB:
		return []string{"information_schema", "mysql", "performance_schema", "sys", "metrics_schema"}
	default:
		return nil
	}
}

func IsSystemDatabase(dbType, dbName string) bool {
	systems := SystemDatabases(dbType)
	for _, s := range systems {
		if strings.EqualFold(s, dbName) {
			return true
		}
	}
	return false
}

func SupportFeature(dbType, feature string) bool {
	t := strings.ToLower(dbType)
	switch strings.ToLower(feature) {
	case "procedure", "trigger", "delimiter", "repair", "fk", "fulltext", "spatial":
		return t == TypeMySQL
	case "analyze", "optimize":
		return t == TypeMySQL || t == TypeTiDB
	case "table-info":
		return true
	default:
		return true
	}
}
