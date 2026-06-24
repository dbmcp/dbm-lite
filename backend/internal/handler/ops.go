/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	username := middleware.GetStr(c, "username")
	if err := h.srvSvc.Create(&s, userId); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "create", s.ServerID, "创建服务器: "+s.Name, model.AuditResultSuccess, "")
	middleware.OK(c, s)
}

func (h *ServerHandler) Get(c *gin.Context) {
	id := getServerId(c)
	srv, err := h.srvSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, http.StatusNotFound, 404, "服务器不存在")
		return
	}
	middleware.OK(c, srv)
}

func (h *ServerHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	env := c.Query("env")
	status := c.Query("status")
	list, total, err := h.srvSvc.List(page, pageSize, keyword, env, status)
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

// ToggleStatus 切换服务器启用 / 禁用
func (h *ServerHandler) ToggleStatus(c *gin.Context) {
	id := getServerId(c)
	newStatus, err := h.srvSvc.ToggleStatus(id)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "toggle_status", id, "切换状态为: "+newStatus, model.AuditResultSuccess, "")
	middleware.OK(c, gin.H{"status": newStatus})
}

// TestConnect 根据服务器 ID 测试 SSH 连接，同时回写连接状态
func (h *ServerHandler) TestConnect(c *gin.Context) {
	id := getServerId(c)
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	latency, info, err := h.srvSvc.TestConnect(id)
	if err != nil {
		h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "test_connect", id, "连接测试失败: "+err.Error(), model.AuditResultFailed, "")
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "test_connect", id, "连接测试成功", model.AuditResultSuccess, "")
	middleware.OK(c, gin.H{"latencyMs": latency, "info": info})
}

// TestConnectByForm 根据前端传入的连接参数测试 SSH（不写入库）
func (h *ServerHandler) TestConnectByForm(c *gin.Context) {
	var body struct {
		Host          string `json:"host"`
		Port          int    `json:"port"`
		Username      string `json:"username"`
		AuthType      string `json:"authType"`
		Password      string `json:"password"`
		PrivateKey    string `json:"privateKey"`
		KeyPassphrase string `json:"keyPassphrase"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	info, err := h.srvSvc.TestConnectByForm(body.Host, body.Port, body.Username, body.AuthType, body.Password, body.PrivateKey, body.KeyPassphrase)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	middleware.OK(c, gin.H{"info": info})
}

// ExecCommand 执行一条命令，返回 stdout / stderr
func (h *ServerHandler) ExecCommand(c *gin.Context) {
	id := getServerId(c)
	var body struct {
		Command string `json:"command"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if body.Command == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "命令不能为空")
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	stdout, stderr, err := h.srvSvc.ExecCommand(id, body.Command)
	if err != nil {
		h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "exec", id, "执行命令失败: "+body.Command, model.AuditResultFailed, "")
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleServer, "exec", id, "执行命令: "+body.Command, model.AuditResultSuccess, "")
	middleware.OK(c, gin.H{"stdout": stdout, "stderr": stderr})
}

// Stats 返回服务器列表页顶部简要统计
func (h *ServerHandler) Stats(c *gin.Context) {
	stats, err := h.srvSvc.Stats()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, stats)
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

// ==================== 项目成员 API ====================

