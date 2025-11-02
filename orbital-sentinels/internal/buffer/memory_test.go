package buffer

import (
	"testing"

	"github.com/celestial/orbital-sentinels/internal/plugin"
)

func TestMemoryBuffer_PushPop(t *testing.T) {
	buf := NewMemoryBuffer(10)

	// 测试推入数据
	metrics := []*plugin.Metric{
		{Name: "test1", Value: 1.0},
		{Name: "test2", Value: 2.0},
	}

	err := buf.Push(metrics)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// 测试大小
	if buf.Size() != 2 {
		t.Errorf("Expected size 2, got %d", buf.Size())
	}

	// 测试弹出数据
	popped, err := buf.Pop(1)
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}

	if len(popped) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(popped))
	}

	if popped[0].Name != "test1" {
		t.Errorf("Expected test1, got %s", popped[0].Name)
	}

	// 验证剩余大小
	if buf.Size() != 1 {
		t.Errorf("Expected size 1, got %d", buf.Size())
	}
}

func TestMemoryBuffer_Overflow(t *testing.T) {
	buf := NewMemoryBuffer(5)

	// 先推入 3 个数据
	metrics1 := []*plugin.Metric{
		{Name: "test", Value: 0.0},
		{Name: "test", Value: 1.0},
		{Name: "test", Value: 2.0},
	}
	err := buf.Push(metrics1)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// 再推入 5 个数据，总共 8 个，超过容量 5
	metrics2 := []*plugin.Metric{
		{Name: "test", Value: 3.0},
		{Name: "test", Value: 4.0},
		{Name: "test", Value: 5.0},
		{Name: "test", Value: 6.0},
		{Name: "test", Value: 7.0},
	}
	err = buf.Push(metrics2)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// 应该只保留最新的 5 个（索引 3-7）
	if buf.Size() != 5 {
		t.Errorf("Expected size 5, got %d", buf.Size())
	}

	// 弹出所有数据，验证是最新的 5 个
	popped, err := buf.Pop(5)
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}

	if len(popped) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(popped))
	}

	// 验证第一个是 3.0
	if popped[0].Value != 3.0 {
		t.Errorf("Expected first value 3.0, got %f", popped[0].Value)
	}

	// 验证最后一个是 7.0
	if popped[4].Value != 7.0 {
		t.Errorf("Expected last value 7.0, got %f", popped[4].Value)
	}
}

func TestMemoryBuffer_Close(t *testing.T) {
	buf := NewMemoryBuffer(10)

	// 关闭缓冲区
	err := buf.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 尝试推入数据应该失败
	metrics := []*plugin.Metric{{Name: "test", Value: 1.0}}
	err = buf.Push(metrics)
	if err == nil {
		t.Error("Expected error when pushing to closed buffer")
	}
}
