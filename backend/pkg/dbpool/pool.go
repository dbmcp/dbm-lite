/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package dbpool

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"

	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"
)

// ConnPool 全局连接池管理器
type ConnPool struct {
	pool   map[string]*sql.DB
	mu     sync.RWMutex
}

var (
	instance *ConnPool
	once     sync.Once
)

// GetPool 返回单例连接池
func GetPool() *ConnPool {
	once.Do(func() {
		instance = &ConnPool{
			pool: make(map[string]*sql.DB),
		}
	})
	return instance
}

// buildDSN 根据数据库类型元信息构建 DSN，返回 driver 名和 DSN 字符串
func buildDSN(ds *model.Datasource, password string) (string, string) {
	dbTypeKey := strings.ToLower(ds.DBType)

	switch dbTypeKey {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		charset := "utf8mb4"
		loc := "Local"
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			ds.Username, password, ds.Host, ds.Port,
			ds.DefaultDB, charset, url.QueryEscape(loc),
		)
		return "mysql", dsn

	case dbtype.TypeSQLite:
		dbFile := ds.DefaultDB
		if dbFile == "" {
			dbFile = ds.Host
		}
		if dbFile == "" {
			// 内存库模式
			return "sqlite", ":memory:"
		}
		return "sqlite", dbFile

	default:
		return "", ""
	}
}

// Get 获取或创建数据库连接
func (p *ConnPool) Get(ds *model.Datasource, password string) (*sql.DB, error) {
	p.mu.RLock()
	db, exists := p.pool[ds.DatasourceID]
	p.mu.RUnlock()

	if exists {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		db.Close()
		p.mu.Lock()
		delete(p.pool, ds.DatasourceID)
		p.mu.Unlock()
	}

	driver, dsn := buildDSN(ds, password)
	if dsn == "" {
		return nil, fmt.Errorf("unsupported db type: %s", ds.DBType)
	}

	newDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open connection error: %w", err)
	}

	// 连接池参数（保守设置，避免连接数过多）
	newDB.SetMaxOpenConns(10)
	newDB.SetMaxIdleConns(3)
	newDB.SetConnMaxLifetime(30 * time.Minute)
	newDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := newDB.Ping(); err != nil {
		newDB.Close()
		return nil, fmt.Errorf("ping error: %w", err)
	}

	p.mu.Lock()
	p.pool[ds.DatasourceID] = newDB
	p.mu.Unlock()

	return newDB, nil
}

// Remove 从连接池移除指定数据源的连接
func (p *ConnPool) Remove(dsID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if db, ok := p.pool[dsID]; ok {
		db.Close()
		delete(p.pool, dsID)
	}
}

// TestConnection 测试数据库连接（不放入连接池）
func (p *ConnPool) TestConnection(ds *model.Datasource, password string) error {
	driver, dsn := buildDSN(ds, password)
	if dsn == "" {
		return fmt.Errorf("unsupported db type: %s", ds.DBType)
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetConnMaxLifetime(5 * time.Second)
	return db.Ping()
}

