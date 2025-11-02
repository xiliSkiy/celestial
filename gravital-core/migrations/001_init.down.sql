-- 回滚初始化脚本

-- 删除表（按依赖关系逆序）
DROP TABLE IF EXISTS forwarder_stats;
DROP TABLE IF EXISTS forwarder_configs;
DROP TABLE IF EXISTS alert_notifications;
DROP TABLE IF EXISTS alert_events;
DROP TABLE IF EXISTS alert_rules;
DROP TABLE IF EXISTS task_executions;
DROP TABLE IF EXISTS collection_tasks;
DROP TABLE IF EXISTS sentinel_heartbeats;
DROP TABLE IF EXISTS sentinels;
DROP TABLE IF EXISTS device_templates;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS device_groups;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

-- 删除扩展
DROP EXTENSION IF EXISTS "uuid-ossp";

