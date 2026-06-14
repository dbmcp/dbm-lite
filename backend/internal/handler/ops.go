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

	"dbm-lite/internal/middleware"
	"dbm-lite/internal/model"
	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

type BusinessHandler struct {
	projSvc  *service.ProjectService
	bizSvc   *service.BusinessService
	auditSvc *service.AuditService
}

func NewBusinessHandler() *BusinessHandler {
	return &BusinessHandler{
		projSvc:  service.NewProjectService(),
		bizSvc:   service.NewBusinessService(),
		auditSvc: service.NewAuditService(),
	}
}

func (h *BusinessHandler) CreateProject(c *gin.Context) {
	var p model.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.projSvc.CreateProject(&p); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "create_project", p.ProjectID, "创建项目: "+p.Name, model.AuditResultSuccess, "")
	middleware.OK(c, p)
}

func (h *BusinessHandler) ListProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	keyword := c.Query("keyword")
	list, total, err := h.projSvc.ListProjects(page, pageSize, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "page": page, "pageSize": pageSize})
}

func (h *BusinessHandler) AllProjects(c *gin.Context) {
	list, err := h.projSvc.AllProjects()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

func (h *BusinessHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.projSvc.UpdateProject(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "update_project", id, "更新项目", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *BusinessHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if err := h.projSvc.DeleteProject(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "delete_project", id, "删除项目", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *BusinessHandler) CreateBusiness(c *gin.Context) {
	var b model.Business
	if err := c.ShouldBindJSON(&b); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.bizSvc.CreateBusiness(&b); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "create_business", b.BusinessID, "创建业务: "+b.Name, model.AuditResultSuccess, "")
	middleware.OK(c, b)
}

func (h *BusinessHandler) ListBusinesses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	keyword := c.Query("keyword")
	projectId := c.Query("projectId")
	list, total, err := h.bizSvc.ListBusinesses(page, pageSize, keyword, projectId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "page": page, "pageSize": pageSize})
}

func (h *BusinessHandler) AllBusinesses(c *gin.Context) {
	list, err := h.bizSvc.AllBusinesses()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

func (h *BusinessHandler) UpdateBusiness(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.bizSvc.UpdateBusiness(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "update_business", id, "更新业务", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *BusinessHandler) DeleteBusiness(c *gin.Context) {
	id := c.Param("id")
	if err := h.bizSvc.DeleteBusiness(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "delete_business", id, "删除业务", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

type ServerHandler struct {
	srvSvc   *service.ServerService
	auditSvc *service.AuditService
}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{
		srvSvc:   service.NewServerService(),
		auditSvc: service.NewAuditService(),
	}
}

func (h *ServerHandler) Create(c *gin.Context) {
	var s model.Server
	if err := c.ShouldBindJSON(&s); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	userId := middleware.GetStr(c, "userId")
	if err := h.srvSvc.Create(&s, userId); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	userId = middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "create", s.ServerID, "创建服务器: "+s.Name, model.AuditResultSuccess, "")
	middleware.OK(c, s)
}

func (h *ServerHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	list, total, err := h.srvSvc.List(page, pageSize, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "page": page, "pageSize": pageSize})
}

func (h *ServerHandler) All(c *gin.Context) {
	list, err := h.srvSvc.All()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

func getServerId(c *gin.Context) string {
	id := c.Param("serverid")
	if id == "" {
		id = c.Param("id")
	}
	return id
}

func (h *ServerHandler) Update(c *gin.Context) {
	id := getServerId(c)
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.srvSvc.Update(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "update", id, "更新服务器", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *ServerHandler) Delete(c *gin.Context) {
	id := getServerId(c)
	if err := h.srvSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "delete", id, "删除服务器", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *ServerHandler) TestConnect(c *gin.Context) {
	// 简单模拟测试连接，实际可扩展为真实SSH连接测试
	middleware.OK(c, gin.H{"success": true, "message": "测试成功（当前为模拟测试）"})
}

type AuditHandler struct {
	auditSvc *service.AuditService
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{auditSvc: service.NewAuditService()}
}

func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	module := c.Query("module")
	action := c.Query("action")
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")
	list, total, err := h.auditSvc.QueryList(page, pageSize, keyword, "", module, action, startTime, endTime)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "page": page, "pageSize": pageSize})
}

