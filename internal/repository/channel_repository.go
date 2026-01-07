package repository

import (
	"context"

	"github.com/869413421/transit/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChannelRepository 渠道仓储接口
type ChannelRepository interface {
	Create(ctx context.Context, channel *models.Channel) error
	FindByID(ctx context.Context, id string) (*models.Channel, error)
	FindAll(ctx context.Context) ([]*models.Channel, error)
	FindActive(ctx context.Context) ([]*models.Channel, error)
	Update(ctx context.Context, channel *models.Channel) error
	Delete(ctx context.Context, id string) error
}

type channelRepository struct {
	db *pgxpool.Pool
}

// NewChannelRepository 创建渠道仓储
func NewChannelRepository(db *pgxpool.Pool) ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(ctx context.Context, channel *models.Channel) error {
	query := `
		INSERT INTO channels (id, name, secret_key, base_url, max_concurrency, weight, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		channel.ID,
		channel.Name,
		channel.SecretKey,
		channel.BaseURL,
		channel.MaxConcurrency,
		channel.Weight,
		channel.IsActive,
		channel.CreatedAt,
		channel.UpdatedAt,
	)
	return err
}

func (r *channelRepository) FindByID(ctx context.Context, id string) (*models.Channel, error) {
	var channel models.Channel
	query := `SELECT id, name, secret_key, base_url, max_concurrency, current_concurrency, weight, is_active, created_at, updated_at FROM channels WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&channel.ID,
		&channel.Name,
		&channel.SecretKey,
		&channel.BaseURL,
		&channel.MaxConcurrency,
		&channel.CurrentConcurrency,
		&channel.Weight,
		&channel.IsActive,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)
	return &channel, err
}

func (r *channelRepository) FindAll(ctx context.Context) ([]*models.Channel, error) {
	query := `SELECT id, name, secret_key, base_url, max_concurrency, current_concurrency, weight, is_active, created_at, updated_at FROM channels`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.Channel
	for rows.Next() {
		var channel models.Channel
		if err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.SecretKey,
			&channel.BaseURL,
			&channel.MaxConcurrency,
			&channel.CurrentConcurrency,
			&channel.Weight,
			&channel.IsActive,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		); err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}
	return channels, rows.Err()
}

func (r *channelRepository) FindActive(ctx context.Context) ([]*models.Channel, error) {
	query := `SELECT id, name, secret_key, base_url, max_concurrency, current_concurrency, weight, is_active, created_at, updated_at FROM channels WHERE is_active = true`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.Channel
	for rows.Next() {
		var channel models.Channel
		if err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.SecretKey,
			&channel.BaseURL,
			&channel.MaxConcurrency,
			&channel.CurrentConcurrency,
			&channel.Weight,
			&channel.IsActive,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		); err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}
	return channels, rows.Err()
}

func (r *channelRepository) Update(ctx context.Context, channel *models.Channel) error {
	query := `
		UPDATE channels 
		SET name = $2, secret_key = $3, base_url = $4, max_concurrency = $5, 
		    weight = $6, is_active = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		channel.ID,
		channel.Name,
		channel.SecretKey,
		channel.BaseURL,
		channel.MaxConcurrency,
		channel.Weight,
		channel.IsActive,
		channel.UpdatedAt,
	)
	return err
}

func (r *channelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channels WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
