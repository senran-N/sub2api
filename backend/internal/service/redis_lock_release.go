package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func runRedisLeaderLockRelease(script *redis.Script, client *redis.Client, key string, holder string, timeout time.Duration) {
	if script == nil || client == nil || key == "" || holder == "" {
		return
	}

	releaseCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, _ = script.Run(releaseCtx, client, []string{key}, holder).Result()
}
