# 采集端自动注册指南

## 概述

采集端(Orbital Sentinel)支持自动注册到中心端(Gravital Core),无需手动配置 `sentinel_id` 和 `api_token`。

### 核心特性

✅ **自动注册**: 首次启动时自动向中心端注册,获取唯一凭证  
✅ **凭证持久化**: 本地保存凭证,重启后无需重新注册  
✅ **凭证验证**: 启动时自动验证凭证有效性  
✅ **智能降级**: 注册失败时自动降级为独立模式  
✅ **重复注册处理**: 基于 Hostname 识别,避免重复注册  

---

## 快速开始

### 1. 配置中心端地址

编辑配置文件 `config/config.yaml`:

```yaml
sentinel:
  name: "sentinel-office-1"
  region: "beijing"
  labels:
    environment: "production"
    datacenter: "dc1"

core:
  url: "http://gravital-core:8080"  # 中心端地址
  # registration_key: "your-key"    # 可选:注册密钥
```

### 2. 启动采集端

```bash
./bin/sentinel start -c config/config.yaml
```

### 3. 查看注册状态

启动日志会显示注册过程:

```
{"level":"INFO","msg":"No valid credentials found, attempting to register to core..."}
{"level":"INFO","msg":"Registering to core","hostname":"my-host","ip":"192.168.1.100"}
{"level":"INFO","msg":"Registration successful","sentinel_id":"sentinel-my-host-abc12345-1699999999"}
{"level":"INFO","msg":"Successfully registered to core","sentinel_id":"sentinel-my-host-abc12345-1699999999"}
```

### 4. 凭证文件

凭证自动保存到:
- **Linux/Mac**: `~/.sentinel/credentials.yaml`
- **Windows**: `%USERPROFILE%\.sentinel\credentials.yaml`

内容示例:

```yaml
sentinel_id: sentinel-my-host-abc12345-1699999999
api_token: sentinel_a1b2c3d4e5f6...
core_url: http://gravital-core:8080
registered_at: 2025-11-13T10:30:00Z
region: beijing
labels:
  environment: production
  datacenter: dc1
```

---

## 高级配置

### 自定义凭证路径

```yaml
credentials_path: "/etc/sentinel/credentials.yaml"
```

### 注册密钥(准入控制)

如果中心端启用了注册密钥验证:

```yaml
core:
  url: "http://gravital-core:8080"
  registration_key: "reg_secret_key_123456"
```

---

## 工作流程

### 首次启动(无凭证)

```
1. 检查本地凭证
   └─> 无凭证
   
2. 收集设备信息
   ├─ Hostname
   ├─ IP Address
   ├─ MAC Address
   ├─ OS/Arch
   └─ Version
   
3. 调用注册 API
   POST /api/v1/sentinels/register
   
4. 保存凭证到本地
   ~/.sentinel/credentials.yaml
   
5. 继续启动流程
```

### 重启(有凭证)

```
1. 检查本地凭证
   └─> 有凭证
   
2. 验证凭证有效性
   └─> 发送测试心跳
       ├─ 有效 → 继续启动
       └─ 无效 → 重新注册
       
3. 使用凭证启动
```

### 注册失败(容错)

```
注册失败
  ├─ 网络错误 → 重试5次
  │   ├─ 0秒
  │   ├─ 5秒
  │   ├─ 10秒
  │   ├─ 30秒
  │   └─ 60秒
  │
  └─ 失败后降级
      ├─ 切换为 direct 模式
      ├─ 使用本地任务配置
      └─ 正常运行
```

---

## 重复注册处理

采集端基于 **Hostname** 识别唯一性:

- 相同 Hostname 再次注册 → 更新已有记录,返回原 Token
- 不同 Hostname → 创建新记录

**场景示例**:

| 场景 | Hostname | 行为 |
|------|----------|------|
| 首次注册 | host-1 | 创建新记录 |
| 重启(有凭证) | host-1 | 使用本地凭证 |
| 重装系统 | host-1 | 检测到重复,返回原凭证 |
| 新机器 | host-2 | 创建新记录 |

---

## 凭证管理

### 查看凭证

```bash
cat ~/.sentinel/credentials.yaml
```

### 删除凭证(重新注册)

```bash
rm ~/.sentinel/credentials.yaml
./bin/sentinel start
```

### 手动指定凭证路径

```bash
./bin/sentinel start -c config/config.yaml --credentials /tmp/creds.yaml
```

---

## 独立模式(无中心端)

如果不配置 `core.url`,采集端会以独立模式运行:

```yaml
# 不配置 core.url
# core:
#   url: ""

sender:
  mode: "direct"  # 直连模式
  direct:
    prometheus:
      enabled: true
      url: "http://prometheus:9090/api/v1/write"
```

