package engine

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

// VMClient VictoriaMetrics 客户端
type VMClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// VMQueryResponse VictoriaMetrics 查询响应
type VMQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"` // [timestamp, value]
		} `json:"result"`
	} `json:"data"`
	ErrorType string `json:"errorType,omitempty"`
	Error     string `json:"error,omitempty"`
}

// MetricResult 指标查询结果
type MetricResult struct {
	Labels map[string]string
	Value  float64
}

// NewVMClient 创建 VictoriaMetrics 客户端
func NewVMClient(baseURL string, logger *zap.Logger) *VMClient {
	return &VMClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Query 执行 PromQL 查询
func (c *VMClient) Query(promQL string) ([]MetricResult, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("VictoriaMetrics URL is not configured")
	}

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		c.baseURL,
		url.QueryEscape(promQL),
		time.Now().Unix())

	c.logger.Debug("Querying VictoriaMetrics",
		zap.String("url", queryURL),
		zap.String("query", promQL))

	// 发送 HTTP 请求
	resp, err := c.client.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VictoriaMetrics returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	var vmResp VMQueryResponse
	if err := json.Unmarshal(body, &vmResp); err != nil {
		return nil, fmt.Errorf("failed to parse VictoriaMetrics response: %w", err)
	}

	// 检查查询状态
	if vmResp.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics query failed: %s - %s", vmResp.ErrorType, vmResp.Error)
	}

	// 转换为 MetricResult
	results := make([]MetricResult, 0, len(vmResp.Data.Result))
	for _, item := range vmResp.Data.Result {
		// 解析值
		if len(item.Value) < 2 {
			c.logger.Warn("Invalid metric value format", zap.Any("value", item.Value))
			continue
		}

		valueStr, ok := item.Value[1].(string)
		if !ok {
			c.logger.Warn("Metric value is not a string", zap.Any("value", item.Value[1]))
			continue
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			c.logger.Warn("Failed to parse metric value",
				zap.String("value", valueStr),
				zap.Error(err))
			continue
		}

		results = append(results, MetricResult{
			Labels: item.Metric,
			Value:  value,
		})
	}

	c.logger.Debug("VictoriaMetrics query result",
		zap.String("query", promQL),
		zap.Int("result_count", len(results)))

	return results, nil
}

// QueryRange 执行范围查询（用于未来扩展）
func (c *VMClient) QueryRange(promQL string, start, end time.Time, step time.Duration) ([]MetricResult, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("VictoriaMetrics URL is not configured")
	}

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%s",
		c.baseURL,
		url.QueryEscape(promQL),
		start.Unix(),
		end.Unix(),
		step.String())

	c.logger.Debug("Querying VictoriaMetrics range",
		zap.String("url", queryURL),
		zap.String("query", promQL))

	// 发送 HTTP 请求
	resp, err := c.client.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VictoriaMetrics returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	var vmResp VMQueryResponse
	if err := json.Unmarshal(body, &vmResp); err != nil {
		return nil, fmt.Errorf("failed to parse VictoriaMetrics response: %w", err)
	}

	// 检查查询状态
	if vmResp.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics query failed: %s - %s", vmResp.ErrorType, vmResp.Error)
	}

	// 对于范围查询，我们只返回最后一个值
	results := make([]MetricResult, 0, len(vmResp.Data.Result))
	for _, item := range vmResp.Data.Result {
		if len(item.Value) < 2 {
			continue
		}

		valueStr, ok := item.Value[1].(string)
		if !ok {
			continue
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		results = append(results, MetricResult{
			Labels: item.Metric,
			Value:  value,
		})
	}

	return results, nil
}

// Health 检查 VictoriaMetrics 健康状态
func (c *VMClient) Health() error {
	if c.baseURL == "" {
		return fmt.Errorf("VictoriaMetrics URL is not configured")
	}

	healthURL := fmt.Sprintf("%s/health", c.baseURL)
	resp, err := c.client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("failed to check VictoriaMetrics health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("VictoriaMetrics health check failed with status %d", resp.StatusCode)
	}

	return nil
}

