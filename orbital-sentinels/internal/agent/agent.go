package agent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/celestial/orbital-sentinels/internal/buffer"
	"github.com/celestial/orbital-sentinels/internal/client"
	"github.com/celestial/orbital-sentinels/internal/credentials"
	"github.com/celestial/orbital-sentinels/internal/heartbeat"
	"github.com/celestial/orbital-sentinels/internal/pkg/config"
	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"github.com/celestial/orbital-sentinels/internal/register"
	"github.com/celestial/orbital-sentinels/internal/scheduler"
	"github.com/celestial/orbital-sentinels/internal/sender"
	ping "github.com/celestial/orbital-sentinels/plugins/ping"
	"go.uber.org/zap"
)

// State 状态
type State int

const (
	StateInitializing State = iota // 初始化
	StateRegistering               // 注册中
	StateRunning                   // 运行中
	StateStopping                  // 停止中
	StateStopped                   // 已停止
	StateError                     // 错误
)

// Agent 主控制器
type Agent struct {
	config       *config.Config
	state        State
	pluginMgr    *plugin.Manager
	scheduler    *scheduler.Scheduler
	buffer       buffer.Buffer
	sender       *sender.Sender
	heartbeatMgr *heartbeat.Manager
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewAgent 创建 Agent
func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		config: cfg,
		state:  StateInitializing,
	}
}

