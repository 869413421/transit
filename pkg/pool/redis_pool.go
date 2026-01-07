package pool

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisPool Redis 并发控制池
type RedisPool struct {
	client *redis.Client
}

// NewRedisPool 创建 Redis 池
func NewRedisPool(client *redis.Client) *RedisPool {
	return &RedisPool{client: client}
}

// Lua 脚本：原子性检查并增加并发位
const luaAcquirePermit = `
local current = tonumber(redis.call('GET', KEYS[1]) or "0")
local max = tonumber(ARGV[1])
if current < max then
    redis.call('INCR', KEYS[1])
    return 1
else
    return 0
end
`

// Acquire 获取并发位
func (p *RedisPool) Acquire(ctx context.Context, channelID string, maxConcurrency int) (bool, error) {
	key := fmt.Sprintf("transit:channel:%s:concurrency", channelID)

	res, err := p.client.Eval(ctx, luaAcquirePermit, []string{key}, maxConcurrency).Result()
	if err != nil {
		return false, err
	}

	acquired := res.(int64) == 1
	if !acquired {
		return false, errors.New("channel concurrency limit reached")
	}

	return true, nil
}

// Release 释放并发位
func (p *RedisPool) Release(ctx context.Context, channelID string) error {
	key := fmt.Sprintf("transit:channel:%s:concurrency", channelID)
	return p.client.Decr(ctx, key).Err()
}

// GetConcurrency 获取当前并发数
func (p *RedisPool) GetConcurrency(ctx context.Context, channelID string) (int, error) {
	key := fmt.Sprintf("transit:channel:%s:concurrency", channelID)
	val, err := p.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}
