package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

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

// UpdateTasks 更新任务列表
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

		// 调用指标处理器
		if s.onMetrics != nil {
			s.onMetrics(metrics, st.Task)
		}

		logger.Debug("Task succeeded",
			zap.String("task_id", st.Task.TaskID),
			zap.String("device_id", st.Task.DeviceID),
			zap.Int("metrics", len(metrics)))
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
