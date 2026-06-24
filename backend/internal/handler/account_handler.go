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

// ====== 账号管理 Handler ======

type AccountHandler struct {
	accountSvc *service.AccountService
	roleSvc    *service.RoleService
}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		accountSvc: service.NewAccountService(),
		roleSvc:    service.NewRoleService(),
	}
}

type accountCreateReq struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Password    string   `json:"password"`
	RoleIDs     []string `json:"roleIds"`
	RoleCodes   []string `json:"roleCodes"`
}

func pickFirst[T any](vals ...T) T {
	var zero T
	for _, v := range vals {
		switch any(v).(type) {
		case string:
			if any(v).(string) != "" {
				return v
			}
		case []string:
			if len(any(v).([]string)) > 0 {
				return v
			}
		}
	}
	return zero
}

func (r *accountCreateReq) getRoleIDs() []string {
	if len(r.RoleIDs) > 0 {
		return r.RoleIDs
	}
	return r.RoleCodes
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req accountCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if req.Username == "" || req.Password == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "用户名和密码不能为空")
		return
	}
	user := &model.User{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Phone:       req.Phone,
		Role:        "member",
		Status:      model.StatusActive,
	}
	if err := h.accountSvc.Create(user, req.Password); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 500, "创建失败: "+err.Error())
		return
	}
	roles := req.getRoleIDs()
	if len(roles) > 0 {
		if err := h.accountSvc.AssignRoles(user.UserID, roles); err != nil {
			middleware.Fail(c, http.StatusInternalServerError, 500, "绑定角色失败: "+err.Error())
			return
		}
	}
	middleware.OK(c, user)
}

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	list, total, err := h.accountSvc.List(page, pageSize, keyword, status)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if err := h.accountSvc.Update(id, updates); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	if roleIDs, ok := updates["roleIds"].([]interface{}); ok {
		ids := make([]string, 0, len(roleIDs))
		for _, rid := range roleIDs {
			if s, ok := rid.(string); ok {
				ids = append(ids, s)
			}
		}
		if err := h.accountSvc.AssignRoles(id, ids); err != nil {
			middleware.Fail(c, http.StatusInternalServerError, 500, "绑定角色失败: "+err.Error())
			return
		}
	}
	middleware.OK(c, nil)
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	if err := h.accountSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

type resetPwdReq struct {
	NewPassword string `json:"newPassword"`
	Password    string `json:"password"`
}

func (h *AccountHandler) ResetAccountPassword(c *gin.Context) {
	id := c.Param("id")
	var req resetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	newPwd := req.NewPassword
	if newPwd == "" {
		newPwd = req.Password
	}
	if newPwd == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "新密码不能为空")
		return
	}
	if err := h.accountSvc.ResetPassword(id, newPwd); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *AccountHandler) ToggleAccountLock(c *gin.Context) {
	id := c.Param("id")
	u, err := h.accountSvc.ToggleLock(id)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, u)
}

func (h *AccountHandler) GetAccountRoles(c *gin.Context) {
	id := c.Param("id")
	roles, err := h.accountSvc.GetUserRoles(id)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, roles)
}

type assignAccountRolesReq struct {
	RoleIDs   []string `json:"roleIds"`
	RoleCodes []string `json:"roleCodes"`
	Roles     []string `json:"roles"`
}

func (r *assignAccountRolesReq) getRoleIDs() []string {
	if len(r.RoleIDs) > 0 {
		return r.RoleIDs
	}
	if len(r.RoleCodes) > 0 {
		return r.RoleCodes
	}
	return r.Roles
}

