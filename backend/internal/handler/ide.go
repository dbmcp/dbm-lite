package handler

import (
	"net/http"
	"strconv"
	"strings"

	"dbm-lite/internal/middleware"
	"dbm-lite/internal/model"
	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

// IDEHandler handles Navicat 风格的 SQL IDE 前端请求。
type IDEHandler struct {
	dsSvc  *service.DatasourceService
	sqlSvc *service.SQLService
}

func NewIDEHandler() *IDEHandler {
	return &IDEHandler{
		dsSvc:  service.NewDatasourceService(),
		sqlSvc: service.NewSQLService(),
	}
}

// ListSystemDatabases 返回通用的系统库名称集合，供前端对不同数据源做灰色/隐藏处理。
func (h *IDEHandler) ListSystemDatabases(c *gin.Context) {
	middleware.OK(c, map[string][]string{
		"mysql":  {"information_schema", "mysql", "performance_schema", "sys"},
		"tidb":   {"information_schema", "mysql", "performance_schema", "metrics_schema", "sys"},
		"sqlite": {"sqlite_master", "sqlite_temp_master"},
	})
}

// GetDatabasesFull 返回数据源下的数据库列表，可选是否包含系统库（默认不包含）。
func (h *IDEHandler) GetDatabasesFull(c *gin.Context) {
	id := c.Param("id")
	includeSystem := c.DefaultQuery("includeSystem", "false") == "true"
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	dbs, err := h.sqlSvc.GetDatabasesWithSystem(ds, includeSystem)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, dbs)
}

// GetTableChildren 返回某表下的子对象（列、索引、外键、触发器），供懒加载树节点使用。
func (h *IDEHandler) GetTableChildren(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	if tableName == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "缺少 table 参数")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	children, err := h.sqlSvc.GetTableChildren(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, children)
}

// GetRoutines 返回指定数据库下的存储过程与函数列表。
func (h *IDEHandler) GetRoutines(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	list, err := h.sqlSvc.GetRoutines(ds, dbName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

// GetTriggersForIde 返回数据库下触发器列表。
func (h *IDEHandler) GetTriggersForIde(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	list, err := h.sqlSvc.GetTriggers(ds, dbName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

// GetIndexes 返回表索引列表。
func (h *IDEHandler) GetIndexes(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	list, err := h.sqlSvc.GetIndexes(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

// GetForeignKeys 返回表外键列表。
func (h *IDEHandler) GetForeignKeys(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	list, err := h.sqlSvc.GetForeignKeys(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

// GetTableTriggers 返回与表相关的触发器列表。
func (h *IDEHandler) GetTableTriggers(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	list, err := h.sqlSvc.GetTableTriggers(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

// GetViewList 返回指定数据库的视图列表（含基础信息）。
func (h *IDEHandler) GetViewList(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	list, err := h.sqlSvc.GetViewList(ds, dbName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

// ExecuteSQLV2 与原有 Execute 语义一致，额外带上执行耗时返回给前端。
func (h *IDEHandler) ExecuteSQLV2(c *gin.Context) {
	var req struct {
		DatasourceID string `json:"datasourceId"`
		Database     string `json:"database"`
		SQL          string `json:"sql"`
		IgnoreRisk   bool   `json:"ignoreRisk"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "SQL 不能为空")
		return
	}
	ds, err := h.dsSvc.GetById(req.DatasourceID)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	results, err := h.sqlSvc.Execute(ds, req.Database, req.SQL, req.IgnoreRisk, userId, username)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, results)
}

// QueryTableV2 分页查询表数据，为 IDE 网格组件提供直接可用的列 + 行数据。
func (h *IDEHandler) QueryTableV2(c *gin.Context) {
	id := c.Param("id")
	dbName := c.DefaultQuery("database", "")
	table := c.DefaultQuery("table", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "500"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 5000 {
		pageSize = 500
	}
	if table == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "缺少 table 参数")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	columns, rows, total, ms, err := h.sqlSvc.QueryTablePaginated(ds, dbName, table, page, pageSize)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	pks, _ := h.sqlSvc.FetchPrimaryKey(ds, dbName, table)
	middleware.OK(c, gin.H{
		"columns":    columns,
		"rows":       rows,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"durationMs": ms,
		"table":      table,
		"database":   dbName,
		"primaryKey": pks,
		"hasMore":    int64(page*pageSize) < total,
	})
}

// TableInfoV2 合并返回表信息 + 主键 + 列信息。
func (h *IDEHandler) TableInfoV2(c *gin.Context) {
	id := c.Param("id")
	dbName := c.DefaultQuery("database", "")
	table := c.DefaultQuery("table", "")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	info, err := h.sqlSvc.GetTableInfo(ds, dbName, table)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	cols, _ := h.sqlSvc.GetColumns(ds, dbName, table)
	pks, _ := h.sqlSvc.FetchPrimaryKey(ds, dbName, table)
	middleware.OK(c, gin.H{
		"info":       info,
		"columns":    cols,
		"primaryKey": pks,
		"database":   dbName,
		"table":      table,
	})
}

// Capabilities 返回数据库类型能力（过程、触发器、外键、修复、分析、全文、空间索引等）。
func (h *IDEHandler) Capabilities(c *gin.Context) {
	dbType := c.DefaultQuery("dbType", "")
	if dbType == "" {
		id := c.Query("datasourceId")
		if id != "" {
			if ds, err := h.dsSvc.GetById(id); err == nil && ds != nil {
				dbType = ds.DBType
			}
		}
	}
	middleware.OK(c, h.sqlSvc.DatabaseCapabilities(dbType))
}

// TableDDLV2 返回表 DDL。
func (h *IDEHandler) TableDDLV2(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	table := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	ddl, err := h.sqlSvc.GetTableDDL(ds, dbName, table)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"ddl": ddl, "table": table, "database": dbName})
}

// InsertRowV2 新增一条数据（以 JSON map 提交），供网格编辑器使用。
func (h *IDEHandler) InsertRowV2(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Row      map[string]interface{} `json:"row"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	affected, err := h.sqlSvc.InsertRow(ds, req.Database, req.Table, req.Row)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"affectedRows": affected})
}

// UpdateRowV2 更新一条数据（通过 where 定位）。
func (h *IDEHandler) UpdateRowV2(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Updates  map[string]interface{} `json:"updates"`
		Where    map[string]interface{} `json:"where"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	affected, err := h.sqlSvc.UpdateRow(ds, req.Database, req.Table, req.Updates, req.Where)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"affectedRows": affected})
}

// DeleteRowV2 删除数据。
func (h *IDEHandler) DeleteRowV2(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Where    map[string]interface{} `json:"where"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	affected, err := h.sqlSvc.DeleteRow(ds, req.Database, req.Table, req.Where)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"affectedRows": affected})
}

// TestConnectionV2 测试数据源连接，用于 IDE 左侧“连接测试”菜单项。
func (h *IDEHandler) TestConnectionV2(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	result := h.sqlSvc.TestConnection(ds)
	if result.Success {
		h.dsSvc.UpdateConnStatus(id, model.ConnStatusOK, result.LatencyMs, result.Version)
	} else {
		h.dsSvc.UpdateConnStatus(id, model.ConnStatusFail, 0, "")
	}
	middleware.OK(c, result)
}