package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/sovereign-fund/sovereign/config"
)

func NewRedis(cfg config.RedisConfig, log *slog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	log.Info("connected to Redis", slog.String("addr", cfg.Addr))

	return client, nil
}
