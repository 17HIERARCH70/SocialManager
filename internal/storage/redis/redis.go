package redis

import (
	"fmt"
	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/go-redis/redis/v8"
)

// NewClient function to initialize a new Redis client.
func NewClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := client.Ping(client.Context()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", cfg.Address, err)
	}

	return client, nil
}
