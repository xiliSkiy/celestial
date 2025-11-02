package auth

import (
	"strings"
)

// Permission 权限检查器
type Permission struct {
	permissions []string
}

// NewPermission 创建权限检查器
func NewPermission(permissions []string) *Permission {
	return &Permission{
		permissions: permissions,
	}
}

// Has 检查是否有权限
func (p *Permission) Has(required string) bool {
	// 检查是否有通配符权限
	for _, perm := range p.permissions {
		if perm == "*" {
			return true
		}
		
		// 检查前缀匹配，如 "devices.*" 匹配 "devices.read"
		if strings.HasSuffix(perm, ".*") {
			prefix := strings.TrimSuffix(perm, ".*")
			if strings.HasPrefix(required, prefix+".") {
				return true
			}
		}
		
		// 检查后缀匹配，如 "*.read" 匹配 "devices.read"
		if strings.HasPrefix(perm, "*.") {
			suffix := strings.TrimPrefix(perm, "*.")
			if strings.HasSuffix(required, "."+suffix) {
				return true
			}
		}
		
		// 精确匹配
		if perm == required {
			return true
		}
	}
	
	return false
}

// HasAny 检查是否有任意一个权限
func (p *Permission) HasAny(required ...string) bool {
	for _, req := range required {
		if p.Has(req) {
			return true
		}
	}
	return false
}

// HasAll 检查是否有所有权限
func (p *Permission) HasAll(required ...string) bool {
	for _, req := range required {
		if !p.Has(req) {
			return false
		}
	}
	return true
}

