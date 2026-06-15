/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
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

type SQLHandler struct {
	sqlSvc *service.SQLService
	dsSvc  *service.DatasourceService
}

func NewSQLHandler() *SQLHandler {
	return &SQLHandler{
		sqlSvc: service.NewSQLService(),
		dsSvc:  service.NewDatasourceService(),
	}
}

type executeReq struct {
	DatasourceID string `json:"datasourceId"`
	Database     string `json:"database"`
	SQL          string `json:"sql"`
	IgnoreRisk   bool   `json:"ignoreRisk"`
}

func (h *SQLHandler) Execute(c *gin.Context) {
	var req executeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	ds, err := h.dsSvc.GetById(req.DatasourceID)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	if ds.ReadOnly {
		upper := strings.ToUpper(strings.TrimSpace(req.SQL))
		if strings.HasPrefix(upper, "INSERT") || strings.HasPrefix(upper, "UPDATE") ||
			strings.HasPrefix(upper, "DELETE") || strings.HasPrefix(upper, "DROP") ||
			strings.HasPrefix(upper, "TRUNCATE") || strings.HasPrefix(upper, "ALTER") ||
			strings.HasPrefix(upper, "CREATE") {
			middleware.Fail(c, http.StatusBadRequest, 400, "只读数据源不允许执行 DDL/DML")
			return
		}
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	result, err := h.sqlSvc.Execute(ds, req.Database, req.SQL, req.IgnoreRisk, userId, username)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, result)
}

func (h *SQLHandler) Explain(c *gin.Context) {
	var req executeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(req.DatasourceID)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	result, err := h.sqlSvc.Explain(ds, req.Database, req.SQL)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, result)
}

func (h *SQLHandler) ReviewSQL(c *gin.Context) {
	var req struct {
		SQL string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	middleware.OK(c, h.sqlSvc.ReviewSQL(req.SQL))
}

func (h *SQLHandler) TestConnection(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	result := h.sqlSvc.TestConnection(ds)
	if result.Success {
		h.dsSvc.UpdateConnStatus(id, model.ConnStatusOK, result.LatencyMs, result.Version)
		middleware.OK(c, gin.H{
			"success":   true,
			"message":   result.Message,
			"latencyMs": result.LatencyMs,
			"version":   result.Version,
		})
		return
	}
	h.dsSvc.UpdateConnStatus(id, model.ConnStatusFail, 0, "")
	middleware.Fail(c, http.StatusBadRequest, 400, result.Message)
}

func (h *SQLHandler) GetDatabases(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	dbs, err := h.sqlSvc.GetDatabases(ds)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, dbs)
}

func (h *SQLHandler) GetTables(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	tables, err := h.sqlSvc.GetTables(ds, dbName)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, tables)
}

func (h *SQLHandler) GetColumns(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	cols, err := h.sqlSvc.GetColumns(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, cols)
}

func (h *SQLHandler) GetFullTree(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	tree, err := h.sqlSvc.GetFullTree(ds)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, tree)
}

func (h *SQLHandler) GetTableInfo(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	info, err := h.sqlSvc.GetTableInfo(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, info)
}

