// Package migrate 提供数据库迁移功能
// 使用 golang-migrate 进行版本化的数据库迁移管理
package migrate

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/869413421/transit/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"go.uber.org/zap"
)

//go:embed *.sql
var migrationsFS embed.FS

// Run 执行数据库迁移
// 参数:
//   - dsn: 数据库连接字符串
//
// 返回:
//   - error: 迁移失败时返回错误
func Run(dsn string) error {
	// 连接数据库
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// 创建迁移驱动
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	// 从嵌入的文件系统创建迁移源
	sourceDriver, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	// 创建迁移实例
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	// 获取当前版本
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("get migration version: %w", err)
	}

	if dirty {
		logger.Warn("Database is in dirty state, forcing version", zap.Uint("version", version))
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("force version: %w", err)
		}
	}

	logger.Info("Current migration version", zap.Uint("version", version))

	// 执行迁移
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("No migration changes to apply")
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	// 获取新版本
	newVersion, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("get new version: %w", err)
	}

	logger.Info("Migrations applied successfully", zap.Uint("new_version", newVersion))
	return nil
}

// Rollback 回滚最后一次迁移
// 参数:
//   - dsn: 数据库连接字符串
//
// 返回:
//   - error: 回滚失败时返回错误
func Rollback(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("rollback migration: %w", err)
	}

	logger.Info("Migration rolled back successfully")
	return nil
}
