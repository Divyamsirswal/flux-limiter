package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

func (l *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {

	rate := float64(limit) / window.Seconds()

	capacity := limit

	now := time.Now().UnixMilli()
	nowSec := float64(now) / 1000.0

	cost := 1

	result, err := l.rdb.Eval(ctx, requestScript, []string{key}, capacity, rate, nowSec, cost).Result()
	if err != nil {
		return false, fmt.Errorf("redis execution failed: %w", err)
	}

	return result.(int64) == 1, nil
}
