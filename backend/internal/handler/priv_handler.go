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

	"dbm-lite/internal/database"
	"dbm-lite/internal/middleware"
	"dbm-lite/internal/model"
	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

// ====== 对象权限管理 Handler ======

type PrivilegeHandler struct {
	applySvc *service.PrivApplyService
	sensSvc  *service.SensitiveColumnService
}

func NewPrivilegeHandler() *PrivilegeHandler {
	return &PrivilegeHandler{
		applySvc: service.NewPrivApplyService(),
		sensSvc:  service.NewSensitiveColumnService(),
	}
}

// ==================== 权限申请与审批 ====================

type privApplySubmitReq struct {
	DatasourceID  string   `json:"datasourceId"`
	DatabaseName  string   `json:"databaseName"`
	TableName     string   `json:"tableName"`
	PrivType      string   `json:"privType"`
	OperationType string   `json:"operationType"`
	Columns       []string `json:"columns"`
	ValidDays     int      `json:"validDays"`
	RowLimit      int      `json:"rowLimit"`
	ApplyRemark   string   `json:"applyRemark"`
}

func (h *PrivilegeHandler) SubmitApply(c *gin.Context) {
	var req privApplySubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if req.DatasourceID == "" || req.DatabaseName == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "数据源与库名必填")
		return
	}
	userID := middleware.GetStr(c, "userId")
	userName := middleware.GetStr(c, "username")

	columnsStr := ""
	if len(req.Columns) > 0 {
		columnsStr = strings.Join(req.Columns, ",")
	}
	a := &model.QueryPrivApply{
		DatasourceID: req.DatasourceID,
		DatabaseName: req.DatabaseName,
		TblName:      req.TableName,
		PrivType:     req.PrivType,
		OperationType: req.OperationType,
		Columns:       columnsStr,
		ApplyRemark:   req.ApplyRemark,
	}
	if a.PrivType == "" {
		a.PrivType = "table"
	}
	if a.OperationType == "" {
		a.OperationType = "select"
	}
	validDays := req.ValidDays
	if validDays <= 0 {
		validDays = 7
	}
	rowLimit := req.RowLimit
	if rowLimit <= 0 {
		rowLimit = 0
	}
	if err := h.applySvc.Submit(a, userID, userName, validDays, rowLimit); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, a)
}

