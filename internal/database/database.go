package database

import (
	"context"
	"fmt"
	"time"

	"github.com/869413421/transit/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool 创建 PostgreSQL 连接池
func NewPostgresPool(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// 配置连接池
	poolConfig.MaxConns = 100
	poolConfig.MinConns = 10
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 测试连接
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