// Start 启动 Agent
func (a *Agent) Start() error {
	a.ctx, a.cancel = context.WithCancel(context.Background())

	// 1. 处理注册和凭证
	a.setState(StateRegistering)
	if err := a.handleRegistration(); err != nil {
		logger.Warn("Failed to handle registration, will try to continue", zap.Error(err))
	}

	// 2. 初始化
	a.setState(StateInitializing)
	if err := a.initialize(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// 3. 启动各组件
	a.setState(StateRunning)
	a.startComponents()

	// 4. 监听信号
	a.handleSignals()

	return nil
}

// Stop 停止 Agent
func (a *Agent) Stop() error {
	a.setState(StateStopping)

	logger.Info("Stopping agent...")

	// 1. 停止心跳
	if a.heartbeatMgr != nil {
		a.heartbeatMgr.Stop()
	}

	// 2. 停止调度器
	if a.scheduler != nil {
		a.scheduler.Stop()
	}

	// 3. 停止发送器
	if a.sender != nil {
		a.sender.Stop()
	}

	// 4. 关闭缓冲区
	if a.buffer != nil {
		a.buffer.Close()
	}

	// 5. 停止插件
	if a.pluginMgr != nil {
		a.pluginMgr.StopAll()
	}

	// 6. 取消上下文
	if a.cancel != nil {
		a.cancel()
	}

	a.setState(StateStopped)
	logger.Info("Agent stopped")

	return nil
}

// initialize 初始化
func (a *Agent) initialize() error {
	logger.Info("Initializing agent...")

	// 1. 创建插件管理器并注册内置插件
	a.pluginMgr = plugin.NewManager(a.config.Plugins.Directory)
	if err := a.pluginMgr.LoadAll(); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// 注册内置插件
	a.registerBuiltinPlugins()

	// 2. 创建缓冲区
	switch a.config.Buffer.Type {
	case "memory":
		a.buffer = buffer.NewMemoryBuffer(a.config.Buffer.Size)
	case "disk":
		// TODO: 实现磁盘缓冲
		return fmt.Errorf("disk buffer not implemented yet")
	default:
		a.buffer = buffer.NewMemoryBuffer(a.config.Buffer.Size)
	}

	// 3. 创建发送器
	// 设置默认值
	flushInterval := a.config.Sender.FlushInterval
	if flushInterval <= 0 {
		flushInterval = 10 * time.Second
		logger.Info("Using default flush_interval",
			zap.Duration("interval", flushInterval))
	}

	batchSize := a.config.Sender.BatchSize
	if batchSize <= 0 {
		batchSize = 100
		logger.Info("Using default batch_size",
			zap.Int("size", batchSize))
	}

	senderConfig := &sender.Config{
		Mode:          sender.SendMode(a.config.Sender.Mode),
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
		Timeout:       a.config.Sender.Timeout,
		RetryTimes:    a.config.Sender.RetryTimes,
		RetryInterval: a.config.Sender.RetryInterval,
	}
	a.sender = sender.NewSender(senderConfig, a.buffer)

	// 配置中心端发送器
	if a.config.Sender.Mode == "core" || a.config.Sender.Mode == "hybrid" {
		coreSender := sender.NewCoreSender(
			a.config.Core.URL,
			a.config.Core.APIToken,
			a.config.Sender.Timeout,
		)
		a.sender.SetCoreSender(coreSender)
	}

	// 配置直连发送器
	if a.config.Sender.Mode == "direct" || a.config.Sender.Mode == "hybrid" {
		directSender := sender.NewDirectSender()

		// 配置 Prometheus
		if a.config.Sender.Direct.Prometheus.Enabled {
			promWriter := sender.NewPrometheusWriter(
				a.config.Sender.Direct.Prometheus.URL,
				a.config.Sender.Timeout,
			)
			if a.config.Sender.Direct.Prometheus.Username != "" {
				promWriter.SetBasicAuth(
					a.config.Sender.Direct.Prometheus.Username,
					a.config.Sender.Direct.Prometheus.Password,
				)
			}
			for k, v := range a.config.Sender.Direct.Prometheus.Headers {
				promWriter.SetHeader(k, v)
			}
			directSender.SetPrometheusWriter(promWriter)
			logger.Info("Prometheus writer configured",
				zap.String("url", a.config.Sender.Direct.Prometheus.URL))
		}

		// 配置 VictoriaMetrics
		if a.config.Sender.Direct.VictoriaMetrics.Enabled {
			vmWriter := sender.NewVictoriaMetricsWriter(
				a.config.Sender.Direct.VictoriaMetrics.URL,
				a.config.Sender.Timeout,
			)
			if a.config.Sender.Direct.VictoriaMetrics.Username != "" {
				vmWriter.SetBasicAuth(
					a.config.Sender.Direct.VictoriaMetrics.Username,
					a.config.Sender.Direct.VictoriaMetrics.Password,
				)
			}
			for k, v := range a.config.Sender.Direct.VictoriaMetrics.Headers {
				vmWriter.SetHeader(k, v)
			}
			directSender.SetVictoriaMetricsWriter(vmWriter)
			logger.Info("VictoriaMetrics writer configured",
				zap.String("url", a.config.Sender.Direct.VictoriaMetrics.URL))
		}

		// 配置 ClickHouse
		if a.config.Sender.Direct.ClickHouse.Enabled {
			tableName := a.config.Sender.Direct.ClickHouse.TableName
			if tableName == "" {
				tableName = "metrics" // 默认表名
			}
			batchSize := a.config.Sender.Direct.ClickHouse.BatchSize
			if batchSize == 0 {
				batchSize = 1000 // 默认批量大小
			}

			chWriter, err := sender.NewClickHouseWriter(&sender.ClickHouseConfig{
				DSN:       a.config.Sender.Direct.ClickHouse.DSN,
				TableName: tableName,
				BatchSize: batchSize,
			})
			if err != nil {
				logger.Error("Failed to create ClickHouse writer", zap.Error(err))
			} else {
				directSender.SetClickHouseWriter(chWriter)
				logger.Info("ClickHouse writer configured",
					zap.String("table", tableName))
			}
		}

		a.sender.SetDirectSender(directSender)
	}

	// 4. 创建调度器
	a.scheduler = scheduler.NewScheduler(
		a.pluginMgr,
		a.config.Collector.WorkerPoolSize,
		a.config.Collector.TaskFetchInterval,
	)

	// 设置指标处理器
	a.scheduler.SetMetricsHandler(func(metrics []*plugin.Metric, task *plugin.CollectionTask) {
		if err := a.buffer.Push(metrics); err != nil {
			logger.Error("Failed to push metrics to buffer", zap.Error(err))
		}
	})

	// 如果配置了中心端 URL 和 Token，创建任务客户端用于从中心端获取任务
	if a.config.Core.URL != "" && a.config.Core.APIToken != "" {
		taskClient := client.NewTaskClient(
			a.config.Core.URL,
			a.config.Core.APIToken,
			a.config.Sentinel.ID,
			a.config.Sender.Timeout,
		)
		a.scheduler.SetTaskClient(taskClient)
		logger.Info("Task client configured for fetching tasks from core",
			zap.String("core_url", a.config.Core.URL),
			zap.Duration("fetch_interval", a.config.Collector.TaskFetchInterval))
	} else {
		logger.Info("Task client not configured (core URL or token missing), will only use local tasks")
	}

	// 5. 创建心跳管理器
	sentinelID := a.config.Sentinel.ID
	if sentinelID == "" {
		sentinelID = generateSentinelID()
	}

	a.heartbeatMgr = heartbeat.NewManager(
		a.config.Core.URL,
		a.config.Core.APIToken,
		sentinelID,
		a.config.Heartbeat.Interval,
		a.config.Heartbeat.Timeout,
		a.config.Heartbeat.RetryTimes,
	)

	logger.Info("Agent initialized",
		zap.String("sentinel_id", sentinelID),
		zap.String("name", a.config.Sentinel.Name))

	return nil
}

// startComponents 启动各组件
func (a *Agent) startComponents() {
	// 启动发送器
	a.sender.Start(a.ctx)

	// 启动调度器
	a.scheduler.Start(a.ctx)

	// 加载本地任务（Direct 模式或配置了本地任务时）
	if len(a.config.Tasks) > 0 {
		logger.Info("Loading local tasks from config", zap.Int("count", len(a.config.Tasks)))
		a.loadLocalTasks()
	}

	// 启动心跳
	a.heartbeatMgr.Start(a.ctx)

	logger.Info("All components started")
}

// registerBuiltinPlugins 注册内置插件
func (a *Agent) registerBuiltinPlugins() {
	// 注册 Ping 插件
	pingPlugin := ping.NewPlugin()
	if err := pingPlugin.Init(nil); err != nil {
		logger.Error("Failed to initialize ping plugin", zap.Error(err))
		return
	}
	if err := a.pluginMgr.RegisterPlugin(pingPlugin); err != nil {
		logger.Error("Failed to register ping plugin", zap.Error(err))
		return
	}
	logger.Info("Registered builtin plugin", zap.String("name", "ping"))
}

// loadLocalTasks 加载本地任务配置
func (a *Agent) loadLocalTasks() {
	successCount := 0
	failedCount := 0

	for _, taskCfg := range a.config.Tasks {
		// 跳过未启用的任务
		if !taskCfg.Enabled {
			logger.Debug("Skipping disabled task", zap.String("task_id", taskCfg.ID))
			continue
		}

		// 解析 interval
		interval, err := time.ParseDuration(taskCfg.Interval)
		if err != nil {
			logger.Error("Invalid task interval",
				zap.String("task_id", taskCfg.ID),
				zap.String("interval", taskCfg.Interval),
				zap.Error(err))
			failedCount++
			continue
		}

		// 解析 timeout（可选）
		timeout := 30 * time.Second // 默认 30 秒
		if taskCfg.Timeout != "" {
			timeout, err = time.ParseDuration(taskCfg.Timeout)
			if err != nil {
				logger.Warn("Invalid task timeout, using default",
					zap.String("task_id", taskCfg.ID),
					zap.String("timeout", taskCfg.Timeout),
					zap.Duration("default", timeout))
			}
		}

		// 创建采集任务
		task := &plugin.CollectionTask{
			TaskID:       taskCfg.ID,
			DeviceID:     taskCfg.DeviceID,
			PluginName:   taskCfg.Plugin,
			DeviceConfig: taskCfg.Config,
			Timeout:      timeout,
		}

		// 添加到调度器
		a.scheduler.AddTask(task, interval)

		logger.Info("Loaded local task",
			zap.String("task_id", taskCfg.ID),
			zap.String("device_id", taskCfg.DeviceID),
			zap.String("plugin", taskCfg.Plugin),
			zap.Duration("interval", interval))

		successCount++
	}

	logger.Info("Local tasks loaded",
		zap.Int("success", successCount),
		zap.Int("failed", failedCount),
		zap.Int("total", len(a.config.Tasks)))
}

// handleSignals 处理信号
func (a *Agent) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal", zap.String("signal", sig.String()))
		a.Stop()
	}()
}

