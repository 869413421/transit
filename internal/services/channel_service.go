package services

import (
	"context"

	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/internal/repository"
	"github.com/869413421/transit/pkg/pool"
)

// ChannelService 渠道服务接口
type ChannelService interface {
	Create(ctx context.Context, channel *models.Channel) error
	GetAll(ctx context.Context) ([]*models.Channel, error)
	GetAllWithConcurrency(ctx context.Context) ([]*models.Channel, error)
	Delete(ctx context.Context, id string) error
}

type channelService struct {
	repo repository.ChannelRepository
	pool *pool.RedisPool
}

// NewChannelService 创建渠道服务
func NewChannelService(repo repository.ChannelRepository, pool *pool.RedisPool) ChannelService {
	return &channelService{
		repo: repo,
		pool: pool,
	}
}

func (s *channelService) Create(ctx context.Context, channel *models.Channel) error {
	return s.repo.Create(ctx, channel)
}

func (s *channelService) GetAll(ctx context.Context) ([]*models.Channel, error) {
	return s.repo.FindAll(ctx)
}

func (s *channelService) GetAllWithConcurrency(ctx context.Context) ([]*models.Channel, error) {
	channels, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// 获取实时并发数
	for i := range channels {
		concurrency, _ := s.pool.GetConcurrency(ctx, channels[i].ID)
		channels[i].CurrentConcurrency = concurrency
	}

	return channels, nil
}

func (s *channelService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
