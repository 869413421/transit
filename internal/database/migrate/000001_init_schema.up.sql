-- Transit API 中转站初始化表结构

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    balance DECIMAL(15, 4) DEFAULT 0.0000,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);

-- 用户 API Key 表
CREATE TABLE IF NOT EXISTS user_api_keys (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key VARCHAR(64) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_api_keys_api_key ON user_api_keys(api_key);
CREATE INDEX idx_user_api_keys_user_id ON user_api_keys(user_id);

-- 上游渠道 Key 池表
CREATE TABLE IF NOT EXISTS channels (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50),
    secret_key VARCHAR(255) NOT NULL,
    base_url VARCHAR(255),
    max_concurrency INTEGER DEFAULT 200,
    current_concurrency INTEGER DEFAULT 0,
    weight INTEGER DEFAULT 10,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_channels_active ON channels(is_active);

-- 任务记录表
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id VARCHAR(36) NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('sync', 'async')),
    model_name VARCHAR(100),
    upstream_task_id VARCHAR(100),
    status VARCHAR(20) DEFAULT 'running',
    cost DECIMAL(15, 4) DEFAULT 0,
    result_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- 账单流水表
CREATE TABLE IF NOT EXISTS billing_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(15, 4) NOT NULL,
    log_type VARCHAR(20) NOT NULL,
    task_id VARCHAR(36),
    remark TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_billing_logs_user_id ON billing_logs(user_id);
CREATE INDEX idx_billing_logs_log_type ON billing_logs(log_type);
CREATE INDEX idx_billing_logs_created_at ON billing_logs(created_at);
