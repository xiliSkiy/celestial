package plugin

import (
	"context"
	"testing"
	"time"
)

// MockPlugin 模拟插件
type MockPlugin struct {
	meta PluginMeta
}

func (m *MockPlugin) Meta() PluginMeta {
	return m.meta
}

func (m *MockPlugin) Schema() PluginSchema {
	return PluginSchema{Meta: m.meta}
}

func (m *MockPlugin) Init(config map[string]interface{}) error {
	return nil
}

func (m *MockPlugin) ValidateConfig(deviceConfig map[string]interface{}) error {
	return nil
}

func (m *MockPlugin) TestConnection(deviceConfig map[string]interface{}) error {
	return nil
}

func (m *MockPlugin) Collect(ctx context.Context, task *CollectionTask) ([]*Metric, error) {
	return []*Metric{
		{
			Name:      "test_metric",
			Value:     42.0,
			Timestamp: time.Now().Unix(),
			Labels:    map[string]string{"test": "true"},
			Type:      MetricTypeGauge,
		},
	}, nil
}

func (m *MockPlugin) Close() error {
	return nil
}

func TestManager_RegisterPlugin(t *testing.T) {
	mgr := NewManager("./plugins")

	plugin := &MockPlugin{
		meta: PluginMeta{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
		},
	}

	err := mgr.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// 验证插件已注册
	p, ok := mgr.GetPlugin("test-plugin")
	if !ok {
		t.Error("Plugin not found after registration")
	}

	if p.Meta().Name != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", p.Meta().Name)
	}
}

func TestManager_GetPlugin(t *testing.T) {
	mgr := NewManager("./plugins")

	plugin := &MockPlugin{
		meta: PluginMeta{Name: "test-plugin"},
	}

	mgr.RegisterPlugin(plugin)

	// 测试获取存在的插件
	p, ok := mgr.GetPlugin("test-plugin")
	if !ok {
		t.Error("Expected to find plugin")
	}
	if p == nil {
		t.Error("Expected non-nil plugin")
	}

	// 测试获取不存在的插件
	_, ok = mgr.GetPlugin("non-existent")
	if ok {
		t.Error("Expected not to find non-existent plugin")
	}
}

func TestManager_ListPlugins(t *testing.T) {
	mgr := NewManager("./plugins")

	// 注册多个插件
	plugins := []Plugin{
		&MockPlugin{meta: PluginMeta{Name: "plugin1"}},
		&MockPlugin{meta: PluginMeta{Name: "plugin2"}},
		&MockPlugin{meta: PluginMeta{Name: "plugin3"}},
	}

	for _, p := range plugins {
		mgr.RegisterPlugin(p)
	}

	// 列出所有插件
	list := mgr.ListPlugins()
	if len(list) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(list))
	}
}

func TestManager_StopAll(t *testing.T) {
	mgr := NewManager("./plugins")

	plugin := &MockPlugin{
		meta: PluginMeta{Name: "test-plugin"},
	}

	mgr.RegisterPlugin(plugin)

	// 停止所有插件（不应该panic）
	mgr.StopAll()
}
