package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/celestial/orbital-sentinels/internal/client"
	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"go.uber.org/zap"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusFailed  TaskStatus = "failed"
)

// ScheduledTask 调度任务
type ScheduledTask struct {
	Task         *plugin.CollectionTask
	NextRun      time.Time
	Interval     time.Duration
	LastStatus   TaskStatus
	LastError    error
	ExecutionCnt int
	mu           sync.Mutex
}

// Scheduler 任务调度器
type Scheduler struct {
	pluginMgr     *plugin.Manager
	tasks         map[string]*ScheduledTask
	workerPool    *WorkerPool
	fetchInterval time.Duration
	taskClient    *client.TaskClient
	onMetrics     func([]*plugin.Metric, *plugin.CollectionTask)
	onReport      func(*ScheduledTask, time.Duration, int)
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewScheduler 创建调度器
func NewScheduler(
	pluginMgr *plugin.Manager,
	workerPoolSize int,
	fetchInterval time.Duration,
) *Scheduler {
	return &Scheduler{
		pluginMgr:     pluginMgr,
		tasks:         make(map[string]*ScheduledTask),
		workerPool:    NewWorkerPool(workerPoolSize),
		fetchInterval: fetchInterval,
	}
}

// SetTaskClient 设置任务客户端（用于从中心端获取任务）
func (s *Scheduler) SetTaskClient(taskClient *client.TaskClient) {
	s.taskClient = taskClient
}

// SetMetricsHandler 设置指标处理器
func (s *Scheduler) SetMetricsHandler(handler func([]*plugin.Metric, *plugin.CollectionTask)) {
	s.onMetrics = handler
}

// SetReportHandler 设置报告处理器
func (s *Scheduler) SetReportHandler(handler func(*ScheduledTask, time.Duration, int)) {
	s.onReport = handler
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// 启动任务获取循环（如果有任务客户端）
	if s.taskClient != nil {
		go s.fetchTasksLoop()
		logger.Info("Task fetch loop started",
			zap.Duration("interval", s.fetchInterval))
	}

	// 启动调度循环
	go s.scheduleLoop()

	logger.Info("Scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}

	// 等待所有任务完成（带超时）
	s.workerPool.Stop(30 * time.Second)

	logger.Info("Scheduler stopped")
}

// AddTask 添加任务
func (s *Scheduler) AddTask(task *plugin.CollectionTask, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := &ScheduledTask{
		Task:       task,
		NextRun:    time.Now(),
		Interval:   interval,
		LastStatus: TaskStatusPending,
	}

	s.tasks[task.TaskID] = st

	logger.Info("Added task",
		zap.String("task_id", task.TaskID),
		zap.String("device_id", task.DeviceID),
		zap.String("plugin", task.PluginName))
}

// RemoveTask 移除任务
func (s *Scheduler) RemoveTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, taskID)

	logger.Info("Removed task", zap.String("task_id", taskID))
}

// UpdateTasks 更新任务列表（使用统一间隔）
func (s *Scheduler) UpdateTasks(tasks []*plugin.CollectionTask, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建新任务映射
	newTasks := make(map[string]bool)
	for _, task := range tasks {
		newTasks[task.TaskID] = true

		// 如果任务已存在，更新配置
		if st, exists := s.tasks[task.TaskID]; exists {
			st.Task = task
			st.Interval = interval
		} else {
			// 新任务
			st := &ScheduledTask{
				Task:       task,
				NextRun:    time.Now(),
				Interval:   interval,
				LastStatus: TaskStatusPending,
			}
			s.tasks[task.TaskID] = st
		}
	}

	// 移除不再存在的任务
	for taskID := range s.tasks {
		if !newTasks[taskID] {
			delete(s.tasks, taskID)
		}
	}

	logger.Info("Updated tasks", zap.Int("count", len(s.tasks)))
}

// UpdateTasksWithIntervals 更新任务列表（每个任务使用不同的间隔）
func (s *Scheduler) UpdateTasksWithIntervals(tasksWithIntervals []client.TaskWithInterval) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建新任务映射
	newTasks := make(map[string]bool)
	for _, twi := range tasksWithIntervals {
		task := twi.Task
		interval := twi.Interval
		newTasks[task.TaskID] = true

		// 如果任务已存在，更新配置
		if st, exists := s.tasks[task.TaskID]; exists {
			st.Task = task
			st.Interval = interval
			logger.Debug("Updated existing task",
				zap.String("task_id", task.TaskID),
				zap.Duration("interval", interval))
		} else {
			// 新任务
			// 如果后端返回了 NextExecutionAt，使用它；否则立即执行
			nextRun := time.Now()
			st := &ScheduledTask{
				Task:       task,
				NextRun:    nextRun,
				Interval:   interval,
				LastStatus: TaskStatusPending,
			}
			s.tasks[task.TaskID] = st
			logger.Info("Added new task",
				zap.String("task_id", task.TaskID),
				zap.String("device_id", task.DeviceID),
				zap.String("plugin", task.PluginName),
				zap.Duration("interval", interval))
		}
	}

	// 移除不再存在的任务
	removedCount := 0
	for taskID := range s.tasks {
		if !newTasks[taskID] {
			delete(s.tasks, taskID)
			removedCount++
			logger.Info("Removed task", zap.String("task_id", taskID))
		}
	}

	logger.Info("Updated tasks from core",
		zap.Int("total", len(tasksWithIntervals)),
		zap.Int("active", len(s.tasks)),
		zap.Int("removed", removedCount))
}

