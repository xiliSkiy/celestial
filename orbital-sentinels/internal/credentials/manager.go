package credentials

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Manager 凭证管理器
type Manager struct {
	credentialsPath string
	credentials     *Credentials
}

// NewManager 创建凭证管理器
func NewManager(credentialsPath string) *Manager {
	if credentialsPath == "" {
		credentialsPath = getDefaultCredentialsPath()
	}
	return &Manager{
		credentialsPath: credentialsPath,
	}
}

// Load 加载凭证
func (m *Manager) Load() (*Credentials, error) {
	data, err := os.ReadFile(m.credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在,返回 nil
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	m.credentials = &creds
	return &creds, nil
}

// Save 保存凭证
func (m *Manager) Save(creds *Credentials) error {
	// 确保目录存在
	dir := filepath.Dir(m.credentialsPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 序列化
	data, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// 写入文件 (0600权限,仅所有者可读写)
	if err := os.WriteFile(m.credentialsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	m.credentials = creds
	return nil
}

// Delete 删除凭证
func (m *Manager) Delete() error {
	if err := os.Remove(m.credentialsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	m.credentials = nil
	return nil
}

// GetCredentials 获取当前凭证
func (m *Manager) GetCredentials() *Credentials {
	return m.credentials
}

// GetPath 获取凭证文件路径
func (m *Manager) GetPath() string {
	return m.credentialsPath
}

// getDefaultCredentialsPath 获取默认凭证文件路径
func getDefaultCredentialsPath() string {
	var basePath string

	if runtime.GOOS == "windows" {
		// Windows: %USERPROFILE%\.sentinel\credentials.yaml
		basePath = os.Getenv("USERPROFILE")
	} else {
		// Linux/Mac: ~/.sentinel/credentials.yaml
		basePath, _ = os.UserHomeDir()
	}

	if basePath == "" {
		// 如果无法获取用户目录,使用当前目录
		basePath = "."
	}

	return filepath.Join(basePath, ".sentinel", "credentials.yaml")
}