---

## 故障排查

### 问题1: 注册失败

**症状**:
```
{"level":"WARN","msg":"Failed to register to core, falling back to standalone mode"}
```

**检查**:
1. 中心端是否运行: `curl http://gravital-core:8080/api/v1/health`
2. 网络是否通畅: `ping gravital-core`
3. 注册密钥是否正确(如果启用)

### 问题2: 凭证验证失败

**症状**:
```
{"level":"WARN","msg":"Credentials validation failed, attempting to re-register..."}
```

**原因**:
- API Token 无效
- 中心端数据库清空

**解决**:
```bash
# 删除旧凭证,重新注册
rm ~/.sentinel/credentials.yaml
./bin/sentinel start
```

### 问题3: 无法写入凭证文件

**症状**:
```
{"level":"ERROR","msg":"failed to save credentials: permission denied"}
```

**解决**:
```bash
# 检查目录权限
mkdir -p ~/.sentinel
chmod 700 ~/.sentinel
```

### 问题4: 中心端不可用

**症状**:
```
{"level":"WARN","msg":"Failed to register to core, falling back to standalone mode"}
{"level":"INFO","msg":"Using direct send mode"}
```

**行为**:
- 采集端自动降级为独立模式
- 使用本地任务配置
- 数据直连发送到 TSDB
- 正常运行,不影响数据采集

---

## 测试

### 运行自动化测试

```bash
cd orbital-sentinels
./scripts/test-registration.sh
```

测试内容:
1. ✅ 检查中心端状态
2. ✅ 清除旧凭证
3. ✅ 首次注册
4. ✅ 凭证保存
5. ✅ 凭证验证
6. ✅ 心跳测试
7. ✅ 中心端查询
8. ✅ 重启(使用已有凭证)

### 手动测试

```bash
# 1. 启动中心端
cd gravital-core
docker-compose up -d

# 2. 清除旧凭证
rm ~/.sentinel/credentials.yaml

# 3. 启动采集端
cd orbital-sentinels
./bin/sentinel start -c config/config.register-test.yaml

# 4. 查看凭证
cat ~/.sentinel/credentials.yaml

# 5. 重启测试
# Ctrl+C 停止
./bin/sentinel start -c config/config.register-test.yaml
# 应该看到 "Using existing credentials"
```

---

## API 接口

### 注册接口

**端点**: `POST /api/v1/sentinels/register`

**请求**:
```json
{
  "name": "sentinel-office-1",
  "hostname": "my-host",
  "ip_address": "192.168.1.100",
  "mac_address": "00:11:22:33:44:55",
  "version": "1.0.0",
  "os": "linux",
  "arch": "amd64",
  "region": "beijing",
  "labels": {
    "environment": "production"
  },
  "registration_key": "optional-key"
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "sentinel_id": "sentinel-my-host-abc12345-1699999999",
    "api_token": "sentinel_a1b2c3d4e5f6...",
    "config": {
      "heartbeat_interval": 30,
      "task_fetch_interval": 60
    },
    "message": "Sentinel registered successfully"
  }
}
```

### 心跳接口

**端点**: `POST /api/v1/sentinels/heartbeat`

**Headers**:
- `X-Sentinel-ID`: sentinel-my-host-abc12345-1699999999
- `X-API-Token`: sentinel_a1b2c3d4e5f6...

**请求**:
```json
{
  "cpu_usage": 45.5,
  "memory_usage": 60.2,
  "disk_usage": 70.0,
  "task_count": 5,
  "plugin_count": 3,
  "uptime_seconds": 3600
}
```

---

## 安全建议

### 1. 凭证文件权限

```bash
# 设置为仅所有者可读写
chmod 600 ~/.sentinel/credentials.yaml
```

### 2. 使用注册密钥

生产环境建议启用注册密钥验证:

```yaml
# 采集端配置
core:
  registration_key: "your-secret-key"

# 中心端配置
sentinel:
  registration:
    mode: "key"  # 需要密钥
    registration_key: "your-secret-key"
```

### 3. Token 管理

- ✅ Token 自动生成,无需手动设置
- ✅ Token 本地加密存储
- ✅ Token 定期轮换(可选功能)

---

## 总结

采集端自动注册功能提供了:

✅ **零配置部署**: 无需手动配置凭证  
✅ **高可用性**: 支持中心端不可用时的降级  
✅ **安全可靠**: 凭证加密存储,支持准入控制  
✅ **易于管理**: 自动处理凭证生命周期  
✅ **生产就绪**: 完善的容错和重试机制  

让大规模采集端的部署和管理变得简单高效! 🚀

