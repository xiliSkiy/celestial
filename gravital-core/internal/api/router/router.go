package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/api/handler"
	"github.com/celestial/gravital-core/internal/api/middleware"
	"github.com/celestial/gravital-core/internal/pkg/auth"
	"github.com/celestial/gravital-core/internal/pkg/config"
	"github.com/celestial/gravital-core/internal/pkg/logger"
	"github.com/celestial/gravital-core/internal/repository"
	"github.com/celestial/gravital-core/internal/service"
	"github.com/celestial/gravital-core/internal/timeseries"
)

// Setup 设置路由
func Setup(cfg *config.Config, db *gorm.DB) (*gin.Engine, service.ForwarderService) {
	r := gin.New()

	// 全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// 健康检查
	r.GET("/health", handler.HealthCheck)
	r.GET("/version", handler.Version)

	// 初始化依赖
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTExpire,
		cfg.Auth.RefreshExpire,
	)

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	sentinelRepo := repository.NewSentinelRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	forwarderRepo := repository.NewForwarderRepository(db)

	// 获取 logger
	log := logger.Get()

	// 初始化时序数据库客户端
	var tsClient *timeseries.Client
	if cfg.TimeSeries.Enabled && cfg.TimeSeries.URL != "" {
		tsClient = timeseries.NewClient(cfg.TimeSeries.URL, log)
		log.Info("Time series database client initialized", zap.String("url", cfg.TimeSeries.URL))

		// 健康检查
		if err := tsClient.Health(); err != nil {
			log.Warn("Time series database health check failed", zap.Error(err))
		}
	} else {
		log.Info("Time series database is not configured, metrics will not be available")
	}

	// 初始化 Service
	authService := service.NewAuthService(userRepo, jwtManager, cfg.Auth.BcryptCost)
	deviceService := service.NewDeviceService(deviceRepo, db, tsClient)
	sentinelService := service.NewSentinelService(sentinelRepo)
	taskService := service.NewTaskService(taskRepo, deviceRepo, sentinelRepo)
	alertService := service.NewAlertService(alertRepo)
	forwarderService := service.NewForwarderService(forwarderRepo, cfg, log)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	sentinelHandler := handler.NewSentinelHandler(sentinelService)
	taskHandler := handler.NewTaskHandler(taskService)
	alertHandler := handler.NewAlertHandler(alertService, db)
	forwarderHandler := handler.NewForwarderHandler(forwarderService, db, log)
	dashboardHandler := handler.NewDashboardHandler(db)
	userHandler := handler.NewUserHandler(db, cfg.Auth.BcryptCost)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.Auth(jwtManager), authHandler.Logout)
			auth.GET("/me", middleware.Auth(jwtManager), authHandler.GetCurrentUser)
		}

		// 需要认证的路由
		authenticated := v1.Group("")
		authenticated.Use(middleware.Auth(jwtManager))
		{
			// 设备管理
			devices := authenticated.Group("/devices")
			{
				devices.GET("", deviceHandler.List)
				devices.GET("/tags", deviceHandler.GetTags)
				devices.GET("/:id", deviceHandler.Get)
				devices.GET("/:id/metrics", deviceHandler.GetMetrics)
				devices.GET("/:id/tasks", deviceHandler.GetTasks)
				devices.GET("/:id/alert-rules", deviceHandler.GetAlertRules)
				devices.GET("/:id/history", deviceHandler.GetHistory)
				devices.POST("", middleware.RequirePermission("devices.write"), deviceHandler.Create)
				devices.PUT("/:id", middleware.RequirePermission("devices.write"), deviceHandler.Update)
				devices.DELETE("/:id", middleware.RequirePermission("devices.delete"), deviceHandler.Delete)
				devices.POST("/batch-import", middleware.RequirePermission("devices.write"), deviceHandler.BatchImport)
				devices.POST("/:id/test-connection", deviceHandler.TestConnection)
			}

			// 设备分组
			groups := authenticated.Group("/device-groups")
			{
				groups.GET("/tree", deviceHandler.GetGroupTree)
				groups.POST("", middleware.RequirePermission("devices.write"), deviceHandler.CreateGroup)
				groups.PUT("/:id", middleware.RequirePermission("devices.write"), deviceHandler.UpdateGroup)
				groups.DELETE("/:id", middleware.RequirePermission("devices.delete"), deviceHandler.DeleteGroup)
			}

			// 告警规则
			alertRules := authenticated.Group("/alert-rules")
			{
				alertRules.GET("", alertHandler.ListRules)
				alertRules.GET("/:id", alertHandler.GetRule)
				alertRules.POST("", middleware.RequirePermission("alerts.write"), alertHandler.CreateRule)
				alertRules.PUT("/:id", middleware.RequirePermission("alerts.write"), alertHandler.UpdateRule)
				alertRules.DELETE("/:id", middleware.RequirePermission("alerts.delete"), alertHandler.DeleteRule)
				alertRules.POST("/:id/toggle", middleware.RequirePermission("alerts.write"), alertHandler.ToggleRule)
				alertRules.POST("/:id/resolve-all", middleware.RequirePermission("alerts.write"), alertHandler.ResolveByRule)
			}

			// 告警事件
			alertEvents := authenticated.Group("/alert-events")
			{
				alertEvents.GET("", alertHandler.ListEvents)
				alertEvents.GET("/:id", alertHandler.GetEvent)
				alertEvents.POST("/:id/acknowledge", alertHandler.AcknowledgeEvent)
				alertEvents.POST("/:id/resolve", alertHandler.ResolveEvent)
				alertEvents.POST("/:id/silence", alertHandler.SilenceEvent)
				alertEvents.POST("/batch-acknowledge", middleware.RequirePermission("alerts.write"), alertHandler.BatchAcknowledge)
				alertEvents.POST("/batch-resolve", middleware.RequirePermission("alerts.write"), alertHandler.BatchResolve)
			}

			// 告警统计和聚合
			authenticated.GET("/alert-stats", alertHandler.GetStats)
			authenticated.GET("/alert-aggregations", alertHandler.GetAggregations)

			// 任务管理（管理端）
			tasks := authenticated.Group("/tasks")
			{
				tasks.GET("", taskHandler.List)
				tasks.GET("/:id", taskHandler.Get)
				tasks.POST("", middleware.RequirePermission("tasks.write"), taskHandler.Create)
				tasks.PUT("/:id", middleware.RequirePermission("tasks.write"), taskHandler.Update)
				tasks.DELETE("/:id", middleware.RequirePermission("tasks.delete"), taskHandler.Delete)
				tasks.PATCH("/:id", middleware.RequirePermission("tasks.write"), taskHandler.Toggle)
				tasks.POST("/:id/trigger", middleware.RequirePermission("tasks.write"), taskHandler.Trigger)
				tasks.GET("/:id/executions", taskHandler.GetExecutions)
			}

			// Sentinel 管理
			sentinels := authenticated.Group("/sentinels")
			{
				sentinels.GET("", sentinelHandler.List)
				sentinels.GET("/:id", sentinelHandler.Get)
				sentinels.DELETE("/:id", middleware.RequirePermission("sentinels.delete"), sentinelHandler.Delete)
				sentinels.POST("/:id/control", middleware.RequirePermission("sentinels.control"), sentinelHandler.Control)
			}

			// 系统管理
			system := authenticated.Group("/system")
			{
				system.GET("/info", handler.SystemInfo)
				system.GET("/config", middleware.RequirePermission("admin.config"), handler.GetConfig)
				system.PUT("/config", middleware.RequirePermission("admin.config"), handler.UpdateConfig)
			}

			// Dashboard
			dashboard := authenticated.Group("/dashboard")
			{
				dashboard.GET("/stats", dashboardHandler.GetStats)
				dashboard.GET("/device-status", dashboardHandler.GetDeviceStatus)
				dashboard.GET("/alert-trend", dashboardHandler.GetAlertTrend)
				dashboard.GET("/sentinel-status", dashboardHandler.GetSentinelStatus)
				dashboard.GET("/forwarder-stats", dashboardHandler.GetForwarderStats)
				dashboard.GET("/activities", dashboardHandler.GetActivities)
			}

			// 用户管理
			users := authenticated.Group("/users")
			{
				users.GET("", userHandler.ListUsers)
				users.GET("/:id", userHandler.GetUser)
				users.POST("", middleware.RequirePermission("admin.config"), userHandler.CreateUser)
				users.PUT("/:id", middleware.RequirePermission("admin.config"), userHandler.UpdateUser)
				users.DELETE("/:id", middleware.RequirePermission("admin.config"), userHandler.DeleteUser)
				users.PATCH("/:id", middleware.RequirePermission("admin.config"), userHandler.ToggleUser)
				users.POST("/:id/reset-password", middleware.RequirePermission("admin.config"), userHandler.ResetPassword)
			}

			// 角色管理
			roles := authenticated.Group("/roles")
			{
				roles.GET("", userHandler.ListRoles)
				roles.GET("/:id", userHandler.GetRole)
				roles.POST("", middleware.RequirePermission("admin.config"), userHandler.CreateRole)
				roles.PUT("/:id", middleware.RequirePermission("admin.config"), userHandler.UpdateRole)
				roles.DELETE("/:id", middleware.RequirePermission("admin.config"), userHandler.DeleteRole)
			}
		}

		// Sentinel API（注意：这些路由在 authenticated 组之外，不需要 JWT 认证）
		// 注册接口（不需要任何认证，因为注册的目的就是获取凭证）
		v1.POST("/sentinels/register", sentinelHandler.Register)

		// Sentinel 心跳接口（需要 Sentinel 认证）
		v1.POST("/sentinels/heartbeat", middleware.SentinelAuth(), sentinelHandler.Heartbeat)

		// 任务 API（Sentinel 调用）- 使用不同的路径避免冲突
		sentinelTasks := v1.Group("/sentinel-tasks")
		sentinelTasks.Use(middleware.SentinelAuth())
		{
			sentinelTasks.GET("", taskHandler.GetSentinelTasks)
			sentinelTasks.POST("/:id/report", taskHandler.ReportExecution)
		}

		// 数据采集 API（Sentinel 调用）
		data := v1.Group("/data")
		data.Use(middleware.SentinelAuth())
		{
			data.POST("/ingest", forwarderHandler.IngestMetrics)
		}

		// 转发器管理（需要认证）
		forwarders := authenticated.Group("/forwarders")
		{
			forwarders.GET("", forwarderHandler.ListForwarders)
			forwarders.GET("/stats", forwarderHandler.GetAllStats)
			forwarders.GET("/:name", forwarderHandler.GetForwarder)
			forwarders.GET("/:name/stats", forwarderHandler.GetForwarderStats)
			forwarders.POST("", middleware.RequirePermission("admin.config"), forwarderHandler.CreateForwarder)
			forwarders.PUT("/:name", middleware.RequirePermission("admin.config"), forwarderHandler.UpdateForwarder)
			forwarders.DELETE("/:name", middleware.RequirePermission("admin.config"), forwarderHandler.DeleteForwarder)
			forwarders.POST("/reload", middleware.RequirePermission("admin.config"), forwarderHandler.ReloadConfig)
			forwarders.POST("/test", middleware.RequirePermission("admin.config"), forwarderHandler.TestConnection)
		}
	}

	return r, forwarderService
}
