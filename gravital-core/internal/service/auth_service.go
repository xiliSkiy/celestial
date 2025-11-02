package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/pkg/auth"
	"github.com/celestial/gravital-core/internal/repository"
)

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, username, password string) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, userID uint) error
	GetUserInfo(ctx context.Context, userID uint) (*model.User, error)
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         *model.User `json:"user"`
}

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
	bcryptCost int
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository, jwtManager *auth.JWTManager, bcryptCost int) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		bcryptCost: bcryptCost,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// 查询用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户是否启用
	if !user.Enabled {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 验证密码
	if !auth.CheckPassword(password, user.PasswordHash) {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 获取用户权限
	var permissions []string
	if user.Role != nil {
		permissions = user.Role.Permissions
	}

	// 生成 Token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Role.Name, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 生成刷新 Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		// 记录错误但不影响登录
		fmt.Printf("failed to update last login: %v\n", err)
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(24 * 3600), // 24小时
		User:         user,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// 验证刷新 Token
	userID, err := s.jwtManager.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// 查询用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 检查用户是否启用
	if !user.Enabled {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 获取用户权限
	var permissions []string
	if user.Role != nil {
		permissions = user.Role.Permissions
	}

	// 生成新的 Token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Role.Name, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 生成新的刷新 Token
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(24 * 3600),
		User:         user,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID uint) error {
	// TODO: 实现 Token 黑名单或 Redis 缓存失效
	return nil
}

func (s *authService) GetUserInfo(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户是否启用
	if !user.Enabled {
		return nil, fmt.Errorf("用户已被禁用")
	}

	return user, nil
}

