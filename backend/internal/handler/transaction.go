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

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	dsSvc       *service.DatasourceService
	txSvc       *service.TransactionService
	editorSvc   *service.DataEditorService
}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{
		dsSvc:     service.NewDatasourceService(),
		txSvc:     service.NewTransactionService(),
		editorSvc: service.NewDataEditorService(),
	}
}

func (h *TransactionHandler) Begin(c *gin.Context) {
	id := c.Param("id")
	if err := h.txSvc.BeginTransaction(id); err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, gin.H{"message": "事务已开启"})
}

func (h *TransactionHandler) Commit(c *gin.Context) {
	id := c.Param("id")
	result, err := h.txSvc.CommitTransaction(id)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	if result.Success {
		middleware.OK(c, result)
	} else {
		middleware.Fail(c, 400, 400, result.Message)
	}
}

func (h *TransactionHandler) Rollback(c *gin.Context) {
	id := c.Param("id")
	result, err := h.txSvc.RollbackTransaction(id)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	if result.Success {
		middleware.OK(c, result)
	} else {
		middleware.Fail(c, 400, 400, result.Message)
	}
}

func (h *TransactionHandler) ExecuteBatch(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	var req struct {
		SQLs []string `json:"sqls"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, 400, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	results, err := h.txSvc.ExecuteInTransaction(ds, dbName, req.SQLs)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	middleware.OK(c, results)
}

func (h *TransactionHandler) InsertRow(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		middleware.Fail(c, 400, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	result, err := h.editorSvc.InsertRow(ds, dbName, tableName, data)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	if result.Success {
		middleware.OK(c, result)
	} else {
		middleware.Fail(c, 400, 400, result.Message)
	}
}

func (h *TransactionHandler) UpdateRow(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	var req struct {
		Data  map[string]interface{} `json:"data"`
		Where map[string]interface{} `json:"where"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Fail(c, 400, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	result, err := h.editorSvc.UpdateRow(ds, dbName, tableName, req.Data, req.Where)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	if result.Success {
		middleware.OK(c, result)
	} else {
		middleware.Fail(c, 400, 400, result.Message)
	}
}

func (h *TransactionHandler) DeleteRow(c *gin.Context) {
	id := c.Param("id")
	dbName := c.Query("database")
	tableName := c.Query("table")
	var where map[string]interface{}
	if err := c.ShouldBindJSON(&where); err != nil {
		middleware.Fail(c, 400, 400, "参数错误")
		return
	}
	ds, err := h.dsSvc.GetById(id)
	if err != nil {
		middleware.Fail(c, 404, 404, "数据源不存在")
		return
	}
	result, err := h.editorSvc.DeleteRow(ds, dbName, tableName, where)
	if err != nil {
		middleware.Fail(c, 500, 500, err.Error())
		return
	}
	if result.Success {
		middleware.OK(c, result)
	} else {
		middleware.Fail(c, 400, 400, result.Message)
	}
}
