package scheduler

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool_Submit(t *testing.T) {
	wp := NewWorkerPool(3)
	defer wp.Stop(5 * time.Second)

	var counter atomic.Int32
	var wg sync.WaitGroup

	// 提交 10 个任务
	for i := 0; i < 10; i++ {
		wg.Add(1)
		wp.Submit(func() {
			defer wg.Done()
			counter.Add(1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 等待所有任务完成
	wg.Wait()

	// 验证所有任务都执行了
	if counter.Load() != 10 {
		t.Errorf("Expected 10 tasks executed, got %d", counter.Load())
	}
}

func TestWorkerPool_Stop(t *testing.T) {
	wp := NewWorkerPool(2)

	var counter atomic.Int32

	// 提交一些任务
	for i := 0; i < 5; i++ {
		wp.Submit(func() {
			counter.Add(1)
			time.Sleep(50 * time.Millisecond)
		})
	}

	// 等待一段时间后停止
	time.Sleep(100 * time.Millisecond)
	wp.Stop(2 * time.Second)

	// 验证至少有一些任务完成了
	if counter.Load() == 0 {
		t.Error("Expected some tasks to complete")
	}
}

func TestWorkerPool_Size(t *testing.T) {
	wp := NewWorkerPool(5)
	defer wp.Stop(1 * time.Second)

	if wp.Size() != 5 {
		t.Errorf("Expected size 5, got %d", wp.Size())
	}
}

func TestWorkerPool_Concurrent(t *testing.T) {
	workers := 3
	wp := NewWorkerPool(workers)
	defer wp.Stop(5 * time.Second)

	var activeCount atomic.Int32
	var maxActive atomic.Int32
	var wg sync.WaitGroup

	// 提交多个长时间运行的任务
	for i := 0; i < 10; i++ {
		wg.Add(1)
		wp.Submit(func() {
			defer wg.Done()

			// 增加活跃计数
			active := activeCount.Add(1)

			// 更新最大活跃数
			for {
				max := maxActive.Load()
				if active <= max || maxActive.CompareAndSwap(max, active) {
					break
				}
			}

			// 模拟工作
			time.Sleep(50 * time.Millisecond)

			// 减少活跃计数
			activeCount.Add(-1)
		})
	}

	wg.Wait()

	// 验证最大并发数不超过工作池大小
	if maxActive.Load() > int32(workers) {
		t.Errorf("Max concurrent workers %d exceeded pool size %d", maxActive.Load(), workers)
	}
}
