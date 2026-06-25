package handler

import (
	"net/http"
	"strconv"

	"dbm-lite/internal/middleware"
	"dbm-lite/internal/model/datasource_auth"
	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

type InternalAuthHandler struct {
	authSvc *service.InternalAuthService
}

func NewInternalAuthHandler() *InternalAuthHandler {
	return &InternalAuthHandler{
		authSvc: service.NewInternalAuthService(),
	}
}

func (h *InternalAuthHandler) ListUsers(c *gin.Context) {
	datasourceID := c.Query("datasourceId")
	keyword := c.Query("keyword")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := h.authSvc.ListInternalUsers(datasourceID, keyword, status, page, pageSize)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.SuccessList(c, list, total, page, pageSize)
}

func (h *InternalAuthHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.authSvc.GetInternalUser(id)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, user)
}

func (h *InternalAuthHandler) CreateUser(c *gin.Context) {
	var req struct {
		DatasourceID string `json:"datasourceId"`
		Username     string `json:"username"`
		Host         string `json:"host"`
		Password     string `json:"password"`
		Remark       string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if req.DatasourceID == "" || req.Username == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "数据源ID和用户名必填")
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	user, err := h.authSvc.CreateInternalUser(struct {
		DatasourceID string
		Username     string
		Host         string
		Password     string
		Remark       string
	}{req.DatasourceID, req.Username, req.Host, req.Password, req.Remark}, operatorID, operatorName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, user)
}

