package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	
	"github.com/celestial/gravital-core/internal/pkg/config"
)

var globalRedis *redis.Client

// Init 初始化 Redis 连接
func Init(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	globalRedis = client
	return client, nil
}

// Get 获取全局 Redis 客户端
func Get() *redis.Client {
	return globalRedis
}

// Close 关闭 Redis 连接
func Close() error {
	if globalRedis != nil {
		return globalRedis.Close()
	}
	return nil
}

