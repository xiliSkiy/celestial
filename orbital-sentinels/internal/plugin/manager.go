package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Manager 插件管理器
type Manager struct {
	plugins   map[string]Plugin
	schemas   map[string]PluginSchema
	directory string
	mu        sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager(directory string) *Manager {
	return &Manager{
		plugins:   make(map[string]Plugin),
		schemas:   make(map[string]PluginSchema),
		directory: directory,
	}
}

// LoadAll 加载所有插件
func (m *Manager) LoadAll() error {
	entries, err := os.ReadDir(m.directory)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(m.directory, entry.Name())
		if err := m.loadPlugin(pluginPath); err != nil {
			logger.Error("Failed to load plugin",
				zap.String("plugin", entry.Name()),
				zap.Error(err))
			continue
		}
	}

	logger.Info("Loaded plugins", zap.Int("count", len(m.plugins)))
	return nil
}

// loadPlugin 加载单个插件
func (m *Manager) loadPlugin(pluginPath string) error {
	// 读取插件配置文件
	schemaPath := filepath.Join(pluginPath, "plugin.yaml")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin schema: %w", err)
	}

	var schema PluginSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse plugin schema: %w", err)
	}

	// 注意：这里需要根据实际的插件加载方式来实现
	// 由于 Go Plugin 在某些平台上有限制，这里先存储 schema
	// 实际的插件实例化需要在具体实现中完成

	m.mu.Lock()
	m.schemas[schema.Meta.Name] = schema
	m.mu.Unlock()

	logger.Info("Loaded plugin schema",
		zap.String("name", schema.Meta.Name),
		zap.String("version", schema.Meta.Version))

	return nil
}

// RegisterPlugin 注册插件实例
func (m *Manager) RegisterPlugin(plugin Plugin) error {
	meta := plugin.Meta()

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[meta.Name]; exists {
		return fmt.Errorf("plugin %s already registered", meta.Name)
	}

	m.plugins[meta.Name] = plugin
	logger.Info("Registered plugin",
		zap.String("name", meta.Name),
		zap.String("version", meta.Version))

	return nil
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, ok := m.plugins[name]
	return plugin, ok
}

// GetSchema 获取插件 Schema
func (m *Manager) GetSchema(name string) (PluginSchema, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schema, ok := m.schemas[name]
	return schema, ok
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []PluginMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metas := make([]PluginMeta, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		metas = append(metas, plugin.Meta())
	}

	return metas
}

// StopAll 停止所有插件
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, plugin := range m.plugins {
		if err := plugin.Close(); err != nil {
			logger.Error("Failed to close plugin",
				zap.String("plugin", name),
				zap.Error(err))
		}
	}

	logger.Info("Stopped all plugins")
}
