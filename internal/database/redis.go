package database

import (
	"context"
	"fmt"

	"github.com/1batu/market-ai/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient Redis bağlantısı oluşturur
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}
