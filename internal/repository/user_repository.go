package repository

import (
	"context"

	"github.com/869413421/transit/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, balance, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Balance,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, balance, status, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Balance,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return &user, err
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET username = $2, balance = $3, status = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Balance,
		user.Status,
		user.UpdatedAt,
	)
	return err
}
