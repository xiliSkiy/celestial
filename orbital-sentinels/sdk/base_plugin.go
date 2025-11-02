package sdk

import (
	"fmt"
	"regexp"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"go.uber.org/zap"
)

// BasePlugin 提供插件基础实现
type BasePlugin struct {
	config map[string]interface{}
}

// ValidateConfig 验证配置（通用实现）
func (bp *BasePlugin) ValidateConfig(deviceConfig map[string]interface{}, schema plugin.PluginSchema) error {
	for _, field := range schema.DeviceFields {
		value, exists := deviceConfig[field.Name]

		// 检查必填字段
		if field.Required && !exists {
			return fmt.Errorf("required field '%s' is missing", field.Name)
		}

		// 如果字段不存在且有默认值，跳过验证
		if !exists {
			continue
		}

		// 类型验证
		if err := bp.validateFieldType(field, value); err != nil {
			return fmt.Errorf("field '%s': %w", field.Name, err)
		}

		// 正则验证
		if field.Validation != "" {
			if strValue, ok := value.(string); ok {
				matched, err := regexp.MatchString(field.Validation, strValue)
				if err != nil {
					return fmt.Errorf("field '%s': invalid validation pattern: %w", field.Name, err)
				}
				if !matched {
					return fmt.Errorf("field '%s': value does not match pattern %s", field.Name, field.Validation)
				}
			}
		}

		// 范围验证
		if field.Type == "int" {
			if intValue, ok := value.(int); ok {
				if field.Min != 0 && intValue < field.Min {
					return fmt.Errorf("field '%s': value %d is less than minimum %d", field.Name, intValue, field.Min)
				}
				if field.Max != 0 && intValue > field.Max {
					return fmt.Errorf("field '%s': value %d is greater than maximum %d", field.Name, intValue, field.Max)
				}
			}
		}
	}

	return nil
}

// validateFieldType 验证字段类型
func (bp *BasePlugin) validateFieldType(field plugin.DeviceField, value interface{}) error {
	switch field.Type {
	case "string", "password":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "int":
		switch value.(type) {
		case int, int64, int32, float64:
			// 允许这些类型
		default:
			return fmt.Errorf("expected int, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case "list":
		// 允许数组或切片
		// 这里简化处理，实际可以更严格
	default:
		// 未知类型，跳过验证
	}

	return nil
}

// Log 获取日志记录器
func (bp *BasePlugin) Log() *zap.Logger {
	return logger.GetLogger()
}

// GetConfig 获取配置值
func (bp *BasePlugin) GetConfig(key string, defaultValue interface{}) interface{} {
	if value, ok := bp.config[key]; ok {
		return value
	}
	return defaultValue
}

// GetConfigString 获取字符串配置
func (bp *BasePlugin) GetConfigString(key string, defaultValue string) string {
	if value, ok := bp.config[key].(string); ok {
		return value
	}
	return defaultValue
}

// GetConfigInt 获取整数配置
func (bp *BasePlugin) GetConfigInt(key string, defaultValue int) int {
	switch v := bp.config[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return defaultValue
	}
}

// GetConfigBool 获取布尔配置
func (bp *BasePlugin) GetConfigBool(key string, defaultValue bool) bool {
	if value, ok := bp.config[key].(bool); ok {
		return value
	}
	return defaultValue
}
