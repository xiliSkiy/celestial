package ping

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/celestial/orbital-sentinels/internal/plugin"
	"github.com/celestial/orbital-sentinels/sdk"
)

// PingPlugin Ping 插件
type PingPlugin struct {
	sdk.BasePlugin
	schema plugin.PluginSchema
}

// NewPlugin 创建插件实例
func NewPlugin() plugin.Plugin {
	return &PingPlugin{}
}

// Meta 返回插件元信息
func (p *PingPlugin) Meta() plugin.PluginMeta {
	return p.schema.Meta
}

// Schema 返回配置 Schema
func (p *PingPlugin) Schema() plugin.PluginSchema {
	return p.schema
}

// Init 初始化插件
func (p *PingPlugin) Init(config map[string]interface{}) error {
	// 加载 schema
	// 在实际使用中，schema 应该从 plugin.yaml 加载
	p.schema = plugin.PluginSchema{
		Meta: plugin.PluginMeta{
			Name:        "ping",
			Version:     "1.0.0",
			Description: "ICMP Ping 连通性检测插件",
			Author:      "Celestial Team",
			DeviceTypes: []string{"server", "switch", "router", "any"},
		},
	}

	return nil
}

// ValidateConfig 验证设备配置
func (p *PingPlugin) ValidateConfig(deviceConfig map[string]interface{}) error {
	// 检查必填字段
	if _, ok := deviceConfig["host"]; !ok {
		return fmt.Errorf("host is required")
	}

	return nil
}

// TestConnection 测试连接
func (p *PingPlugin) TestConnection(deviceConfig map[string]interface{}) error {
	host := deviceConfig["host"].(string)

	// 执行一次 ping
	_, err := p.ping(host, 1, 5)
	return err
}

// Collect 采集数据
func (p *PingPlugin) Collect(ctx context.Context, task *plugin.CollectionTask) ([]*plugin.Metric, error) {
	host := task.DeviceConfig["host"].(string)
	count := p.getInt(task.DeviceConfig, "count", 4)
	timeout := p.getInt(task.DeviceConfig, "timeout", 5)

	// 执行 ping
	result, err := p.ping(host, count, timeout)
	if err != nil {
		// Ping 失败，返回不可达指标
		return []*plugin.Metric{
			{
				Name:      "ping_reachable",
				Value:     0,
				Timestamp: time.Now().Unix(),
				Labels: map[string]string{
					"device_id": task.DeviceID,
					"host":      host,
				},
				Type: plugin.MetricTypeGauge,
			},
		}, nil
	}

	// 构建指标
	metrics := []*plugin.Metric{
		{
			Name:      "ping_reachable",
			Value:     1,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"device_id": task.DeviceID,
				"host":      host,
			},
			Type: plugin.MetricTypeGauge,
		},
		{
			Name:      "ping_rtt_ms",
			Value:     result.AvgRTT,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"device_id": task.DeviceID,
				"host":      host,
			},
			Type: plugin.MetricTypeGauge,
		},
		{
			Name:      "ping_packet_loss",
			Value:     result.PacketLoss,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"device_id": task.DeviceID,
				"host":      host,
			},
			Type: plugin.MetricTypeGauge,
		},
	}

	return metrics, nil
}

// Close 关闭插件
func (p *PingPlugin) Close() error {
	return nil
}

// PingResult Ping 结果
type PingResult struct {
	AvgRTT     float64
	PacketLoss float64
}

// ping 执行 ping 命令
func (p *PingPlugin) ping(host string, count, timeout int) (*PingResult, error) {
	var cmd *exec.Cmd

	// 根据操作系统构建命令
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-W", strconv.Itoa(timeout), host)
	case "windows":
		cmd = exec.Command("ping", "-n", strconv.Itoa(count), "-w", strconv.Itoa(timeout*1000), host)
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	// 解析输出
	return p.parseOutput(string(output))
}

// parseOutput 解析 ping 输出
func (p *PingPlugin) parseOutput(output string) (*PingResult, error) {
	result := &PingResult{}

	// 解析平均 RTT (Linux/macOS)
	// 示例: rtt min/avg/max/mdev = 0.123/0.456/0.789/0.012 ms
	rttRegex := regexp.MustCompile(`rtt min/avg/max/[a-z]+ = [\d.]+/([\d.]+)/[\d.]+/[\d.]+ ms`)
	if matches := rttRegex.FindStringSubmatch(output); len(matches) > 1 {
		if avg, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.AvgRTT = avg
		}
	}

	// 解析丢包率
	// 示例: 4 packets transmitted, 4 received, 0% packet loss
	lossRegex := regexp.MustCompile(`(\d+)% packet loss`)
	if matches := lossRegex.FindStringSubmatch(output); len(matches) > 1 {
		if loss, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.PacketLoss = loss
		}
	}

	return result, nil
}

// getInt 获取整数配置
func (p *PingPlugin) getInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// 加载 schema 的辅助函数
func loadSchema(path string) (plugin.PluginSchema, error) {
	var schema plugin.PluginSchema
	// 这里应该读取 plugin.yaml 文件
	// 为了简化，暂时返回空
	return schema, nil
}

// 确保实现了接口
var _ plugin.Plugin = (*PingPlugin)(nil)
