package register

// Config 注册配置
type Config struct {
	Name    string            // 显示名称
	Version string            // 版本号
	Region  string            // 区域
	Labels  map[string]string // 标签
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name            string            `json:"name"`
	Hostname        string            `json:"hostname"`
	IPAddress       string            `json:"ip_address"`
	MACAddress      string            `json:"mac_address"`
	Version         string            `json:"version"`
	OS              string            `json:"os"`
	Arch            string            `json:"arch"`
	Region          string            `json:"region,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	RegistrationKey string            `json:"registration_key,omitempty"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	SentinelID string                 `json:"sentinel_id"`
	APIToken   string                 `json:"api_token"`
	Config     map[string]interface{} `json:"config"`
	Message    string                 `json:"message,omitempty"`
}

