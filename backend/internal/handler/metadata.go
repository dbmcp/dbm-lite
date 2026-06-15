/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package handler

import (
	"dbm-lite/internal/middleware"
	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

type MetadataHandler struct {
	dsSvc *service.DatasourceService
	metaSvc *service.ObjectMetadataService
	maintainSvc *service.TableMaintenanceService
}

func NewMetadataHandler() *MetadataHandler {
	return &MetadataHandler{
		dsSvc: service.NewDatasourceService(),
		metaSvc: service.NewObjectMetadataService(),
		maintainSvc: service.NewTableMaintenanceService(),
	}
}

func (h *MetadataHandler) GetProcedures(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	procs, err := h.metaSvc.GetProcedures(ds, dbName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, procs)
}

func (h *MetadataHandler) GetTriggers(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	triggers, err := h.metaSvc.GetTriggers(ds, dbName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, triggers)
}

func (h *MetadataHandler) GetIndexes(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	indexes, err := h.metaSvc.GetIndexes(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, indexes)
}

func (h *MetadataHandler) AnalyzeTable(c *gin.Context) {
	h.executeMaintenance(c, "analyze")
}

func (h *MetadataHandler) CheckTable(c *gin.Context) {
	h.executeMaintenance(c, "check")
}

func (h *MetadataHandler) OptimizeTable(c *gin.Context) {
	h.executeMaintenance(c, "optimize")
}

func (h *MetadataHandler) RepairTable(c *gin.Context) {
	h.executeMaintenance(c, "repair")
}

func (h *MetadataHandler) GetRowCount(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	count, err := h.maintainSvc.GetRowCount(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"count": count})
}

func (h *MetadataHandler) executeMaintenance(c *gin.Context, op string) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	var result *service.MaintenanceResult
	switch op {
	case "analyze":
		result, err = h.maintainSvc.AnalyzeTable(ds, dbName, tableName)
	case "check":
		result, err = h.maintainSvc.CheckTable(ds, dbName, tableName)
	case "optimize":
		result, err = h.maintainSvc.OptimizeTable(ds, dbName, tableName)
	case "repair":
		result, err = h.maintainSvc.RepairTable(ds, dbName, tableName)
	default:
		middleware.Fail(c, 400, 400, "未知操作类型")
		return
	}
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	if result.Success {
		middleware.OK(c, result)
	} else {
		middleware.Fail(c, 400, 400, result.Message)
	}
}

type TableDesignHandler struct {
	dsSvc *service.DatasourceService
	sqlSvc *service.SQLService
}

func NewTableDesignHandler() *TableDesignHandler {
	return &TableDesignHandler{
		dsSvc: service.NewDatasourceService(),
		sqlSvc: service.NewSQLService(),
	}
}

func (h *TableDesignHandler) GetTableDDL(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	ddl, err := h.sqlSvc.GetTableDDL(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"ddl": ddl})
}

func (h *TableDesignHandler) GetTableInfo(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	info, err := h.sqlSvc.GetTableInfo(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, info)
}

func (h *TableDesignHandler) GetColumns(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	cols, err := h.sqlSvc.GetColumns(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, cols)
}

type ViewDesignHandler struct {
	dsSvc *service.DatasourceService
	sqlSvc *service.SQLService
}

func NewViewDesignHandler() *ViewDesignHandler {
	return &ViewDesignHandler{
		dsSvc: service.NewDatasourceService(),
		sqlSvc: service.NewSQLService(),
	}
}

func (h *ViewDesignHandler) GetViewDefinition(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	viewName := c.Query("view")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	info, err := h.sqlSvc.GetTableInfo(ds, dbName, viewName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"definition": info["ddl"]})
}

type DBSettingsHandler struct{}

func NewDBSettingsHandler() *DBSettingsHandler { return &DBSettingsHandler{} }

func (h *DBSettingsHandler) GetDatabaseTypes(c *gin.Context) {
	types := []map[string]interface{}{
		{
			"type":        "mysql",
			"name":        "MySQL",
			"defaultPort": 3306,
			"features": map[string]bool{
				"storedProcedure": true,
				"trigger":         true,
				"foreignKey":      true,
				"fullTextIndex":   true,
				"spatialIndex":    true,
				"repairTable":     true,
			},
			"systemDatabases": []string{"information_schema", "mysql", "performance_schema", "sys"},
		},
		{
			"type":        "tidb",
			"name":        "TiDB",
			"defaultPort": 4000,
			"features": map[string]bool{
				"storedProcedure": false,
				"trigger":         false,
				"foreignKey":      false,
				"fullTextIndex":   false,
				"spatialIndex":    false,
				"repairTable":     false,
			},
			"systemDatabases": []string{"information_schema", "mysql", "performance_schema", "sys", "metrics_schema"},
		},
	}
	middleware.OK(c, types)
}

type ExecutePlanHandler struct {
	dsSvc *service.DatasourceService
	sqlSvc *service.SQLService
}

func NewExecutePlanHandler() *ExecutePlanHandler {
	return &ExecutePlanHandler{
		dsSvc: service.NewDatasourceService(),
		sqlSvc: service.NewSQLService(),
	}
}

func (h *ExecutePlanHandler) Explain(c *gin.Context) {
	var req struct {
		DatasourceID string `json:"datasourceId"`
		Database     string `json:"database"`
		SQL          string `json:"sql"`
		Analyze      bool   `json:"analyze"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, 400, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(req.DatasourceID)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	plan, err := h.sqlSvc.Explain(ds, req.Database, req.SQL)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"plan": plan, "analyze": req.Analyze})
}

type DataExportHandler struct {
	dsSvc *service.DatasourceService
	sqlSvc *service.SQLService
}

func NewDataExportHandler() *DataExportHandler {
	return &DataExportHandler{
		dsSvc: service.NewDatasourceService(),
		sqlSvc: service.NewSQLService(),
	}
}

func (h *DataExportHandler) ExportCSV(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	_ = c.DefaultQuery("page", "1")
	_ = c.DefaultQuery("pageSize", "1000")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	// 简化实现：返回数据供前端处理
	cols, err := h.sqlSvc.GetColumns(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	// 实际导出逻辑应由前端实现或通过流式响应
	middleware.OK(c, gin.H{"columns": cols, "message": "请使用前端导出功能"})
}