func (h *AuditHandler) Stats(c *gin.Context) {
	stats, err := h.auditSvc.Stats(30)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, stats)
}

type MaintenanceHandler struct {
	backupSvc  *service.BackupService
	inspectSvc *service.InspectService
	slowSvc    *service.SlowLogService
	haSvc      *service.HAService
	pluginSvc  *service.PluginService
	permSvc    *service.DBPermissionService
	capSvc     *service.CapacityService
	auditSvc   *service.AuditService
}

func NewMaintenanceHandler() *MaintenanceHandler {
	return &MaintenanceHandler{
		backupSvc:  service.NewBackupService(),
		inspectSvc: service.NewInspectService(),
		slowSvc:    service.NewSlowLogService(),
		haSvc:      service.NewHAService(),
		pluginSvc:  service.NewPluginService(),
		permSvc:    service.NewDBPermissionService(),
		capSvc:     service.NewCapacityService(),
		auditSvc:   service.NewAuditService(),
	}
}

// 备份策略
func (h *MaintenanceHandler) CreateBackupPolicy(c *gin.Context) {
	var p model.BackupPolicy
	if err := c.ShouldBindJSON(&p); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	userId := middleware.GetStr(c, "userId")
	if err := h.backupSvc.CreatePolicy(&p, userId); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBackup, "create_policy", p.PolicyID, "创建备份策略: "+p.Name, model.AuditResultSuccess, "")
	middleware.OK(c, p)
}

func (h *MaintenanceHandler) ListBackupPolicies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	datasourceId := c.Query("datasourceId")
	status := c.Query("status")
	list, total, err := h.backupSvc.ListPolicies(page, pageSize, keyword, datasourceId, status)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *MaintenanceHandler) UpdateBackupPolicy(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.backupSvc.UpdatePolicy(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBackup, "update_policy", id, "更新备份策略", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *MaintenanceHandler) DeleteBackupPolicy(c *gin.Context) {
	id := c.Param("id")
	if err := h.backupSvc.DeletePolicy(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBackup, "delete_policy", id, "删除备份策略", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *MaintenanceHandler) TriggerBackup(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	policyId, _ := req["policyId"].(string)
	policyName, _ := req["name"].(string)
	backupType, _ := req["backupType"].(string)
	if backupType == "" {
		backupType = "full"
	}
	record, err := h.backupSvc.TriggerBackup(policyId, policyName, backupType)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBackup, "trigger", record.RecordID, "触发备份", model.AuditResultSuccess, "")
	middleware.OK(c, record)
}

func (h *MaintenanceHandler) ListBackupRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	policyId := c.Query("policyId")
	list, total, err := h.backupSvc.ListRecords(page, pageSize, policyId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

// 巡检任务
func (h *MaintenanceHandler) CreateInspectTask(c *gin.Context) {
	var t model.InspectTask
	if err := c.ShouldBindJSON(&t); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	userId := middleware.GetStr(c, "userId")
	if err := h.inspectSvc.CreateTask(&t, userId); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, t)
}

func (h *MaintenanceHandler) ListInspectTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	datasourceId := c.Query("datasourceId")
	list, total, err := h.inspectSvc.ListTasks(page, pageSize, keyword, datasourceId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *MaintenanceHandler) UpdateInspectTask(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.inspectSvc.UpdateTask(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *MaintenanceHandler) DeleteInspectTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.inspectSvc.DeleteTask(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *MaintenanceHandler) TriggerInspect(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	taskId, _ := req["taskId"].(string)
	datasourceId, _ := req["datasourceId"].(string)
	report, err := h.inspectSvc.TriggerInspect(taskId, datasourceId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, report)
}

func (h *MaintenanceHandler) ListInspectReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	datasourceId := c.Query("datasourceId")
	list, total, err := h.inspectSvc.ListReports(page, pageSize, datasourceId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

// 慢日志
func (h *MaintenanceHandler) ListSlowLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	datasourceId := c.Query("datasourceId")
	keyword := c.Query("keyword")
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")
	list, total, err := h.slowSvc.List(page, pageSize, datasourceId, keyword, startTime, endTime)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *MaintenanceHandler) TopSlow(c *gin.Context) {
	datasourceId := c.Query("datasourceId")
	top, err := h.slowSvc.TopSlow(datasourceId, 10)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, top)
}

// HA管理
func (h *MaintenanceHandler) CreateHACluster(c *gin.Context) {
	var c2 model.HaCluster
	if err := c.ShouldBindJSON(&c2); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.haSvc.CreateCluster(&c2); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, c2)
}

