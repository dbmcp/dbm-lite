/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package handler

import (
	"strconv"
	"strings"

	"dbm-lite/internal/middleware"
	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

type SQLWindowHandler struct {
	svc *service.SQLWindowService
}

func NewSQLWindowHandler() *SQLWindowHandler {
	return &SQLWindowHandler{svc: service.NewSQLWindowService()}
}

type windowReq struct {
	WindowID       string `json:"windowId"`
	Title          string `json:"title"`
	SQL            string `json:"sql"`
	DatasourceID   string `json:"datasourceId"`
	DatasourceName string `json:"datasourceName"`
	DatabaseName   string `json:"databaseName"`
	SortOrder      int    `json:"sortOrder"`
	IsActive       bool   `json:"isActive"`
}

func (h *SQLWindowHandler) List(c *gin.Context) {
	userID := middleware.GetStr(c, "userId")
	list, total, err := h.svc.ListByUser(userID)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": total, "current": 1, "pageSize": int(total)})
}

func (h *SQLWindowHandler) Get(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetStr(c, "userId")
	w, err := h.svc.GetById(id, userID)
	if err != nil {
		middleware.Fail(c, 404, 404, "窗口不存在")
		return
	}
	middleware.OK(c, w)
}

func (h *SQLWindowHandler) Create(c *gin.Context) {
	var req windowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, 400, 400, "参数错误: "+err.Error())
		return
	}
	userID := middleware.GetStr(c, "userId")
	username := middleware.GetStr(c, "username")
	w, err := h.svc.Create(userID, username, req.Title, req.SQL, req.DatasourceID, req.DatasourceName, req.DatabaseName, req.SortOrder)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, w)
}

func (h *SQLWindowHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req windowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, 400, 400, "参数错误: "+err.Error())
		return
	}
	userID := middleware.GetStr(c, "userId")
	w, err := h.svc.Update(id, userID, req.Title, req.SQL, req.DatasourceID, req.DatasourceName, req.DatabaseName)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	if req.IsActive {
		_ = h.svc.SetActive(id, userID)
	}
	middleware.OK(c, w)
}

func (h *SQLWindowHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetStr(c, "userId")
	if err := h.svc.Delete(id, userID); err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"windowId": id})
}

func (h *SQLWindowHandler) BatchDelete(c *gin.Context) {
	var req struct {
		Ids []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, 400, 400, "参数错误: "+err.Error())
		return
	}
	userID := middleware.GetStr(c, "userId")
	if err := h.svc.BatchDelete(req.Ids, userID); err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"count": len(req.Ids)})
}

func (h *SQLWindowHandler) SetActive(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetStr(c, "userId")
	if err := h.svc.SetActive(id, userID); err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"windowId": id})
}

func (h *SQLWindowHandler) Recent(c *gin.Context) {
	limit, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("limit", "20")))
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	userID := middleware.GetStr(c, "userId")
	list, err := h.svc.Recent(userID, limit)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"list": list, "total": len(list)})
}
