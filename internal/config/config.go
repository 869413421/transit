package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Admin    AdminConfig    `mapstructure:"admin"`
	Models   ModelsConfig   // 模型配置,单独加载
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        string `mapstructure:"port"`
	Environment string `mapstructure:"environment"` // development, production
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Token string `mapstructure:"token"` // 管理员 API Token
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 读取配置文件（可选）
	_ = viper.ReadInConfig()

	// 绑定环境变量
	viper.AutomaticEnv()
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.environment", "SERVER_ENVIRONMENT")
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "DATABASE_DBNAME")
	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("admin.token", "ADMIN_TOKEN")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 加载模型配置
	if err := loadModels(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load models config: %w", err)
	}

	return &cfg, nil
}

// loadModels 加载模型配置
func loadModels(cfg *Config) error {
	modelsViper := viper.New()
	modelsViper.SetConfigName("models")
	modelsViper.SetConfigType("yaml")
	modelsViper.AddConfigPath("./configs")
	modelsViper.AddConfigPath(".")

	if err := modelsViper.ReadInConfig(); err != nil {
		return fmt.Errorf("read models config: %w", err)
	}

	var allModels AllModels
	if err := modelsViper.Unmarshal(&allModels); err != nil {
		return fmt.Errorf("unmarshal models config: %w", err)
	}

	cfg.Models = allModels.Models
	return nil
}
