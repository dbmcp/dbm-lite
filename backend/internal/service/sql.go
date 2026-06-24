/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
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
	Extra        map[string]interface{}   `json:"extra,omitempty"`
	SQL          string                   `json:"sql"` // 执行的SQL语句
}

func (s *SQLService) executeStatements(ctx context.Context, ds *model.Datasource, dbName string, sqls []string, ignoreRisk bool, userId, username string) (results []*ExecResult, err error) {
	// 全局 panic 恢复，防止任何底层连接驱动或数据库操作导致整个服务崩溃
	defer func() {
		if r := recover(); r != nil {
			results = []*ExecResult{{
				Success: false,
				Message: fmt.Sprintf("SQL执行发生未预期异常: %v", r),
			}}
			err = fmt.Errorf("recover from panic: %v", r)
		}
	}()

	params := &dbtype.ConnectionParams{
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  dbName,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
	}

	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return []*ExecResult{{
			Success: false,
			Message: fmt.Sprintf("连接失败: %v", err),
		}}, nil
	}

	combined := strings.Join(sqls, ";\n")
	review := sqllint.Review(combined)
	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite

	// 显式切换数据库（MySQL/TiDB 需要）
	if !isSQLite && dbName != "" {
		_, _ = conn.DB.Exec(fmt.Sprintf("USE `%s`", dbName))
	}

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

	// isQueryLike 通过关键字判断是否为可能返回结果集的语句（兜底识别）
	isQueryLike := func(sql string) bool {
		upper := strings.ToUpper(strings.TrimSpace(sql))
		prefixes := []string{"SELECT", "SHOW", "DESCRIBE", "DESC", "EXPLAIN", "ANALYZE", "WITH", "PRAGMA", "CHECK TABLE"}
		for _, p := range prefixes {
			if strings.HasPrefix(upper, p) {
				return true
			}
		}
		return false
	}

	// 循环执行所有语句
	results = make([]*ExecResult, 0, len(cleanSqls))
	for _, execSQL := range cleanSqls {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		start := time.Now()
		stmtReview := sqllint.Review(execSQL)
		// 关键字兜底：如果 Review 没有识别，但语句看起来会返回结果集，则走 Query 路径
		useQuery := stmtReview.IsSelect || isQueryLike(execSQL)
		result := &ExecResult{
			IsSelect: useQuery,
			Review:   stmtReview,
			Success:  true,
			SQL:      execSQL, // 保存当前执行的SQL语句
		}

		if useQuery {
			rows, err := conn.DB.QueryContext(ctx, execSQL)
			if err != nil {
				if ctx.Err() != nil {
					return results, ctx.Err()
				}
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
				select {
				case <-ctx.Done():
					rows.Close()
					return results, ctx.Err()
				default:
				}

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
			res, err := conn.DB.ExecContext(ctx, execSQL)
			if err != nil {
				if ctx.Err() != nil {
					return results, ctx.Err()
				}
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
			HistoryID:      uuid.New().String(),
			UserID:         userId,
			Username:       username,
			DatasourceID:   ds.DatasourceID,
			DatasourceName: ds.Name,
			DatabaseName:   dbName,
			SqlText:        execSQL,
			RowsAffected:   result.AffectedRows,
			DurationMs:     result.DurationMs,
			IsHighRisk:     stmtReview.IsHighRisk,
			Status:         model.SQLStatusSuccess,
			CreatedAt:      time.Now(),
		}
		if err := database.DB.Create(history).Error; err != nil {
			// 历史记录写入失败不影响 SQL 执行结果，静默忽略
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *SQLService) Execute(ds *model.Datasource, dbName, sql string, ignoreRisk bool, userId, username string) ([]*ExecResult, error) {
	stmts := splitSQL(sql)
	if len(stmts) == 0 {
		return []*ExecResult{{Success: true, Message: "SQL为空"}}, nil
	}
	return s.executeStatements(context.Background(), ds, dbName, stmts, ignoreRisk, userId, username)
}

func (s *SQLService) ExecuteWithCancel(ds *model.Datasource, dbName, sql string, ignoreRisk bool, userId, username string) (ExecutionID, []*ExecResult, error) {
	stmts := splitSQL(sql)
	if len(stmts) == 0 {
		return "", []*ExecResult{{Success: true, Message: "SQL为空"}}, nil
	}

	ec := NewExecutionContext(ds.DatasourceID, sql)
	defer ec.Done()

	results, err := s.executeStatements(ec.Ctx, ds, dbName, stmts, ignoreRisk, userId, username)
	if err != nil {
		if ec.Ctx.Err() == context.Canceled {
			return ec.ID, []*ExecResult{{Success: false, Message: "SQL执行已被取消"}}, nil
		}
		return ec.ID, results, err
	}

	return ec.ID, results, nil
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
	return s.GetDatabasesWithSystem(ds, true)
}

// GetDatabasesWithSystem 支持 includeSystem 开关控制系统库是否暴露
func (s *SQLService) GetDatabasesWithSystem(ds *model.Datasource, includeSystem bool) ([]string, error) {
	params := &dbtype.ConnectionParams{
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  ds.DefaultDB,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		var rows *sql.Rows
		var err error

		// 优先使用 SHOW DATABASES 获取所有数据库（包括特殊数据库如 __cdb_recycle_bin__）
		rows, err = conn.DB.Query("SHOW DATABASES")
		if err != nil {
			// 权限不足时尝试多种查询方式
			// 方式1: 尝试 INFORMATION_SCHEMA
			rows, err = conn.DB.Query("SELECT SCHEMA_NAME FROM information_schema.SCHEMATA")
			if err != nil {
				// 方式2: 尝试获取所有可见数据库（适用于云数据库）
				rows, err = conn.DB.Query("SELECT DISTINCT table_schema FROM information_schema.tables")
				if err != nil {
					return nil, err
				}
			}
		}
		defer rows.Close()
		dbs := []string{}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				continue
			}
			if !includeSystem && dbtype.IsSystemDatabase(ds.DBType, name) {
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
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  dbName,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}

	var query string
	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		query = fmt.Sprintf("SELECT TABLE_NAME as `name`, TABLE_TYPE as `type`, TABLE_ROWS as `rows`, ROUND((IFNULL(DATA_LENGTH, 0) + IFNULL(INDEX_LENGTH, 0))/1024/1024, 2) as `size_mb` FROM information_schema.TABLES WHERE TABLE_SCHEMA = '%s' ORDER BY TABLE_TYPE, TABLE_NAME", dbName)
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
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
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
			if upperType == "VIEW" || strings.Contains(upperType, "VIEW") {
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

	type rawItem struct {
		objType  string
		name     string
		tblName  string
		rootpage int64
		sqlStmt  string
	}
	rawItems := []rawItem{}
	for rows.Next() {
		var objType, name, tblName sql.NullString
		var rootpage sql.NullInt64
		var sqlStmt sql.NullString
		if err := rows.Scan(&objType, &name, &tblName, &rootpage, &sqlStmt); err != nil {
			continue
		}
		if strings.HasPrefix(name.String, "sqlite_") && objType.String == "table" {
			continue
		}
		rawItems = append(rawItems, rawItem{
			objType:  objType.String,
			name:     name.String,
			tblName:  tblName.String,
			rootpage: rootpage.Int64,
			sqlStmt:  sqlStmt.String,
		})
	}
	rows.Close()

	tableStats := make(map[string]int64)
	statsQuery := fmt.Sprintf("SELECT tbl, stat FROM %s.sqlite_stat1", schema)
	statsRows, err := conn.Query(statsQuery)
	if err == nil {
		for statsRows.Next() {
			var tbl string
			var stat string
			if err := statsRows.Scan(&tbl, &stat); err == nil {
				parts := strings.Split(stat, " ")
				if len(parts) > 0 {
					if cnt, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
						tableStats[tbl] = cnt
					}
				}
			}
		}
		statsRows.Close()
	}

	tableSizes := make(map[string]float64)
	var pageSize int64 = 4096
	if pageSizeRow := conn.QueryRow("PRAGMA page_size"); pageSizeRow != nil {
		if err := pageSizeRow.Scan(&pageSize); err != nil {
			pageSize = 4096
		}
	}
	dbStatQuery := fmt.Sprintf("SELECT name, SUM(pgsize) as total_size FROM %s.dbstat GROUP BY name", schema)
	dbStatRows, err := conn.Query(dbStatQuery)
	if err == nil {
		for dbStatRows.Next() {
			var name string
			var size int64
			if err := dbStatRows.Scan(&name, &size); err == nil {
				tableSizes[name] = float64(size) / 1024 / 1024
			}
		}
		dbStatRows.Close()
	}

	result := []map[string]interface{}{}
	for _, r := range rawItems {
		item := map[string]interface{}{
			"name":    r.name,
			"type":    r.objType,
			"tblName": r.tblName,
			"schema":  schema,
		}
		if r.objType == "table" {
			var rowCount int64 = 0
			if cnt, ok := tableStats[r.name]; ok && cnt > 0 {
				rowCount = cnt
			} else {
				countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schema, r.name)
				countRow := conn.QueryRow(countQuery)
				if countRow != nil {
					if err := countRow.Scan(&rowCount); err != nil {
						rowCount = 0
					}
				}
			}
			item["rows"] = rowCount

			var sizeMb float64 = 0
			if size, ok := tableSizes[r.name]; ok && size > 0 {
				sizeMb = size
			} else {
				pageCountQuery := fmt.Sprintf("PRAGMA %s.page_count", schema)
				var pageCount int64 = 0
				if pageRow := conn.QueryRow(pageCountQuery); pageRow != nil {
					if err := pageRow.Scan(&pageCount); err == nil && pageCount > 0 {
						sizeMb = float64(pageCount*pageSize) / 1024 / 1024
					}
				}
			}
			item["sizeMb"] = sizeMb
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *SQLService) GetColumns(ds *model.Datasource, dbName, tableName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  dbName,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
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

func (s *SQLService) GetHistory(page, pageSize int, datasourceId, keyword, userId string) ([]model.SQLHistory, int64, error) {
	var list []model.SQLHistory
	var total int64
	q := database.DB.Model(&model.SQLHistory{})
	if userId != "" {
		q = q.Where("user_id = ?", userId)
	}
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

func (s *SQLService) CountHistory(datasourceId, keyword, userId string) (int64, error) {
	var count int64
	q := database.DB.Model(&model.SQLHistory{})
	if userId != "" {
		q = q.Where("user_id = ?", userId)
	}
	if datasourceId != "" {
		q = q.Where("datasource_id = ?", datasourceId)
	}
	if keyword != "" {
		q = q.Where("sql LIKE ?", "%"+keyword+"%")
	}
	err := q.Count(&count).Error
	return count, err
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

// buildTableChildren 为指定表构建子节点（列/索引），用于树形结构展示
func (s *SQLService) buildTableChildren(ds *model.Datasource, dbName, tableName string) []map[string]interface{} {
	var children []map[string]interface{}
	// 列
	if cols, err := s.GetColumns(ds, dbName, tableName); err == nil && len(cols) > 0 {
		colNodes := make([]interface{}, 0, len(cols))
		for _, c := range cols {
			name, _ := c["name"].(string)
			colType, _ := c["type"].(string)
			nullable, _ := c["nullable"].(string)
			keyInfo, _ := c["key"].(string)
			comment, _ := c["comment"].(string)
			isPK := keyInfo == "PK" || strings.ToUpper(keyInfo) == "PRI" || strings.Contains(strings.ToUpper(keyInfo), "PRI")
			colNodes = append(colNodes, map[string]interface{}{
				"name":     name,
				"type":     "column",
				"database": dbName,
				"table":    tableName,
				"colType":  colType,
				"nullable": nullable,
				"keyInfo":  keyInfo,
				"pk":       isPK,
				"comment":  comment,
			})
		}
		children = append(children, map[string]interface{}{
			"name":     "字段",
			"type":     "group",
			"group":    "column",
			"database": dbName,
			"table":    tableName,
			"children": colNodes,
		})
	}
	// 索引
	if idxs, err := s.GetIndexes(ds, dbName, tableName); err == nil && len(idxs) > 0 {
		idxNodes := make([]interface{}, 0, len(idxs))
		seen := map[string]bool{}
		for _, idx := range idxs {
			name := ""
			if v, ok := idx["name"].(string); ok && v != "" {
				name = v
			} else if v, ok := idx["Key_name"].(string); ok && v != "" {
				name = v
			}
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			idxNodes = append(idxNodes, map[string]interface{}{
				"name":     name,
				"type":     "index",
				"database": dbName,
				"table":    tableName,
			})
		}
		if len(idxNodes) > 0 {
			children = append(children, map[string]interface{}{
				"name":     "索引",
				"type":     "group",
				"group":    "index",
				"database": dbName,
				"table":    tableName,
				"children": idxNodes,
			})
		}
	}
	return children
}

// GetTableChildren 返回表的子节点（字段、索引）信息。
func (s *SQLService) GetTableChildren(ds *model.Datasource, dbName, tableName string) ([]map[string]interface{}, error) {
	children := s.buildTableChildren(ds, dbName, tableName)
	return children, nil
}

// GetRoutines 获取存储过程和函数列表（导出）
func (s *SQLService) GetRoutines(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
	return s.getRoutines(ds, dbName)
}

// GetTriggers 获取数据库级触发器列表（导出）
func (s *SQLService) GetTriggers(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
	return s.getTriggers(ds, dbName)
}

// GetForeignKeys 获取指定表的外键列表
func (s *SQLService) GetForeignKeys(ds *model.Datasource, dbName, tableName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  dbName,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	var query string
	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		query = fmt.Sprintf(
			"SELECT CONSTRAINT_NAME as fk_name, TABLE_NAME as table_name, COLUMN_NAME as column_name, REFERENCED_TABLE_NAME as ref_table, REFERENCED_COLUMN_NAME as ref_column FROM information_schema.KEY_COLUMN_USAGE WHERE TABLE_SCHEMA='%s' AND TABLE_NAME='%s' AND REFERENCED_TABLE_NAME IS NOT NULL ORDER BY CONSTRAINT_NAME, ORDINAL_POSITION",
			dbName, tableName,
		)
	default:
		return []map[string]interface{}{}, nil
	}
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []map[string]interface{}{}
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		item := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				item[c] = string(b)
			} else {
				item[c] = vals[i]
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// GetTableTriggers 获取与指定表相关的触发器列表
func (s *SQLService) GetTableTriggers(ds *model.Datasource, dbName, tableName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  dbName,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	var query string
	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		query = fmt.Sprintf(
			"SELECT TRIGGER_NAME as name, EVENT_MANIPULATION as event, ACTION_TIMING as timing, EVENT_OBJECT_TABLE as table_name, ACTION_STATEMENT as statement FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA='%s' AND EVENT_OBJECT_TABLE='%s' ORDER BY TRIGGER_NAME",
			dbName, tableName,
		)
	case dbtype.TypeSQLite:
		query = fmt.Sprintf("SELECT name, tbl_name, sql FROM sqlite_master WHERE type='trigger' AND tbl_name='%s'", tableName)
	default:
		return []map[string]interface{}{}, nil
	}
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []map[string]interface{}{}
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		item := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				item[c] = string(b)
			} else {
				item[c] = vals[i]
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// GetViewList 获取数据库的视图列表
func (s *SQLService) GetViewList(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  dbName,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	var query string
	switch strings.ToLower(ds.DBType) {
	case dbtype.TypeMySQL, dbtype.TypeTiDB:
		query = fmt.Sprintf(
			"SELECT TABLE_NAME as name, TABLE_TYPE as type, TABLE_COMMENT as comment FROM information_schema.TABLES WHERE TABLE_SCHEMA='%s' AND TABLE_TYPE='VIEW' ORDER BY TABLE_NAME",
			dbName,
		)
	case dbtype.TypeSQLite:
		query = "SELECT name, type, tbl_name FROM sqlite_master WHERE type='view' ORDER BY name"
	default:
		return []map[string]interface{}{}, nil
	}
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []map[string]interface{}{}
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		item := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				item[c] = string(b)
			} else {
				item[c] = vals[i]
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// getRoutines 获取指定库下的存储过程/函数列表（MySQL / TiDB 通用）
func (s *SQLService) getRoutines(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(
		"SELECT ROUTINE_NAME as name, ROUTINE_TYPE as type FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA='%s' ORDER BY ROUTINE_TYPE, ROUTINE_NAME",
		dbName)
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []map[string]interface{}{}
	for rows.Next() {
		var nameBytes, typeBytes []byte
		if err := rows.Scan(&nameBytes, &typeBytes); err != nil {
			continue
		}
		nameStr := string(nameBytes)
		typeStr := strings.ToLower(string(typeBytes))
		if typeStr == "procedure" || strings.Contains(typeStr, "procedure") {
			result = append(result, map[string]interface{}{
				"name":         nameStr,
				"type":         "procedure",
				"database":     dbName,
				"datasourceId": ds.DatasourceID,
				"table":        nameStr,
			})
		} else if typeStr == "function" || strings.Contains(typeStr, "function") {
			result = append(result, map[string]interface{}{
				"name":         nameStr,
				"type":         "function",
				"database":     dbName,
				"datasourceId": ds.DatasourceID,
				"table":        nameStr,
			})
		}
	}
	return result, nil
}

// getTriggers 获取指定库下的触发器列表（MySQL / TiDB 通用）
func (s *SQLService) getTriggers(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(
		"SELECT TRIGGER_NAME as name, EVENT_OBJECT_TABLE as tbl_name FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA='%s' ORDER BY TRIGGER_NAME",
		dbName)
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []map[string]interface{}{}
	for rows.Next() {
		var nameBytes, tblBytes []byte
		if err := rows.Scan(&nameBytes, &tblBytes); err != nil {
			continue
		}
		if len(nameBytes) == 0 {
			continue
		}
		nameStr := string(nameBytes)
		tblStr := string(tblBytes)
		result = append(result, map[string]interface{}{
			"name":         nameStr,
			"type":         "trigger",
			"database":     dbName,
			"datasourceId": ds.DatasourceID,
			"table":        tblStr,
			"tblName":      tblStr,
		})
	}
	return result, nil
}

// getEvents 获取指定库下的事件调度器列表（MySQL / TiDB 通用）
func (s *SQLService) getEvents(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(
		"SELECT EVENT_NAME as name, EVENT_TYPE as event_type, EXECUTE_AT as execute_at, INTERVAL_VALUE as iv, INTERVAL_FIELD as iv_field FROM information_schema.EVENTS WHERE EVENT_SCHEMA='%s' ORDER BY EVENT_NAME",
		dbName)
	rows, err := conn.DB.Query(query)
	if err != nil {
		// 某些受限账号或旧版 TiDB 可能没有 EVENTS 表，静默失败返回空
		return []map[string]interface{}{}, nil
	}
	defer rows.Close()
	result := []map[string]interface{}{}
	for rows.Next() {
		var nameBytes, evtTypeBytes, executeAtBytes, ivBytes, ivFieldBytes []byte
		if err := rows.Scan(&nameBytes, &evtTypeBytes, &executeAtBytes, &ivBytes, &ivFieldBytes); err != nil {
			continue
		}
		if len(nameBytes) == 0 {
			continue
		}
		nameStr := string(nameBytes)
		result = append(result, map[string]interface{}{
			"name":         nameStr,
			"type":         "event",
			"database":     dbName,
			"datasourceId": ds.DatasourceID,
			"table":        nameStr,
		})
	}
	return result, nil
}

// buildGroupNode 辅助函数：根据分组类型构造分组节点
func buildGroupNode(groupKey, dbName string, children []interface{}, labels map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"name":     labels[groupKey],
		"type":     "group",
		"group":    groupKey,
		"database": dbName,
		"children": children,
	}
}

// 获取完整树结构: 数据库 -> 分组(表/视图/存储过程/触发器) -> 对象
func (s *SQLService) GetFullTree(ds *model.Datasource) ([]map[string]interface{}, error) {
	dbs, err := s.GetDatabases(ds)
	if err != nil {
		return nil, err
	}

	dbTypeLower := strings.ToLower(ds.DBType)
	isSQLite := dbTypeLower == dbtype.TypeSQLite
	isMySQL := dbTypeLower == dbtype.TypeMySQL || dbTypeLower == dbtype.TypeTiDB

	tree := make([]map[string]interface{}, 0, len(dbs))
	for _, db := range dbs {
		dbNode := map[string]interface{}{
			"name":         db,
			"type":         "database",
			"database":     db,
			"datasourceId": ds.DatasourceID,
			"children":     []interface{}{},
		}
		objects, _ := s.GetTables(ds, db)

		groupLabels := map[string]string{
			"table":     "数据表",
			"view":      "视图",
			"index":     "索引",
			"procedure": "存储过程",
			"function":  "函数",
			"trigger":   "触发器",
			"event":     "事件",
		}

		if isSQLite {
			typeGroups := map[string][]interface{}{
				"table":   {},
				"view":    {},
				"index":   {},
				"trigger": {},
			}
			for _, obj := range objects {
				t, _ := obj["type"].(string)
				name, _ := obj["name"].(string)
				if t == "" {
					t = "table"
				}
				child := map[string]interface{}{
					"name":         name,
					"type":         t,
					"table":        name,
					"database":     db,
					"datasourceId": ds.DatasourceID,
					"tblName":      obj["tblName"],
					"rows":         obj["rows"],
					"sizeMb":       obj["sizeMb"],
				}
				if t == "table" || t == "view" {
					if sub := s.buildTableChildren(ds, db, name); len(sub) > 0 {
						childList := make([]interface{}, 0, len(sub))
						for _, c := range sub {
							childList = append(childList, c)
						}
						child["children"] = childList
					}
				}
				if list, ok := typeGroups[t]; ok {
					typeGroups[t] = append(list, child)
				} else {
					typeGroups["table"] = append(typeGroups["table"], child)
				}
			}
			groupOrder := []string{"table", "view", "index", "trigger"}
			dbChildren := []interface{}{}
			for _, g := range groupOrder {
				list := typeGroups[g]
				if len(list) == 0 {
					continue
				}
				dbChildren = append(dbChildren, buildGroupNode(g, db, list, groupLabels))
			}
			dbNode["children"] = dbChildren
		} else if isMySQL {
			// MySQL / TiDB：表、视图、存储过程、函数、触发器、事件
			typeGroups := map[string][]interface{}{
				"table":     {},
				"view":      {},
				"procedure": {},
				"function":  {},
				"trigger":   {},
				"event":     {},
			}
			for _, obj := range objects {
				t, _ := obj["type"].(string)
				name, _ := obj["name"].(string)
				if t == "" {
					t = "table"
				}
				if t != "table" && t != "view" {
					continue
				}
				child := map[string]interface{}{
					"name":         name,
					"type":         t,
					"table":        name,
					"database":     db,
					"datasourceId": ds.DatasourceID,
					"rows":         obj["rows"],
					"sizeMb":       obj["sizeMb"],
				}
				if sub := s.buildTableChildren(ds, db, name); len(sub) > 0 {
					childList := make([]interface{}, 0, len(sub))
					for _, c := range sub {
						childList = append(childList, c)
					}
					child["children"] = childList
				}
				typeGroups[t] = append(typeGroups[t], child)
			}
			// 独立查询存储过程/函数/触发器
			if routines, err := s.getRoutines(ds, db); err == nil && len(routines) > 0 {
				for _, r := range routines {
					rtype, _ := r["type"].(string)
					if rtype == "function" {
						typeGroups["function"] = append(typeGroups["function"], r)
					} else {
						typeGroups["procedure"] = append(typeGroups["procedure"], r)
					}
				}
			}
			if triggers, err := s.getTriggers(ds, db); err == nil && len(triggers) > 0 {
				for _, tr := range triggers {
					typeGroups["trigger"] = append(typeGroups["trigger"], tr)
				}
			}
			// 查询事件调度器
			if events, err := s.getEvents(ds, db); err == nil && len(events) > 0 {
				for _, ev := range events {
					typeGroups["event"] = append(typeGroups["event"], ev)
				}
			}
			groupOrder := []string{"table", "view", "procedure", "function", "trigger", "event"}
			dbChildren := []interface{}{}
			for _, g := range groupOrder {
				list := typeGroups[g]
				if len(list) == 0 {
					continue
				}
				dbChildren = append(dbChildren, buildGroupNode(g, db, list, groupLabels))
			}
			dbNode["children"] = dbChildren
		} else {
			// 其他类型：退化到 tables 列表
			list := []interface{}{}
			for _, obj := range objects {
				name, _ := obj["name"].(string)
				list = append(list, map[string]interface{}{
					"name":     name,
					"type":     "table",
					"table":    name,
					"database": db,
				})
			}
			dbNode["children"] = list
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
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   ds.Password,
		Database:   ds.DefaultDB,
		FilePath:   ds.FilePath,
		OpenMode:   ds.OpenMode,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
		TimeoutSec: ds.Timeout,
	}
	return dbtype.TestConnect(params)
}

// 测试连接简洁版
func (s *SQLService) TestConnectionSimple(ds *model.Datasource) (bool, string) {
	result := s.TestConnection(ds)
	return result.Success, result.Message
}

// 直接测试连接（不需要已有数据源记录）
func (s *SQLService) TestConnectionDirect(dbType, host string, port int, username, password, database string, filePath, openMode, charset, timezone, sslMode, sslCAFile string, timeoutSec int) *dbtype.TestResult {
	if strings.ToLower(dbType) == dbtype.TypeSQLite {
		if filePath != "" && filePath != ":memory:" {
			if err := dbtype.ValidateSQLiteFile(filePath); err != nil {
				return &dbtype.TestResult{Success: false, Message: err.Error()}
			}
		}
	}
	params := &dbtype.ConnectionParams{
		Type:       dbType,
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		Database:   database,
		FilePath:   filePath,
		OpenMode:   openMode,
		Charset:    charset,
		Timezone:   timezone,
		SSLMode:    sslMode,
		SSLCAFile:  sslCAFile,
		TimeoutSec: timeoutSec,
	}
	return dbtype.TestConnect(params)
}

// 直接测试连接简洁版
func (s *SQLService) TestConnectionDirectSimple(dbType, host string, port int, username, password, database string, filePath, openMode string, timeoutSec int) (bool, string) {
	result := s.TestConnectionDirect(dbType, host, port, username, password, database, filePath, openMode, "", "", "", "", timeoutSec)
	return result.Success, result.Message
}

// Explain 执行计划
func (s *SQLService) Explain(ds *model.Datasource, dbName, sql string) (result []*ExecResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("explain panic: %v", r)
		}
	}()
	params := &dbtype.ConnectionParams{
		Type:      ds.DBType,
		Host:      ds.Host,
		Port:      ds.Port,
		Username:  ds.Username,
		Password:  ds.Password,
		Database:  dbName,
		FilePath:  ds.FilePath,
		OpenMode:  ds.OpenMode,
		Charset:   ds.Charset,
		Timezone:  ds.Timezone,
		SSLMode:   ds.SSLMode,
		SSLCAFile: ds.SSLCAFile,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	isSQLite := strings.ToLower(ds.DBType) == dbtype.TypeSQLite

	sqlStatements := strings.Split(strings.TrimSpace(sql), ";")
	result = make([]*ExecResult, 0, len(sqlStatements))

	for _, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		var query string
		if isSQLite {
			query = "EXPLAIN QUERY PLAN " + stmt
		} else {
			query = "EXPLAIN " + stmt
		}

		rows, err := conn.DB.Query(query)
		if err != nil {
			result = append(result, &ExecResult{
				Success: false,
				Message: err.Error(),
				SQL:     stmt,
			})
			continue
		}

		cols, _ := rows.Columns()
		dataRows := []map[string]interface{}{}
		for rows.Next() {
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if scanErr := rows.Scan(ptrs...); scanErr != nil {
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
			dataRows = append(dataRows, row)
		}
		rows.Close()

		result = append(result, &ExecResult{
			Success:      true,
			Columns:      cols,
			Rows:         dataRows,
			AffectedRows: int64(len(dataRows)),
			IsSelect:     true,
			SQL:          stmt,
		})
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
		rows, err := conn.DB.Query(fmt.Sprintf("SELECT sql FROM %s.sqlite_master WHERE name='%s'", schema, tableName))
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
	// MySQL / TiDB: 依次尝试表/视图/函数/存储过程/触发器/事件
	candidates := []string{
		fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", dbName, tableName),
		fmt.Sprintf("SHOW CREATE VIEW `%s`.`%s`", dbName, tableName),
		fmt.Sprintf("SHOW CREATE FUNCTION `%s`.`%s`", dbName, tableName),
		fmt.Sprintf("SHOW CREATE PROCEDURE `%s`.`%s`", dbName, tableName),
		fmt.Sprintf("SHOW CREATE TRIGGER `%s`.`%s`", dbName, tableName),
		fmt.Sprintf("SHOW CREATE EVENT `%s`.`%s`", dbName, tableName),
	}
	for _, q := range candidates {
		rows, err := conn.DB.Query(q)
		if err != nil {
			continue
		}
		got := func() string {
			defer rows.Close()
			if !rows.Next() {
				return ""
			}
			cols, _ := rows.Columns()
			if len(cols) < 2 {
				return ""
			}
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				return ""
			}
			if b, ok := vals[1].([]byte); ok && b != nil {
				return string(b)
			}
			if s2, ok := vals[1].(string); ok && s2 != "" {
				return s2
			}
			parts := []string{}
			for _, v := range vals {
				if v == nil {
					continue
				}
				if b, ok := v.([]byte); ok {
					parts = append(parts, string(b))
				} else if s2, ok := v.(string); ok {
					parts = append(parts, s2)
				} else {
					parts = append(parts, fmt.Sprintf("%v", v))
				}
			}
			if len(parts) > 1 {
				return strings.Join(parts[1:], "\n")
			}
			return ""
		}()
		if got != "" {
			return got, nil
		}
	}
	return "", nil
}

// 简单的JSON辅助
func (s *SQLService) toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// TableMaintenanceResult 表维护操作统一返回结构
type TableMaintenanceResult struct {
	Success bool        `json:"success"`
	Action  string      `json:"action"`
	Table   string      `json:"table"`
	Rows    int64       `json:"rows,omitempty"`
	Message string      `json:"message,omitempty"`
	Detail  interface{} `json:"detail,omitempty"`
	Ms      int64       `json:"durationMs"`
}

// execMaintenance 执行指定的维护语句
func (s *SQLService) execMaintenance(ds *model.Datasource, dbName, sql, action, table string) *TableMaintenanceResult {
	start := time.Now()
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return &TableMaintenanceResult{Success: false, Action: action, Table: table, Message: err.Error(), Ms: time.Since(start).Milliseconds()}
	}
	_, err = conn.DB.Exec(sql)
	if err != nil {
		return &TableMaintenanceResult{Success: false, Action: action, Table: table, Message: err.Error(), Ms: time.Since(start).Milliseconds()}
	}
	return &TableMaintenanceResult{Success: true, Action: action, Table: table, Message: "完成", Ms: time.Since(start).Milliseconds()}
}

// AnalyzeTable 分析表
func (s *SQLService) AnalyzeTable(ds *model.Datasource, dbName, table string) *TableMaintenanceResult {
	if !dbtype.SupportFeature(ds.DBType, "analyze") {
		return &TableMaintenanceResult{Success: false, Action: "analyze", Table: table, Message: "当前数据库类型不支持 ANALYZE TABLE"}
	}
	return s.execMaintenance(ds, dbName, fmt.Sprintf("ANALYZE TABLE `%s`.`%s`", dbName, table), "analyze", table)
}

// CheckTable 检查表
func (s *SQLService) CheckTable(ds *model.Datasource, dbName, table string) *TableMaintenanceResult {
	return s.execMaintenance(ds, dbName, fmt.Sprintf("CHECK TABLE `%s`.`%s`", dbName, table), "check", table)
}

// OptimizeTable 优化表
func (s *SQLService) OptimizeTable(ds *model.Datasource, dbName, table string) *TableMaintenanceResult {
	if !dbtype.SupportFeature(ds.DBType, "optimize") {
		return &TableMaintenanceResult{Success: false, Action: "optimize", Table: table, Message: "当前数据库类型不支持 OPTIMIZE TABLE"}
	}
	return s.execMaintenance(ds, dbName, fmt.Sprintf("OPTIMIZE TABLE `%s`.`%s`", dbName, table), "optimize", table)
}

// RepairTable 修复表（仅 MySQL）
func (s *SQLService) RepairTable(ds *model.Datasource, dbName, table string) *TableMaintenanceResult {
	if !dbtype.SupportFeature(ds.DBType, "repair") {
		return &TableMaintenanceResult{Success: false, Action: "repair", Table: table, Message: "当前数据库类型不支持 REPAIR TABLE"}
	}
	return s.execMaintenance(ds, dbName, fmt.Sprintf("REPAIR TABLE `%s`.`%s`", dbName, table), "repair", table)
}

// GetTableRowCount 获取行数
func (s *SQLService) GetTableRowCount(ds *model.Datasource, dbName, table string) *TableMaintenanceResult {
	start := time.Now()
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return &TableMaintenanceResult{Success: false, Action: "count", Table: table, Message: err.Error(), Ms: time.Since(start).Milliseconds()}
	}
	row := conn.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", dbName, table))
	var cnt int64
	if err := row.Scan(&cnt); err != nil {
		return &TableMaintenanceResult{Success: false, Action: "count", Table: table, Message: err.Error(), Ms: time.Since(start).Milliseconds()}
	}
	return &TableMaintenanceResult{Success: true, Action: "count", Table: table, Rows: cnt, Ms: time.Since(start).Milliseconds()}
}

// DatabaseCapabilities 返回数据源的能力描述
func (s *SQLService) DatabaseCapabilities(dbType string) map[string]interface{} {
	t := strings.ToLower(dbType)
	return map[string]interface{}{
		"dbType":          t,
		"defaultPort":     dbtype.DefaultPort(t),
		"systemDatabases": dbtype.SystemDatabases(t),
		"features": map[string]bool{
			"procedure": dbtype.SupportFeature(t, "procedure"),
			"trigger":   dbtype.SupportFeature(t, "trigger"),
			"fk":        dbtype.SupportFeature(t, "fk"),
			"delimiter": dbtype.SupportFeature(t, "delimiter"),
			"repair":    dbtype.SupportFeature(t, "repair"),
			"analyze":   dbtype.SupportFeature(t, "analyze"),
			"optimize":  dbtype.SupportFeature(t, "optimize"),
			"fulltext":  dbtype.SupportFeature(t, "fulltext"),
			"spatial":   dbtype.SupportFeature(t, "spatial"),
		},
	}
}

// InsertRow 数据编辑器：插入行
func (s *SQLService) InsertRow(ds *model.Datasource, dbName, table string, row map[string]interface{}) (int64, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return 0, err
	}
	cols := make([]string, 0, len(row))
	placeholders := make([]string, 0, len(row))
	vals := make([]interface{}, 0, len(row))
	for k, v := range row {
		cols = append(cols, "`"+k+"`")
		placeholders = append(placeholders, "?")
		vals = append(vals, v)
	}
	sql := fmt.Sprintf("INSERT INTO `%s`.`%s` (%s) VALUES (%s)", dbName, table, strings.Join(cols, ","), strings.Join(placeholders, ","))
	res, err := conn.DB.Exec(sql, vals...)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}

// UpdateRow 数据编辑器：按条件更新行
func (s *SQLService) UpdateRow(ds *model.Datasource, dbName, table string, updates, where map[string]interface{}) (int64, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return 0, err
	}
	setClauses := make([]string, 0, len(updates))
	vals := make([]interface{}, 0, len(updates)+len(where))
	for k, v := range updates {
		setClauses = append(setClauses, "`"+k+"` = ?")
		vals = append(vals, v)
	}
	whereClauses := make([]string, 0, len(where))
	for k, v := range where {
		whereClauses = append(whereClauses, "`"+k+"` = ?")
		vals = append(vals, v)
	}
	if len(whereClauses) == 0 {
		return 0, fmt.Errorf("更新语句必须包含 WHERE 条件")
	}
	sql := fmt.Sprintf("UPDATE `%s`.`%s` SET %s WHERE %s LIMIT 1", dbName, table, strings.Join(setClauses, ","), strings.Join(whereClauses, " AND "))
	res, err := conn.DB.Exec(sql, vals...)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}

// DeleteRow 数据编辑器：按条件删除行
func (s *SQLService) DeleteRow(ds *model.Datasource, dbName, table string, where map[string]interface{}) (int64, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return 0, err
	}
	whereClauses := make([]string, 0, len(where))
	vals := make([]interface{}, 0, len(where))
	for k, v := range where {
		whereClauses = append(whereClauses, "`"+k+"` = ?")
		vals = append(vals, v)
	}
	if len(whereClauses) == 0 {
		return 0, fmt.Errorf("删除语句必须包含 WHERE 条件")
	}
	sql := fmt.Sprintf("DELETE FROM `%s`.`%s` WHERE %s LIMIT 1", dbName, table, strings.Join(whereClauses, " AND "))
	res, err := conn.DB.Exec(sql, vals...)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}

// QueryTablePaginated 分页读取表数据
func (s *SQLService) QueryTablePaginated(ds *model.Datasource, dbName, table string, page, pageSize int) (columns []string, rows []map[string]interface{}, total int64, ms int64, err error) {
	start := time.Now()
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	countRow := conn.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", dbName, table))
	_ = countRow.Scan(&total)

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT ? OFFSET ?", dbName, table)
	qRows, err := conn.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, nil, 0, time.Since(start).Milliseconds(), err
	}
	defer qRows.Close()
	cols, _ := qRows.Columns()
	columns = cols
	for qRows.Next() {
		rawPtrs := make([]interface{}, len(cols))
		rawVals := make([]interface{}, len(cols))
		for i := range rawPtrs {
			rawPtrs[i] = &rawVals[i]
		}
		if err := qRows.Scan(rawPtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := rawVals[i].([]byte); ok {
				row[c] = string(b)
			} else {
				row[c] = rawVals[i]
			}
		}
		rows = append(rows, row)
	}
	return columns, rows, total, time.Since(start).Milliseconds(), nil
}

// FetchPrimaryKey 获取表主键列
func (s *SQLService) FetchPrimaryKey(ds *model.Datasource, dbName, table string) ([]string, error) {
	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: ds.Password,
		Database: dbName,
	}
	conn, err := dbtype.Connect(ds.DatasourceID, params)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SHOW KEYS FROM `%s`.`%s` WHERE Key_name = 'PRIMARY'", dbName, table)
	rows, err := conn.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pks := []string{}
	cols, _ := rows.Columns()
	colIdx := -1
	for i, c := range cols {
		if strings.EqualFold(c, "Column_name") || strings.EqualFold(c, "column_name") {
			colIdx = i
			break
		}
	}
	for rows.Next() {
		rawPtrs := make([]interface{}, len(cols))
		rawVals := make([]interface{}, len(cols))
		for i := range rawPtrs {
			rawPtrs[i] = &rawVals[i]
		}
		if err := rows.Scan(rawPtrs...); err != nil {
			continue
		}
		if colIdx >= 0 {
			if b, ok := rawVals[colIdx].([]byte); ok {
				pks = append(pks, string(b))
			} else if s, ok := rawVals[colIdx].(string); ok {
				pks = append(pks, s)
			}
		}
	}
	return pks, nil
}
