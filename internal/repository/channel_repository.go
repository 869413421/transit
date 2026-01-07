package repository

import (
	"github.com/869413421/transit/internal/models"
	"gorm.io/gorm"
)

// ChannelRepository 渠道仓储接口
type ChannelRepository interface {
	Create(channel *models.Channel) error
	FindByID(id string) (*models.Channel, error)
	FindAll() ([]*models.Channel, error)
	FindActive() ([]*models.Channel, error)
	Update(channel *models.Channel) error
	Delete(id string) error
}

type channelRepository struct {
	db *gorm.DB
}

// NewChannelRepository 创建渠道仓储
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(channel *models.Channel) error {
	return r.db.Create(channel).Error
}

func (r *channelRepository) FindByID(id string) (*models.Channel, error) {
	var channel models.Channel
	err := r.db.Where("id = ?", id).First(&channel).Error
	return &channel, err
}

func (r *channelRepository) FindAll() ([]*models.Channel, error) {
	var channels []*models.Channel
	err := r.db.Find(&channels).Error
	return channels, err
}

func (r *channelRepository) FindActive() ([]*models.Channel, error) {
	var channels []*models.Channel
	err := r.db.Where("is_active = ?", true).Find(&channels).Error
	return channels, err
}

func (r *channelRepository) Update(channel *models.Channel) error {
	return r.db.Save(channel).Error
}

func (r *channelRepository) Delete(id string) error {
	return r.db.Delete(&models.Channel{}, "id = ?", id).Error
}
