package models

import "time"

// User 用户模型
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Balance   float64   `json:"balance" gorm:"type:decimal(15,4);default:0"`
	Status    int       `json:"status" gorm:"default:1"` // 1:正常 0:禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserAPIKey 用户API密钥
type UserAPIKey struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	APIKey    string    `json:"api_key" gorm:"unique;not null;index"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
}

// Channel 上游渠道
type Channel struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	Name               string    `json:"name"`
	SecretKey          string    `json:"secret_key" gorm:"not null"`
	BaseURL            string    `json:"base_url"`
	MaxConcurrency     int       `json:"max_concurrency" gorm:"default:200"`
	CurrentConcurrency int       `json:"current_concurrency" gorm:"default:0"`
	Weight             int       `json:"weight" gorm:"default:10"`
	IsActive           bool      `json:"is_active" gorm:"default:true;index"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Task 任务记录
type Task struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	UserID         string    `json:"user_id" gorm:"not null;index"`
	ChannelID      string    `json:"channel_id" gorm:"not null"`
	Type           string    `json:"type" gorm:"type:enum('sync','async');not null"` // sync/async
	ModelName      string    `json:"model_name"`
	UpstreamTaskID string    `json:"upstream_task_id"`
	Status         string    `json:"status" gorm:"default:'running';index"` // running/completed/failed
	Cost           float64   `json:"cost" gorm:"type:decimal(15,4);default:0"`
	ResultURL      string    `json:"result_url" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BillingLog 账单流水
type BillingLog struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Amount    float64   `json:"amount" gorm:"type:decimal(15,4);not null"`
	LogType   string    `json:"log_type" gorm:"not null;index"` // recharge/consume/refund
	TaskID    string    `json:"task_id"`
	Remark    string    `json:"remark" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}
