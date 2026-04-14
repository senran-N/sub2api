package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/senran-N/sub2api/internal/service"
)

// RPM 计数器缓存常量定义
//
// 设计说明：
// 使用 Redis Hash 跟踪每个账号最近 1-2 分钟的请求数：
// - Key: rpm:account:{accountID}
// - Field: minuteTimestamp
// - Value: 当前分钟内的请求计数
// - TTL: 120 秒（覆盖当前分钟窗口 + 一定冗余）
//
// Reserve/Increment 使用 Lua + Redis TIME 在服务端完成“取当前分钟 + 递增/限流判断”，
// 消除客户端 TIME/INCR 之间的分钟边界竞态，同时只触达单个 key。
const (
	// RPM 计数器键前缀
	// 格式: rpm:account:{accountID}
	rpmKeyPrefix = "rpm:account:"

	// RPM 计数器 TTL（120 秒，覆盖当前分钟窗口 + 冗余）
	rpmKeyTTL = 120 * time.Second
)

var (
	reserveRPMScript = redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local ttl = tonumber(ARGV[2])

		local timeResult = redis.call('TIME')
		local minuteField = tostring(math.floor(tonumber(timeResult[1]) / 60))
		local current = tonumber(redis.call('HGET', key, minuteField) or '0')

		if current >= limit then
			if current > 0 then
				redis.call('EXPIRE', key, ttl)
			end
			return {0, current}
		end

		current = tonumber(redis.call('HINCRBY', key, minuteField, 1))
		redis.call('EXPIRE', key, ttl)
		return {1, current}
	`)

	incrementRPMScript = redis.NewScript(`
		local key = KEYS[1]
		local ttl = tonumber(ARGV[1])

		local timeResult = redis.call('TIME')
		local minuteField = tostring(math.floor(tonumber(timeResult[1]) / 60))
		local current = tonumber(redis.call('HINCRBY', key, minuteField, 1))
		redis.call('EXPIRE', key, ttl)
		return current
	`)
)

// RPMCacheImpl RPM 计数器缓存 Redis 实现
type RPMCacheImpl struct {
	rdb *redis.Client
}

// NewRPMCache 创建 RPM 计数器缓存
func NewRPMCache(rdb *redis.Client) service.RPMCache {
	ctx := context.Background()
	for _, script := range []*redis.Script{reserveRPMScript, incrementRPMScript} {
		if err := script.Load(ctx, rdb).Err(); err != nil {
			log.Printf("[RPMCache] Failed to preload Lua script: %v", err)
		}
	}
	return &RPMCacheImpl{rdb: rdb}
}

func rpmAccountKey(accountID int64) string {
	return fmt.Sprintf("%s%d", rpmKeyPrefix, accountID)
}

// currentMinuteField 获取当前分钟 field（供查询路径使用）
// 使用 rdb.Time() 获取 Redis 服务端时间
func (c *RPMCacheImpl) currentMinuteField(ctx context.Context) (string, error) {
	serverTime, err := c.rdb.Time(ctx).Result()
	if err != nil {
		return "", fmt.Errorf("redis TIME: %w", err)
	}
	minuteTS := serverTime.Unix() / 60
	return strconv.FormatInt(minuteTS, 10), nil
}

// ReserveRPM 原子预留一个 RPM 配额位。
func (c *RPMCacheImpl) ReserveRPM(ctx context.Context, accountID int64, limit int) (bool, int, error) {
	if limit <= 0 {
		return true, 0, nil
	}

	raw, err := reserveRPMScript.Run(ctx, c.rdb, []string{rpmAccountKey(accountID)}, limit, int(rpmKeyTTL.Seconds())).Result()
	if err != nil {
		return false, 0, fmt.Errorf("rpm reserve: %w", err)
	}

	values, ok := raw.([]any)
	if !ok || len(values) != 2 {
		return false, 0, fmt.Errorf("rpm reserve: unexpected script result %T", raw)
	}

	allowed, err := redisInt(values[0])
	if err != nil {
		return false, 0, fmt.Errorf("rpm reserve: parse allowed: %w", err)
	}
	count, err := redisInt(values[1])
	if err != nil {
		return false, 0, fmt.Errorf("rpm reserve: parse count: %w", err)
	}

	return allowed == 1, count, nil
}

// IncrementRPM 原子递增并返回当前分钟的计数。
func (c *RPMCacheImpl) IncrementRPM(ctx context.Context, accountID int64) (int, error) {
	count, err := incrementRPMScript.Run(ctx, c.rdb, []string{rpmAccountKey(accountID)}, int(rpmKeyTTL.Seconds())).Int()
	if err != nil {
		return 0, fmt.Errorf("rpm increment: %w", err)
	}
	return count, nil
}

// GetRPM 获取当前分钟的 RPM 计数
func (c *RPMCacheImpl) GetRPM(ctx context.Context, accountID int64) (int, error) {
	field, err := c.currentMinuteField(ctx)
	if err != nil {
		return 0, fmt.Errorf("rpm get: %w", err)
	}

	val, err := c.rdb.HGet(ctx, rpmAccountKey(accountID), field).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil // 当前分钟无记录
	}
	if err != nil {
		return 0, fmt.Errorf("rpm get: %w", err)
	}
	return val, nil
}

// GetRPMBatch 批量获取多个账号的 RPM 计数（使用 Pipeline）
func (c *RPMCacheImpl) GetRPMBatch(ctx context.Context, accountIDs []int64) (map[int64]int, error) {
	if len(accountIDs) == 0 {
		return map[int64]int{}, nil
	}

	minuteField, err := c.currentMinuteField(ctx)
	if err != nil {
		return nil, fmt.Errorf("rpm batch get: %w", err)
	}

	pipe := c.rdb.Pipeline()
	cmds := make(map[int64]*redis.StringCmd, len(accountIDs))
	for _, id := range accountIDs {
		cmds[id] = pipe.HGet(ctx, rpmAccountKey(id), minuteField)
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("rpm batch get: %w", err)
	}

	result := make(map[int64]int, len(accountIDs))
	for id, cmd := range cmds {
		if val, err := cmd.Int(); err == nil {
			result[id] = val
		} else {
			result[id] = 0
		}
	}
	return result, nil
}

func redisInt(raw any) (int, error) {
	switch v := raw.(type) {
	case int64:
		return int(v), nil
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unexpected value type %T", raw)
	}
}
