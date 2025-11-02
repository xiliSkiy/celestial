-- 初始化测试数据

-- 插入默认角色
INSERT INTO roles (name, permissions, description) VALUES
    ('admin', '["*"]', '管理员，拥有所有权限'),
    ('operator', '["devices.read", "devices.write", "sentinels.read", "tasks.read", "tasks.write", "alerts.read"]', '运维人员'),
    ('viewer', '["devices.read", "sentinels.read", "tasks.read", "alerts.read"]', '只读用户')
ON CONFLICT (name) DO NOTHING;

-- 插入默认管理员用户 (密码: admin123)
-- 密码哈希是 bcrypt 加密的 "admin123"
INSERT INTO users (username, email, password_hash, role_id, enabled) 
SELECT 
    'admin',
    'admin@gravital-core.local',
    '$2a$10$YourBcryptHashHere',  -- 需要替换为实际的 bcrypt 哈希
    (SELECT id FROM roles WHERE name = 'admin'),
    true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');

-- 插入测试用户 (密码: operator123)
INSERT INTO users (username, email, password_hash, role_id, enabled) 
SELECT 
    'operator',
    'operator@gravital-core.local',
    '$2a$10$YourBcryptHashHere',  -- 需要替换为实际的 bcrypt 哈希
    (SELECT id FROM roles WHERE name = 'operator'),
    true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'operator');

-- 插入只读用户 (密码: viewer123)
INSERT INTO users (username, email, password_hash, role_id, enabled) 
SELECT 
    'viewer',
    'viewer@gravital-core.local',
    '$2a$10$YourBcryptHashHere',  -- 需要替换为实际的 bcrypt 哈希
    (SELECT id FROM roles WHERE name = 'viewer'),
    true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'viewer');

-- 插入设备分组
INSERT INTO device_groups (name, description) VALUES
    ('生产环境', '生产环境设备'),
    ('测试环境', '测试环境设备')
ON CONFLICT DO NOTHING;

-- 显示创建的角色和用户
SELECT 'Roles:' as info;
SELECT id, name, permissions FROM roles;

SELECT 'Users:' as info;
SELECT id, username, email, enabled FROM users;

