package loadbalancer

import (
	"context"
	"errors"
	"math/rand"

	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/internal/repository"
	"github.com/869413421/transit/pkg/logger"
	"github.com/869413421/transit/pkg/pool"
	"go.uber.org/zap"
)

// Selector 渠道选择器
type Selector struct {
	channelRepo repository.ChannelRepository
	pool        *pool.RedisPool
}

// NewSelector 创建渠道选择器
func NewSelector(channelRepo repository.ChannelRepository, pool *pool.RedisPool) *Selector {
	return &Selector{
		channelRepo: channelRepo,
		pool:        pool,
	}
}

// SelectChannel 选择可用渠道并获取并发位
// 使用加权轮询算法,优先选择权重高且并发未满的渠道
func (s *Selector) SelectChannel(ctx context.Context) (*models.Channel, error) {
	// 获取所有激活的渠道
	channels, err := s.channelRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		return nil, errors.New("no available channels")
	}

	// 过滤出激活的渠道
	var activeChannels []*models.Channel
	for _, ch := range channels {
		if ch.IsActive {
			activeChannels = append(activeChannels, ch)
		}
	}

	if len(activeChannels) == 0 {
		return nil, errors.New("no active channels")
	}

	// 尝试获取并发位(最多尝试所有渠道)
	tried := make(map[string]bool)
	for len(tried) < len(activeChannels) {
		// 加权随机选择
		channel := s.weightedRandomSelect(activeChannels, tried)
		if channel == nil {
			break
		}

		tried[channel.ID] = true

		// 尝试获取并发位
		acquired, err := s.pool.Acquire(ctx, channel.ID, channel.MaxConcurrency)
		if err != nil {
			logger.Warn("Failed to acquire concurrency slot",
				zap.String("channel_id", channel.ID),
				zap.Error(err),
			)
			continue
		}

		if acquired {
			logger.Info("Channel selected",
				zap.String("channel_id", channel.ID),
				zap.String("channel_name", channel.Name),
			)
			return channel, nil
		}

		logger.Debug("Channel concurrency limit reached",
			zap.String("channel_id", channel.ID),
			zap.Int("max_concurrency", channel.MaxConcurrency),
		)
	}

	return nil, errors.New("all channels are at capacity")
}

// ReleaseChannel 释放渠道并发位
func (s *Selector) ReleaseChannel(ctx context.Context, channelID string) error {
	if err := s.pool.Release(ctx, channelID); err != nil {
		logger.Error("Failed to release concurrency slot",
			zap.String("channel_id", channelID),
			zap.Error(err),
		)
		return err
	}

	logger.Debug("Channel released", zap.String("channel_id", channelID))
	return nil
}

// weightedRandomSelect 加权随机选择渠道
func (s *Selector) weightedRandomSelect(channels []*models.Channel, tried map[string]bool) *models.Channel {
	// 计算未尝试渠道的总权重
	var totalWeight int
	for _, ch := range channels {
		if !tried[ch.ID] {
			totalWeight += ch.Weight
		}
	}

	if totalWeight == 0 {
		return nil
	}

	// 随机选择
	r := rand.Intn(totalWeight)
	var cumulative int
	for _, ch := range channels {
		if tried[ch.ID] {
			continue
		}
		cumulative += ch.Weight
		if r < cumulative {
			return ch
		}
	}

	return nil
}
