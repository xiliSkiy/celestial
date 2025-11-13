package credentials

import (
	"time"
)

// Credentials 采集端凭证
type Credentials struct {
	SentinelID   string            `yaml:"sentinel_id"`
	APIToken     string            `yaml:"api_token"`
	CoreURL      string            `yaml:"core_url"`
	RegisteredAt time.Time         `yaml:"registered_at"`
	Region       string            `yaml:"region,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty"`
}

// IsValid 检查凭证是否有效
func (c *Credentials) IsValid() bool {
	return c != nil && c.SentinelID != "" && c.APIToken != ""
}