// scheduleLoop 调度循环
func (s *Scheduler) scheduleLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndExecuteTasks()
		case <-s.ctx.Done():
			return
		}
	}
}

// checkAndExecuteTasks 检查并执行任务
func (s *Scheduler) checkAndExecuteTasks() {
	now := time.Now()

	s.mu.RLock()
	tasksToExecute := make([]*ScheduledTask, 0)

	for _, st := range s.tasks {
		st.mu.Lock()
		if st.NextRun.Before(now) && st.LastStatus != TaskStatusRunning {
			tasksToExecute = append(tasksToExecute, st)
		}
		st.mu.Unlock()
	}
	s.mu.RUnlock()

	// 提交任务到工作池
	for _, st := range tasksToExecute {
		s.executeTask(st)
	}
}

// executeTask 执行任务
func (s *Scheduler) executeTask(st *ScheduledTask) {
	s.workerPool.Submit(func() {
		s.runTask(st)
	})
}

// runTask 运行任务
func (s *Scheduler) runTask(st *ScheduledTask) {
	st.mu.Lock()
	st.LastStatus = TaskStatusRunning
	st.mu.Unlock()

	startTime := time.Now()

	// 获取插件
	p, ok := s.pluginMgr.GetPlugin(st.Task.PluginName)
	if !ok {
		st.mu.Lock()
		st.LastStatus = TaskStatusFailed
		st.LastError = fmt.Errorf("plugin not found: %s", st.Task.PluginName)
		st.NextRun = time.Now().Add(st.Interval)
		st.mu.Unlock()

		logger.Error("Plugin not found",
			zap.String("task_id", st.Task.TaskID),
			zap.String("plugin", st.Task.PluginName))
		return
	}

	// 执行采集
	taskCtx, cancel := context.WithTimeout(s.ctx, st.Task.Timeout)
	defer cancel()

	metrics, err := p.Collect(taskCtx, st.Task)

	// 生成设备状态指标（用于时序库和 PostgreSQL）
	statusMetric := s.createDeviceStatusMetric(st.Task, err)
	if metrics == nil {
		metrics = make([]*plugin.Metric, 0)
	}

	// 添加调试日志
	logger.Info("Generated device_status metric",
		zap.String("device_id", st.Task.DeviceID),
		zap.Float64("value", statusMetric.Value),
		zap.Int("metrics_before", len(metrics)))

	metrics = append(metrics, statusMetric)

	logger.Info("After appending device_status",
		zap.Int("metrics_after", len(metrics)))

	st.mu.Lock()
	if err != nil {
		st.LastStatus = TaskStatusFailed
		st.LastError = err
		logger.Error("Task failed",
			zap.String("task_id", st.Task.TaskID),
			zap.String("device_id", st.Task.DeviceID),
			zap.Error(err))
	} else {
		st.LastStatus = TaskStatusSuccess
		st.LastError = nil

		logger.Info("Task succeeded",
			zap.String("task_id", st.Task.TaskID),
			zap.String("device_id", st.Task.DeviceID),
			zap.Int("metrics", len(metrics)))
	}

	// 调用指标处理器（包含状态指标）
	if s.onMetrics != nil {
		s.onMetrics(metrics, st.Task)
	}

	// 更新下次执行时间
	st.NextRun = time.Now().Add(st.Interval)
	st.ExecutionCnt++
	st.mu.Unlock()

	// 上报执行结果
	if s.onReport != nil {
		s.onReport(st, time.Since(startTime), len(metrics))
	}
}

// createDeviceStatusMetric 创建设备状态指标
func (s *Scheduler) createDeviceStatusMetric(task *plugin.CollectionTask, collectErr error) *plugin.Metric {
	// 状态值：1=online, 0=offline
	statusValue := 1.0
	if collectErr != nil {
		statusValue = 0.0
	}

	// 获取设备类型（从 DeviceConfig 中提取，如果没有则为 unknown）
	deviceType := "unknown"
	if dt, ok := task.DeviceConfig["device_type"].(string); ok {
		deviceType = dt
	}

	return &plugin.Metric{
		Name:      "device_status",
		Value:     statusValue,
		Timestamp: time.Now().Unix(), // 使用秒级时间戳，与其他指标保持一致
		Labels: map[string]string{
			"device_id":   task.DeviceID,
			"device_type": deviceType,
			"task_id":     task.TaskID,
			"plugin":      task.PluginName,
		},
		Type: plugin.MetricTypeGauge,
	}
}

// GetTaskStatus 获取任务状态
func (s *Scheduler) GetTaskStatus(taskID string) (*ScheduledTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, ok := s.tasks[taskID]
	return st, ok
}

// GetAllTasks 获取所有任务
func (s *Scheduler) GetAllTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, st := range s.tasks {
		tasks = append(tasks, st)
	}

	return tasks
}

// fetchTasksLoop 任务获取循环
func (s *Scheduler) fetchTasksLoop() {
	// 立即获取一次任务
	s.fetchTasks()

	// 设置定时器
	ticker := time.NewTicker(s.fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.fetchTasks()
		case <-s.ctx.Done():
			return
		}
	}
}

// fetchTasks 从中心端获取任务
func (s *Scheduler) fetchTasks() {
	if s.taskClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	tasksWithIntervals, err := s.taskClient.GetTasks(ctx)
	if err != nil {
		logger.Warn("Failed to fetch tasks from core",
			zap.Error(err))
		return
	}

	if len(tasksWithIntervals) == 0 {
		logger.Debug("No tasks fetched from core")
		return
	}

	// 更新任务列表（每个任务使用自己的间隔）
	s.UpdateTasksWithIntervals(tasksWithIntervals)
}
