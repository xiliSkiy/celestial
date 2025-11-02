# 登录问题修复说明

## 问题描述

前端登录时报错：
```json
{
    "code": 20001,
    "message": "failed to get user: sql: Scan error on column index 2, name \"permissions\": json: cannot unmarshal array into Go value of type map[string]interface {}"
}
```

## 问题原因

数据库中 `roles.permissions` 字段存储的是 JSON 数组格式（如 `["devices.read", "devices.write"]`），但 Go 代码中 `JSONB` 类型被定义为 `map[string]interface{}`，导致反序列化失败。

## 修复方案

### 1. 创建 `StringArray` 类型

在 `internal/model/user.go` 中添加了新的类型来处理 JSON 数组：

```go
// StringArray 字符串数组类型，用于存储 JSON 数组
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
    if value == nil {
        *s = []string{}
        return nil
    }
    
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("failed to unmarshal JSONB value")
    }
    
    return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
    if len(s) == 0 {
        return "[]", nil
    }
    return json.Marshal(s)
}
```

### 2. 更新模型定义

将 `Role` 和 `APIToken` 的 `Permissions` 字段类型从 `JSONB` 改为 `StringArray`：

```go
type Role struct {
    ID          uint        `gorm:"primaryKey" json:"id"`
    Name        string      `gorm:"uniqueIndex;size:64;not null" json:"name"`
    Permissions StringArray `gorm:"type:jsonb" json:"permissions"`  // 改为 StringArray
    Description string      `gorm:"type:text" json:"description"`
    CreatedAt   time.Time   `json:"created_at"`
}

type APIToken struct {
    // ...
    Permissions StringArray `gorm:"type:jsonb" json:"permissions"`  // 改为 StringArray
    // ...
}
```

### 3. 更新 AuthService

简化权限获取逻辑，因为 `Permissions` 现在直接是 `[]string` 类型：

```go
// 修改前
var permissions []string
if user.Role != nil && user.Role.Permissions != nil {
    for key := range user.Role.Permissions {
        permissions = append(permissions, key)
    }
}

// 修改后
var permissions []string
if user.Role != nil {
    permissions = user.Role.Permissions
}
```

## 测试数据

已创建测试用户和角色：

### 角色
- **admin**: 管理员，权限 `["*"]`
- **operator**: 运维人员，权限 `["devices.read", "devices.write", "sentinels.read", "tasks.read", "tasks.write", "alerts.read"]`
- **viewer**: 只读用户，权限 `["devices.read", "sentinels.read", "tasks.read", "alerts.read"]`

### 用户
- **用户名**: admin
- **密码**: admin123
- **邮箱**: admin@gravital-core.local
- **角色**: admin

## 验证步骤

1. 确保服务已重新编译并启动：
```bash
cd gravital-core
make build
./bin/gravital-core -c config/config.yaml
```

2. 访问前端页面：
```
http://localhost:5173
```

3. 使用以下凭据登录：
- 用户名: `admin`
- 密码: `admin123`

## 创建更多用户

如果需要创建更多用户，可以使用以下 SQL：

```sql
-- 创建运维用户 (密码: operator123)
INSERT INTO users (username, email, password_hash, role_id, enabled) 
SELECT 
    'operator',
    'operator@gravital-core.local',
    '$2a$10$YourBcryptHashHere',  -- 需要生成 bcrypt 哈希
    (SELECT id FROM roles WHERE name = 'operator'),
    true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'operator');
```

或者使用 Docker 命令：

```bash
docker-compose exec -T postgres psql -U postgres -d gravital <<'SQL'
INSERT INTO users (username, email, password_hash, role_id, enabled) 
SELECT 
    'operator',
    'operator@gravital-core.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    (SELECT id FROM roles WHERE name = 'operator'),
    true
ON CONFLICT (username) DO NOTHING;
SQL
```

## 相关文件

- `internal/model/user.go` - 用户模型定义
- `internal/service/auth_service.go` - 认证服务
- `scripts/create-admin-simple.sh` - 创建管理员脚本
- `migrations/001_init.up.sql` - 数据库初始化脚本

## 注意事项

1. **密码哈希**: 示例中使用的密码哈希对应密码 `admin123`，生产环境请使用更强的密码
2. **权限格式**: 权限必须以 JSON 数组格式存储，如 `["permission1", "permission2"]`
3. **数据库**: 确保 PostgreSQL 和 Redis 服务正常运行

---

**修复日期**: 2025-11-02  
**修复版本**: v1.0.0

