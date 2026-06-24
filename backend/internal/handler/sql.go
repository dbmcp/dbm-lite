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
	sqlSvc     *service.SQLService
	dsSvc      *service.DatasourceService
	savedSvc   *service.SavedQueryService
	favoriteSvc *service.SQLFavoriteService
}

func NewSQLHandler() *SQLHandler {
	return &SQLHandler{
		sqlSvc:      service.NewSQLService(),
		dsSvc:       service.NewDatasourceService(),
		savedSvc:    service.NewSavedQueryService(),
		favoriteSvc: service.NewSQLFavoriteService(),
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

	executionId, result, err := h.sqlSvc.ExecuteWithCancel(ds, req.Database, req.SQL, req.IgnoreRisk, userId, username)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, gin.H{
		"executionId": string(executionId),
		"results":     result,
	})
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

func (h *SQLHandler) CancelExecute(c *gin.Context) {
	var req struct {
		ExecutionID string `json:"executionId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	
	if req.ExecutionID == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "执行ID不能为空")
		return
	}
	
	success := service.CancelExecution(service.ExecutionID(req.ExecutionID))
	if success {
		middleware.OK(c, gin.H{"success": true, "message": "已发送取消请求"})
	} else {
		middleware.OK(c, gin.H{"success": false, "message": "未找到执行任务或已完成"})
	}
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
	userId := middleware.GetStr(c, "userId")
	list, total, err := h.sqlSvc.GetHistory(page, pageSize, datasourceId, keyword, userId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

// ===== 保存查询（SavedQuery）接口 =====

type savedQueryReq struct {
	DatasourceID string `json:"datasourceId"`
	Database     string `json:"database"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	SQL          string `json:"sql"`
}

func (h *SQLHandler) ListSavedQueries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	datasourceId := c.Query("datasourceId")
	keyword := c.Query("keyword")
	list, total, err := h.savedSvc.List(datasourceId, page, pageSize, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

func (h *SQLHandler) GetSavedQuery(c *gin.Context) {
	qid := c.Param("id")
	q, err := h.savedSvc.Get(qid)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "查询不存在")
		return
	}
	middleware.OK(c, q)
}

func (h *SQLHandler) SaveQuery(c *gin.Context) {
	var req savedQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "SQL 内容不能为空")
		return
	}
	userID, _ := c.Get("userId")
	username, _ := c.Get("username")
	uid, _ := userID.(string)
	uname, _ := username.(string)
	q, err := h.savedSvc.Save(uid, uname, req.DatasourceID, req.Database, req.Title, req.Description, req.SQL)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, q)
}

func (h *SQLHandler) UpdateSavedQuery(c *gin.Context) {
	qid := c.Param("id")
	var req savedQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	if err := h.savedSvc.Update(qid, req.Title, req.Description, req.SQL, req.Database); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"queryId": qid})
}

func (h *SQLHandler) DeleteSavedQuery(c *gin.Context) {
	qid := c.Param("id")
	if err := h.savedSvc.Delete(qid); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"queryId": qid})
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

func getCurrentUser(c *gin.Context) (string, string) {
	userId := ""
	if v, ok := c.Get("userId"); ok {
		if s, ok := v.(string); ok {
			userId = s
		}
	}
	username := ""
	if v, ok := c.Get("username"); ok {
		if s, ok := v.(string); ok {
			username = s
		}
	}
	return userId, username
}

