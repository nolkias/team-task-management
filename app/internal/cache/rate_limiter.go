package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{client: client, limit: limit, window: window}
}

func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	bucket := time.Now().Unix() / int64(r.window.Seconds())
	redisKey := fmt.Sprintf("rate_limit:%s:%d", key, bucket)

	count, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		r.client.Expire(ctx, redisKey, r.window)
	}

	return count <= int64(r.limit), nil
}