// setState 设置状态
func (a *Agent) setState(state State) {
	a.state = state
	logger.Debug("State changed", zap.Int("state", int(state)))
}

// GetState 获取状态
func (a *Agent) GetState() State {
	return a.state
}

// generateSentinelID 生成 Sentinel ID
func generateSentinelID() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("sentinel-%s-%d", hostname, time.Now().Unix())
}

// handleRegistration 处理注册和凭证
func (a *Agent) handleRegistration() error {
	// 1. 初始化凭证管理器
	credsMgr := credentials.NewManager(a.config.CredentialsPath)

	// 2. 尝试加载本地凭证
	creds, err := credsMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// 3. 如果没有凭证或凭证无效,进行注册
	if creds == nil || !creds.IsValid() {
		logger.Info("No valid credentials found, attempting to register to core...")

		// 检查是否配置了中心端 URL
		if a.config.Core.URL == "" {
			logger.Warn("Core URL not configured, skipping registration")
			return nil
		}

		registerMgr := register.NewManager(
			a.config.Core.URL,
			a.config.Core.RegistrationKey,
			credsMgr,
		)

		registerConfig := &register.Config{
			Name:    a.config.Sentinel.Name,
			Version: "1.0.0", // TODO: 从编译时注入
			Region:  a.config.Sentinel.Region,
			Labels:  a.config.Sentinel.Labels,
		}

		resp, err := registerMgr.RegisterWithRetry(a.ctx, registerConfig)
		if err != nil {
			logger.Warn("Failed to register to core, falling back to standalone mode",
				zap.Error(err))
			// 降级为独立模式
			a.config.Sender.Mode = "direct"
			return err
		}

		logger.Info("Successfully registered to core",
			zap.String("sentinel_id", resp.SentinelID))
		creds = credsMgr.GetCredentials()
	} else {
		logger.Info("Using existing credentials",
			zap.String("sentinel_id", creds.SentinelID),
			zap.Time("registered_at", creds.RegisteredAt))

		// 验证凭证有效性
		if a.config.Core.URL != "" && !a.validateCredentials(creds) {
			logger.Warn("Credentials validation failed, attempting to re-register...")
			// 删除旧凭证并重新注册
			credsMgr.Delete()

			// 递归调用重新注册
			return a.handleRegistration()
		}
	}

	// 4. 使用凭证更新配置
	if creds != nil && creds.IsValid() {
		a.config.Sentinel.ID = creds.SentinelID
		a.config.Core.APIToken = creds.APIToken
		if creds.CoreURL != "" {
			a.config.Core.URL = creds.CoreURL
		}
	}

	return nil
}

// validateCredentials 验证凭证有效性
func (a *Agent) validateCredentials(creds *credentials.Credentials) bool {
	logger.Debug("Validating credentials", zap.String("sentinel_id", creds.SentinelID))

	// 发送一次心跳测试
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("POST",
		creds.CoreURL+"/api/v1/sentinels/heartbeat",
		bytes.NewReader([]byte("{}")))
	if err != nil {
		logger.Debug("Failed to create validation request", zap.Error(err))
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sentinel-ID", creds.SentinelID)
	req.Header.Set("X-API-Token", creds.APIToken)

	resp, err := client.Do(req)
	if err != nil {
		logger.Debug("Credentials validation failed", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	isValid := resp.StatusCode == http.StatusOK
	logger.Debug("Credentials validation result",
		zap.Bool("valid", isValid),
		zap.Int("status_code", resp.StatusCode))

	return isValid
}