type createReq struct {
	Name             string      `json:"name"`
	DBType           string      `json:"dbType"`
	Host             string      `json:"host"`
	Port             int         `json:"port"`
	Username         string      `json:"username"`
	Password         string      `json:"password"`
	DefaultDB        string      `json:"defaultDatabase"`
	FilePath         string      `json:"filePath"`
	OpenMode         string      `json:"openMode"`
	Charset          string      `json:"charset"`
	Timezone         string      `json:"timezone"`
	SSLMode          interface{} `json:"sslMode"`
	SSLCAFile        string      `json:"sslCaFile"`
	ColorLabel       string      `json:"colorLabel"`
	Tags             string      `json:"tags"`
	BusinessID       string      `json:"businessId"`
	ServerID         string      `json:"serverId"`
	ProjectID        string      `json:"projectId"`
	Env              string      `json:"env"`
	Remark           string      `json:"remark"`
	Timeout          int         `json:"timeout"`
	AutoTest         bool        `json:"autoTest"`
	AutoCreateServer bool        `json:"autoCreateServer"`
	ConnStatus       string      `json:"connStatus"`
	ConnLatencyMs    int64       `json:"connLatencyMs"`
	Status           string      `json:"status"`
	CreatedAt        string      `json:"createdAt"`
	UpdatedAt        string      `json:"updatedAt"`
	OwnerID          string      `json:"ownerId"`
	OrgID            string      `json:"orgId"`
	DatasourceID     string      `json:"datasourceId"`
	DatasourceType   string      `json:"datasourceType"`
	Type             string      `json:"type"`
	ReadOnly         bool        `json:"readOnly"`
	Version          string      `json:"version"`
	CreatedBy        string      `json:"createdBy"`
	CreatedByName    string      `json:"createdByName"`
	LastConnTestAt   string      `json:"lastConnTestAt,omitempty"`
	LastUseTime      string      `json:"lastUseTime,omitempty"`
}

func convertSSLMode(v interface{}) string {
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case string:
		return val
	default:
		return "false"
	}
}

func (h *DatasourceHandler) Create(c *gin.Context) {
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "名称不能为空")
		return
	}
	if !model.IsSupportedDBType(strings.ToLower(req.DBType)) {
		middleware.Fail(c, http.StatusBadRequest, 400, "不支持的数据库类型")
		return
	}
	userId, username := getCurrentUser(c)
	// autoCreateServer: 如果启用且未选择已有服务器，则根据主机自动创建
	if req.AutoCreateServer && req.ServerID == "" && req.Host != "" {
		if srv, err := h.srvSvc.EnsureByHost(req.Host, userId); err == nil && srv != nil {
			req.ServerID = srv.ServerID
		}
	}
	ds := &model.Datasource{
		Name:       req.Name,
		DBType:     strings.ToLower(req.DBType),
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		DefaultDB:  req.DefaultDB,
		FilePath:   req.FilePath,
		OpenMode:   req.OpenMode,
		Charset:    req.Charset,
		Timezone:   req.Timezone,
		SSLMode:    convertSSLMode(req.SSLMode),
		SSLCAFile:  req.SSLCAFile,
		ColorLabel: req.ColorLabel,
		Tags:       req.Tags,
		BusinessID: req.BusinessID,
		ServerID:   req.ServerID,
		ProjectID:  req.ProjectID,
		Env:        req.Env,
		Remark:     req.Remark,
		Timeout:    req.Timeout,
		ReadOnly:   req.ReadOnly,
		ConnStatus: model.ConnStatusNone,
		Status:     "active",
	}
	if err := h.dsSvc.Create(ds, req.Password, userId, username); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, "创建失败: "+err.Error())
		return
	}
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), "datasource", "create", ds.DatasourceID, "创建数据源: "+ds.Name, model.AuditResultSuccess, "")
	middleware.OK(c, gin.H{"datasourceId": ds.DatasourceID})
}

func (h *DatasourceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	dbType := c.Query("dbType")
	status := c.Query("status")
	sortBy := c.Query("sortBy")
	list, total, err := h.dsSvc.List(page, pageSize, keyword, dbType, status, sortBy)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *DatasourceHandler) AllSimple(c *gin.Context) {
	list, _, err := h.dsSvc.List(1, 1000, "", "", "", "name")
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	result := make([]map[string]interface{}, 0, len(list))
	for _, d := range list {
		result = append(result, map[string]interface{}{
			"datasourceId": d.DatasourceID,
			"name":         d.Name,
			"dbType":       d.DBType,
		})
	}
	middleware.OK(c, result)
}

