package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/celestial/gravital-core/internal/alert/engine"
	"github.com/celestial/gravital-core/internal/api/router"
	"github.com/celestial/gravital-core/internal/pkg/cache"
	"github.com/celestial/gravital-core/internal/pkg/config"
	"github.com/celestial/gravital-core/internal/pkg/database"
	"github.com/celestial/gravital-core/internal/pkg/logger"
	"github.com/celestial/gravital-core/internal/repository"
	"github.com/celestial/gravital-core/internal/service"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

var (
	configPath  = flag.String("c", "config/config.yaml", "配置文件路径")
	showVersion = flag.Bool("v", false, "显示版本信息")
)

func main() {
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Gravital Core\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Go Version: %s\n", GoVersion)
		return
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(&logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		Output:     cfg.Logging.Output,
		FilePath:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Gravital Core",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
	)

	// 初始化数据库
	logger.Info("Connecting to database...")
	db, err := database.Init(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect database", zap.Error(err))
	}
	defer database.Close()
	logger.Info("Database connected")

	// 初始化 Redis
	logger.Info("Connecting to Redis...")
	if _, err := cache.Init(&cfg.Redis); err != nil {
		logger.Fatal("Failed to connect Redis", zap.Error(err))
	}
	defer cache.Close()
	logger.Info("Redis connected")

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r, forwarderService := router.Setup(cfg, db)

	// 启动转发服务
	logger.Info("Starting forwarder service...")
	if err := forwarderService.Start(); err != nil {
		logger.Fatal("Failed to start forwarder service", zap.Error(err))
	}
	logger.Info("Forwarder service started")

	// 启动设备监控服务
	logger.Info("Starting device monitor...")
	deviceMonitor := service.NewDeviceMonitor(db, logger.Get(), &service.DeviceMonitorConfig{
		CheckInterval:  1 * time.Minute,
		OfflineTimeout: 5 * time.Minute,
	})
	deviceMonitor.Start()
	logger.Info("Device monitor started")

	// 启动拓扑自动发现调度器
	logger.Info("Starting topology discovery scheduler...")
	// 需要从 router 中获取服务，或者在这里重新创建
	// 为了简化，我们在这里重新创建必要的依赖
	topologyRepo := repository.NewTopologyRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	topologyDiscoveryService := service.NewTopologyDiscoveryService(topologyRepo, deviceRepo, logger.Get())
	topologyDiscoveryScheduler := service.NewTopologyDiscoveryScheduler(
		topologyRepo,
		topologyDiscoveryService,
		db,
		logger.Get(),
		&service.TopologyDiscoverySchedulerConfig{
			CheckInterval:   5 * time.Minute, // 每 5 分钟检查一次
			CleanupInterval: 1 * time.Hour,   // 每小时清理一次过期数据
		},
	)
	topologyDiscoveryScheduler.Start()
	logger.Info("Topology discovery scheduler started")

	// 启动告警引擎
	logger.Info("Starting alert engine...")

	// 从转发器配置中查找 VictoriaMetrics 端点
	vmURL := ""
	for _, target := range cfg.Forwarder.Targets {
		if target.Type == "victoriametrics" && target.Enabled {
			vmURL = target.Endpoint
			break
		}
	}

	alertEngine := engine.NewAlertEngine(db, logger.Get(), &engine.Config{
		VMURL:         vmURL,
		CheckInterval: 30 * time.Second, // 每 30 秒检查一次
	})
	alertEngine.Start()
	logger.Info("Alert engine started")

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:           cfg.Server.GetAddr(),
		Handler:        r,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// 启动服务器
	go func() {
		logger.Info("Server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止告警引擎
	logger.Info("Stopping alert engine...")
	alertEngine.Stop()

	// 停止设备监控服务
	logger.Info("Stopping device monitor...")
	deviceMonitor.Stop()

	// 停止拓扑自动发现调度器
	logger.Info("Stopping topology discovery scheduler...")
	topologyDiscoveryScheduler.Stop()

	// 停止转发服务
	logger.Info("Stopping forwarder service...")
	if err := forwarderService.Stop(); err != nil {
		logger.Error("Failed to stop forwarder service", zap.Error(err))
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
