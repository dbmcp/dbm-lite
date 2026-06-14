/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/sqllint"

	"github.com/google/uuid"
)

type SQLService struct{}

func NewSQLService() *SQLService { return &SQLService{} }

type ExecResult struct {
	Columns      []string                 `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	AffectedRows int64                    `json:"affectedRows"`
	IsSelect     bool                     `json:"isSelect"`
	DurationMs   int64                    `json:"durationMs"`
	Message      string                   `json:"message"`
	Success      bool                     `json:"success"`
	Review       *sqllint.ReviewResult    `json:"review"`
}

func (s *SQLService) executeStatements(ds *model.Datasource, dbName string, sqls []string, ignoreRisk bool, userId, username string) ([]*ExecResult, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}

	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	combined := strings.Join(sqls, ";\n")
	review := sqllint.Review(combined)
	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite

	// SQLite 高危SQL额外检查: 无 WHERE 的 DELETE/UPDATE, DROP, TRUNCATE
	if isSQLite {
		upper := strings.ToUpper(combined)
		if strings.Contains(upper, "DELETE") || strings.Contains(upper, "UPDATE") || strings.Contains(upper, "DROP") || strings.Contains(upper, "TRUNCATE") {
			if !strings.Contains(upper, "WHERE") && !ignoreRisk {
				review.IsHighRisk = true
				review.Reasons = append(review.Reasons, "SQLite 写入/删除/修改操作需二次确认")
			}
		}
	}

	if review.IsHighRisk && !ignoreRisk {
		return []*ExecResult{{
			Success: false,
			Message: "高危SQL需要二次确认: " + strings.Join(review.Reasons, "; "),
			Review:  review,
		}}, nil
	}

	// 过滤非空语句
	cleanSqls := make([]string, 0, len(sqls))
	for _, s2 := range sqls {
		s2 = strings.TrimSpace(s2)
		if s2 != "" {
			cleanSqls = append(cleanSqls, s2)
		}
	}
	if len(cleanSqls) == 0 {
		return []*ExecResult{{Success: true, Message: "没有可执行的SQL", Review: review}}, nil
	}

	// 循环执行所有语句
	results := make([]*ExecResult, 0, len(cleanSqls))
	for _, execSQL := range cleanSqls {
		start := time.Now()
		stmtReview := sqllint.Review(execSQL)
		result := &ExecResult{
			IsSelect: stmtReview.IsSelect,
			Review:   stmtReview,
			Success:  true,
		}

		if stmtReview.IsSelect {
			rows, err := conn.DB.Query(execSQL)
			if err != nil {
				results = append(results, &ExecResult{
					Success:  false,
					Message:  fmt.Sprintf("执行失败: %v", err),
					IsSelect: true,
					Review:   stmtReview,
				})
				continue
			}

			cols, err := rows.Columns()
			if err != nil {
				rows.Close()
				results = append(results, &ExecResult{Success: false, Message: err.Error(), IsSelect: true, Review: stmtReview})
				continue
			}
			result.Columns = cols
			result.Rows = []map[string]interface{}{}

			limit := 1000
			count := 0
			for rows.Next() && count < limit {
				colVals := make([]interface{}, len(cols))
				colPtrs := make([]interface{}, len(cols))
				for i := range colVals {
					colPtrs[i] = &colVals[i]
				}
				if err := rows.Scan(colPtrs...); err != nil {
					break
				}
				row := make(map[string]interface{})
				for i, col := range cols {
					v := colVals[i]
					if b, ok := v.([]byte); ok {
						row[col] = string(b)
					} else {
						row[col] = v
					}
				}
				result.Rows = append(result.Rows, row)
				count++
			}
			rows.Close()
		} else {
			res, err := conn.DB.Exec(execSQL)
			if err != nil {
				results = append(results, &ExecResult{
					Success: false,
					Message: fmt.Sprintf("执行失败: %v", err),
					Review:  stmtReview,
				})
				continue
			}
			affected, _ := res.RowsAffected()
			result.AffectedRows = affected
			if isSQLite {
				result.Message = fmt.Sprintf("SQLite 执行成功，影响 %d 行", affected)
			} else {
				result.Message = fmt.Sprintf("执行成功，影响 %d 行", affected)
			}
		}

		result.DurationMs = time.Since(start).Milliseconds()

		// 记录SQL历史
		history := &model.SQLHistory{
			HistoryID:    uuid.New().String(),
			UserID:       userId,
			Username:     username,
			DatasourceID: ds.DatasourceID,
			DatabaseName: dbName,
			SQL:          execSQL,
			RowsAffected: result.AffectedRows,
			DurationMs:   result.DurationMs,
			IsHighRisk:   stmtReview.IsHighRisk,
			Status:       model.SQLStatusSuccess,
			CreatedAt:    time.Now(),
		}
		database.DB.Create(history)

		results = append(results, result)
	}

	return results, nil
}

func (s *SQLService) Execute(ds *model.Datasource, dbName, sql string, ignoreRisk bool, userId, username string) ([]*ExecResult, error) {
	stmts := splitSQL(sql)
	if len(stmts) == 0 {
		return []*ExecResult{{Success: true, Message: "SQL为空"}}, nil
	}
	return s.executeStatements(ds, dbName, stmts, ignoreRisk, userId, username)
}

func splitSQL(sql string) []string {
	parts := strings.Split(sql, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func (s *SQLService) GetDatabases(ds *model.Datasource) ([]string, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: ds.DefaultDB,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		rows, err := conn.DB.Query("SHOW DATABASES")
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		dbs := []string{}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				continue
			}
			if name == "information_schema" || name == "mysql" || name == "performance_schema" || name == "sys" {
				continue
			}
			dbs = append(dbs, name)
		}
		return dbs, nil
	case dbtype.TypeSQLite:
		// 通过 PRAGMA database_list 获取当前所有库（main/temp/附加库）
		rows, err := conn.DB.Query("PRAGMA database_list")
		if err != nil {
			return []string{"main"}, nil
		}
		defer rows.Close()
		dbs := []string{}
		for rows.Next() {
			var seq int
			var name, file string
			if err := rows.Scan(&seq, &name, &file); err != nil {
				continue
			}
			dbs = append(dbs, name)
		}
		if len(dbs) == 0 {
			dbs = []string{"main"}
		}
		return dbs, nil
	default:
		return []string{ds.DefaultDB}, nil
	}
}

func (s *SQLService) GetTables(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}

	var query string
	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		query = fmt.Sprintf("SELECT TABLE_NAME as name, TABLE_TYPE as type, TABLE_ROWS as rows, ROUND((DATA_LENGTH + INDEX_LENGTH)/1024/1024, 2) as size_mb FROM information_schema.TABLES WHERE TABLE_SCHEMA = '%s' ORDER BY TABLE_TYPE, TABLE_NAME", dbName)
	case dbtype.TypeSQLite:
		// SQLite: 从指定 schema 中查询所有对象类型
		schema := dbName
		if schema == "" {
			schema = "main"
		}
		return s.getSQLiteObjects(conn.DB, schema)
	default:
		return []map[string]interface{}{}, nil
	}
	if query == "" {
		return []map[string]interface{}{}, nil
	}

	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []map[string]interface{}{}
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		m := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				m[c] = string(b)
			} else {
				m[c] = vals[i]
			}
		}
		name, _ := m["name"].(string)
		if name == "" {
			name, _ = m["TABLE_NAME"].(string)
		}
		// 推断类型：MySQL/TiDB 的 TABLE_TYPE 是 "BASE TABLE" 或 "VIEW"；SQLite 来自 sqlite_master
		objType := "table"
		if typeVal, ok := m["type"].(string); ok && typeVal != "" {
			upperType := strings.ToUpper(typeVal)
			if upperType == "VIEW" {
				objType = "view"
			} else if upperType == "INDEX" {
				objType = "index"
			} else if upperType == "TRIGGER" {
				objType = "trigger"
			} else {
				objType = "table"
			}
		}
		t := map[string]interface{}{"name": name, "type": objType}
		if v, ok := m["rows"]; ok {
			t["rows"] = v
		}
		if v, ok := m["size_mb"]; ok {
			t["sizeMb"] = v
		}
		// 保留原始 tbl_name（SQLite 中 index/trigger 指向所属表）
		if v, ok := m["tbl_name"]; ok {
			t["tblName"] = v
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// getSQLiteObjects 返回指定 schema 下的 tables/views/indexes/triggers 列表
func (s *SQLService) getSQLiteObjects(conn *sql.DB, schema string) ([]map[string]interface{}, error) {
	// 查询 schema.sqlite_master 表
	query := fmt.Sprintf("SELECT type, name, tbl_name, rootpage, sql FROM %s.sqlite_master WHERE type IN ('table','view','index','trigger') ORDER BY type, name", schema)
	rows, err := conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []map[string]interface{}{}
	for rows.Next() {
		var objType, name, tblName sql.NullString
		var rootpage sql.NullInt64
		var sqlStmt sql.NullString
		if err := rows.Scan(&objType, &name, &tblName, &rootpage, &sqlStmt); err != nil {
			continue
		}
		// 过滤 sqlite_ 等特殊表
		if strings.HasPrefix(name.String, "sqlite_") && objType.String == "table" {
			continue
		}
		item := map[string]interface{}{
			"name":    name.String,
			"type":    objType.String, // table/view/index/trigger
			"tblName": tblName.String, // 对于 index/trigger，指向所属表
			"schema":  schema,
		}
		// 仅对表对象查询行数（sqlite_master 不维护行数）
		if objType.String == "table" {
			countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schema, name.String)
			var cnt int64
			if err := conn.QueryRow(countQuery).Scan(&cnt); err == nil {
				item["rows"] = cnt
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *SQLService) GetColumns(ds *model.Datasource, dbName, tableName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}

	var query string
	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		query = fmt.Sprintf("SELECT COLUMN_NAME as name, DATA_TYPE as type, COLUMN_KEY as key_info, IS_NULLABLE as nullable, COLUMN_DEFAULT as default_val, COLUMN_COMMENT as comment FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='%s' AND TABLE_NAME='%s'", dbName, tableName)
	case dbtype.TypeSQLite:
		// 使用 schema.table 形式查询指定库的表字段
		schema := dbName
		if schema == "" {
			schema = "main"
		}
		query = fmt.Sprintf("PRAGMA %s.table_info(%s)", schema, tableName)
	default:
		return []map[string]interface{}{}, nil
	}

	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := []map[string]interface{}{}
	for rows.Next() {
		// cid, name, type, notnull, dflt_value, pk
		var cid, notnull, pk sql.NullInt64
		var name, colType, dflt sql.NullString
		err := rows.Scan(&cid, &name, &colType, &notnull, &dflt, &pk)
		if err != nil {
			// 兼容 MySQL 结果
			var n, ct, ki, nu, dv, cm interface{}
			if err := rows.Scan(&n, &ct, &ki, &nu, &dv, &cm); err != nil {
				continue
			}
			c := map[string]interface{}{
				"name":     stringify(n),
				"type":     stringify(ct),
				"key":      stringify(ki),
				"nullable": stringify(nu),
				"default":  stringify(dv),
				"comment":  stringify(cm),
			}
			cols = append(cols, c)
		} else {
			isNullable := "YES"
			if notnull.Int64 == 1 {
				isNullable = "NO"
			}
			isKey := ""
			if pk.Int64 == 1 {
				isKey = "PK"
			}
			c := map[string]interface{}{
				"name":     name.String,
				"type":     colType.String,
				"key":      isKey,
				"nullable": isNullable,
				"default":  dflt.String,
				"comment":  "",
				"cid":      cid.Int64,
			}
			cols = append(cols, c)
		}
	}
	return cols, nil
}

func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func (s *SQLService) GetHistory(page, pageSize int, datasourceId, keyword string) ([]model.SQLHistory, int64, error) {
	var list []model.SQLHistory
	var total int64
	q := database.DB.Model(&model.SQLHistory{})
	if datasourceId != "" {
		q = q.Where("datasource_id = ?", datasourceId)
	}
	if keyword != "" {
		q = q.Where("sql LIKE ?", "%"+keyword+"%")
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *SQLService) AnalyzeCapacity(ds *model.Datasource, dbName string) (interface{}, error) {
	tables, err := s.GetTables(ds, dbName)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"database":   dbName,
		"tables":     tables,
		"tableCount": len(tables),
	}
	return result, nil
}

func (s *SQLService) ReviewSQL(sql string) *sqllint.ReviewResult {
	return sqllint.Review(sql)
}

// 获取完整树结构: 数据库 -> 表/视图/索引/触发器 -> 字段
func (s *SQLService) GetFullTree(ds *model.Datasource) ([]map[string]interface{}, error) {
	dbs, err := s.GetDatabases(ds)
	if err != nil {
		return nil, err
	}

	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite

	tree := make([]map[string]interface{}, 0, len(dbs))
	for _, db := range dbs {
		dbNode := map[string]interface{}{
			"name":     db,
			"database": db,
			"children": []interface{}{},
		}
		objects, _ := s.GetTables(ds, db)

		if isSQLite {
			// SQLite: 按类型分组 tables/views/indexes/triggers
			typeGroups := map[string][]interface{}{
				"table":   {},
				"view":    {},
				"index":   {},
				"trigger": {},
			}
			for _, obj := range objects {
				t, _ := obj["type"].(string)
				name, _ := obj["name"].(string)
				child := map[string]interface{}{
					"name":     name,
					"type":     t,
					"table":    name,
					"database": db,
					"tblName":  obj["tblName"],
					"rows":     obj["rows"],
					"sizeMb":   obj["sizeMb"],
				}
				if list, ok := typeGroups[t]; ok {
					typeGroups[t] = append(list, child)
				} else {
					typeGroups["table"] = append(typeGroups["table"], child)
				}
			}
			// 固定顺序: tables, views, indexes, triggers
			groupOrder := []string{"table", "view", "index", "trigger"}
			groupLabels := map[string]string{
				"table":   "数据表",
				"view":    "视图",
				"index":   "索引",
				"trigger": "触发器",
			}
			dbChildren := []interface{}{}
			for _, g := range groupOrder {
				list := typeGroups[g]
				if len(list) == 0 {
					continue
				}
				dbChildren = append(dbChildren, map[string]interface{}{
					"name":     groupLabels[g],
					"group":    g,
					"database": db,
					"children": list,
				})
			}
			dbNode["children"] = dbChildren
		} else {
			// MySQL / TiDB：按类型分组 tables / views
			typeGroups := map[string][]interface{}{
				"table": {},
				"view":  {},
			}
			for _, obj := range objects {
				t, _ := obj["type"].(string)
				name, _ := obj["name"].(string)
				if t == "" {
					t = "table"
				}
				child := map[string]interface{}{
					"name":     name,
					"type":     t,
					"table":    name,
					"database": db,
					"rows":     obj["rows"],
					"sizeMb":   obj["sizeMb"],
				}
				if list, ok := typeGroups[t]; ok {
					typeGroups[t] = append(list, child)
				} else {
					typeGroups["table"] = append(typeGroups["table"], child)
				}
			}
			groupOrder := []string{"table", "view"}
			groupLabels := map[string]string{
				"table": "数据表",
				"view":  "视图",
			}
			dbChildren := []interface{}{}
			for _, g := range groupOrder {
				list := typeGroups[g]
				if len(list) == 0 {
					continue
				}
				dbChildren = append(dbChildren, map[string]interface{}{
					"name":     groupLabels[g],
					"group":    g,
					"database": db,
					"children": list,
				})
			}
			dbNode["children"] = dbChildren
		}
		tree = append(tree, dbNode)
	}
	return tree, nil
}

// 测试连接 - 返回完整结果（延迟/版本）
func (s *SQLService) TestConnection(ds *model.Datasource) *dbtype.TestResult {
	if strings.ToLower(ds.DBType) == dbtype.TypeSQLite {
		if err := dbtype.ValidateSQLiteFile(ds.FilePath); err != nil {
			return &dbtype.TestResult{Success: false, Message: err.Error()}
		}
	}
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: ds.DefaultDB,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
		Charset:  ds.Charset,
		Timezone: ds.Timezone,
		SSLMode:  ds.SSLMode,
	}
	return dbtype.TestConnect(params)
}

// 测试连接简洁版
func (s *SQLService) TestConnectionSimple(ds *model.Datasource) (bool, string) {
	result := s.TestConnection(ds)
	return result.Success, result.Message
}

// 直接测试连接（不需要已有数据源记录）
func (s *SQLService) TestConnectionDirect(dbType, host string, port int, username, password, database string, filePath, openMode, charset, timezone, sslMode string) *dbtype.TestResult {
	if strings.ToLower(dbType) == dbtype.TypeSQLite {
		if filePath != "" && filePath != ":memory:" {
			if err := dbtype.ValidateSQLiteFile(filePath); err != nil {
				return &dbtype.TestResult{Success: false, Message: err.Error()}
			}
		}
	}
	params := &dbtype.ConnectionParams{
		Type:     dbType,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
		FilePath: filePath,
		OpenMode: openMode,
		Charset:  charset,
		Timezone: timezone,
		SSLMode:  sslMode,
	}
	return dbtype.TestConnect(params)
}

// 直接测试连接简洁版
func (s *SQLService) TestConnectionDirectSimple(dbType, host string, port int, username, password, database string, filePath, openMode string) (bool, string) {
	result := s.TestConnectionDirect(dbType, host, port, username, password, database, filePath, openMode, "", "", "")
	return result.Success, result.Message
}

// Explain 执行计划
func (s *SQLService) Explain(ds *model.Datasource, dbName, sql string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite
	var query string
	if isSQLite {
		query = "EXPLAIN QUERY PLAN " + sql
	} else {
		query = "EXPLAIN " + sql
	}
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	result := []map[string]interface{}{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				row[c] = string(b)
			} else {
				row[c] = vals[i]
			}
		}
		result = append(result, row)
	}
	return result, nil
}

// GetTableInfo 返回表的基础信息（基本信息+列+索引+DDL）
func (s *SQLService) GetTableInfo(ds *model.Datasource, dbName, tableName string) (map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite
	result := map[string]interface{}{
		"database": dbName,
		"table":    tableName,
	}

	// 1. 基本信息（行数/大小）
	if isSQLite {
		schema := dbName
		if schema == "" {
			schema = "main"
		}
		if rows, err := conn.DB.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schema, tableName)); err == nil {
			var cnt int64
			if rows.Next() {
				rows.Scan(&cnt)
			}
			rows.Close()
			result["rows"] = cnt
		}
	} else {
		if rows, err := conn.DB.Query(fmt.Sprintf(
			"SELECT TABLE_ROWS, ROUND((DATA_LENGTH+INDEX_LENGTH)/1024/1024,2) as size_mb, TABLE_COMMENT, ENGINE "+
				"FROM information_schema.TABLES WHERE TABLE_SCHEMA='%s' AND TABLE_NAME='%s'", dbName, tableName)); err == nil {
			if rows.Next() {
				var r sql.NullInt64
				var size sql.NullFloat64
				var comment, engine sql.NullString
				rows.Scan(&r, &size, &comment, &engine)
				result["rows"] = r.Int64
				result["sizeMb"] = size.Float64
				result["comment"] = comment.String
				result["engine"] = engine.String
			}
			rows.Close()
		}
	}

	// 2. 列信息
	cols, err := s.GetColumns(ds, dbName, tableName)
	if err == nil {
		result["columns"] = cols
	}

	// 3. 索引信息
	idxs, err := s.GetIndexes(ds, dbName, tableName)
	if err == nil {
		result["indexes"] = idxs
	}

	// 4. DDL
	ddl, err := s.GetTableDDL(ds, dbName, tableName)
	if err == nil {
		result["ddl"] = ddl
	}
	return result, nil
}

// GetIndexes 获取表索引信息
func (s *SQLService) GetIndexes(ds *model.Datasource, dbName, tableName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite
	var query string
	if isSQLite {
		schema := dbName
		if schema == "" {
			schema = "main"
		}
		query = fmt.Sprintf("PRAGMA %s.index_list(%s)", schema, tableName)
	} else {
		query = fmt.Sprintf("SHOW INDEX FROM `%s`.`%s`", dbName, tableName)
	}
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	result := []map[string]interface{}{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				row[c] = string(b)
			} else {
				row[c] = vals[i]
			}
		}
		result = append(result, row)
	}
	return result, nil
}

// GetTableDDL 获取表建表语句
func (s *SQLService) GetTableDDL(ds *model.Datasource, dbName, tableName string) (string, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
		FilePath: ds.FilePath,
		OpenMode: ds.OpenMode,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return "", err
	}
	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite
	if isSQLite {
		schema := dbName
		if schema == "" {
			schema = "main"
		}
		rows, err := conn.DB.Query(fmt.Sprintf("SELECT sql FROM %s.sqlite_master WHERE type='table' AND name='%s'", schema, tableName))
		if err != nil {
			return "", err
		}
		defer rows.Close()
		if rows.Next() {
			var ddl sql.NullString
			rows.Scan(&ddl)
			return ddl.String, nil
		}
		return "", nil
	}
	rows, err := conn.DB.Query(fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", dbName, tableName))
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if rows.Next() {
		vals := make([]interface{}, 2)
		ptrs := make([]interface{}, 2)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err == nil {
			if b, ok := vals[1].([]byte); ok {
				return string(b), nil
			}
			if s, ok := vals[1].(string); ok {
				return s, nil
			}
		}
	}
	return "", nil
}

// 简单的JSON辅助
func (s *SQLService) toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
