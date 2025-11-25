package timeseries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Client 时序数据库客户端
type Client struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// QueryRangeResponse 范围查询响应
type QueryRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"` // [[timestamp, value], ...]
		} `json:"result"`
	} `json:"data"`
	ErrorType string `json:"errorType,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	Metric     map[string]string
	Timestamps []string
	Values     []float64
}

// DeviceMetrics 设备监控指标
type DeviceMetrics struct {
	DeviceID string                     `json:"device_id"`
	Metrics  map[string]*TimeSeriesData `json:"metrics"`
}

// NewClient 创建时序数据库客户端
func NewClient(baseURL string, logger *zap.Logger) *Client {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// QueryRange 执行范围查询
func (c *Client) QueryRange(promQL string, start, end time.Time, step time.Duration) (*QueryRangeResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("time series database URL is not configured")
	}

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%s",
		c.baseURL,
		url.QueryEscape(promQL),
		start.Unix(),
		end.Unix(),
		step.String())

	c.logger.Debug("Querying time series database",
		zap.String("url", queryURL),
		zap.String("query", promQL),
		zap.Time("start", start),
		zap.Time("end", end))

	// 发送 HTTP 请求
	resp, err := c.client.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query time series database: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("time series database returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	var response QueryRangeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查查询状态
	if response.Status != "success" {
		return nil, fmt.Errorf("query failed: %s - %s", response.ErrorType, response.Error)
	}

	c.logger.Debug("Query result",
		zap.String("query", promQL),
		zap.Int("result_count", len(response.Data.Result)))

	return &response, nil
}

// GetDeviceMetrics 获取设备监控指标
func (c *Client) GetDeviceMetrics(deviceID string, hours int) (*DeviceMetrics, error) {
	end := time.Now()
	start := end.Add(-time.Duration(hours) * time.Hour)

	// 根据时间范围计算合适的步长
	step := c.calculateStep(hours)

	metrics := &DeviceMetrics{
		DeviceID: deviceID,
		Metrics:  make(map[string]*TimeSeriesData),
	}

	// 先查询该设备所有可用的指标
	// 使用通配符查询所有带 device_id 标签的指标
	allMetricsQuery := fmt.Sprintf(`{device_id="%s"}`, deviceID)

	c.logger.Debug("Querying all metrics for device",
		zap.String("device_id", deviceID),
		zap.String("query", allMetricsQuery))

	// 查询所有指标
	response, err := c.QueryRange(allMetricsQuery, start, end, step)
	if err != nil {
		c.logger.Warn("Failed to query metrics for device",
			zap.String("device_id", deviceID),
			zap.Error(err))
		return metrics, nil // 返回空指标而不是错误
	}

	// 解析所有时间序列
	for _, result := range response.Data.Result {
		// 从 metric 标签中获取指标名称
		metricName := result.Metric["__name__"]
		if metricName == "" {
			continue
		}

		// 解析时间序列数据
		data := &TimeSeriesData{
			Metric:     result.Metric,
			Timestamps: make([]string, 0, len(result.Values)),
			Values:     make([]float64, 0, len(result.Values)),
		}

		for _, value := range result.Values {
			if len(value) < 2 {
				continue
			}

			// 解析时间戳
			timestamp, ok := value[0].(float64)
			if !ok {
				continue
			}
			timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
			data.Timestamps = append(data.Timestamps, timeStr)

			// 解析值
			valueStr, ok := value[1].(string)
			if !ok {
				continue
			}

			floatValue, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				c.logger.Warn("Failed to parse metric value",
					zap.String("metric", metricName),
					zap.String("value", valueStr),
					zap.Error(err))
				continue
			}

			data.Values = append(data.Values, floatValue)
		}

		// 只添加有数据的指标
		if len(data.Values) > 0 {
			// 使用指标名称作为 key，去掉常见的前缀以简化
			simpleName := c.simplifyMetricName(metricName)
			metrics.Metrics[simpleName] = data

			c.logger.Debug("Found metric for device",
				zap.String("device_id", deviceID),
				zap.String("metric", metricName),
				zap.String("simple_name", simpleName),
				zap.Int("data_points", len(data.Values)))
		}
	}

	c.logger.Info("Retrieved metrics for device",
		zap.String("device_id", deviceID),
		zap.Int("metric_count", len(metrics.Metrics)))

	return metrics, nil
}

// simplifyMetricName 简化指标名称
func (c *Client) simplifyMetricName(metricName string) string {
	// 常见的指标名称映射
	nameMap := map[string]string{
		"cpu_usage":         "cpu",
		"cpu_percent":       "cpu",
		"memory_usage":      "memory",
		"memory_percent":    "memory",
		"disk_usage":        "disk",
		"disk_percent":      "disk",
		"network_in_bytes":  "network_in",
		"network_out_bytes": "network_out",
		"net_bytes_recv":    "network_in",
		"net_bytes_sent":    "network_out",
	}

	// 如果有映射，使用映射后的名称
	if simpleName, exists := nameMap[metricName]; exists {
		return simpleName
	}

	// 否则直接使用原名称
	return metricName
}

// queryMetric 查询单个指标
func (c *Client) queryMetric(promQL string, start, end time.Time, step time.Duration) (*TimeSeriesData, error) {
	response, err := c.QueryRange(promQL, start, end, step)
	if err != nil {
		return nil, err
	}

	// 如果没有数据，返回 nil
	if len(response.Data.Result) == 0 {
		return nil, nil
	}

	// 取第一个结果（通常一个设备一个指标只有一个时间序列）
	result := response.Data.Result[0]

	data := &TimeSeriesData{
		Metric:     result.Metric,
		Timestamps: make([]string, 0, len(result.Values)),
		Values:     make([]float64, 0, len(result.Values)),
	}

	// 解析时间序列数据
	for _, value := range result.Values {
		if len(value) < 2 {
			continue
		}

		// 解析时间戳
		timestamp, ok := value[0].(float64)
		if !ok {
			continue
		}
		timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
		data.Timestamps = append(data.Timestamps, timeStr)

		// 解析值
		valueStr, ok := value[1].(string)
		if !ok {
			continue
		}

		floatValue, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			c.logger.Warn("Failed to parse metric value",
				zap.String("value", valueStr),
				zap.Error(err))
			continue
		}

		data.Values = append(data.Values, floatValue)
	}

	return data, nil
}

// calculateStep 根据时间范围计算合适的步长
func (c *Client) calculateStep(hours int) time.Duration {
	switch {
	case hours <= 1:
		return 30 * time.Second
	case hours <= 6:
		return 1 * time.Minute
	case hours <= 24:
		return 5 * time.Minute
	case hours <= 168: // 7 days
		return 30 * time.Minute
	default:
		return 1 * time.Hour
	}
}

// Health 检查时序数据库健康状态
func (c *Client) Health() error {
	if c.baseURL == "" {
		return fmt.Errorf("time series database URL is not configured")
	}

	healthURL := fmt.Sprintf("%s/health", c.baseURL)

	resp, err := c.client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
