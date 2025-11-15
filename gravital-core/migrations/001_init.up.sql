-- Gravital Core 数据库初始化脚本

-- 启用 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- 用户和权限表
-- ============================================================

-- 角色表
CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) UNIQUE NOT NULL,
    permissions JSONB,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 用户表
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    role_id BIGINT REFERENCES roles(id),
    enabled BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- API Token 表
CREATE TABLE api_tokens (
    id BIGSERIAL PRIMARY KEY,
    token VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    user_id BIGINT REFERENCES users(id),
    sentinel_id VARCHAR(64),
    permissions JSONB,
    expires_at TIMESTAMP,
    last_used TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
-- 设备管理表
-- ============================================================

-- 设备分组表
CREATE TABLE device_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    parent_id BIGINT REFERENCES device_groups(id),
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 设备表
CREATE TABLE devices (
    id BIGSERIAL PRIMARY KEY,
    device_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(64) NOT NULL,
    group_id BIGINT REFERENCES device_groups(id),
    sentinel_id VARCHAR(64),
    connection_config JSONB,
    labels JSONB,
    status VARCHAR(32) DEFAULT 'unknown',
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 设备模板表
CREATE TABLE device_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(64) NOT NULL,
    connection_schema JSONB,
    default_config JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
-- Sentinel 管理表
-- ============================================================

-- Sentinel 注册表
CREATE TABLE sentinels (
    id BIGSERIAL PRIMARY KEY,
    sentinel_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255),
    hostname VARCHAR(255),
    ip_address VARCHAR(64),
    version VARCHAR(32),
    os VARCHAR(64),
    arch VARCHAR(32),
    region VARCHAR(64),
    labels JSONB,
    api_token VARCHAR(255),
    status VARCHAR(32),
    last_heartbeat TIMESTAMP,
    registered_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 心跳记录表
CREATE TABLE sentinel_heartbeats (
    id BIGSERIAL PRIMARY KEY,
    sentinel_id VARCHAR(64) NOT NULL,
    cpu_usage FLOAT,
    memory_usage FLOAT,
    disk_usage FLOAT,
    task_count INTEGER,
    plugin_count INTEGER,
    uptime_seconds BIGINT,
    received_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
-- 采集任务表
-- ============================================================

-- 采集任务表
CREATE TABLE collection_tasks (
    id BIGSERIAL PRIMARY KEY,
    task_id VARCHAR(64) UNIQUE NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    sentinel_id VARCHAR(64) NOT NULL,
    plugin_name VARCHAR(64) NOT NULL,
    config JSONB,
    interval_seconds INTEGER,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 5,
    retry_count INTEGER DEFAULT 3,
    timeout_seconds INTEGER DEFAULT 30,
    last_executed_at TIMESTAMP,
    next_execution_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (device_id) REFERENCES devices(device_id)
);

-- 任务执行记录
CREATE TABLE task_executions (
    id BIGSERIAL PRIMARY KEY,
    task_id VARCHAR(64) NOT NULL,
    sentinel_id VARCHAR(64),
    status VARCHAR(32),
    metrics_collected INTEGER,
    error_message TEXT,
    execution_time_ms INTEGER,
    executed_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
-- 告警管理表
-- ============================================================

-- 告警规则表
CREATE TABLE alert_rules (
    id BIGSERIAL PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    severity VARCHAR(32),
    condition TEXT NOT NULL,
    filters JSONB,
    duration INTEGER,
    notification_config JSONB,
    inhibit_rules JSONB,
    mute_periods JSONB,
    description TEXT,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 告警记录表
CREATE TABLE alert_events (
    id BIGSERIAL PRIMARY KEY,
    alert_id VARCHAR(64) UNIQUE NOT NULL,
    rule_id BIGINT REFERENCES alert_rules(id),
    device_id VARCHAR(64),
    metric_name VARCHAR(255),
    severity VARCHAR(32),
    message TEXT,
    labels JSONB,
    triggered_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    status VARCHAR(32),
    notification_sent BOOLEAN DEFAULT false,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by BIGINT,
    acknowledged_at TIMESTAMP,
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 告警通知记录
CREATE TABLE alert_notifications (
    id BIGSERIAL PRIMARY KEY,
    alert_event_id BIGINT REFERENCES alert_events(id),
    channel VARCHAR(64),
    recipient VARCHAR(255),
    status VARCHAR(32),
    sent_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
-- 数据转发表
-- ============================================================

-- 转发配置表
CREATE TABLE forwarder_configs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(64) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    endpoint TEXT,
    auth_config JSONB,
    batch_size INTEGER,
    flush_interval INTEGER,
    retry_times INTEGER,
    timeout_seconds INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 转发统计表
CREATE TABLE forwarder_stats (
    id BIGSERIAL PRIMARY KEY,
    forwarder_name VARCHAR(255),
    success_count BIGINT,
    failed_count BIGINT,
    total_bytes BIGINT,
    avg_latency_ms INTEGER,
    recorded_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
-- 索引
-- ============================================================

-- 设备索引
CREATE INDEX idx_devices_sentinel ON devices(sentinel_id);
CREATE INDEX idx_devices_type ON devices(device_type);
CREATE INDEX idx_devices_status ON devices(status);

-- Sentinel 索引
CREATE INDEX idx_sentinels_status ON sentinels(status);
CREATE INDEX idx_sentinels_region ON sentinels(region);
CREATE INDEX idx_heartbeats_sentinel_time ON sentinel_heartbeats(sentinel_id, received_at DESC);

-- 任务索引
CREATE INDEX idx_tasks_sentinel ON collection_tasks(sentinel_id);
CREATE INDEX idx_tasks_device ON collection_tasks(device_id);
CREATE INDEX idx_tasks_next_exec ON collection_tasks(next_execution_at);
CREATE INDEX idx_task_executions_task ON task_executions(task_id);
CREATE INDEX idx_task_executions_time ON task_executions(executed_at DESC);

-- 告警索引
CREATE INDEX idx_alert_events_status ON alert_events(status);
CREATE INDEX idx_alert_events_device ON alert_events(device_id);
CREATE INDEX idx_alert_events_time ON alert_events(triggered_at DESC);
CREATE INDEX idx_alert_events_rule ON alert_events(rule_id);

-- 转发统计索引
CREATE INDEX idx_forwarder_stats_name ON forwarder_stats(forwarder_name);
CREATE INDEX idx_forwarder_stats_time ON forwarder_stats(recorded_at DESC);

-- ============================================================
-- 初始数据
-- ============================================================

-- 插入默认角色
INSERT INTO roles (name, permissions, description) VALUES
('admin', '["*"]'::jsonb, '管理员，拥有所有权限'),
('operator', '["devices.*", "alerts.*", "tasks.read", "sentinels.read"]'::jsonb, '运维人员'),
('viewer', '["*.read"]'::jsonb, '只读用户');

-- 插入默认管理员用户 (密码: admin123)
-- bcrypt hash of "admin123"
INSERT INTO users (username, email, password_hash, role_id, enabled) VALUES
('admin', 'admin@example.com', '$2a$10$C/d6qPp3yGedbXT9kxnTieZYlwboRXIy.FcFjrie/yghKedWwR8yG', 1, true);

-- 插入默认设备分组
INSERT INTO device_groups (name, parent_id, description) VALUES
('全部设备', NULL, '根分组'),
('网络设备', 1, '交换机、路由器等网络设备'),
('服务器', 1, '物理服务器和虚拟机'),
('存储设备', 1, '存储阵列、NAS等');

-- ============================================================
-- 注释
-- ============================================================

COMMENT ON TABLE devices IS '设备表';
COMMENT ON TABLE device_groups IS '设备分组表';
COMMENT ON TABLE sentinels IS 'Sentinel 采集端注册表';
COMMENT ON TABLE collection_tasks IS '采集任务表';
COMMENT ON TABLE alert_rules IS '告警规则表';
COMMENT ON TABLE alert_events IS '告警事件表';
COMMENT ON TABLE users IS '用户表';
COMMENT ON TABLE roles IS '角色表';