func (h *BusinessHandler) ListProjectMembers(c *gin.Context) {
	projectId := c.Param("id")
	list, err := h.projSvc.ListProjectMembers(projectId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

type assignProjectMembersReq struct {
	UserIds []string `json:"userIds"`
	Role    string   `json:"role"`
}

func (h *BusinessHandler) AssignProjectMembers(c *gin.Context) {
	projectId := c.Param("id")
	var req assignProjectMembersReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if err := h.projSvc.AssignProjectMembers(projectId, req.UserIds, req.Role); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "assign_project_members", projectId, "分配项目成员", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *BusinessHandler) RemoveProjectMember(c *gin.Context) {
	projectId := c.Param("id")
	userId := c.Param("userId")
	if err := h.projSvc.RemoveProjectMember(projectId, userId); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	operatorId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(operatorId, username, middleware.GetClientIP(c), model.ModuleBusiness, "remove_project_member", projectId, "移除项目成员: "+userId, model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

// ==================== 业务成员 API ====================

func (h *BusinessHandler) ListBusinessMembers(c *gin.Context) {
	businessId := c.Param("id")
	list, err := h.bizSvc.ListBusinessMembers(businessId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

type assignBusinessMembersReq struct {
	UserIds []string `json:"userIds"`
	Role    string   `json:"role"`
}

func (h *BusinessHandler) AssignBusinessMembers(c *gin.Context) {
	businessId := c.Param("id")
	var req assignBusinessMembersReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if err := h.bizSvc.AssignBusinessMembers(businessId, req.UserIds, req.Role); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleBusiness, "assign_business_members", businessId, "分配业务成员", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *BusinessHandler) RemoveBusinessMember(c *gin.Context) {
	businessId := c.Param("id")
	userId := c.Param("userId")
	if err := h.bizSvc.RemoveBusinessMember(businessId, userId); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	operatorId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(operatorId, username, middleware.GetClientIP(c), model.ModuleBusiness, "remove_business_member", businessId, "移除业务成员: "+userId, model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

// ==================== 概览 API ====================

func (h *BusinessHandler) ProjectOverview(c *gin.Context) {
	projectId := c.Param("id")
	info, err := h.projSvc.Overview(projectId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, info)
}

func (h *BusinessHandler) BusinessOverview(c *gin.Context) {
	businessId := c.Param("id")
	info, err := h.bizSvc.Overview(businessId)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, info)
}

// ===================== OpsHandler 管理域页面补充接口 =====================

type OpsHandler struct {
	auditSvc *service.AuditService
}

func NewOpsHandler() *OpsHandler {
	return &OpsHandler{auditSvc: service.NewAuditService()}
}

func (h *OpsHandler) log(c *gin.Context, module, action, object, content, result string) {
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), module, action, object, content, result, "")
}

// ------------ 导入导出 ------------
func (h *OpsHandler) CreateImportExportTask(c *gin.Context) {
	var body struct {
		DatasourceID string   `json:"datasourceId"`
		Mode         string   `json:"mode"` // import / export
		Scope        string   `json:"scope"` // schema / data / all
		Tables       []string `json:"tables"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if body.DatasourceID == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "请先选择数据源")
		return
	}
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	record := gin.H{
		"id":           "TASK" + time.Now().Format("20060102150405"),
		"type":         body.Mode,
		"datasourceId": body.DatasourceID,
		"scope":        body.Scope,
		"tables":       strings.Join(body.Tables, ", "),
		"status":       "执行中",
		"createTime":   nowTime,
	}
	h.log(c, model.ModuleImportExport, "create_task", record["id"].(string), fmt.Sprintf("提交导入导出任务"), model.AuditResultSuccess)
	middleware.OK(c, record)
}

func (h *OpsHandler) ListImportExportTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if pageSize > 200 {
		pageSize = 200
	}
	list := []gin.H{}
	statuses := []string{"已完成", "执行中", "失败", "已完成", "已完成"}
	modes := []string{"export", "import", "export", "import", "export"}
	scopes := []string{"表结构", "数据", "结构+数据", "表结构", "数据"}
	for i := 0; i < 5; i++ {
		list = append(list, gin.H{
			"id":           "TASK" + fmt.Sprintf("%04d", i+1),
			"type":         modes[i],
			"datasourceId": "ds-" + fmt.Sprintf("%03d", i+1),
			"datasource":   fmt.Sprintf("演示数据源 %d (MySQL)", i+1),
			"scope":        scopes[i],
			"tables":       fmt.Sprintf("tbl_%d, tbl_%d_sub", i+1, i+1),
			"status":       statuses[i],
			"createTime":   time.Now().Format("2006-01-02 15:04:05"),
		})
	}
	middleware.OK(c, gin.H{
		"list":     list,
		"total":    len(list),
		"current":  page,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ------------ SQL 审核 ------------
func (h *OpsHandler) ListAuditFlows(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list := []gin.H{
		{"id": "SQLA-0001", "changeType": "DDL", "env": "prod", "sql": "ALTER TABLE orders ADD COLUMN remark VARCHAR(255);", "risk": "中等", "status": "已通过", "createTime": time.Now().Format("2006-01-02 15:04:05")},
		{"id": "SQLA-0002", "changeType": "DML", "env": "test", "sql": "UPDATE products SET price = price * 1.1;", "risk": "低危", "status": "审核中", "createTime": time.Now().Format("2006-01-02 15:04:05")},
		{"id": "SQLA-0003", "changeType": "DML", "env": "test", "sql": "DELETE FROM logs WHERE create_time < NOW();", "risk": "高危", "status": "已拒绝", "createTime": time.Now().Format("2006-01-02 15:04:05")},
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list), "current": page, "page": page, "pageSize": pageSize})
}

func (h *OpsHandler) CreateAuditFlow(c *gin.Context) {
	var body struct {
		DatasourceID string `json:"datasourceId"`
		ChangeType   string `json:"changeType"`
		Env          string `json:"env"`
		SQL          string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if strings.TrimSpace(body.SQL) == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "SQL 不能为空")
		return
	}
	risk := "低危"
	if ok, _ := regexp.MatchString(`(?i)DROP|TRUNCATE`, body.SQL); ok {
		risk = "高危"
	} else if ok, _ := regexp.MatchString(`(?i)ALTER|UPDATE|DELETE`, body.SQL); ok {
		risk = "中等"
	}
	record := gin.H{
		"id":         "SQLA" + time.Now().Format("20060102150405"),
		"changeType": body.ChangeType,
		"env":        body.Env,
		"sql":        body.SQL,
		"risk":       risk,
		"status":     "待审核",
		"createTime": time.Now().Format("2006-01-02 15:04:05"),
	}
	h.log(c, model.ModuleSQLAudit, "create_flow", record["id"].(string), "提交 SQL 审核", model.AuditResultSuccess)
	middleware.OK(c, record)
}

func (h *OpsHandler) ListAuditRules(c *gin.Context) {
	rules := []gin.H{
		{"name": "禁用无 WHERE 的 DELETE/UPDATE", "scope": "DML", "desc": "禁止无 WHERE 条件的全表更新/删除。", "level": "ERROR", "enabled": true},
		{"name": "禁止 DROP TABLE/TRUNCATE", "scope": "DDL", "desc": "在生产环境禁用危险的删除/清空表操作。", "level": "ERROR", "enabled": true},
		{"name": "ALTER TABLE 必须在线执行", "scope": "DDL", "desc": "大表 ALTER 需走在线 DDL/工具执行。", "level": "WARNING", "enabled": true},
		{"name": "表必须有主键", "scope": "DDL", "desc": "新建表必须包含主键。", "level": "ERROR", "enabled": true},
		{"name": "字段命名规范", "scope": "DDL", "desc": "字段名只能包含小写字母、数字与下划线。", "level": "WARNING", "enabled": false},
		{"name": "禁止 SELECT *", "scope": "DQL", "desc": "禁止 SELECT *，需显式列出字段。", "level": "WARNING", "enabled": true},
		{"name": "字符集与排序规则", "scope": "DDL", "desc": "新建表字符集需为 utf8mb4。", "level": "WARNING", "enabled": true},
	}
	middleware.OK(c, gin.H{"list": rules, "total": len(rules)})
}

// ------------ 敏感数据维护 ------------
func (h *OpsHandler) ListSensitiveData(c *gin.Context) {
	keyword := strings.ToLower(c.Query("keyword"))
	items := []gin.H{
		{"datasourceId": "ds-1", "datasource": "生产-订单库", "table": "users", "column": "phone", "dataType": "varchar(20)", "level": "L3", "maskRule": "手机号脱敏", "owner": "李工"},
		{"datasourceId": "ds-1", "datasource": "生产-订单库", "table": "users", "column": "id_card", "dataType": "varchar(20)", "level": "L4", "maskRule": "身份证脱敏", "owner": "李工"},
		{"datasourceId": "ds-1", "datasource": "生产-订单库", "table": "users", "column": "email", "dataType": "varchar(64)", "level": "L2", "maskRule": "邮箱脱敏", "owner": "李工"},
		{"datasourceId": "ds-2", "datasource": "测试-商品库", "table": "products", "column": "cost_price", "dataType": "decimal(10,2)", "level": "L3", "maskRule": "金额扰动", "owner": "王工"},
		{"datasourceId": "ds-3", "datasource": "本地-SQLite", "table": "customers", "column": "address", "dataType": "text", "level": "L2", "maskRule": "全部遮蔽", "owner": "赵工"},
	}
	filtered := make([]gin.H, 0, len(items))
	for _, it := range items {
		if keyword == "" {
			filtered = append(filtered, it)
			continue
		}
		tbl, _ := it["table"].(string)
		col, _ := it["column"].(string)
		if strings.Contains(strings.ToLower(tbl), keyword) || strings.Contains(strings.ToLower(col), keyword) {
			filtered = append(filtered, it)
		}
	}
	middleware.OK(c, gin.H{"list": filtered, "total": len(filtered)})
}

func (h *OpsHandler) CreateSensitiveData(c *gin.Context) {
	var body struct {
		DatasourceID string `json:"datasourceId"`
		Table        string `json:"table"`
		Column       string `json:"column"`
		DataType     string `json:"dataType"`
		Level        string `json:"level"`
		MaskRule     string `json:"maskRule"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if strings.TrimSpace(body.Table) == "" || strings.TrimSpace(body.Column) == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "表名和字段必填")
		return
	}
	record := gin.H{
		"datasourceId": body.DatasourceID,
		"table":        body.Table,
		"column":       body.Column,
		"dataType":     body.DataType,
		"level":        body.Level,
		"maskRule":     body.MaskRule,
		"owner":        middleware.GetStr(c, "username"),
		"createTime":   time.Now().Format("2006-01-02 15:04:05"),
	}
	h.log(c, model.ModuleSensitiveData, "create", body.Table+"."+body.Column, "新增敏感字段", model.AuditResultSuccess)
	middleware.OK(c, record)
}

func (h *OpsHandler) DeleteSensitiveData(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "id 不能为空")
		return
	}
	h.log(c, model.ModuleSensitiveData, "delete", id, "移除敏感字段", model.AuditResultSuccess)
	middleware.OK(c, nil)
}

// ------------ DB 生命周期管理 ------------
func (h *OpsHandler) ListLifecycleNodes(c *gin.Context) {
	nodes := []gin.H{
		{"key": "create", "title": "数据库创建与初始化", "status": "活跃", "color": "#409EFF", "icon": "DataLine", "desc": "新建数据库实例、表结构初始化、元数据注册。"},
		{"key": "upgrade", "title": "版本管理与升级", "status": "近期操作", "color": "#67C23A", "icon": "MagicStick", "desc": "记录数据库版本、Patch 历史、执行小版本升级或大版本迁移。"},
		{"key": "backup", "title": "备份与恢复", "status": "每日运行", "color": "#E6A23C", "icon": "Coin", "desc": "全量/增量备份、备份集校验与一键恢复演练。"},
		{"key": "monitor", "title": "运行监控", "status": "实时", "color": "#2E6BA8", "icon": "Histogram", "desc": "连接数、QPS、慢查询、主从延迟等核心指标监控。"},
		{"key": "capacity", "title": "容量与配额", "status": "本月达标", "color": "#8E44AD", "icon": "DataLine", "desc": "容量趋势预测、空间配额管理、存储回收建议。"},
		{"key": "params", "title": "参数维护", "status": "待优化 3 项", "color": "#F56C6C", "icon": "Setting", "desc": "集中维护核心参数 (innodb_buffer_pool_size/max_connections 等)。"},
		{"key": "ha", "title": "高可用管理", "status": "M-S 架构", "color": "#13C2C2", "icon": "RefreshRight", "desc": "主从/MGR/集群状态巡检、自动切换演练与切换历史。"},
		{"key": "offline", "title": "下线与销毁", "status": "流程审核中", "color": "#909399", "icon": "Delete", "desc": "归档、只读、下线、销毁全流程审批与数据清理。"},
	}
	middleware.OK(c, nodes)
}

func (h *OpsHandler) ListLifecycleDBs(c *gin.Context) {
	dbs := []gin.H{
		{"name": "dbm_orders_prod", "dbType": "MySQL", "version": "8.0.34", "env": "生产", "currentPhase": "参数维护", "haMode": "M-S", "params": "innodb_buffer_pool_size=8G, max_connections=2000", "status": "运行中", "createTime": time.Now().Format("2006-01-02 15:04:05")},
		{"name": "dbm_products_prod", "dbType": "TiDB", "version": "7.1.1", "env": "生产", "currentPhase": "运行监控", "haMode": "TiDB Cluster", "params": "tidb_mem_quota_query=8G, raftstore.apply-pool-size=4", "status": "运行中", "createTime": time.Now().Format("2206-01-02 15:04:05")},
		{"name": "dbm_uat", "dbType": "MySQL", "version": "5.7.42", "env": "预发布", "currentPhase": "版本管理与升级", "haMode": "M-S", "params": "innodb_buffer_pool_size=4G, slow_query_log=1", "status": "运行中", "createTime": time.Now().Format("2206-01-02 15:04:05")},
		{"name": "dbm_test_local", "dbType": "SQLite", "version": "3.41", "env": "测试", "currentPhase": "数据库创建与初始化", "haMode": "单机", "params": "journal_mode=WAL, synchronous=NORMAL", "status": "运行中", "createTime": time.Now().Format("2206-01-02 15:04:05")},
		{"name": "dbm_log_prod", "dbType": "MySQL", "version": "8.0.33", "env": "生产", "currentPhase": "容量与配额", "haMode": "MGR", "params": "innodb_buffer_pool_size=16G, group_concat_max_len=10240", "status": "运行中", "createTime": time.Now().Format("2206-01-02 15:04:05")},
		{"name": "dbm_archive", "dbType": "MySQL", "version": "5.7.42", "env": "预发布", "currentPhase": "下线与销毁", "haMode": "单机", "params": "read_only=1, innodb_buffer_pool_size=2G", "status": "维护中", "createTime": time.Now().Format("2206-01-02 15:04:05")},
	}
	middleware.OK(c, gin.H{"list": dbs, "total": len(dbs)})
}

// ------------ 健康巡检 ------------
func (h *OpsHandler) HealthMetrics(c *gin.Context) {
	middleware.OK(c, []gin.H{
		{"title": "在线实例", "value": "6", "sub": "共 6 台", "percent": 100, "color": "#67C23A"},
		{"title": "平均 QPS", "value": "8,243", "sub": "近 5 分钟", "percent": 70, "color": "#409EFF"},
		{"title": "总连接数", "value": "1,234", "sub": "峰值 2,180", "percent": 56, "color": "#E6A23C"},
		{"title": "慢查询", "value": "18", "sub": "近 1 小时", "percent": 24, "color": "#F56C6C"},
	})
}

func (h *OpsHandler) HealthInstances(c *gin.Context) {
	list := []gin.H{
		{"name": "dbm_orders_prod (10.0.1.12:3306)", "qps": 1234, "connections": 324, "bufferHit": 98, "replicationLag": 0, "slowQueries": 3, "diskUsage": 62, "status": "正常"},
		{"name": "dbm_products_prod (10.0.2.21:4000)", "qps": 3421, "connections": 612, "bufferHit": 96, "replicationLag": 1, "slowQueries": 6, "diskUsage": 48, "status": "正常"},
		{"name": "dbm_log_prod (10.0.3.8:3306)", "qps": 892, "connections": 188, "bufferHit": 91, "replicationLag": 5, "slowQueries": 8, "diskUsage": 83, "status": "关注"},
		{"name": "dbm_uat (10.0.4.4:3306)", "qps": 324, "connections": 66, "bufferHit": 94, "replicationLag": 0, "slowQueries": 1, "diskUsage": 35, "status": "正常"},
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *OpsHandler) HealthInspectResults(c *gin.Context) {
	list := []gin.H{
		{"time": time.Now().Format("2006-01-02 15:04:05"), "env": "生产", "instance": "dbm_orders_prod", "item": "参数合规", "level": "INFO", "detail": "innodb_buffer_pool_size / max_connections 均在推荐范围"},
		{"time": time.Now().Format("2006-01-02 15:04:05"), "env": "生产", "instance": "dbm_log_prod", "item": "磁盘空间", "level": "WARN", "detail": "/data 使用率 83%, 建议启动扩容评估"},
		{"time": time.Now().Format("2006-01-02 15:04:05"), "env": "生产", "instance": "dbm_products_prod", "item": "备份状态", "level": "INFO", "detail": "最近一次全量备份 12 小时前, 校验通过"},
		{"time": time.Now().Format("2006-01-02 15:04:05"), "env": "预发布", "instance": "dbm_uat", "item": "主从延迟", "level": "ERROR", "detail": "从库延迟 58s, 超过阈值 30s"},
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *OpsHandler) TriggerHealthInspect(c *gin.Context) {
	var body struct {
		Env    string   `json:"env"`
		Items  []string `json:"items"`
		Target string   `json:"target"`
	}
	_ = c.ShouldBindJSON(&body)
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	items := "参数合规/备份状态/连接状态/主从延迟/磁盘空间"
	middleware.OK(c, gin.H{"time": nowTime, "items": items, "target": body.Target, "status": "已提交"})
}

// ------------ 数据迁移 ------------
func (h *OpsHandler) ListMigrationTasks(c *gin.Context) {
	list := []gin.H{
		{"id": "MIG-0001", "source": "生产-订单库 (MySQL)", "target": "测试-商品库 (TiDB)", "mode": "全量", "progress": 82, "status": "运行中", "createTime": time.Now().Format("2006-01-02 15:04:05")},
		{"id": "MIG-0002", "source": "归档库 (MySQL)", "target": "本地-SQLite", "mode": "增量", "progress": 100, "status": "完成", "createTime": time.Now().Format("2006-01-02 15:04:05")},
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *OpsHandler) CreateMigrationTask(c *gin.Context) {
	var body struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Mode   string `json:"mode"`
		Tables string `json:"tables"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if strings.TrimSpace(body.Source) == "" || strings.TrimSpace(body.Target) == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "源端和目标端必填")
		return
	}
	record := gin.H{
		"id":         "MIG" + time.Now().Format("20060102150405"),
		"source":     body.Source,
		"target":     body.Target,
		"mode":       body.Mode,
		"tables":     body.Tables,
		"progress":   5,
		"status":     "运行中",
		"createTime": time.Now().Format("2006-01-02 15:04:05"),
	}
	h.log(c, model.ModuleMigration, "create_task", record["id"].(string), "提交数据迁移任务", model.AuditResultSuccess)
	middleware.OK(c, record)
}

func (h *OpsHandler) SchemaDiff(c *gin.Context) {
	list := []gin.H{
		{"table": "orders", "type": "一致", "detail": "字段、索引、约束均一致"},
		{"table": "order_items", "type": "字段不一致", "detail": "源端多出字段 remark VARCHAR(255)"},
		{"table": "products", "type": "新增", "detail": "源端存在，目标端缺失该表"},
		{"table": "users", "type": "索引不一致", "detail": "目标端缺失 idx_user_name 索引"},
		{"table": "payments", "type": "一致", "detail": "字段、索引、约束均一致"},
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *OpsHandler) DataDiff(c *gin.Context) {
	list := []gin.H{
		{"table": "orders", "sourceCount": 134567, "targetCount": 134512, "diff": 55, "percent": 99},
		{"table": "order_items", "sourceCount": 5891243, "targetCount": 5891243, "diff": 0, "percent": 100},
		{"table": "users", "sourceCount": 218341, "targetCount": 218301, "diff": 40, "percent": 99},
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list)})
}

// ------------ 平台配置 介质维护 ------------
func (h *OpsHandler) ListMediums(c *gin.Context) {
	list := []gin.H{
		{"name": "MySQL 社区版", "type": "MySQL", "version": "8.0.36", "os": "Linux x86_64", "size": "486 MB", "uploader": "admin", "uploadTime": time.Now().Format("2006-01-02 15:04:05"), "status": "已发布"},
		{"name": "MySQL 社区版", "type": "MySQL", "version": "5.7.44", "os": "Linux x86_64", "size": "412 MB", "uploader": "admin", "uploadTime": time.Now().Format("2006-01-02 15:04:05"), "status": "已发布"},
		{"name": "TiDB 社区版", "type": "TiDB", "version": "7.1.2", "os": "Linux x86_64", "size": "1.2 GB", "uploader": "admin", "uploadTime": time.Now().Format("2006-01-02 15:04:05"), "status": "已发布"},
		{"name": "TiDB 社区版", "type": "TiDB", "version": "6.5.4", "os": "Linux x86_64", "size": "1.1 GB", "uploader": "admin", "uploadTime": time.Now().Format("2006-01-02 15:04:05"), "status": "已归档"},
		{"name": "SQLite 预编译", "type": "SQLite", "version": "3.44.2", "os": "跨平台", "size": "4.2 MB", "uploader": "admin", "uploadTime": time.Now().Format("2006-01-02 15:04:05"), "status": "已发布"},
		{"name": "MySQL Percona Server", "type": "MySQL", "version": "8.0.34-26", "os": "Linux x86_64", "size": "512 MB", "uploader": "admin", "uploadTime": time.Now().Format("2006-01-02 15:04:05"), "status": "已发布"},
	}
	usage := []gin.H{
		{"type": "MySQL", "version": "8.0.36", "instances": 12, "percent": 60, "latest": true},
		{"type": "MySQL", "version": "5.7.44", "instances": 5, "percent": 25, "latest": false},
		{"type": "TiDB", "version": "7.1.2", "instances": 2, "percent": 10, "latest": true},
		{"type": "SQLite", "version": "3.44", "instances": 1, "percent": 5, "latest": true},
	}
	middleware.OK(c, gin.H{"list": list, "usage": usage, "total": len(list)})
}