func (h *SQLHandler) GetHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	datasourceId := c.Query("datasourceId")
	keyword := c.Query("keyword")
	list, total, err := h.sqlSvc.GetHistory(page, pageSize, datasourceId, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

type DatasourceHandler struct {
	dsSvc    *service.DatasourceService
	sqlSvc   *service.SQLService
	auditSvc *service.AuditService
	srvSvc   *service.ServerService
}

func NewDatasourceHandler() *DatasourceHandler {
	return &DatasourceHandler{
		dsSvc:    service.NewDatasourceService(),
		sqlSvc:   service.NewSQLService(),
		auditSvc: service.NewAuditService(),
		srvSvc:   service.NewServerService(),
	}
}

type createDsReq struct {
	Name            string `json:"name"`
	DBType          string `json:"dbType"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	DefaultDB       string `json:"defaultDatabase"`
	DefaultDatabase string `json:"defaultDb"`
	FilePath        string `json:"filePath"`
	OpenMode        string `json:"openMode"`
	Charset         string `json:"charset"`
	Timezone        string `json:"timezone"`
	SSLMode         string `json:"sslMode"`
	SSLCAFile       string `json:"sslCaFile"`
	ReadOnly        bool   `json:"readOnly"`
	ColorLabel      string `json:"colorLabel"`
	Version         string `json:"version"`
	Tags            string `json:"tags"`
	BusinessID      string `json:"businessId"`
	ServerID        string `json:"serverId"`
	ProjectID       string `json:"projectId"`
	Env             string `json:"env"`
	Remark          string `json:"remark"`
	Status          string `json:"status"`
	AutoCreateSrv   bool   `json:"autoCreateServer"`
}

func (h *DatasourceHandler) Create(c *gin.Context) {
	var req createDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")

	openMode := strings.ToLower(strings.TrimSpace(req.OpenMode))

	// 如果没有关联 serverId，尝试自动创建一个
	if req.ServerID == "" && req.Host != "" && strings.ToLower(req.DBType) != model.DBTypeSQLite && req.AutoCreateSrv {
		if srv, err := h.srvSvc.EnsureByHost(req.Host, userId); err == nil && srv != nil {
			req.ServerID = srv.ServerID
		}
	}

	defaultDB := req.DefaultDB
	if defaultDB == "" {
		defaultDB = req.DefaultDatabase
	}
	if defaultDB == "" && strings.ToLower(req.DBType) == model.DBTypeSQLite {
		defaultDB = "main"
	}
	ds := &model.Datasource{
		Name:       req.Name,
		DBType:     strings.ToLower(req.DBType),
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		DefaultDB:  defaultDB,
		FilePath:   req.FilePath,
		OpenMode:   openMode,
		Charset:    req.Charset,
		Timezone:   req.Timezone,
		SSLMode:    req.SSLMode,
		SSLCAFile:  req.SSLCAFile,
		ReadOnly:   req.ReadOnly,
		ColorLabel: req.ColorLabel,
		Version:    "",
		Tags:       req.Tags,
		BusinessID: req.BusinessID,
		ServerID:   req.ServerID,
		ProjectID:  req.ProjectID,
		Env:        req.Env,
		Remark:     req.Remark,
		Status:     req.Status,
	}
	if err := h.dsSvc.Create(ds, req.Password, userId, username); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleDatasource, "create", ds.DatasourceID, "创建数据源: "+ds.Name, model.AuditResultSuccess, "")
	middleware.OK(c, ds)
}

func (h *DatasourceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if p, err := strconv.Atoi(c.Query("size")); err == nil && p > 0 {
		pageSize = p
	}
	keyword := c.Query("keyword")
	dbType := c.Query("dbType")
	if dbType == "" {
		dbType = c.Query("type")
	}
	status := c.Query("status")
	sortBy := c.Query("sortBy")
	businessId := c.Query("businessId")
	env := c.Query("env")

	list, total, err := h.dsSvc.List(page, pageSize, keyword, dbType, status, sortBy, businessId, env)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{
		"list":     list,
		"total":    total,
		"current":  page,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Get 返回单个数据源详情（脱敏密码）
func (h *DatasourceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetByIdNoDecrypt(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	middleware.OK(c, ds)
}

// GetDetail 返回带完整信息的数据源详情（脱敏密码），包含状态、延迟、版本等
func (h *DatasourceHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetByIdNoDecrypt(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}

	// 返回详情 + 最近测试信息
	middleware.OK(c, gin.H{
		"datasource":   ds,
		"connStatus":   ds.ConnStatus,
		"lastTestAt":   ds.LastConnTestAt,
		"latencyMs":    ds.ConnLatencyMs,
		"version":      ds.Version,
		"passwordSet":  ds.HasPassword(),
		"connectionOK": ds.IsConnectionOK(),
	})
}

func (h *DatasourceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req createDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	rawPwd := req.Password
	openMode := strings.ToLower(strings.TrimSpace(req.OpenMode))
	defaultDB := req.DefaultDB
	if defaultDB == "" {
		defaultDB = req.DefaultDatabase
	}
	if defaultDB == "" && strings.ToLower(req.DBType) == model.DBTypeSQLite {
		defaultDB = "main"
	}
	updates := map[string]interface{}{
		"name":        req.Name,
		"db_type":     strings.ToLower(req.DBType),
		"host":        req.Host,
		"port":        req.Port,
		"username":    req.Username,
		"default_db":  defaultDB,
		"file_path":   req.FilePath,
		"open_mode":   openMode,
		"charset":     req.Charset,
		"timezone":    req.Timezone,
		"ssl_mode":    req.SSLMode,
		"ssl_ca_file": req.SSLCAFile,
		"read_only":   req.ReadOnly,
		"color_label": req.ColorLabel,
		"tags":        req.Tags,
		"business_id": req.BusinessID,
		"server_id":   req.ServerID,
		"project_id":  req.ProjectID,
		"env":         req.Env,
		"remark":      req.Remark,
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if err := h.dsSvc.Update(id, updates, rawPwd); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleDatasource, "update", id, "更新数据源: "+req.Name, model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetByIdNoDecrypt(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	if err := h.dsSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleDatasource, "delete", id, "删除数据源: "+ds.Name, model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *DatasourceHandler) Copy(c *gin.Context) {
	id := c.Param("id")
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	newDs, err := h.dsSvc.Copy(id, userId, username)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleDatasource, "copy", newDs.DatasourceID, "复制数据源: "+newDs.Name, model.AuditResultSuccess, "")
	middleware.OK(c, newDs)
}

func (h *DatasourceHandler) Stats(c *gin.Context) {
	stats, err := h.dsSvc.Stats()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, stats)
}

func (h *DatasourceHandler) AllSimple(c *gin.Context) {
	keyword := c.Query("keyword")
	dbType := c.Query("dbType")
	list, _, err := h.dsSvc.List(1, 5000, keyword, dbType, "", "", "", "")
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	result := make([]map[string]interface{}, 0, len(list))
	for _, ds := range list {
		addr := ""
		if strings.ToLower(ds.DBType) == model.DBTypeSQLite {
			addr = ds.FilePath
			if addr == "" {
				addr = "(内存库)"
			}
		} else {
			addr = ds.Host
			if ds.Port > 0 {
				addr = addr + ":" + strconv.Itoa(ds.Port)
			}
		}
		result = append(result, map[string]interface{}{
			"id":         ds.DatasourceID,
			"name":       ds.Name,
			"dbType":     ds.DBType,
			"host":       ds.Host,
			"port":       ds.Port,
			"address":    addr,
			"env":        ds.Env,
			"defaultDb":  ds.DefaultDB,
			"connStatus": ds.ConnStatus,
			"colorLabel": ds.ColorLabel,
			"latencyMs":  ds.ConnLatencyMs,
			"version":    ds.Version,
			"lastTestAt": ds.LastConnTestAt,
			"readOnly":   ds.ReadOnly,
			"username":   ds.Username,
			"tags":       ds.Tags,
		})
	}
	middleware.OK(c, result)
}

// TestConnectionById 测试已有数据源连接，返回包含 success/message/latencyMs/version
func (h *DatasourceHandler) TestConnectionById(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	result := h.sqlSvc.TestConnection(ds)
	if result.Success {
		h.dsSvc.UpdateConnStatus(id, model.ConnStatusOK, result.LatencyMs, result.Version)
		middleware.OK(c, gin.H{
			"success":   true,
			"message":   result.Message,
			"latencyMs": result.LatencyMs,
			"version":   result.Version,
		})
		return
	}
	h.dsSvc.UpdateConnStatus(id, model.ConnStatusFail, 0, "")
	middleware.Fail(c, http.StatusBadRequest, 400, result.Message)
}

// TestConnectionFromForm 根据提交的表单直接测试连接
func (h *DatasourceHandler) TestConnectionFromForm(c *gin.Context) {
	var req createDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	defaultDB := req.DefaultDB
	if defaultDB == "" {
		defaultDB = req.DefaultDatabase
	}
	result := h.sqlSvc.TestConnectionDirect(
		strings.ToLower(req.DBType),
		req.Host,
		req.Port,
		req.Username,
		req.Password,
		defaultDB,
		req.FilePath,
		strings.ToLower(req.OpenMode),
		req.Charset,
		req.Timezone,
		req.SSLMode,
	)
	if result.Success {
		middleware.OK(c, gin.H{
			"success":   true,
			"message":   result.Message,
			"latencyMs": result.LatencyMs,
			"version":   result.Version,
		})
		return
	}
	middleware.Fail(c, http.StatusBadRequest, 400, result.Message)
}

// ==================== V2 API 处理器 (success/data/message 统一格式) ====================

type v2CreateDsReq struct {
	Name           string   `json:"name"`
	DatasourceType string   `json:"datasourceType"`
	Type           string   `json:"type"`
	DBType         string   `json:"dbType"`
	Env            string   `json:"env"`
	Host           string   `json:"host"`
	Port           *int     `json:"port"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	DatabaseName   string   `json:"databaseName"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	FilePath       string   `json:"filePath"`
	OpenMode       string   `json:"openMode"`
	ColorLabel     string   `json:"colorLabel"`
}

type v2UpdateDsReq struct {
	Name         string   `json:"name"`
	Env          string   `json:"env"`
	Host         string   `json:"host"`
	Port         *int     `json:"port"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	DatabaseName string   `json:"databaseName"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	FilePath     string   `json:"filePath"`
	OpenMode     string   `json:"openMode"`
	ColorLabel   string   `json:"colorLabel"`
	Status       string   `json:"status"`
}

type v2TestConnReq struct {
	DBType       string `json:"dbType"`
	Type         string `json:"type"`
	DatasourceType string `json:"datasourceType"`
	Host         string `json:"host"`
	Port         *int   `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"databaseName"`
	FilePath     string `json:"filePath"`
}

// GetMatrix - GET /api/datasource/matrix
func (h *DatasourceHandler) GetMatrix(c *gin.Context) {
	result, err := h.dsSvc.GetMatrix()
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, result)
}

// ListDatasourceV2 - GET /api/datasource/listDatasource
func (h *DatasourceHandler) ListDatasourceV2(c *gin.Context) {
	keyword := c.Query("keyword")
	dbType := c.Query("type")
	if dbType == "" {
		dbType = c.Query("dbType")
	}
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := h.dsSvc.ListDatasource(keyword, dbType, current, pageSize)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.SuccessList(c, list, total, current, pageSize)
}

// GetDatasourceInfoV2 - GET /api/datasource/:id/datasourceInfo
func (h *DatasourceHandler) GetDatasourceInfoV2(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetDatasourceInfo(id)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	middleware.Success(c, ds)
}

// CreateDatasourceV2 - POST /api/datasource/createDatasource
func (h *DatasourceHandler) CreateDatasourceV2(c *gin.Context) {
	var req v2CreateDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "请求参数错误: "+err.Error())
		return
	}

	userId := middleware.GetStr(c, "userId")

	// 转换请求参数到内部模型
	dbType := strings.ToLower(req.DBType)
	if dbType == "" {
		dbType = strings.ToLower(req.Type)
	}
	if dbType == "" {
		dbType = strings.ToLower(req.DatasourceType)
	}

	serviceReq := &service.CreateDatasourceReq{
		Name:           req.Name,
		DatasourceType: dbType,
		Type:           dbType,
		DBType:         dbType,
		Env:            req.Env,
		Host:           req.Host,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		DatabaseName:   req.DatabaseName,
		Description:    req.Description,
		Tags:           req.Tags,
		FilePath:       req.FilePath,
		OpenMode:       req.OpenMode,
		ColorLabel:     req.ColorLabel,
	}

	result, err := h.dsSvc.CreateDatasource(serviceReq, userId)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, result)
}

// UpdateDatasourceV2 - POST /api/datasource/:id/updateDatasource
func (h *DatasourceHandler) UpdateDatasourceV2(c *gin.Context) {
	id := c.Param("id")
	var req v2UpdateDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "请求参数错误")
		return
	}

	serviceReq := &service.UpdateDatasourceReq{
		Name:         req.Name,
		Env:          req.Env,
		Host:         req.Host,
		Port:         req.Port,
		Username:     req.Username,
		Password:     req.Password,
		DatabaseName: req.DatabaseName,
		Description:  req.Description,
		Tags:         req.Tags,
		FilePath:     req.FilePath,
		OpenMode:     req.OpenMode,
		ColorLabel:   req.ColorLabel,
		Status:       req.Status,
	}

	result, err := h.dsSvc.UpdateDatasource(id, serviceReq)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, result)
}

// DeleteDatasourceV2 - POST /api/datasource/:id/deleteDatasource
func (h *DatasourceHandler) DeleteDatasourceV2(c *gin.Context) {
	id := c.Param("id")
	err := h.dsSvc.DeleteDatasource(id)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, nil)
}

// TestConnectionV2 - POST /api/datasource/testConnection
func (h *DatasourceHandler) TestConnectionV2(c *gin.Context) {
	var req v2TestConnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "请求参数错误")
		return
	}

	dbType := strings.ToLower(req.DBType)
	if dbType == "" {
		dbType = strings.ToLower(req.Type)
	}
	if dbType == "" {
		dbType = strings.ToLower(req.DatasourceType)
	}

	serviceReq := &service.TestConnectionReq{
		DBType:       dbType,
		Host:         req.Host,
		Port:         req.Port,
		Username:     req.Username,
		Password:     req.Password,
		DatabaseName: req.DatabaseName,
		FilePath:     req.FilePath,
	}

	result, err := h.dsSvc.TestConnection(serviceReq)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, result)
}

// ListRecentlyDatasourceV2 - GET /api/datasource/listRecentlyDatasource
func (h *DatasourceHandler) ListRecentlyDatasourceV2(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))
	result, err := h.dsSvc.ListRecentlyDatasource(limit)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, result)
}

// ==================== SQL IDE V2 API 处理器 ====================

// ExecuteSQLV2 - POST /api/dataquery/sql/execute
func (h *SQLHandler) ExecuteSQLV2(c *gin.Context) {
	var req executeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "请求参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(req.DatasourceID)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	result, err := h.sqlSvc.Execute(ds, req.Database, req.SQL, req.IgnoreRisk, userId, username)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, result)
}

// HistoryV2 - GET /api/dataquery/sqlHistory/list
func (h *SQLHandler) HistoryV2(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	datasourceId := c.Query("datasourceId")
	keyword := c.Query("keyword")
	list, total, err := h.sqlSvc.GetHistory(current, pageSize, datasourceId, keyword)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.SuccessList(c, list, total, current, pageSize)
}

// GetDatabasesV2 - GET /api/dataquery/databases/:id
func (h *SQLHandler) GetDatabasesV2(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	dbs, err := h.sqlSvc.GetDatabases(ds)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, dbs)
}

// GetTablesV2 - GET /api/dataquery/tables/:id
func (h *SQLHandler) GetTablesV2(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	tables, err := h.sqlSvc.GetTables(ds, dbName)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, tables)
}

// GetColumnsV2 - GET /api/dataquery/columns/:id
func (h *SQLHandler) GetColumnsV2(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	cols, err := h.sqlSvc.GetColumns(ds, dbName, tableName)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, cols)
}

// GetFullTreeV2 - GET /api/dataquery/tree/:id
func (h *SQLHandler) GetFullTreeV2(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	tree, err := h.sqlSvc.GetFullTree(ds)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, tree)
}

// GetTableInfoV2 - GET /api/dataquery/table-info/:id
func (h *SQLHandler) GetTableInfoV2(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	info, err := h.sqlSvc.GetTableInfo(ds, dbName, tableName)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, info)
}

// ExplainSQLV2 - POST /api/dataquery/sql/explain
func (h *SQLHandler) ExplainSQLV2(c *gin.Context) {
	var req executeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "请求参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(req.DatasourceID)
	if err != nil {
		middleware.Error(c, "数据源不存在")
		return
	}
	result, err := h.sqlSvc.Explain(ds, req.Database, req.SQL)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}
	middleware.Success(c, result)
}
