package billing

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// Service 计费服务
type Service struct {
	redis *redis.Client
}

// NewService 创建计费服务
func NewService(redis *redis.Client) *Service {
	return &Service{redis: redis}
}

// Lua 脚本：原子性检查余额并扣费
const luaDeductBalance = `
local balance = tonumber(redis.call('GET', KEYS[1]) or "0")
local cost = tonumber(ARGV[1])
if balance >= cost then
    redis.call('INCRBYFLOAT', KEYS[1], -cost)
    return 1
else
    return 0
end
`

// PreDeduct 预扣费（异步任务：视频/图片）
func (s *Service) PreDeduct(ctx context.Context, userID string, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	balanceKey := fmt.Sprintf("transit:user:%s:balance", userID)

	res, err := s.redis.Eval(ctx, luaDeductBalance, []string{balanceKey}, amount).Result()
	if err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}

	if res.(int64) == 0 {
		return errors.New("insufficient balance")
	}

	return nil
}

// PostDeduct 后扣费（同步任务：文本）
func (s *Service) PostDeduct(ctx context.Context, userID string, tokens int, pricePerToken float64) error {
	cost := float64(tokens) * pricePerToken
	if cost <= 0 {
		return nil
	}

	balanceKey := fmt.Sprintf("transit:user:%s:balance", userID)
	return s.redis.IncrByFloat(ctx, balanceKey, -cost).Err()
}

// Refund 退费（任务失败）
func (s *Service) Refund(ctx context.Context, userID string, amount float64) error {
	if amount <= 0 {
		return errors.New("refund amount must be positive")
	}

	balanceKey := fmt.Sprintf("transit:user:%s:balance", userID)
	return s.redis.IncrByFloat(ctx, balanceKey, amount).Err()
}

// GetBalance 获取余额
func (s *Service) GetBalance(ctx context.Context, userID string) (float64, error) {
	balanceKey := fmt.Sprintf("transit:user:%s:balance", userID)
	val, err := s.redis.Get(ctx, balanceKey).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// Recharge 充值
func (s *Service) Recharge(ctx context.Context, userID string, amount float64) error {
	if amount <= 0 {
		return errors.New("recharge amount must be positive")
	}

	balanceKey := fmt.Sprintf("transit:user:%s:balance", userID)
	return s.redis.IncrByFloat(ctx, balanceKey, amount).Err()
}