func (h *InternalAuthHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Host   string `json:"host"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.UpdateInternalUser(id, struct {
		Host   string
		Remark string
	}{req.Host, req.Remark}, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.DeleteInternalUser(id, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.ResetPassword(id, req.Password, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) ToggleUserStatus(c *gin.Context) {
	id := c.Param("id")
	enable, _ := strconv.ParseBool(c.Query("enable"))
	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.ToggleUserStatus(id, enable, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) SyncUsers(c *gin.Context) {
	datasourceID := c.Query("datasourceId")
	if err := h.authSvc.SyncDBUsers(datasourceID); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) ListRoles(c *gin.Context) {
	datasourceID := c.Query("datasourceId")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := h.authSvc.ListInternalRoles(datasourceID, keyword, page, pageSize)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.SuccessList(c, list, total, page, pageSize)
}

func (h *InternalAuthHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.authSvc.GetInternalRole(id)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, role)
}

func (h *InternalAuthHandler) CreateRole(c *gin.Context) {
	var req struct {
		DatasourceID string `json:"datasourceId"`
		Name         string `json:"name"`
		Description  string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if req.DatasourceID == "" || req.Name == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "数据源ID和角色名称必填")
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	role, err := h.authSvc.CreateInternalRole(struct {
		DatasourceID string
		Name         string
		Description  string
	}{req.DatasourceID, req.Name, req.Description}, operatorID, operatorName)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, role)
}

func (h *InternalAuthHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.UpdateInternalRole(id, struct {
		Name        string
		Description string
	}{req.Name, req.Description}, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.DeleteInternalRole(id, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) GetRoleUserCount(c *gin.Context) {
	roleID := c.Param("id")
	count, err := h.authSvc.GetRoleUserCount(roleID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"count": count})
}

func (h *InternalAuthHandler) AssignRoles(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		RoleIDs []string `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.AssignRoles(userID, req.RoleIDs, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	roles, err := h.authSvc.GetUserRoles(userID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, roles)
}

func (h *InternalAuthHandler) GrantPermission(c *gin.Context) {
	var req struct {
		DatasourceID  string   `json:"datasourceId"`
		PrincipalType string   `json:"principalType"`
		PrincipalID   string   `json:"principalId"`
		PrivilegeType string   `json:"privilegeType"`
		ObjectLevel   string   `json:"objectLevel"`
		DatabaseName  string   `json:"databaseName"`
		TableName     string   `json:"tableName"`
		Columns       []string `json:"columns"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.GrantPermission(struct {
		DatasourceID  string
		PrincipalType string
		PrincipalID   string
		PrivilegeType string
		ObjectLevel   string
		DatabaseName  string
		TableName     string
		Columns       []string
	}{req.DatasourceID, req.PrincipalType, req.PrincipalID, req.PrivilegeType, req.ObjectLevel, req.DatabaseName, req.TableName, req.Columns}, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) BatchGrantPermission(c *gin.Context) {
	var req struct {
		DatasourceID  string `json:"datasourceId"`
		PrincipalType string `json:"principalType"`
		PrincipalID   string `json:"principalId"`
		PrivilegeType string `json:"privilegeType"`
		Rules         []struct {
			ObjectLevel  string   `json:"objectLevel"`
			DatabaseName string   `json:"databaseName"`
			TableName    string   `json:"tableName"`
			Columns      []string `json:"columns"`
		} `json:"rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	rules := make([]struct {
		ObjectLevel  string
		DatabaseName string
		TableName    string
		Columns      []string
	}, len(req.Rules))
	for i, r := range req.Rules {
		rules[i] = struct {
			ObjectLevel  string
			DatabaseName string
			TableName    string
			Columns      []string
		}{r.ObjectLevel, r.DatabaseName, r.TableName, r.Columns}
	}
	if err := h.authSvc.BatchGrantPermission(struct {
		DatasourceID  string
		PrincipalType string
		PrincipalID   string
		PrivilegeType string
		Rules         []struct {
			ObjectLevel  string
			DatabaseName string
			TableName    string
			Columns      []string
		}
	}{req.DatasourceID, req.PrincipalType, req.PrincipalID, req.PrivilegeType, rules}, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) RevokePermission(c *gin.Context) {
	id := c.Param("id")
	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.RevokePermission(id, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) BatchRevokePermission(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetStr(c, "userId")
	operatorName := middleware.GetStr(c, "username")

	if err := h.authSvc.BatchRevokePermission(req.IDs, operatorID, operatorName); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *InternalAuthHandler) ListPermissionRules(c *gin.Context) {
	datasourceID := c.Query("datasourceId")
	principalType := c.Query("principalType")
	principalID := c.Query("principalId")
	privilegeType := c.Query("privilegeType")
	objectLevel := c.Query("objectLevel")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := h.authSvc.ListPermissionRules(datasourceID, principalType, principalID, privilegeType, objectLevel, page, pageSize)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.SuccessList(c, list, total, page, pageSize)
}

func (h *InternalAuthHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.authSvc.GetInternalUser(userID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	rules, err := h.authSvc.GetUserEffectivePermissions(user.DatasourceID, userID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, rules)
}

func (h *InternalAuthHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")

	role, err := h.authSvc.GetInternalRole(roleID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	rules, _, err := h.authSvc.ListPermissionRules(role.DatasourceID, datasource_auth.PrincipalTypeRole, roleID, "", "", 1, 1000)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, rules)
}

func (h *InternalAuthHandler) GetUserGrants(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.authSvc.GetInternalUser(userID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	grants, err := h.authSvc.GetUserGrants(user.DatasourceID, user.Username, user.Host)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"grants": grants})
}

func (h *InternalAuthHandler) GetRoleGrants(c *gin.Context) {
	roleID := c.Param("id")

	role, err := h.authSvc.GetInternalRole(roleID)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	grants, err := h.authSvc.GetRoleGrants(role.DatasourceID, role.Name)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"grants": grants})
}

func (h *InternalAuthHandler) CheckSQLPermission(c *gin.Context) {
	var req struct {
		DatasourceID string `json:"datasourceId"`
		UserID       string `json:"userId"`
		SQLText      string `json:"sqlText"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	allowed, msg := h.authSvc.CheckSQLPermission(req.DatasourceID, req.UserID, req.SQLText)
	if !allowed {
		middleware.Fail(c, http.StatusForbidden, 403, msg)
		return
	}
	middleware.OK(c, gin.H{"allowed": true})
}

func (h *InternalAuthHandler) ListAuditLogs(c *gin.Context) {
	datasourceID := c.Query("datasourceId")
	operator := c.Query("operator")
	operType := c.Query("operType")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := h.authSvc.ListAuditLogs(datasourceID, operator, operType, page, pageSize)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.SuccessList(c, list, total, page, pageSize)
}