package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/celestial/orbital-sentinels/internal/credentials"
	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"go.uber.org/zap"
)

// Manager 注册管理器
type Manager struct {
	client          *http.Client
	coreURL         string
	registrationKey string
	credsMgr        *credentials.Manager
}

// NewManager 创建注册管理器
func NewManager(coreURL, registrationKey string, credsMgr *credentials.Manager) *Manager {
	return &Manager{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		coreURL:         coreURL,
		registrationKey: registrationKey,
		credsMgr:        credsMgr,
	}
}

// Register 注册到中心端
func (m *Manager) Register(ctx context.Context, config *Config) (*RegisterResponse, error) {
	// 1. 构建注册请求
	req, err := m.buildRegisterRequest(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build register request: %w", err)
	}

	logger.Info("Registering to core",
		zap.String("hostname", req.Hostname),
		zap.String("ip", req.IPAddress),
		zap.String("core_url", m.coreURL))

	// 2. 发送注册请求
	resp, err := m.sendRegisterRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send register request: %w", err)
	}

	// 3. 保存凭证
	creds := &credentials.Credentials{
		SentinelID:   resp.SentinelID,
		APIToken:     resp.APIToken,
		CoreURL:      m.coreURL,
		RegisteredAt: time.Now(),
		Region:       config.Region,
		Labels:       config.Labels,
	}

	if err := m.credsMgr.Save(creds); err != nil {
		return nil, fmt.Errorf("failed to save credentials: %w", err)
	}

	logger.Info("Registration successful",
		zap.String("sentinel_id", resp.SentinelID),
		zap.String("credentials_path", m.credsMgr.GetPath()))

	return resp, nil
}

// RegisterWithRetry 带重试的注册
func (m *Manager) RegisterWithRetry(ctx context.Context, config *Config) (*RegisterResponse, error) {
	retryIntervals := []time.Duration{
		0,               // 立即尝试
		5 * time.Second, // 第2次: 等待5秒
		10 * time.Second, // 第3次: 等待10秒
		30 * time.Second, // 第4次: 等待30秒
		60 * time.Second, // 第5次: 等待60秒
	}

	var lastErr error
	for i, interval := range retryIntervals {
		if i > 0 {
			logger.Info("Retrying registration",
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", len(retryIntervals)),
				zap.Duration("wait", interval))

			select {
			case <-time.After(interval):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := m.Register(ctx, config)
		if err == nil {
			if i > 0 {
				logger.Info("Registration successful after retry",
					zap.Int("attempt", i+1))
			}
			return resp, nil
		}

		lastErr = err
		logger.Warn("Registration attempt failed",
			zap.Int("attempt", i+1),
			zap.Error(err))
	}

	return nil, fmt.Errorf("registration failed after %d attempts: %w",
		len(retryIntervals), lastErr)
}

// buildRegisterRequest 构建注册请求
func (m *Manager) buildRegisterRequest(config *Config) (*RegisterRequest, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	ipAddress := getLocalIP()
	macAddress := getMACAddress()

	return &RegisterRequest{
		Name:            config.Name,
		Hostname:        hostname,
		IPAddress:       ipAddress,
		MACAddress:      macAddress,
		Version:         config.Version,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Region:          config.Region,
		Labels:          config.Labels,
		RegistrationKey: m.registrationKey,
	}, nil
}

// sendRegisterRequest 发送注册请求
func (m *Manager) sendRegisterRequest(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		m.coreURL+"/api/v1/sentinels/register",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if m.registrationKey != "" {
		httpReq.Header.Set("X-Registration-Key", m.registrationKey)
	}

	httpResp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed: status=%d, body=%s",
			httpResp.StatusCode, string(body))
	}

	var apiResp struct {
		Code int              `json:"code"`
		Data RegisterResponse `json:"data"`
		Msg  string           `json:"message"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("registration failed: code=%d, message=%s",
			apiResp.Code, apiResp.Msg)
	}

	return &apiResp.Data, nil
}

