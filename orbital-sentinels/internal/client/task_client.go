package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"go.uber.org/zap"
)

// TaskClient 任务客户端
type TaskClient struct {
	client     *http.Client
	coreURL    string
	apiToken   string
	sentinelID string
}

// NewTaskClient 创建任务客户端
func NewTaskClient(coreURL, apiToken, sentinelID string, timeout time.Duration) *TaskClient {
	return &TaskClient{
		client: &http.Client{
			Timeout: timeout,
		},
		coreURL:    coreURL,
		apiToken:   apiToken,
		sentinelID: sentinelID,
	}
}

// CoreTask 中心端返回的任务格式
type CoreTask struct {
	ID              uint                   `json:"id"`
	TaskID          string                 `json:"task_id"`
	DeviceID        string                 `json:"device_id"`
	SentinelID      string                 `json:"sentinel_id"`
	PluginName      string                 `json:"plugin_name"`
	Config          map[string]interface{} `json:"config"`
	IntervalSeconds int                    `json:"interval_seconds"`
	Enabled         bool                   `json:"enabled"`
	Priority        int                    `json:"priority"`
	RetryCount      int                    `json:"retry_count"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	NextExecutionAt *string                `json:"next_execution_at"`
}

// GetTasksResponse 获取任务响应
type GetTasksResponse struct {
	Tasks         []CoreTask `json:"tasks"`
	ConfigVersion int        `json:"config_version"`
}

// TaskWithInterval 任务和间隔
type TaskWithInterval struct {
	Task     *plugin.CollectionTask
	Interval time.Duration
}

// GetTasks 从中心端获取任务列表
func (c *TaskClient) GetTasks(ctx context.Context) ([]TaskWithInterval, error) {
	url := c.coreURL + "/api/v1/sentinel-tasks"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Sentinel-ID", c.sentinelID)
	req.Header.Set("X-API-Token", c.apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Code int `json:"code"`
		Data struct {
			Tasks         []CoreTask `json:"tasks"`
			ConfigVersion int        `json:"config_version"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, message=%s", apiResp.Code, apiResp.Message)
	}

	// 转换为 plugin.CollectionTask
	tasks := make([]TaskWithInterval, 0, len(apiResp.Data.Tasks))
	for _, ct := range apiResp.Data.Tasks {
		// 只处理启用的任务
		if !ct.Enabled {
			logger.Debug("Skipping disabled task",
				zap.String("task_id", ct.TaskID))
			continue
		}

		// 解析超时时间
		timeout := time.Duration(ct.TimeoutSeconds) * time.Second
		if timeout == 0 {
			timeout = 30 * time.Second // 默认 30 秒
		}

		// 解析间隔时间
		interval := time.Duration(ct.IntervalSeconds) * time.Second
		if interval == 0 {
			interval = 60 * time.Second // 默认 60 秒
		}

		// 解析配置
		deviceConfig := make(map[string]interface{})
		if ct.Config != nil {
			deviceConfig = ct.Config
		}

		task := &plugin.CollectionTask{
			TaskID:       ct.TaskID,
			DeviceID:     ct.DeviceID,
			PluginName:   ct.PluginName,
			DeviceConfig: deviceConfig,
			PluginConfig: make(map[string]interface{}),
			Timeout:      timeout,
		}

		tasks = append(tasks, TaskWithInterval{
			Task:     task,
			Interval: interval,
		})
	}

	logger.Info("Fetched tasks from core",
		zap.Int("total", len(apiResp.Data.Tasks)),
		zap.Int("enabled", len(tasks)),
		zap.Int("config_version", apiResp.Data.ConfigVersion))

	return tasks, nil
}