func (h *AccountHandler) AssignAccountRoles(c *gin.Context) {
	id := c.Param("id")
	var req assignAccountRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if err := h.accountSvc.AssignRoles(id, req.getRoleIDs()); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

// ====== 角色管理 Handler ======

type roleCreateReq struct {
	Name           string   `json:"name"`
	RoleName       string   `json:"roleName"`
	Description    string   `json:"description"`
	Codes          []string `json:"codes"`
	PermissionCodes []string `json:"permissionCodes"`
}

func (r *roleCreateReq) getName() string {
	if r.RoleName != "" {
		return r.RoleName
	}
	return r.Name
}

func (r *roleCreateReq) getCodes() []string {
	if len(r.PermissionCodes) > 0 {
		return r.PermissionCodes
	}
	return r.Codes
}

func (h *AccountHandler) CreateRole(c *gin.Context) {
	var req roleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if req.getName() == "" {
		middleware.Fail(c, http.StatusBadRequest, 400, "角色名称不能为空")
		return
	}
	role := &model.Role{Name: req.getName(), Description: req.Description}
	if err := h.roleSvc.Create(role, req.getCodes()); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, role)
}

func (h *AccountHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	list, total, err := h.roleSvc.List(page, pageSize, keyword)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "pageSize": pageSize})
}

func (h *AccountHandler) GetAllRoles(c *gin.Context) {
	list, err := h.roleSvc.All()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, list)
}

func (h *AccountHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if codes, ok := updates["codes"].([]interface{}); ok {
		cs := make([]string, 0, len(codes))
		for _, c := range codes {
			if s, ok := c.(string); ok {
				cs = append(cs, s)
			}
		}
		updates["codes"] = cs
	}
	if err := h.roleSvc.Update(id, updates); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *AccountHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.roleSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

type assignRolePermReq struct {
	Codes           []string `json:"codes"`
	PermissionCodes []string `json:"permissionCodes"`
}

func (r *assignRolePermReq) getCodes() []string {
	if len(r.PermissionCodes) > 0 {
		return r.PermissionCodes
	}
	return r.Codes
}

func (h *AccountHandler) AssignRolePermissions(c *gin.Context) {
	id := c.Param("id")
	var req assignRolePermReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if err := h.roleSvc.AssignPermissions(id, req.getCodes()); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

// ====== 权限码清单 ======

func (h *AccountHandler) ListPermissionCodes(c *gin.Context) {
	codes := []struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Type string `json:"type"`
	}{
		{Code: model.PermMenuAccount, Name: "账号管理菜单", Type: "menu"},
		{Code: model.PermAccountEdit, Name: "账号新增/编辑", Type: "button"},
		{Code: model.PermAccountResetPwd, Name: "账号重置密码", Type: "button"},
		{Code: model.PermAccountLock, Name: "账号锁定/解锁", Type: "button"},
		{Code: model.PermMenuRole, Name: "角色管理菜单", Type: "menu"},
		{Code: model.PermRoleEdit, Name: "角色编辑", Type: "button"},
		{Code: model.PermRoleAssignUser, Name: "绑定用户", Type: "button"},
		{Code: model.PermRoleAssignPerm, Name: "分配权限码", Type: "button"},
		{Code: model.PermMenuAuditLog, Name: "操作审计菜单", Type: "menu"},
		{Code: model.PermMenuPrivGroup, Name: "资源组管理菜单", Type: "menu"},
		{Code: model.PermMenuPrivApply, Name: "权限申请菜单", Type: "menu"},
		{Code: model.PermMenuPrivMy, Name: "我的权限菜单", Type: "menu"},
		{Code: model.PermMenuPrivAudit, Name: "权限审计日志菜单", Type: "menu"},
		{Code: model.PermPrivGroupEdit, Name: "资源组编辑", Type: "button"},
		{Code: model.PermPrivApplyOper, Name: "权限申请审批", Type: "button"},
		{Code: model.PermPrivAuditOper, Name: "权限审计操作", Type: "button"},
		{Code: model.PermSqlQueryAll, Name: "SQL超级查询(豁免所有表)", Type: "sql"},
		{Code: model.PermSqlQueryGroup, Name: "SQL资源组整库查询", Type: "sql"},
	}
	middleware.OK(c, codes)
}
