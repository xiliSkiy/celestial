#!/bin/bash

# 创建管理员用户脚本

set -e

echo "🔧 创建 Gravital Core 管理员用户..."

# 数据库配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-gravital}
DB_PASSWORD=${DB_PASSWORD:-gravital123}
DB_NAME=${DB_NAME:-gravital_core}

# 用户信息
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
ADMIN_EMAIL=${ADMIN_EMAIL:-admin@gravital-core.local}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-admin123}

echo "📝 生成密码哈希..."
# 使用 Go 生成 bcrypt 哈希
PASSWORD_HASH=$(cd .. && go run -ldflags "-s -w" - <<EOF
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "os"
)
func main() {
    hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), 10)
    if err != nil {
        panic(err)
    }
    fmt.Print(string(hash))
}
EOF
$ADMIN_PASSWORD)

echo "📊 连接数据库..."

# 插入角色
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
    '$ADMIN_USERNAME',
    '$ADMIN_EMAIL',
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
SELECT id, username, email, enabled FROM users WHERE username = '$ADMIN_USERNAME';
SQL

echo ""
echo "✅ 管理员用户创建成功！"
echo ""
echo "📋 登录信息:"
echo "   用户名: $ADMIN_USERNAME"
echo "   密码:   $ADMIN_PASSWORD"
echo "   邮箱:   $ADMIN_EMAIL"
echo ""
echo "🌐 现在可以使用以下地址登录:"
echo "   http://localhost:3000"
echo ""

