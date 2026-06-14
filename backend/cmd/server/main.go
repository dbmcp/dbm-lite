/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/database"
	"dbm-lite/internal/handler"
	"dbm-lite/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Printf("config load warning: %v", err)
	}

	if err := database.Init(); err != nil {
		log.Fatalf("init database failed: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"service": "dbm-lite-api",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 静态文件服务 - 前端UI
	setupStaticRoutes(r)

	// API路由组 - 必须在静态文件之后，以免 /api/* 被当作静态文件
	api := r.Group("/api")
	{
		authH := handler.NewAuthHandler()
		userH := handler.NewUserHandler()
		sqlH := handler.NewSQLHandler()
		dsH := handler.NewDatasourceHandler()
		bizH := handler.NewBusinessHandler()
		srvH := handler.NewServerHandler()
		auditH := handler.NewAuditHandler()
		opsH := handler.NewMaintenanceHandler()

		// 认证 - 无token即可访问
		auth := api.Group("/auth")
		{
			auth.POST("/login", authH.Login)
			auth.GET("/me", middleware.AuthRequired(), authH.Me)
			auth.PUT("/password", middleware.AuthRequired(), authH.ChangePassword)
		}

		// 用户管理
		user := api.Group("/users")
		user.Use(middleware.AuthRequired())
		{
			user.POST("", userH.Create)
			user.GET("", userH.List)
			user.PUT("/:id", userH.Update)
			user.DELETE("/:id", userH.Delete)
			user.POST("/:id/reset-password", userH.ResetPassword)
		}
		// accounts 别名，兼容前端
		accounts := api.Group("/accounts")
		accounts.Use(middleware.AuthRequired())
		{
			accounts.POST("", userH.Create)
			accounts.GET("", userH.List)
			accounts.PUT("/:id", userH.Update)
			accounts.DELETE("/:id", userH.Delete)
			accounts.POST("/:id/resetPassword", userH.ResetPassword)
		}

		// SQL工作台相关
		sql := api.Group("/sql")
		sql.Use(middleware.AuthRequired())
		{
			sql.POST("/execute", sqlH.Execute)
			sql.POST("/explain", sqlH.Explain)
			sql.POST("/review", sqlH.ReviewSQL)
			sql.POST("/datasources/:id/test", sqlH.TestConnection)
			sql.GET("/datasources/:id/databases", sqlH.GetDatabases)
			sql.GET("/datasources/:id/tables", sqlH.GetTables)
			sql.GET("/datasources/:id/columns", sqlH.GetColumns)
			sql.GET("/datasources/:id/tree", sqlH.GetFullTree)
			sql.GET("/datasources/:id/table-info", sqlH.GetTableInfo)
			sql.GET("/history", sqlH.GetHistory)
		}

		// 数据源管理
		ds := api.Group("/datasources")
	ds.Use(middleware.AuthRequired())
	{
		ds.POST("", dsH.Create)
		ds.GET("", dsH.List)
		ds.GET("/all", dsH.AllSimple)
		ds.GET("/simple", dsH.AllSimple)
		ds.GET("/stats", dsH.Stats)
		ds.GET("/:id", dsH.Get)
		ds.GET("/:id/detail", dsH.GetDetail)
		ds.PUT("/:id", dsH.Update)
		ds.DELETE("/:id", dsH.Delete)
		ds.POST("/:id/copy", dsH.Copy)
		ds.POST("/:id/test", dsH.TestConnectionById)
		ds.POST("/testConnection", dsH.TestConnectionFromForm)
		ds.GET("/:id/databases", sqlH.GetDatabases)
		ds.GET("/:id/tables", sqlH.GetTables)
		ds.GET("/:id/columns", sqlH.GetColumns)
		ds.GET("/:id/tree", sqlH.GetFullTree)
	}

		// 项目和业务管理
	proj := api.Group("/projects")
	proj.Use(middleware.AuthRequired())
	{
		proj.POST("", bizH.CreateProject)
		proj.GET("", bizH.ListProjects)
		proj.GET("/all", bizH.AllProjects)
		proj.PUT("/:id", bizH.UpdateProject)
		proj.DELETE("/:id", bizH.DeleteProject)

		// 项目作用域下的服务器
		projSrv := proj.Group("/:id/servers")
		{
			projSrv.GET("", srvH.ListByProject)
			projSrv.POST("", srvH.CreateByProject)
			projSrv.PUT("/:serverid", srvH.Update)
			projSrv.DELETE("/:serverid", srvH.Delete)
			projSrv.POST("/testConnection", srvH.TestConnect)
		}

		// 项目作用域下的业务
		projBiz := proj.Group("/:id/businesses")
		{
			projBiz.GET("", bizH.ListBusinessesByProject)
			projBiz.POST("", bizH.CreateBusinessByProject)
		}
	}
	biz := api.Group("/businesses")
	biz.Use(middleware.AuthRequired())
	{
		biz.POST("", bizH.CreateBusiness)
		biz.GET("", bizH.ListBusinesses)
		biz.GET("/all", bizH.AllBusinesses)
		biz.PUT("/:id", bizH.UpdateBusiness)
		biz.DELETE("/:id", bizH.DeleteBusiness)
	}

	// 服务器管理
	srv := api.Group("/servers")
	srv.Use(middleware.AuthRequired())
	{
		srv.POST("", srvH.Create)
		srv.GET("", srvH.List)
		srv.GET("/all", srvH.All)
		srv.PUT("/:id", srvH.Update)
		srv.DELETE("/:id", srvH.Delete)
		srv.POST("/:id/test", srvH.TestConnect)
	}

	// 插件管理
	plg := api.Group("/plugins")
	plg.Use(middleware.AuthRequired())
	{
		plg.POST("", opsH.CreatePlugin)
		plg.GET("", opsH.ListPlugins)
		plg.PUT("/:id", opsH.UpdatePlugin)
		plg.DELETE("/:id", opsH.DeletePlugin)
	}

	// 审计日志
	audit := api.Group("/audit")
	audit.Use(middleware.AuthRequired())
	{
		audit.GET("", auditH.List)
		audit.GET("/logs", auditH.List)
		audit.GET("/stats", auditH.Stats)
	}
}

	log.Printf("dbm-lite server starting on :%s", config.App.ServerPort)
	if err := r.Run(":" + config.App.ServerPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// setupStaticRoutes 配置前端静态文件服务
// 查找顺序：
// 1. 当前目录下的 frontend/dist
// 2. 上级目录下的 frontend/dist
// 3. 同级目录下的 frontend/dist
func setupStaticRoutes(r *gin.Engine) {
	candidates := []string{
		filepath.Join(".", "frontend", "dist"),
		filepath.Join("..", "frontend", "dist"),
		filepath.Join(".", "..", "frontend", "dist"),
	}

	var staticDir string
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if stat, err := os.Stat(abs); err == nil && stat.IsDir() {
				staticDir = abs
				log.Printf("frontend static dir: %s", staticDir)
				break
			}
		}
	}

	if staticDir == "" {
		log.Printf("warning: frontend dist not found, serving API only")
		// 根路径返回简单信息页
		r.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "dbm-lite API server is running",
				"api":     "/api/*",
				"health":  "/health",
			})
		})
		return
	}

	// 服务 index.html 和其他静态资源
	// 对 SPA (Single Page Application) 使用 NoRoute 回退到 index.html
	r.StaticFile("/", filepath.Join(staticDir, "index.html"))
	r.Static("/assets", filepath.Join(staticDir, "assets"))
	r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))

	// SPA 回退：所有非 /api/* 且未匹配到文件的路径，都返回 index.html
	r.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求，正常返回 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "route not found",
			})
			return
		}
		// 其他路径返回 index.html 以支持前端路由
		c.File(filepath.Join(staticDir, "index.html"))
	})
}

