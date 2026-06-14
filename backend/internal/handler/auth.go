/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
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

type AuthHandler struct {
	authSvc  *service.AuthService
	auditSvc *service.AuditService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authSvc:  service.NewAuthService(),
		auditSvc: service.NewAuditService(),
	}
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	user, token, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		middleware.Fail(c, http.StatusUnauthorized, 401, err.Error())
		h.auditSvc.Log("", req.Username, middleware.GetClientIP(c), model.ModuleAuth, "login", "", err.Error(), model.AuditResultFailed, "")
		return
	}
	h.auditSvc.Log(user.UserID, user.Username, middleware.GetClientIP(c), model.ModuleAuth, "login", user.UserID, "登录成功", model.AuditResultSuccess, "")
	middleware.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"userId":      user.UserID,
			"username":    user.Username,
			"displayName": user.DisplayName,
			"role":        user.Role,
			"email":       user.Email,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	role := middleware.GetStr(c, "role")
	displayName := middleware.GetStr(c, "displayName")
	middleware.OK(c, gin.H{
		"userId":      userId,
		"username":    username,
		"role":        role,
		"displayName": displayName,
	})
}

type changePwdReq struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	if err := h.authSvc.ChangePassword(userId, req.OldPassword, req.NewPassword); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleAuth, "change_password", userId, err.Error(), model.AuditResultFailed, "")
		return
	}
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleAuth, "change_password", userId, "修改密码成功", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

type UserHandler struct {
	userSvc  *service.UserService
	auditSvc *service.AuditService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userSvc:  service.NewUserService(),
		auditSvc: service.NewAuditService(),
	}
}

type createUserReq struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	user := &model.User{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Role:        req.Role,
	}
	if err := h.userSvc.Create(user, req.Password); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleUser, "create", user.UserID, "创建用户: "+user.Username, model.AuditResultSuccess, "")
	middleware.OK(c, user)
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p, err := strconv.Atoi(c.Query("current")); err == nil && p > 0 {
		page = p
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	role := c.Query("role")
	list, total, err := h.userSvc.List(page, pageSize, keyword, role)
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": page, "page": page, "pageSize": pageSize})
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if err := h.userSvc.Update(id, req); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleUser, "update", id, "更新用户", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.userSvc.Delete(id); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	userId := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	h.auditSvc.Log(userId, username, middleware.GetClientIP(c), model.ModuleUser, "delete", id, "删除用户", model.AuditResultSuccess, "")
	middleware.OK(c, nil)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	newPwd := req["newPassword"]
	if newPwd == "" {
		newPwd = req["password"]
	}
	if err := h.userSvc.ResetPassword(id, newPwd); err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, nil)
}

