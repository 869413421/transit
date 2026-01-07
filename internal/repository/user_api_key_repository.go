package repository

import (
	"context"

	"github.com/869413421/transit/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserAPIKeyRepository 用户API Key仓储接口
type UserAPIKeyRepository interface {
	Create(ctx context.Context, key *models.UserAPIKey) error
	FindByAPIKey(ctx context.Context, apiKey string) (*models.UserAPIKey, error)
	FindByUserID(ctx context.Context, userID string) ([]*models.UserAPIKey, error)
	Delete(ctx context.Context, id string) error
}

type userAPIKeyRepository struct {
	db *pgxpool.Pool
}

// NewUserAPIKeyRepository 创建用户API Key仓储
func NewUserAPIKeyRepository(db *pgxpool.Pool) UserAPIKeyRepository {
	return &userAPIKeyRepository{db: db}
}

func (r *userAPIKeyRepository) Create(ctx context.Context, key *models.UserAPIKey) error {
	query := `
		INSERT INTO user_api_keys (id, user_id, api_key, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		key.ID,
		key.UserID,
		key.APIKey,
		key.IsActive,
		key.CreatedAt,
	)
	return err
}

func (r *userAPIKeyRepository) FindByAPIKey(ctx context.Context, apiKey string) (*models.UserAPIKey, error) {
	var key models.UserAPIKey
	query := `SELECT id, user_id, api_key, is_active, created_at FROM user_api_keys WHERE api_key = $1`
	err := r.db.QueryRow(ctx, query, apiKey).Scan(
		&key.ID,
		&key.UserID,
		&key.APIKey,
		&key.IsActive,
		&key.CreatedAt,
	)
	return &key, err
}

func (r *userAPIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*models.UserAPIKey, error) {
	query := `SELECT id, user_id, api_key, is_active, created_at FROM user_api_keys WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.UserAPIKey
	for rows.Next() {
		var key models.UserAPIKey
		if err := rows.Scan(&key.ID, &key.UserID, &key.APIKey, &key.IsActive, &key.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}

func (r *userAPIKeyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM user_api_keys WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
