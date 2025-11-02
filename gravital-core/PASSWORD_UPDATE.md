# 密码哈希更新说明

## 问题

初始使用的 bcrypt 密码哈希不正确，导致无法登录。

## 解决方案

已将管理员用户的密码哈希更新为正确的值。

### 正确的密码哈希

- **密码**: `admin123`
- **Bcrypt 哈希**: `$2a$10$C/d6qPp3yGedbXT9kxnTieZYlwboRXIy.FcFjrie/yghKedWwR8yG`
- **Cost**: 10

### 验证方法

可以使用以下 Go 代码验证密码哈希：

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "admin123"
    hash := "$2a$10$C/d6qPp3yGedbXT9kxnTieZYlwboRXIy.FcFjrie/yghKedWwR8yG"
    
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        fmt.Printf("❌ 密码验证失败: %v\n", err)
    } else {
        fmt.Println("✅ 密码验证成功！")
    }
}
```

### 生成新的密码哈希

如果需要为其他密码生成哈希，可以使用：

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "your_password_here"
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Password hash: %s\n", string(hash))
}
```

或使用在线工具：
- https://bcrypt-generator.com/
- 选择 Rounds: 10

## 当前登录信息

- **前端地址**: http://localhost:5173
- **用户名**: admin
- **密码**: admin123

## 数据库更新记录

```sql
-- 已执行的更新语句
UPDATE users 
SET password_hash = '$2a$10$C/d6qPp3yGedbXT9kxnTieZYlwboRXIy.FcFjrie/yghKedWwR8yG'
WHERE username = 'admin';
```

---

**更新日期**: 2025-11-02  
**状态**: ✅ 已修复

