package cache

import (
	"fmt"

	"github.com/redis/go-redis/v9"

	"teamtask/internal/config"
)

func NewClient(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
	})
}
