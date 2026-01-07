package repository

import (
	"context"

	"github.com/869413421/transit/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	FindByID(ctx context.Context, id string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	FindPendingTasks(ctx context.Context, limit int) ([]*models.Task, error)
}

type taskRepository struct {
	db *pgxpool.Pool
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository(db *pgxpool.Pool) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (id, user_id, channel_id, type, model_name, upstream_task_id, status, cost, result_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		task.ID,
		task.UserID,
		task.ChannelID,
		task.Type,
		task.ModelName,
		task.UpstreamTaskID,
		task.Status,
		task.Cost,
		task.ResultURL,
		task.CreatedAt,
		task.UpdatedAt,
	)
	return err
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	var task models.Task
	query := `
		SELECT id, user_id, channel_id, type, model_name, upstream_task_id, status, cost, result_url, created_at, updated_at
		FROM tasks WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.ChannelID,
		&task.Type,
		&task.ModelName,
		&task.UpstreamTaskID,
		&task.Status,
		&task.Cost,
		&task.ResultURL,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	return &task, err
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
	query := `
		UPDATE tasks
		SET status = $2, cost = $3, result_url = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		task.ID,
		task.Status,
		task.Cost,
		task.ResultURL,
		task.UpdatedAt,
	)
	return err
}

func (r *taskRepository) FindPendingTasks(ctx context.Context, limit int) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, channel_id, type, model_name, upstream_task_id, status, cost, result_url, created_at, updated_at
		FROM tasks
		WHERE status = 'running'
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.ChannelID,
			&task.Type,
			&task.ModelName,
			&task.UpstreamTaskID,
			&task.Status,
			&task.Cost,
			&task.ResultURL,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	return tasks, rows.Err()
}
