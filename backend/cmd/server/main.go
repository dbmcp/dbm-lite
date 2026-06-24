package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/database"
	"dbm-lite/internal/ferry"
	"dbm-lite/internal/handler"
	"dbm-lite/internal/migrate"
	"dbm-lite/internal/middleware"
	"dbm-lite/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if err := database.Init(); err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	if err := migrate.RunAutoMigrate(database.DB); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logger.Backend()))

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    io.MultiWriter(os.Stdout, logger.Backend()),
		SkipPaths: []string{"/healthz", "/favicon.ico"},
	}))

	r.Use(middleware.StaticCacheControl())
	r.Use(middleware.Gzip())

	r.Use(corsMiddleware())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "DBM-Lite API", "time": time.Now().Format(time.RFC3339)})
	})
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "ok"})
	})

	api := r.Group("/api")
	{
		api.POST("/frontend/log", frontendLogHandler)

		authH := handler.NewAuthHandler()
		api.POST("/auth/login", authH.Login)

		authorized := api.Group("/")
		authorized.Use(middleware.AuthRequired())
		{
			authorized.GET("/auth/me", authH.Me)
			authorized.POST("/auth/password", authH.ChangePassword)

			userH := handler.NewUserHandler()
			users := authorized.Group("/users")
			users.POST("", middleware.AdminRequired(), userH.Create)
			users.GET("", userH.List)
			users.PUT("/:id", middleware.AdminRequired(), userH.Update)
			users.DELETE("/:id", middleware.AdminRequired(), userH.Delete)
			users.POST("/:id/reset-password", middleware.AdminRequired(), userH.ResetPassword)

			sqlH := handler.NewSQLHandler()
			dataquery := authorized.Group("/dataquery")
			{
				dataquery.POST("/sql/execute", sqlH.Execute)
				dataquery.POST("/sql/explain", sqlH.Explain)
				dataquery.POST("/sql/review", sqlH.ReviewSQL)
				dataquery.POST("/sql/cancel", sqlH.CancelExecute)
				historyAPI := dataquery.Group("/sqlHistory")
				historyAPI.GET("/list", sqlH.GetHistory)

				savedAPI := dataquery.Group("/savedQueries")
				savedAPI.GET("", sqlH.ListSavedQueries)
				savedAPI.GET("/:id", sqlH.GetSavedQuery)
				savedAPI.POST("", sqlH.SaveQuery)
				savedAPI.PUT("/:id", sqlH.UpdateSavedQuery)
				savedAPI.DELETE("/:id", sqlH.DeleteSavedQuery)

				favoritesAPI := dataquery.Group("/favorites")
				favoritesAPI.GET("", sqlH.ListFavorites)
				favoritesAPI.GET("/:id", sqlH.GetFavorite)
				favoritesAPI.POST("", sqlH.CreateFavorite)
				favoritesAPI.PUT("/:id", sqlH.UpdateFavorite)
				favoritesAPI.DELETE("/:id", sqlH.DeleteFavorite)

				ds := dataquery.Group("/datasources")
				ds.GET("/:id/databases", middleware.CacheReadThrough(3*time.Minute), sqlH.GetDatabases)
				ds.GET("/:id/tables", middleware.CacheReadThrough(3*time.Minute), sqlH.GetTables)
				ds.GET("/:id/columns", sqlH.GetColumns)
				ds.GET("/:id/tree", middleware.CacheReadThrough(3*time.Minute), sqlH.GetFullTree)
				ds.GET("/:id/table-info", sqlH.GetTableInfo)
				ds.GET("/:id/table-info-full", sqlH.GetTableInfoFull)
				ds.GET("/:id/data/query", sqlH.QueryTable)
				ds.GET("/:id/primary-key", sqlH.PrimaryKey)
				ds.POST("/:id/maintenance/:action", sqlH.TableMaintenance)
				// 表格子节点懒加载
				ds.GET("/:id/table-children", sqlH.GetTableChildren)
				ds.GET("/:id/routines", sqlH.GetRoutines)
				ds.GET("/:id/triggers-list", sqlH.GetTriggersForIde)
				// 系统数据库列表
				dataquery.GET("/system-databases", sqlH.ListSystemDatabases)
				// 完整数据库列表（含/不含系统库）
				dataquery.GET("/:id/databases-full", sqlH.GetDatabasesFull)
				// 表结构信息（列、索引、外键、触发器）
				ds.GET("/:id/table-columns", sqlH.GetColumns)
				ds.GET("/:id/table-indexes", sqlH.GetIndexes)
				ds.GET("/:id/table-foreign-keys", sqlH.GetForeignKeys)
				ds.GET("/:id/table-triggers", sqlH.GetTableTriggers)
				// 数据表视图列表
				ds.GET("/:id/view-list", sqlH.GetViewList)

				ds.POST("/:id/data/insert", sqlH.InsertRow)
				ds.POST("/:id/data/update", sqlH.UpdateRow)
				ds.POST("/:id/data/delete", sqlH.DeleteRow)
				ds.GET("/:id/data/get-columns", sqlH.GetColumns)

				dataquery.GET("/capabilities", middleware.CacheReadThrough(10*time.Minute), sqlH.DatabaseCapabilities)
				ds.GET("/:id/ddl", sqlH.GetTableDDL)
				ds.POST("/:id/test-connection", sqlH.TestConnection)
			}

			dsH := handler.NewDatasourceHandler()
			datasources := authorized.Group("/datasources")
			{
				datasources.POST("", dsH.Create)
				datasources.GET("", middleware.CacheReadThrough(3*time.Minute), dsH.List)
				datasources.GET("/all", middleware.CacheReadThrough(3*time.Minute), dsH.AllSimple)
				datasources.GET("/:id", dsH.Get)
				datasources.GET("/:id/detail", dsH.GetDetail)
				datasources.GET("/:id/databases", sqlH.GetDatabases)
				datasources.GET("/:id/tables", sqlH.GetTables)
				datasources.PUT("/:id", dsH.Update)
				datasources.DELETE("/:id", dsH.Delete)
				datasources.POST("/:id/copy", dsH.Copy)
				datasources.POST("/:id/test", dsH.TestConnectionById)
			}
			authorized.POST("/datasources/test", dsH.TestConnectionFromForm)
			authorized.GET("/datasource-stats", dsH.Stats)

			metaH := handler.NewMetadataHandler()
			meta := authorized.Group("/metadata")
			{
				meta.GET("/:id/procedures", metaH.GetProcedures)
				meta.GET("/:id/triggers", metaH.GetTriggers)
				meta.GET("/:id/indexes", metaH.GetIndexes)
				meta.POST("/:id/maintenance/analyze", metaH.AnalyzeTable)
				meta.POST("/:id/maintenance/check", metaH.CheckTable)
				meta.POST("/:id/maintenance/optimize", metaH.OptimizeTable)
				meta.POST("/:id/maintenance/repair", metaH.RepairTable)
				meta.GET("/:id/maintenance/row-count", metaH.GetRowCount)
			}

			designH := handler.NewTableDesignHandler()
			design := authorized.Group("/table-design")
			{
				design.GET("/:id/ddl", designH.GetTableDDL)
				design.GET("/:id/info", designH.GetTableInfo)
				design.GET("/:id/columns", designH.GetColumns)
			}

			settingH := handler.NewDBSettingsHandler()
			settings := authorized.Group("/settings")
			{
				settings.GET("/db-types", settingH.GetDatabaseTypes)
			}

			planH := handler.NewExecutePlanHandler()
			authorized.POST("/exec-plan/explain", planH.Explain)

			exportH := handler.NewDataExportHandler()
			authorized.GET("/export/:id/csv", exportH.ExportCSV)

			bizH := handler.NewBusinessHandler()
			projects := authorized.Group("/projects")
			{
				projects.POST("", bizH.CreateProject)
				projects.GET("", bizH.ListProjects)
				projects.GET("/all", bizH.AllProjects)
				projects.PUT("/:id", bizH.UpdateProject)
				projects.DELETE("/:id", bizH.DeleteProject)
				projects.GET("/:id/overview", bizH.ProjectOverview)
				projects.GET("/:id/businesses", bizH.ListBusinessesByProject)
				projects.POST("/:id/businesses", bizH.CreateBusinessByProject)
				projects.GET("/:id/members", bizH.ListProjectMembers)
				projects.POST("/:id/members", bizH.AssignProjectMembers)
				projects.DELETE("/:id/members/:userId", bizH.RemoveProjectMember)
			}

			businesses := authorized.Group("/businesses")
			{
				businesses.POST("", bizH.CreateBusiness)
				businesses.GET("", bizH.ListBusinesses)
				businesses.GET("/all", bizH.AllBusinesses)
				businesses.PUT("/:id", bizH.UpdateBusiness)
				businesses.DELETE("/:id", bizH.DeleteBusiness)
				businesses.GET("/:id/overview", bizH.BusinessOverview)
				businesses.GET("/:id/members", bizH.ListBusinessMembers)
				businesses.POST("/:id/members", bizH.AssignBusinessMembers)
				businesses.DELETE("/:id/members/:userId", bizH.RemoveBusinessMember)
			}

			srvH := handler.NewServerHandler()
			servers := authorized.Group("/servers")
			{
				servers.POST("", srvH.Create)
				servers.GET("", srvH.List)
				servers.GET("/all", srvH.All)
				servers.GET("/stats", srvH.Stats)
				servers.GET("/:id", srvH.Get)
				servers.PUT("/:id", srvH.Update)
				servers.DELETE("/:id", srvH.Delete)
				servers.POST("/:id/test", srvH.TestConnect)
				servers.POST("/:id/toggle", srvH.ToggleStatus)
				servers.POST("/:id/exec", srvH.ExecCommand)
				servers.POST("/test", srvH.TestConnectByForm)
			}

			auditH := handler.NewAuditHandler()
			audit := authorized.Group("/audit")
			{
				audit.GET("", auditH.List)
				audit.GET("/stats", auditH.Stats)
			}

			// 首页概览聚合统计
			dashH := handler.NewDashboardHandler()
			authorized.GET("/dashboard/stats", dashH.Summary)

			maintH := handler.NewMaintenanceHandler()
			opsH := handler.NewOpsHandler()

			backup := authorized.Group("/backup-policies")
			{
				backup.POST("", maintH.CreateBackupPolicy)
				backup.GET("", maintH.ListBackupPolicies)
				backup.PUT("/:id", maintH.UpdateBackupPolicy)
				backup.DELETE("/:id", maintH.DeleteBackupPolicy)
			}
			authorized.POST("/backup/trigger", maintH.TriggerBackup)
			authorized.GET("/backup/records", maintH.ListBackupRecords)

			inspect := authorized.Group("/inspect-tasks")
			{
				inspect.POST("", maintH.CreateInspectTask)
				inspect.GET("", maintH.ListInspectTasks)
				inspect.PUT("/:id", maintH.UpdateInspectTask)
				inspect.DELETE("/:id", maintH.DeleteInspectTask)
			}
			authorized.POST("/inspect/trigger", maintH.TriggerInspect)
			authorized.GET("/inspect/reports", maintH.ListInspectReports)

			slow := authorized.Group("/slow-logs")
			{
				slow.GET("", maintH.ListSlowLogs)
				slow.GET("/top", maintH.TopSlow)
			}

			ha := authorized.Group("/ha-clusters")
			{
				ha.POST("", maintH.CreateHACluster)
				ha.GET("", maintH.ListHAClusters)
				ha.PUT("/:id", maintH.UpdateHACluster)
				ha.DELETE("/:id", maintH.DeleteHACluster)
			}

			plugins := authorized.Group("/plugins")
			{
				plugins.POST("", maintH.CreatePlugin)
				plugins.GET("", maintH.ListPlugins)
				plugins.PUT("/:id", maintH.UpdatePlugin)
				plugins.DELETE("/:id", maintH.DeletePlugin)
			}

			authorized.GET("/capacity/analyze", maintH.AnalyzeCapacity)

			dbusers := authorized.Group("/db-users")
			{
				dbusers.POST("", maintH.CreateDBUser)
				dbusers.GET("", maintH.ListDBUsers)
				dbusers.DELETE("/:id", maintH.DeleteDBUser)
			}

			authorized.GET("/platform-audit", maintH.PlatformAudit)

			// ====== 账号与角色管理（RBAC） ======
			accountH := handler.NewAccountHandler()
			accounts := authorized.Group("/account")
			{
				accounts.POST("", middleware.AdminRequired(), accountH.CreateAccount)
				accounts.GET("", middleware.AdminRequired(), accountH.ListAccounts)
				accounts.PUT("/:id", middleware.AdminRequired(), accountH.UpdateAccount)
				accounts.DELETE("/:id", middleware.AdminRequired(), accountH.DeleteAccount)
				accounts.POST("/:id/pwd", middleware.AdminRequired(), accountH.ResetAccountPassword)
				accounts.POST("/:id/lock", middleware.AdminRequired(), accountH.ToggleAccountLock)
				accounts.GET("/:id/roles", accountH.GetAccountRoles)
				accounts.POST("/:id/roles", middleware.AdminRequired(), accountH.AssignAccountRoles)
			}
			roles := authorized.Group("/role")
			{
				roles.POST("", middleware.AdminRequired(), accountH.CreateRole)
				roles.GET("", accountH.ListRoles)
				roles.GET("/all", accountH.GetAllRoles)
				roles.PUT("/:id", middleware.AdminRequired(), accountH.UpdateRole)
				roles.DELETE("/:id", middleware.AdminRequired(), accountH.DeleteRole)
				roles.POST("/:id/perm", middleware.AdminRequired(), accountH.AssignRolePermissions)
			}
			authorized.GET("/permission-codes", accountH.ListPermissionCodes)

			// ====== 管理域补充接口 ======
			importExport := authorized.Group("/import-export")
			{
				importExport.POST("/tasks", opsH.CreateImportExportTask)
				importExport.GET("/tasks", opsH.ListImportExportTasks)
			}
			sqlAudit := authorized.Group("/sql-audit")
			{
				sqlAudit.GET("/flows", opsH.ListAuditFlows)
				sqlAudit.POST("/flows", opsH.CreateAuditFlow)
				sqlAudit.GET("/rules", opsH.ListAuditRules)
			}
			sensitive := authorized.Group("/sensitive-data")
			{
				sensitive.GET("", opsH.ListSensitiveData)
				sensitive.POST("", opsH.CreateSensitiveData)
				sensitive.DELETE("/:id", opsH.DeleteSensitiveData)
			}
			lifecycle := authorized.Group("/db-lifecycle")
			{
				lifecycle.GET("/nodes", opsH.ListLifecycleNodes)
				lifecycle.GET("/dbs", opsH.ListLifecycleDBs)
			}
			health := authorized.Group("/health")
			{
				health.GET("/metrics", opsH.HealthMetrics)
				health.GET("/instances", opsH.HealthInstances)
				health.GET("/inspect", opsH.HealthInspectResults)
				health.POST("/inspect", opsH.TriggerHealthInspect)
			}
			mig := authorized.Group("/migration")
			{
				mig.GET("/tasks", opsH.ListMigrationTasks)
				mig.POST("/tasks", opsH.CreateMigrationTask)
				mig.GET("/schema-diff", opsH.SchemaDiff)
				mig.GET("/data-diff", opsH.DataDiff)
			}
			platform := authorized.Group("/platform")
			{
				platform.GET("/mediums", opsH.ListMediums)
			}

			// ====== Ferry 原生工作流 API（dbm-lite/ferry 整合实现） ======
			ferryAPI := authorized.Group("/ferry")
			{
				// 统计
				ferryAPI.GET("/statistics", ferry.StatisticsHandler)

				// 用户管理（管理员）
				ferryAPI.GET("/users", middleware.AdminRequired(), ferry.ListUsersHandler)
				ferryAPI.GET("/users/:id", middleware.AdminRequired(), ferry.GetUserHandler)
				ferryAPI.POST("/users", ferry.CreateUserHandler)
				ferryAPI.PUT("/users/:id", ferry.UpdateUserHandler)
				ferryAPI.DELETE("/users/:id", ferry.DeleteUserHandler)
				ferryAPI.POST("/users/:id/password", ferry.ResetUserPasswordHandler)

				// 角色管理（管理员）
				ferryAPI.GET("/roles", ferry.ListRolesHandler)
				ferryAPI.GET("/roles/all", ferry.ListAllRolesHandler)
				ferryAPI.POST("/roles", ferry.CreateRoleHandler)
				ferryAPI.PUT("/roles/:id", ferry.UpdateRoleHandler)
				ferryAPI.DELETE("/roles/:id", ferry.DeleteRoleHandler)

				// 岗位管理（管理员）
				ferryAPI.GET("/posts", ferry.ListPostsHandler)
				ferryAPI.POST("/posts", ferry.CreatePostHandler)
				ferryAPI.PUT("/posts/:id", ferry.UpdatePostHandler)
				ferryAPI.DELETE("/posts/:id", ferry.DeletePostHandler)

				// 系统配置（管理员）
				ferryAPI.GET("/system-settings", ferry.ListSystemSettingsHandler)
				ferryAPI.POST("/system-settings", ferry.SaveSystemSettingsHandler)

				// 部门管理（管理员）
				ferryAPI.GET("/departments", ferry.ListDepartmentsHandler)
				ferryAPI.GET("/departments/tree", ferry.ListAllDepartmentsHandler)
				ferryAPI.POST("/departments", ferry.CreateDepartmentHandler)
				ferryAPI.PUT("/departments/:id", ferry.UpdateDepartmentHandler)
				ferryAPI.DELETE("/departments/:id", ferry.DeleteDepartmentHandler)

				// 菜单管理（管理员）
				ferryAPI.GET("/menus", ferry.ListMenusHandler)
				ferryAPI.POST("/menus", ferry.CreateMenuHandler)
				ferryAPI.PUT("/menus/:id", ferry.UpdateMenuHandler)
				ferryAPI.DELETE("/menus/:id", ferry.DeleteMenuHandler)

				// 字典管理（管理员）
				ferryAPI.GET("/dictionaries", ferry.ListDictionariesHandler)
				ferryAPI.POST("/dictionaries", ferry.CreateDictionaryHandler)
				ferryAPI.PUT("/dictionaries/:id", ferry.UpdateDictionaryHandler)
				ferryAPI.DELETE("/dictionaries/:id", ferry.DeleteDictionaryHandler)
				ferryAPI.GET("/dict-items", ferry.ListDictItemsHandler)
				ferryAPI.POST("/dict-items", ferry.CreateDictItemHandler)
				ferryAPI.PUT("/dict-items/:id", ferry.UpdateDictItemHandler)
				ferryAPI.DELETE("/dict-items/:id", ferry.DeleteDictItemHandler)

				// 参数配置（管理员）
				ferryAPI.GET("/parameters", ferry.ListParametersHandler)
				ferryAPI.POST("/parameters", ferry.CreateParameterHandler)
				ferryAPI.PUT("/parameters/:id", ferry.UpdateParameterHandler)
				ferryAPI.DELETE("/parameters/:id", ferry.DeleteParameterHandler)

				// 流程定义
				ferryAPI.GET("/processes", ferry.ListProcessesHandler)
				ferryAPI.GET("/processes/enabled", ferry.ListEnabledProcessesHandler)
				ferryAPI.GET("/processes/:id", ferry.GetProcessHandler)
				ferryAPI.POST("/processes", middleware.AdminRequired(), ferry.CreateProcessHandler)
				ferryAPI.PUT("/processes/:id", middleware.AdminRequired(), ferry.UpdateProcessHandler)
				ferryAPI.DELETE("/processes/:id", middleware.AdminRequired(), ferry.DeleteProcessHandler)
				ferryAPI.POST("/processes/:id/clone", middleware.AdminRequired(), ferry.CloneProcessHandler)
				ferryAPI.POST("/processes/:id/enabled", middleware.AdminRequired(), ferry.ToggleProcessEnabledHandler)

				// 流程分类
				ferryAPI.GET("/classifies", ferry.ListClassifiesHandler)
				ferryAPI.POST("/classifies", middleware.AdminRequired(), ferry.CreateClassifyHandler)
				ferryAPI.PUT("/classifies/:id", middleware.AdminRequired(), ferry.UpdateClassifyHandler)
				ferryAPI.DELETE("/classifies/:id", middleware.AdminRequired(), ferry.DeleteClassifyHandler)

				// 表单 schema
				ferryAPI.GET("/processes/:id/apply-form", ferry.ApplyFormSchemaHandler)

				// 工单
				ferryAPI.POST("/work-orders", ferry.SubmitWorkOrderHandler)
				ferryAPI.GET("/work-orders/:id", ferry.GetWorkOrderDetailHandler)
				ferryAPI.POST("/work-orders/:id/tasks/:taskId/handle", ferry.HandleTaskHandler)
				ferryAPI.POST("/work-orders/:id/revoke", ferry.RevokeWorkOrderHandler)
				ferryAPI.POST("/work-orders/:id/urge", ferry.UrgeWorkOrderHandler)
				ferryAPI.GET("/work-orders/my/todo", ferry.TodoListHandler)
				ferryAPI.GET("/work-orders/my/apply", ferry.ApplyListHandler)
				ferryAPI.GET("/work-orders/my/related", ferry.RelatedListHandler)
				ferryAPI.GET("/work-orders", middleware.AdminRequired(), ferry.AllWorkOrdersHandler)
			}

			// ====== 对象权限与 SQL 校验 ======
			privH := handler.NewPrivilegeHandler()
			privApplies := authorized.Group("/priv/apply")
			{
				privApplies.POST("", privH.SubmitApply)
				privApplies.GET("", privH.ListApplies)
				privApplies.POST("/:id/approve", middleware.AdminRequired(), privH.ApproveApply)
				privApplies.POST("/:id/reject", middleware.AdminRequired(), privH.RejectApply)
			}
			authorized.GET("/priv/my", privH.MyPrivileges)
			authorized.GET("/priv/privileges", middleware.AdminRequired(), privH.ListAllPrivileges)
			authorized.POST("/priv/grant", middleware.AdminRequired(), privH.GrantPrivilege)
			authorized.POST("/priv/grant-batch", middleware.AdminRequired(), privH.BatchGrantPrivilege)
			authorized.DELETE("/priv/priv/:id", middleware.AdminRequired(), privH.RevokePrivilege)
			authorized.POST("/priv/revoke-batch", middleware.AdminRequired(), privH.BatchRevokePrivilege)
			authorized.GET("/priv/audit", middleware.AdminRequired(), privH.ListAuditLogs)
			authorized.POST("/priv/cleanup", middleware.AdminRequired(), privH.CleanupExpired)
			// 敏感列配置
			authorized.GET("/priv/sensitive-columns", privH.ListSensitiveColumns)
			authorized.POST("/priv/sensitive-columns", middleware.AdminRequired(), privH.CreateSensitiveColumn)
			authorized.DELETE("/priv/sensitive-columns/:id", middleware.AdminRequired(), privH.DeleteSensitiveColumn)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "API not found: " + c.Request.URL.Path})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "DBM-Lite API"})
	})

	port := config.App.ServerPort
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	timeout := time.Duration(config.App.APITimeout) * time.Second
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	go func() {
		log.Printf("DBM-Lite backend started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("DBM-Lite shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced shutdown: %v", err)
	}
	database.Shutdown()
	log.Println("DBM-Lite service has stopped")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "43200")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func frontendLogHandler(c *gin.Context) {
	var body struct {
		Level   string      `json:"level"`
		Message interface{} `json:"message"`
		Page    string      `json:"page"`
		User    string      `json:"user"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "param error"})
		return
	}
	level := strings.ToUpper(strings.TrimSpace(body.Level))
	if level == "" {
		level = "INFO"
	}
	user := body.User
	if user == "" {
		user = "anonymous"
	}
	page := body.Page
	if page == "" {
		page = c.Request.Referer()
	}
	logger.FrontendLog("%s [%s] user=%s page=%s %v", level, c.ClientIP(), user, page, body.Message)
	c.JSON(http.StatusOK, gin.H{"success": true})
}
