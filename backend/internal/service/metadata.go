/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"database/sql"
	"fmt"
	"strings"

	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"
)

type TableMaintenanceService struct{}

func NewTableMaintenanceService() *TableMaintenanceService { return &TableMaintenanceService{} }

type MaintenanceResult struct {
	Operation string `json:"operation"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Duration  int64  `json:"durationMs"`
}

func (s *TableMaintenanceService) AnalyzeTable(ds *model.Datasource, dbName, tableName string) (*MaintenanceResult, error) {
	return s.executeMaintenance(ds, dbName, fmt.Sprintf("ANALYZE TABLE `%s`.`%s`", dbName, tableName), "ANALYZE")
}

func (s *TableMaintenanceService) CheckTable(ds *model.Datasource, dbName, tableName string) (*MaintenanceResult, error) {
	return s.executeMaintenance(ds, dbName, fmt.Sprintf("CHECK TABLE `%s`.`%s`", dbName, tableName), "CHECK")
}

func (s *TableMaintenanceService) OptimizeTable(ds *model.Datasource, dbName, tableName string) (*MaintenanceResult, error) {
	return s.executeMaintenance(ds, dbName, fmt.Sprintf("OPTIMIZE TABLE `%s`.`%s`", dbName, tableName), "OPTIMIZE")
}

func (s *TableMaintenanceService) RepairTable(ds *model.Datasource, dbName, tableName string) (*MaintenanceResult, error) {
	return s.executeMaintenance(ds, dbName, fmt.Sprintf("REPAIR TABLE `%s`.`%s`", dbName, tableName), "REPAIR")
}

func (s *TableMaintenanceService) GetRowCount(ds *model.Datasource, dbName, tableName string) (int64, error) {
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
		return 0, err
	}
	var count int64
	err = conn.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", dbName, tableName)).Scan(&count)
	return count, err
}

func (s *TableMaintenanceService) executeMaintenance(ds *model.Datasource, dbName, sql, operation string) (*MaintenanceResult, error) {
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
	_, err = conn.DB.Exec(sql)
	if err != nil {
		return &MaintenanceResult{
			Operation: operation,
			Success:   false,
			Message:   err.Error(),
		}, nil
	}
	return &MaintenanceResult{
		Operation: operation,
		Success:   true,
		Message:   fmt.Sprintf("%s TABLE `%s`.`%s` 成功", operation, dbName, strings.Split(sql, "`")[3]),
	}, nil
}

type ObjectMetadataService struct{}

func NewObjectMetadataService() *ObjectMetadataService { return &ObjectMetadataService{} }

func (s *ObjectMetadataService) GetProcedures(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
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
	rows, err := conn.DB.Query(fmt.Sprintf("SHOW PROCEDURE STATUS WHERE Db = '%s'", dbName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.parseShowResults(rows)
}

func (s *ObjectMetadataService) GetTriggers(ds *model.Datasource, dbName string) ([]map[string]interface{}, error) {
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
	rows, err := conn.DB.Query(fmt.Sprintf("SHOW TRIGGERS FROM `%s`", dbName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.parseShowResults(rows)
}

func (s *ObjectMetadataService) GetIndexes(ds *model.Datasource, dbName, tableName string) ([]map[string]interface{}, error) {
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
	rows, err := conn.DB.Query(fmt.Sprintf("SHOW INDEX FROM `%s`.`%s`", dbName, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.parseShowResults(rows)
}

func (s *ObjectMetadataService) parseShowResults(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
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
