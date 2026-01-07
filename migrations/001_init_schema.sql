-- Transit API 中转站数据库架构
-- 支持高并发多模型 API 转发

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    balance DECIMAL(15, 4) DEFAULT 0.0000 COMMENT '账户余额',
    status TINYINT DEFAULT 1 COMMENT '1:正常 0:禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 用户 API Key 表
CREATE TABLE IF NOT EXISTS user_api_keys (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    api_key VARCHAR(64) UNIQUE NOT NULL COMMENT '用户的 API Key (sk-xxxx)',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_api_key (api_key),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户API密钥表';

-- 上游渠道 Key 池表
CREATE TABLE IF NOT EXISTS channels (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) COMMENT '供应商名称',
    secret_key VARCHAR(255) NOT NULL COMMENT '上游 API Key',
    base_url VARCHAR(255) COMMENT 'API 基础地址',
    max_concurrency INT DEFAULT 200 COMMENT '最大并发数',
    current_concurrency INT DEFAULT 0 COMMENT '当前并发数(实时由Redis维护)',
    weight INT DEFAULT 10 COMMENT '负载均衡权重',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='上游渠道Key池';

-- 任务记录表 (支持同步和异步)
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    channel_id VARCHAR(36) NOT NULL,
    type ENUM('sync', 'async') NOT NULL COMMENT 'sync:文本 async:图片视频',
    model_name VARCHAR(100) COMMENT '模型名称',
    upstream_task_id VARCHAR(100) COMMENT '上游任务ID(仅异步)',
    status VARCHAR(20) DEFAULT 'running' COMMENT 'running/completed/failed',
    cost DECIMAL(15, 4) DEFAULT 0 COMMENT '消耗金额',
    result_url TEXT COMMENT '结果链接(异步任务)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务记录表';

-- 账单流水表
CREATE TABLE IF NOT EXISTS billing_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    amount DECIMAL(15, 4) NOT NULL COMMENT '金额(正数:充值 负数:扣费)',
    log_type VARCHAR(20) NOT NULL COMMENT 'recharge/consume/refund',
    task_id VARCHAR(36) COMMENT '关联任务ID',
    remark TEXT COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_log_type (log_type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账单流水表';