func (h *PrivilegeHandler) ListApplies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.Query("status")
	applicantID := c.Query("applicantId")
	keyword := c.Query("keyword")
	list, total, err := h.applySvc.List(page, pageSize, status, applicantID, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

type privApprovalReq struct {
	Remark string `json:"remark"`
}

func (h *PrivilegeHandler) ApproveApply(c *gin.Context) {
	id := c.Param("id")
	var req privApprovalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	userID := middleware.GetStr(c, "userId")
	userName := middleware.GetStr(c, "username")
	_, err := h.applySvc.Approve(id, userID, userName, req.Remark)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *PrivilegeHandler) RejectApply(c *gin.Context) {
	id := c.Param("id")
	var req privApprovalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	userID := middleware.GetStr(c, "userId")
	userName := middleware.GetStr(c, "username")
	_, err := h.applySvc.Reject(id, userID, userName, req.Remark)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

// ==================== 权限清单与直接授权 ====================

func (h *PrivilegeHandler) MyPrivileges(c *gin.Context) {
	userID := middleware.GetStr(c, "userId")
	dsID := c.Query("datasourceId")
	var list []model.QueryPrivilege
	if dsID != "" {
		privs, err := h.applySvc.GetUserEffectivePrivileges(userID, dsID)
		if err != nil {
			middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
			return
		}
		list = privs
	} else {
		var err error
		list, _, err = h.applySvc.ListAllEffectivePrivileges(1, 1000, userID, "", "", "", false)
		if err != nil {
			middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
			return
		}
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *PrivilegeHandler) ListAllPrivileges(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	userID := c.Query("userId")
	dsID := c.Query("datasourceId")
	dbName := c.Query("databaseName")
	tableName := c.Query("tableName")
	status := c.Query("status")
	onlyExpired := status == "expired"
	list, total, err := h.applySvc.ListAllEffectivePrivileges(page, pageSize, userID, dsID, dbName, tableName, onlyExpired)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

func (h *PrivilegeHandler) RevokePrivilege(c *gin.Context) {
	id := c.Param("id")
	if err := h.applySvc.Revoke(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

type batchRevokeReq struct {
	Ids []string `json:"ids"`
}

func (h *PrivilegeHandler) BatchRevokePrivilege(c *gin.Context) {
	var req batchRevokeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if err := h.applySvc.RevokeByIDs(req.Ids); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

type privGrantReq struct {
	UserID        string   `json:"userId"`
	UserName      string   `json:"userName"`
	DatasourceID  string   `json:"datasourceId"`
	DatabaseName  string   `json:"databaseName"`
	TableName     string   `json:"tableName"`
	PrivType      string   `json:"privType"`
	OperationType string   `json:"operationType"`
	Columns       []string `json:"columns"`
	RowLimit      int      `json:"rowLimit"`
	ValidDays     int      `json:"validDays"`
}

// GrantPrivilege 直接授权单用户
func (h *PrivilegeHandler) GrantPrivilege(c *gin.Context) {
	var req privGrantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if req.UserID == "" || req.DatasourceID == "" || req.DatabaseName == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "用户ID、数据源、库名必填")
		return
	}
	userID := middleware.GetStr(c, "userId")
	userName := middleware.GetStr(c, "username")
	// 被授权用户若未传 name，回退为当前用户
	grantUserName := req.UserName
	if grantUserName == "" {
		grantUserName = req.UserID
	}
	columnsStr := ""
	if len(req.Columns) > 0 {
		columnsStr = strings.Join(req.Columns, ",")
	}
	priv, err := h.applySvc.Grant(req.UserID, grantUserName, req.DatasourceID, req.DatabaseName, req.TableName,
		req.PrivType, req.OperationType, columnsStr, req.RowLimit, req.ValidDays)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"privId": priv.PrivID, "operator": userName, "operatorId": userID})
}

// BatchGrantPrivilege 批量授予多用户同样的权限规则
func (h *PrivilegeHandler) BatchGrantPrivilege(c *gin.Context) {
	var req struct {
		UserIDs       []string `json:"userIds"`
		DatasourceID  string   `json:"datasourceId"`
		DatabaseName  string   `json:"databaseName"`
		TableName     string   `json:"tableName"`
		PrivType      string   `json:"privType"`
		OperationType string   `json:"operationType"`
		Columns       []string `json:"columns"`
		RowLimit      int      `json:"rowLimit"`
		ValidDays     int      `json:"validDays"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if len(req.UserIDs) == 0 || req.DatasourceID == "" || req.DatabaseName == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "用户列表、数据源、库名不能为空")
		return
	}
	columnsStr := ""
	if len(req.Columns) > 0 {
		columnsStr = strings.Join(req.Columns, ",")
	}
	createdIDs := make([]string, 0, len(req.UserIDs))
	for _, uid := range req.UserIDs {
		priv, err := h.applySvc.Grant(uid, uid, req.DatasourceID, req.DatabaseName, req.TableName,
			req.PrivType, req.OperationType, columnsStr, req.RowLimit, req.ValidDays)
		if err != nil {
			middleware.Fail(c, http.StatusInternalServerError, 500, "用户 "+uid+" 授权失败: "+err.Error())
			return
		}
		createdIDs = append(createdIDs, priv.PrivID)
	}
	middleware.OK(c, gin.H{"created": createdIDs})
}

// ==================== 敏感列管理 ====================

type sensitiveColumnReq struct {
	DatasourceID string `json:"datasourceId"`
	DatabaseName string `json:"databaseName"`
	TableName    string `json:"tableName"`
	ColumnName   string `json:"columnName"`
	Rule         string `json:"rule"`
}

func (h *PrivilegeHandler) ListSensitiveColumns(c *gin.Context) {
	dsID := c.Query("datasourceId")
	list, err := h.sensSvc.List(dsID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

func (h *PrivilegeHandler) CreateSensitiveColumn(c *gin.Context) {
	var req sensitiveColumnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if req.DatasourceID == "" || req.DatabaseName == "" || req.TableName == "" || req.ColumnName == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数不完整")
		return
	}
	rule := req.Rule
	if rule == "" {
		rule = "mask"
	}
	sc := &model.SensitiveColumn{
		DatasourceID: req.DatasourceID,
		DatabaseName: req.DatabaseName,
		TblName:      req.TableName,
		ColumnName:   req.ColumnName,
		Rule:         rule,
	}
	if err := h.sensSvc.Create(sc); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, sc)
}

func (h *PrivilegeHandler) DeleteSensitiveColumn(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "id 参数错误")
		return
	}
	if err := h.sensSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

// ==================== 权限审计日志 ====================

func (h *PrivilegeHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	operType := c.Query("operType")
	operator := c.Query("operator")

	query := database.DB.Model(&model.PrivAuditLog{})
	if operType != "" {
		query = query.Where("oper_type = ?", operType)
	}
	if operator != "" {
		query = query.Where("operator LIKE ?", "%"+operator+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	var list []model.PrivAuditLog
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

// ==================== 定时清理入口（可选由 cron 触发） ====================

func (h *PrivilegeHandler) CleanupExpired(c *gin.Context) {
	n, err := h.applySvc.CleanupExpired()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"reclaimed": n})
}