func (h *DatasourceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	// 支持 withPassword 查询参数控制是否返回解密密码，编辑页面会传 1
	wantPassword := false
	if v := c.Query("withPassword"); v == "1" || strings.ToLower(v) == "true" {
		wantPassword = true
	}
	var (
		ds      *model.Datasource
		err     error
		pwdText string
	)
	if wantPassword {
		ds, err = h.dsSvc.GetById(id)
		if ds != nil {
			// GetById 已解密，保存明文用于返回
			pwdText = ds.Password
		}
	} else {
		ds, err = h.dsSvc.GetByIdNoDecrypt(id)
	}
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	if wantPassword && pwdText != "" {
		middleware.OK(c, gin.H{
			"datasourceId":    ds.DatasourceID,
			"name":            ds.Name,
			"dbType":          ds.DBType,
			"host":            ds.Host,
			"port":            ds.Port,
			"username":        ds.Username,
			"password":        pwdText,
			"defaultDatabase": ds.DefaultDB,
			"filePath":        ds.FilePath,
			"openMode":        ds.OpenMode,
			"charset":         ds.Charset,
			"timezone":        ds.Timezone,
			"sslMode":         ds.SSLMode,
			"sslCaFile":       ds.SSLCAFile,
			"readOnly":        ds.ReadOnly,
			"colorLabel":      ds.ColorLabel,
			"tags":            ds.Tags,
			"businessId":      ds.BusinessID,
			"serverId":        ds.ServerID,
			"projectId":       ds.ProjectID,
			"env":             ds.Env,
			"remark":          ds.Remark,
			"timeout":         ds.Timeout,
			"connStatus":      ds.ConnStatus,
			"connLatencyMs":   ds.ConnLatencyMs,
			"status":          ds.Status,
			"version":         ds.Version,
			"createdAt":       ds.CreatedAt,
			"updatedAt":       ds.UpdatedAt,
			"lastConnTestAt":  ds.LastConnTestAt,
			"ownerId":         ds.OwnerID,
			"createdBy":       ds.CreatedBy,
		})
		return
	}
	middleware.OK(c, ds)
}

func (h *DatasourceHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetByIdNoDecrypt(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	middleware.OK(c, ds)
}

func (h *DatasourceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	// 使用 raw map 解析以避免 model.Datasource 中 Password 的 json:"-" tag 导致丢失密码
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	// 单独提取密码
	var newPassword string
	if p, ok := raw["password"]; ok {
		if s, ok := p.(string); ok {
			newPassword = s
		}
	}
	// 处理 autoCreateServer 字段：如果启用且 serverId 为空，自动创建对应服务器
	shouldAutoCreate := false
	if v, ok := raw["autoCreateServer"]; ok {
		if b, ok := v.(bool); ok {
			shouldAutoCreate = b
		}
	}
	if shouldAutoCreate {
		serverID := ""
		if s, ok := raw["serverId"]; ok {
			if s2, ok := s.(string); ok {
				serverID = s2
			}
		}
		host := ""
		if h2, ok := raw["host"]; ok {
			if s2, ok := h2.(string); ok {
				host = s2
			}
		}
		if serverID == "" && host != "" {
			userId := middleware.GetStr(c, "userId")
			if srv, err := h.srvSvc.EnsureByHost(host, userId); err == nil && srv != nil {
				raw["serverId"] = srv.ServerID
			}
		}
	}
	if err := h.dsSvc.Update(id, raw, newPassword); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "数据源 ID 不能为空")
		return
	}
	if err := h.dsSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, "删除失败: "+err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), "datasource", "delete", id, "删除数据源: "+id, model.AuditResultSuccess, "")
	middleware.OK(c, gin.H{"deleted": id})
}

func (h *DatasourceHandler) Copy(c *gin.Context) {
	id := c.Param("id")
	userId, username := getCurrentUser(c)
	ds, err := h.dsSvc.Copy(id, userId, username)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, ds)
}

func (h *DatasourceHandler) TestConnectionById(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	result := h.sqlSvc.TestConnection(ds)
	latencyMs := int64(0)
	if result.LatencyMs > 0 {
		latencyMs = result.LatencyMs
	}
	connStatus := model.ConnStatusFail
	if result.Success {
		connStatus = model.ConnStatusOK
	}
	_ = h.dsSvc.UpdateConnStatus(id, connStatus, latencyMs, result.Version)
	middleware.OK(c, gin.H{"success": result.Success, "message": result.Message, "version": result.Version, "latencyMs": result.LatencyMs})
}

