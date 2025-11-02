#!/bin/bash

# 简单的创建管理员用户脚本

set -e

echo "🔧 创建 Gravital Core 管理员用户..."

# 数据库配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-gravital}
DB_PASSWORD=${DB_PASSWORD:-gravital123}
DB_NAME=${DB_NAME:-gravital_core}

# 管理员密码 "admin123" 的 bcrypt 哈希 (cost=10)
# 你可以使用在线工具或 Go 代码生成: bcrypt.GenerateFromPassword([]byte("admin123"), 10)
PASSWORD_HASH='$2a$10$C/d6qPp3yGedbXT9kxnTieZYlwboRXIy.FcFjrie/yghKedWwR8yG'

echo "📊 连接数据库并创建用户..."

PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME <<SQL
-- 插入管理员角色
INSERT INTO roles (name, permissions, description) VALUES
    ('admin', '["*"]', '管理员，拥有所有权限')
ON CONFLICT (name) DO UPDATE SET 
    permissions = EXCLUDED.permissions,
    description = EXCLUDED.description;

-- 插入运维角色
INSERT INTO roles (name, permissions, description) VALUES
    ('operator', '["devices.read", "devices.write", "sentinels.read", "tasks.read", "tasks.write", "alerts.read"]', '运维人员')
ON CONFLICT (name) DO NOTHING;

-- 插入只读角色
INSERT INTO roles (name, permissions, description) VALUES
    ('viewer', '["devices.read", "sentinels.read", "tasks.read", "alerts.read"]', '只读用户')
ON CONFLICT (name) DO NOTHING;

-- 插入或更新管理员用户
INSERT INTO users (username, email, password_hash, role_id, enabled) 
SELECT 
    'admin',
    'admin@gravital-core.local',
    '$PASSWORD_HASH',
    (SELECT id FROM roles WHERE name = 'admin'),
    true
ON CONFLICT (username) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    email = EXCLUDED.email,
    enabled = EXCLUDED.enabled;

-- 显示结果
SELECT '✅ 角色创建成功:' as info;
SELECT id, name, description FROM roles;

SELECT '✅ 用户创建成功:' as info;
SELECT id, username, email, enabled FROM users WHERE username = 'admin';
SQL

echo ""
echo "✅ 管理员用户创建成功！"
echo ""
echo "📋 登录信息:"
echo "   用户名: admin"
echo "   密码:   admin123"
echo "   邮箱:   admin@gravital-core.local"
echo ""
echo "🌐 现在可以使用以下地址登录:"
echo "   http://localhost:5173"
echo ""

