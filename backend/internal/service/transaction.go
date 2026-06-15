/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"fmt"
	"strings"

	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/sqllint"
)

type TransactionService struct{}

func NewTransactionService() *TransactionService { return &TransactionService{} }

type TransactionResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RowsCount int64  `json:"rowsCount"`
}

func (s *TransactionService) BeginTransaction(dsId string) error {
	return dbtype.BeginTransaction(dsId)
}

func (s *TransactionService) CommitTransaction(dsId string) (*TransactionResult, error) {
	err := dbtype.CommitTransaction(dsId)
	if err != nil {
		return &TransactionResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &TransactionResult{
		Success: true,
		Message: "事务提交成功",
	}, nil
}

func (s *TransactionService) RollbackTransaction(dsId string) (*TransactionResult, error) {
	err := dbtype.RollbackTransaction(dsId)
	if err != nil {
		return &TransactionResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &TransactionResult{
		Success: true,
		Message: "事务回滚成功",
	}, nil
}

func (s *TransactionService) ExecuteInTransaction(ds *model.Datasource, dbName string, sqls []string) ([]*ExecResult, error) {
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

	tx, err := conn.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}

	results := make([]*ExecResult, 0, len(sqls))
	for _, execSQL := range sqls {
		execSQL = strings.TrimSpace(execSQL)
		if execSQL == "" {
			continue
		}

		stmtReview := sqllint.Review(execSQL)
		result := &ExecResult{
			IsSelect: stmtReview.IsSelect,
			Review:   stmtReview,
			Success:  true,
		}

		if stmtReview.IsSelect {
			rows, err := tx.Query(execSQL)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("执行查询失败: %w", err)
			}
			cols, _ := rows.Columns()
			result.Columns = cols
			result.Rows = []map[string]interface{}{}
			for rows.Next() {
				colVals := make([]interface{}, len(cols))
				colPtrs := make([]interface{}, len(cols))
				for i := range colVals {
					colPtrs[i] = &colVals[i]
				}
				rows.Scan(colPtrs...)
				row := make(map[string]interface{})
				for i, col := range cols {
					if b, ok := colVals[i].([]byte); ok {
						row[col] = string(b)
					} else {
						row[col] = colVals[i]
					}
				}
				result.Rows = append(result.Rows, row)
			}
			rows.Close()
		} else {
			res, err := tx.Exec(execSQL)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("执行失败: %w", err)
			}
			affected, _ := res.RowsAffected()
			result.AffectedRows = affected
		}
		results = append(results, result)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return results, nil
}

type DataEditorService struct{}

func NewDataEditorService() *DataEditorService { return &DataEditorService{} }

func (s *DataEditorService) InsertRow(ds *model.Datasource, dbName, tableName string, data map[string]interface{}) (*TransactionResult, error) {
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
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for k, v := range data {
		columns = append(columns, "`"+k+"`")
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}
	sql := fmt.Sprintf("INSERT INTO `%s`.`%s` (%s) VALUES (%s)", dbName, tableName, strings.Join(columns, ","), strings.Join(placeholders, ","))
	res, err := conn.DB.Exec(sql, values...)
	if err != nil {
		return &TransactionResult{Success: false, Message: err.Error()}, nil
	}
	affected, _ := res.RowsAffected()
	return &TransactionResult{Success: true, Message: "插入成功", RowsCount: affected}, nil
}

func (s *DataEditorService) UpdateRow(ds *model.Datasource, dbName, tableName string, data map[string]interface{}, where map[string]interface{}) (*TransactionResult, error) {
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
	setParts := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+len(where))
	for k, v := range data {
		setParts = append(setParts, "`"+k+"`=?")
		values = append(values, v)
	}
	whereParts := make([]string, 0, len(where))
	for k, v := range where {
		whereParts = append(whereParts, "`"+k+"`=?")
		values = append(values, v)
	}
	sql := fmt.Sprintf("UPDATE `%s`.`%s` SET %s WHERE %s", dbName, tableName, strings.Join(setParts, ","), strings.Join(whereParts, " AND "))
	res, err := conn.DB.Exec(sql, values...)
	if err != nil {
		return &TransactionResult{Success: false, Message: err.Error()}, nil
	}
	affected, _ := res.RowsAffected()
	return &TransactionResult{Success: true, Message: "更新成功", RowsCount: affected}, nil
}

func (s *DataEditorService) DeleteRow(ds *model.Datasource, dbName, tableName string, where map[string]interface{}) (*TransactionResult, error) {
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
	whereParts := make([]string, 0, len(where))
	values := make([]interface{}, 0, len(where))
	for k, v := range where {
		whereParts = append(whereParts, "`"+k+"`=?")
		values = append(values, v)
	}
	sql := fmt.Sprintf("DELETE FROM `%s`.`%s` WHERE %s", dbName, tableName, strings.Join(whereParts, " AND "))
	res, err := conn.DB.Exec(sql, values...)
	if err != nil {
		return &TransactionResult{Success: false, Message: err.Error()}, nil
	}
	affected, _ := res.RowsAffected()
	return &TransactionResult{Success: true, Message: "删除成功", RowsCount: affected}, nil
}