func (h *DatasourceHandler) TestConnectionFromForm(c *gin.Context) {
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	ds := &model.Datasource{
		DBType:    strings.ToLower(req.DBType),
		Host:      req.Host,
		Port:      req.Port,
		Username:  req.Username,
		Password:  req.Password,
		DefaultDB: req.DefaultDB,
		FilePath:  req.FilePath,
		OpenMode:  req.OpenMode,
		Charset:   req.Charset,
		Timezone:  req.Timezone,
		SSLMode:   convertSSLMode(req.SSLMode),
		SSLCAFile: req.SSLCAFile,
		Timeout:   req.Timeout,
	}
	result := h.sqlSvc.TestConnection(ds)
	middleware.OK(c, gin.H{"success": result.Success, "message": result.Message, "version": result.Version, "latencyMs": result.LatencyMs})
}

func (h *DatasourceHandler) Stats(c *gin.Context) {
	stats, err := h.dsSvc.Stats()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, stats)
}

// SQLHandler 扩展方法：IDE 相关表结构查询
func (h *SQLHandler) GetTableChildren(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
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

func (h *SQLHandler) GetRoutines(c *gin.Context) {
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

func (h *SQLHandler) GetTriggersForIde(c *gin.Context) {
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

func (h *SQLHandler) ListSystemDatabases(c *gin.Context) {
	dbType := c.DefaultQuery("dbType", "mysql")
	dbs := []string{}
	switch strings.ToLower(dbType) {
	case "mysql", "tidb":
		dbs = []string{"information_schema", "mysql", "performance_schema", "sys"}
	case "sqlite":
		dbs = []string{"sqlite_master"}
	}
	middleware.OK(c, gin.H{"databases": dbs})
}

func (h *SQLHandler) GetDatabasesFull(c *gin.Context) {
	id := c.Param("id")
	includeSystem, _ := strconv.ParseBool(c.DefaultQuery("includeSystem", "false"))
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
	middleware.OK(c, gin.H{"databases": dbs})
}

func (h *SQLHandler) GetIndexes(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	idxs, err := h.sqlSvc.GetIndexes(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, idxs)
}

func (h *SQLHandler) GetForeignKeys(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	fks, err := h.sqlSvc.GetForeignKeys(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, fks)
}

func (h *SQLHandler) GetTableTriggers(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	triggers, err := h.sqlSvc.GetTableTriggers(ds, dbName, tableName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, triggers)
}

func (h *SQLHandler) GetViewList(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "数据源不存在")
		return
	}
	views, err := h.sqlSvc.GetViewList(ds, dbName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, views)
}

// ===== 收藏管理接口 =====

func (h *SQLHandler) ListFavorites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	keyword := c.Query("keyword")
	userId := middleware.GetStr(c, "userId")
	
	list, total, err := h.favoriteSvc.List(userId, keyword, page, pageSize)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

func (h *SQLHandler) GetFavorite(c *gin.Context) {
	id := c.Param("id")
	userId := middleware.GetStr(c, "userId")
	
	fav, err := h.favoriteSvc.Get(userId, id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "收藏不存在")
		return
	}
	middleware.OK(c, fav)
}

func (h *SQLHandler) CreateFavorite(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		SQL         string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	
	if err := h.favoriteSvc.Create(userId, username, req.Title, req.Description, req.SQL); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"message": "添加成功"})
}

func (h *SQLHandler) UpdateFavorite(c *gin.Context) {
	id := c.Param("id")
	
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		SQL         string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	
	userId := middleware.GetStr(c, "userId")
	
	if err := h.favoriteSvc.Update(userId, id, req.Title, req.Description, req.SQL); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"message": "更新成功"})
}

func (h *SQLHandler) DeleteFavorite(c *gin.Context) {
	id := c.Param("id")
	userId := middleware.GetStr(c, "userId")
	
	if err := h.favoriteSvc.Delete(userId, id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"message": "删除成功"})
}
