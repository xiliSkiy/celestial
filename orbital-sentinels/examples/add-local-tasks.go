package main

// 示例：如何在 Direct 模式下添加本地采集任务
// 
// 使用方法：
// 1. 将此文件的代码复制到 internal/agent/agent.go
// 2. 在 Start() 方法中调用 addLocalTasks()
// 3. 重新编译：make build
// 4. 启动：./bin/sentinel start -c config.yaml

import (
	"time"

	"github.com/celestial/orbital-sentinels/internal/plugin"
	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"go.uber.org/zap"
)

// addLocalTasks 添加本地采集任务
// 在 Agent.Start() 方法中调用此函数
func addLocalTasks(scheduler interface{}, config interface{}) {
	logger.Info("Adding local collection tasks")

	// ============================================================
	// 示例 1: Ping 本地网关
	// ============================================================
	task1 := &plugin.CollectionTask{
		TaskID:     "local-ping-gateway",
		DeviceID:   "192.168.1.1",
		PluginName: "ping",
		DeviceConfig: map[string]interface{}{
			"host":     "192.168.1.1",
			"count":    4,
			"interval": "1s",
			"timeout":  "5s",
		},
		Timeout: 10 * time.Second,
	}
	// scheduler.AddTask(task1, 60*time.Second) // 每 60 秒执行一次
	logger.Info("Added task",
		zap.String("task_id", "local-ping-gateway"),
		zap.String("device", "192.168.1.1"),
		zap.String("interval", "60s"))

	// ============================================================
	// 示例 2: Ping Google DNS
	// ============================================================
	task2 := &plugin.CollectionTask{
		TaskID:     "local-ping-google-dns",
		DeviceID:   "8.8.8.8",
		PluginName: "ping",
		DeviceConfig: map[string]interface{}{
			"host":     "8.8.8.8",
			"count":    4,
			"interval": "1s",
			"timeout":  "5s",
		},
		Timeout: 10 * time.Second,
	}
	// scheduler.AddTask(task2, 300*time.Second) // 每 5 分钟执行一次
	logger.Info("Added task",
		zap.String("task_id", "local-ping-google-dns"),
		zap.String("device", "8.8.8.8"),
		zap.String("interval", "300s"))

	// ============================================================
	// 示例 3: Ping 阿里云 DNS
	// ============================================================
	task3 := &plugin.CollectionTask{
		TaskID:     "local-ping-aliyun-dns",
		DeviceID:   "223.5.5.5",
		PluginName: "ping",
		DeviceConfig: map[string]interface{}{
			"host":     "223.5.5.5",
			"count":    4,
			"interval": "1s",
			"timeout":  "5s",
		},
		Timeout: 10 * time.Second,
	}
	// scheduler.AddTask(task3, 300*time.Second) // 每 5 分钟执行一次
	logger.Info("Added task",
		zap.String("task_id", "local-ping-aliyun-dns"),
		zap.String("device", "223.5.5.5"),
		zap.String("interval", "300s"))

	// ============================================================
	// 示例 4: 批量添加多个设备
	// ============================================================
	devices := []struct {
		ID   string
		Host string
	}{
		{"server-1", "192.168.1.10"},
		{"server-2", "192.168.1.11"},
		{"server-3", "192.168.1.12"},
		{"switch-1", "192.168.1.100"},
		{"router-1", "192.168.1.1"},
	}

	for _, device := range devices {
		task := &plugin.CollectionTask{
			TaskID:     "local-ping-" + device.ID,
			DeviceID:   device.Host,
			PluginName: "ping",
			DeviceConfig: map[string]interface{}{
				"host":     device.Host,
				"count":    4,
				"interval": "1s",
				"timeout":  "5s",
			},
			Timeout: 10 * time.Second,
		}
		// scheduler.AddTask(task, 60*time.Second) // 每 60 秒执行一次
		logger.Info("Added task",
			zap.String("task_id", "local-ping-"+device.ID),
			zap.String("device", device.Host))
	}

	logger.Info("All local tasks added", zap.Int("total", 3+len(devices)))
}

// ============================================================
// 集成到 Agent 的示例代码
// ============================================================

/*
// 在 internal/agent/agent.go 中添加以下代码：

// Start 启动 Agent
func (a *Agent) Start() error {
	logger.Info("Initializing agent...")

	// 初始化各个组件
	if err := a.initialize(); err != nil {
		return err
	}

	// ⭐ 添加这段代码 ⭐
	// Direct 模式下添加本地任务
	if a.config.Sender.Mode == "direct" {
		logger.Info("Running in direct mode, adding local tasks")
		a.addLocalTasks()
	}

	// 启动各个组件
	a.sender.Start(a.ctx)
	a.scheduler.Start(a.ctx)
	a.heartbeat.Start(a.ctx)

	logger.Info("All components started")

	// 监听信号
	go a.handleSignals()

	return nil
}

// ⭐ 添加这个方法 ⭐
func (a *Agent) addLocalTasks() {
	logger.Info("Adding local collection tasks")

	// 任务 1: Ping 网关
	task1 := &plugin.CollectionTask{
		TaskID:     "local-ping-gateway",
		DeviceID:   "192.168.1.1",
		PluginName: "ping",
		DeviceConfig: map[string]interface{}{
			"host":     "192.168.1.1",
			"count":    4,
			"interval": "1s",
			"timeout":  "5s",
		},
		Timeout: 10 * time.Second,
	}
	a.scheduler.AddTask(task1, 60*time.Second)
	logger.Info("Added task", zap.String("task_id", "local-ping-gateway"))

	// 任务 2: Ping DNS
	task2 := &plugin.CollectionTask{
		TaskID:     "local-ping-dns",
		DeviceID:   "8.8.8.8",
		PluginName: "ping",
		DeviceConfig: map[string]interface{}{
			"host":     "8.8.8.8",
			"count":    4,
			"interval": "1s",
			"timeout":  "5s",
		},
		Timeout: 10 * time.Second,
	}
	a.scheduler.AddTask(task2, 300*time.Second)
	logger.Info("Added task", zap.String("task_id", "local-ping-dns"))

	logger.Info("All local tasks added", zap.Int("total", 2))
}
*/

// ============================================================
// 配置文件示例
// ============================================================

/*
# config.yaml

sentinel:
  name: "sentinel-standalone"
  region: "local"

# Direct 模式配置
sender:
  mode: "direct"  # 使用 direct 模式
  batch_size: 1000
  flush_interval: 10s
  timeout: 30s
  
  direct:
    prometheus:
      enabled: true
      url: "http://localhost:9090/api/v1/write"

collector:
  worker_pool_size: 10
  task_fetch_interval: 60s  # 这个参数在 direct 模式下不使用

buffer:
  type: "memory"
  size: 10000
  flush_interval: 10s

plugins:
  directory: "./plugins"

logging:
  level: info
  format: json
  output: both
  file_path: "./logs/sentinel.log"
*/

// ============================================================
// 使用步骤
// ============================================================

/*
1. 复制 addLocalTasks() 方法到 internal/agent/agent.go

2. 在 Start() 方法中添加调用：
   if a.config.Sender.Mode == "direct" {
       a.addLocalTasks()
   }

3. 配置 config.yaml 为 direct 模式

4. 重新编译：
   make build

5. 启动服务：
   ./bin/sentinel start -c config.yaml

6. 查看日志确认任务已添加：
   tail -f logs/sentinel.log | grep "Added task"

7. 查看任务执行情况：
   tail -f logs/sentinel.log | grep "Task succeeded"

8. 在 Prometheus 中查询数据：
   ping_rtt_ms{host="192.168.1.1"}
*/

