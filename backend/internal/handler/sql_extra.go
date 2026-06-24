package handler

import (
	"strconv"
	"strings"

	"dbm-lite/internal/middleware"
	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *SQLHandler) DatabaseCapabilities(c *gin.Context) {
	dbType := c.DefaultQuery("dbType", "")
	if dbType == "" {
		id := c.Query("datasourceId")
		if id != "" {
			if ds, err := h.dsSvc.GetById(id); err == nil && ds != nil {
				dbType = ds.DBType
			}
		}
	}
	middleware.Success(c, h.sqlSvc.DatabaseCapabilities(dbType))
}

func (h *SQLHandler) TableMaintenance(c *gin.Context) {
	id := c.Param("id")
	action := strings.ToLower(c.Param("action"))
	var req struct {
		Database string `json:"database"`
		Table    string `json:"table"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "param error")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	var result *service.TableMaintenanceResult
	switch action {
	case "analyze":
		result = h.sqlSvc.AnalyzeTable(ds, req.Database, req.Table)
	case "check":
		result = h.sqlSvc.CheckTable(ds, req.Database, req.Table)
	case "optimize":
		result = h.sqlSvc.OptimizeTable(ds, req.Database, req.Table)
	case "repair":
		result = h.sqlSvc.RepairTable(ds, req.Database, req.Table)
	case "count":
		result = h.sqlSvc.GetTableRowCount(ds, req.Database, req.Table)
	default:
		middleware.Error(c, "unsupported action: "+action)
		return
	}
	middleware.Success(c, result)
}

func (h *SQLHandler) GetTableDDL(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	table := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	ddl, err := h.sqlSvc.GetTableDDL(ds, dbName, table)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, gin.H{"ddl": ddl})
}

func (h *SQLHandler) QueryTable(c *gin.Context) {
	id := c.Param("id")
	dbName := c.DefaultQuery("database", "")
	table := c.DefaultQuery("table", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "300"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 300
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	columns, rows, total, ms, err := h.sqlSvc.QueryTablePaginated(ds, dbName, table, page, pageSize)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, gin.H{
		"columns":    columns,
		"rows":       rows,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"durationMs": ms,
		"table":      table,
		"database":   dbName,
		"hasMore":    int64(page*pageSize) < total,
	})
}

func (h *SQLHandler) InsertRow(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Row      map[string]interface{} `json:"row"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "param error")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	affected, err := h.sqlSvc.InsertRow(ds, req.Database, req.Table, req.Row)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, gin.H{"affectedRows": affected})
}

func (h *SQLHandler) UpdateRow(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Updates  map[string]interface{} `json:"updates"`
		Where    map[string]interface{} `json:"where"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "param error")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	affected, err := h.sqlSvc.UpdateRow(ds, req.Database, req.Table, req.Updates, req.Where)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, gin.H{"affectedRows": affected})
}

func (h *SQLHandler) DeleteRow(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Where    map[string]interface{} `json:"where"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "param error")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	affected, err := h.sqlSvc.DeleteRow(ds, req.Database, req.Table, req.Where)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, gin.H{"affectedRows": affected})
}

func (h *SQLHandler) PrimaryKey(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	table := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	pks, err := h.sqlSvc.FetchPrimaryKey(ds, dbName, table)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, pks)
}

func (h *SQLHandler) GetTableInfoFull(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	table := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "datasource not found")
		return
	}
	info, err := h.sqlSvc.GetTableInfo(ds, dbName, table)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	pks, _ := h.sqlSvc.FetchPrimaryKey(ds, dbName, table)
	middleware.Success(c, map[string]interface{}{
		"info":       info,
		"primaryKey": pks,
		"database":   dbName,
		"table":      table,
	})
}