func (h *MaintenanceHandler) ListHAClusters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	list, total, err := h.haSvc.ListClusters(page, pageSize, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *MaintenanceHandler) UpdateHACluster(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.haSvc.UpdateCluster(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *MaintenanceHandler) DeleteHACluster(c *gin.Context) {
	id := c.Param("id")
	if err := h.haSvc.DeleteCluster(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

// 插件管理
func (h *MaintenanceHandler) CreatePlugin(c *gin.Context) {
	var p model.Plugin
	if err := c.ShouldBindJSON(&p); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	userId := middleware.GetStr(c, "userId")
	if err := h.pluginSvc.Create(&p, userId); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, p)
}

func (h *MaintenanceHandler) ListPlugins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	module := c.Query("module")
	status := c.Query("status")
	list, total, err := h.pluginSvc.List(page, pageSize, keyword, module, status)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *MaintenanceHandler) UpdatePlugin(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.pluginSvc.Update(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *MaintenanceHandler) DeletePlugin(c *gin.Context) {
	id := c.Param("id")
	if err := h.pluginSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

// 容量管理
func (h *MaintenanceHandler) AnalyzeCapacity(c *gin.Context) {
	datasourceId := c.Query("datasourceId")
	dbName := c.Query("database")
	if datasourceId == "" {
		middleware.OK(c, gin.H{
			"summary":   gin.H{"totalDatabases": 0, "totalTables": 0, "totalSizeMB": 0, "totalRows": 0},
			"databases": []interface{}{},
			"message":   "请指定 datasourceId 参数以查询容量信息",
		})
		return
	}
	result, err := h.capSvc.Analyze(datasourceId, dbName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, result)
}

// DB用户权限
func (h *MaintenanceHandler) CreateDBUser(c *gin.Context) {
	var u model.DBUser
	if err := c.ShouldBindJSON(&u); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.permSvc.Create(&u); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, u)
}

func (h *MaintenanceHandler) ListDBUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	datasourceId := c.Query("datasourceId")
	keyword := c.Query("keyword")
	list, total, err := h.permSvc.List(page, pageSize, datasourceId, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *MaintenanceHandler) DeleteDBUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.permSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

// 平台操作审计（和平台级别与SQL审计分开，这里复用）
func (h *MaintenanceHandler) PlatformAudit(c *gin.Context) {
	h.auditSvc = service.NewAuditService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	module := c.Query("module")
	action := c.Query("action")
	list, total, err := h.auditSvc.QueryList(page, pageSize, keyword, "", module, action, "", "")
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

// ===================== 项目作用域路由 =====================

func (h *ServerHandler) ListByProject(c *gin.Context) {
	projectId := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	keyword := c.Query("keyword")
	list, total, err := h.srvSvc.ListByProject(page, pageSize, keyword, projectId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *ServerHandler) CreateByProject(c *gin.Context) {
	projectId := c.Param("id")
	var s model.Server
	if err := c.ShouldBindJSON(&s); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	s.ProjectID = projectId
	userId := middleware.GetStr(c, "userId")
	if err := h.srvSvc.Create(&s, userId); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "create", s.ServerID, "项目中创建服务器: "+s.Name, model.AuditResultSuccess, "")
	middleware.OK(c, s)
}

func (h *BusinessHandler) ListBusinessesByProject(c *gin.Context) {
	projectId := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	keyword := c.Query("keyword")
	list, total, err := h.bizSvc.ListBusinesses(page, pageSize, keyword, projectId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (h *BusinessHandler) CreateBusinessByProject(c *gin.Context) {
	projectId := c.Param("id")
	var b model.Business
	if err := c.ShouldBindJSON(&b); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	b.ProjectID = projectId
	if err := h.bizSvc.CreateBusiness(&b); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "create", b.BusinessID, "项目中创建业务: "+b.Name, model.AuditResultSuccess, "")
	middleware.OK(c, b)
}
