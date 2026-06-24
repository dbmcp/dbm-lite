/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package handler

import (
	"dbm-lite/internal/middleware"
	"dbm-lite/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dsSvc *service.DashboardService
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{dsSvc: service.NewDashboardService()}
}

// Summary 返回首页概览聚合统计数据
func (h *DashboardHandler) Summary(c *gin.Context) {
	stats, err := h.dsSvc.Summary()
	if err != nil {
		middleware.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	middleware.OK(c, stats)
}
