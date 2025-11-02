package scheduler

import (
	"sync"
	"time"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"go.uber.org/zap"
)

// WorkerPool 工作池
type WorkerPool struct {
	workers   int
	taskQueue chan func()
	wg        sync.WaitGroup
	quit      chan struct{}
	once      sync.Once
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workers int) *WorkerPool {
	wp := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan func(), workers*2),
		quit:      make(chan struct{}),
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	logger.Info("Worker pool started", zap.Int("workers", workers))

	return wp
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskQueue:
			// 执行任务
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("Worker panic recovered",
							zap.Int("worker_id", id),
							zap.Any("panic", r))
					}
				}()
				task()
			}()

		case <-wp.quit:
			return
		}
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task func()) {
	select {
	case wp.taskQueue <- task:
		// 任务已提交
	case <-wp.quit:
		// 工作池已停止
		logger.Warn("Worker pool is stopped, task rejected")
	default:
		// 队列已满，阻塞等待
		select {
		case wp.taskQueue <- task:
		case <-wp.quit:
			logger.Warn("Worker pool is stopped, task rejected")
		}
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop(timeout time.Duration) {
	wp.once.Do(func() {
		close(wp.quit)

		// 等待所有任务完成（带超时）
		done := make(chan struct{})
		go func() {
			wp.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			logger.Info("Worker pool stopped gracefully")
		case <-time.After(timeout):
			logger.Warn("Worker pool stop timeout",
				zap.Duration("timeout", timeout))
		}
	})
}

// Size 获取工作池大小
func (wp *WorkerPool) Size() int {
	return wp.workers
}

// QueueLength 获取队列长度
func (wp *WorkerPool) QueueLength() int {
	return len(wp.taskQueue)
}
