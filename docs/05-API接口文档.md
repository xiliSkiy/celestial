# API 接口文档

## 1. 概述

### 1.1 基本信息
- **Base URL**: `https://gravital-core.example.com/api/v1`
- **协议**: HTTPS
- **认证方式**: JWT Token / API Token
- **数据格式**: JSON
- **编码**: UTF-8

### 1.2 通用响应格式

#### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 业务数据
  }
}
```

#### 错误响应
```json
{
  "code": 40001,
  "message": "参数错误: device_id 不能为空",
  "error": "ValidationError",
  "request_id": "req-123456"
}
```

#### 分页响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "items": [
      // 数据列表
    ]
  }
}
```

### 1.3 错误码定义
```yaml
通用错误码:
  0: 成功
  10001: 系统内部错误
  10002: 服务不可用
  10003: 请求超时

认证错误码:
  20001: 未认证
  20002: Token 无效
  20003: Token 过期
  20004: 权限不足

参数错误码:
  40001: 参数错误
  40002: 参数缺失
  40003: 参数格式错误
  40004: 参数值无效

资源错误码:
  50001: 资源不存在
  50002: 资源已存在
  50003: 资源冲突
  50004: 资源已删除
```

### 1.4 认证方式

#### JWT Token (用户认证)
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### API Token (Sentinel 认证)
```http
X-API-Token: sentinel_1a2b3c4d5e6f7g8h9i0j
```

## 2. 认证相关 API

### 2.1 用户登录
```http
POST /auth/login
```

**请求体**:
```json
{
  "username": "admin",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "refresh_token": "refresh_token_here",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin"
    }
  }
}
```

### 2.2 刷新 Token
```http
POST /auth/refresh
```

**请求体**:
```json
{
  "refresh_token": "refresh_token_here"
}
```

### 2.3 退出登录
```http
POST /auth/logout
Authorization: Bearer {token}
```

## 3. 设备管理 API

### 3.1 获取设备列表
```http
GET /devices
```

**查询参数**:
```yaml
page: 1                          # 页码，默认 1
page_size: 20                    # 每页数量，默认 20
group_id: 123                    # 设备组 ID
device_type: switch              # 设备类型
status: online                   # 设备状态: online, offline, error, unknown
keyword: switch01                # 关键字搜索（名称、IP）
sort: created_at                 # 排序字段
order: desc                      # 排序方向: asc, desc
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "device_id": "dev-001",
        "name": "Core Switch 01",
        "device_type": "switch",
        "group_id": 10,
        "group_name": "核心网络",
        "sentinel_id": "sentinel-001",
        "connection_config": {
          "host": "192.168.1.1",
          "port": 161,
          "community": "public"
        },
        "labels": {
          "env": "production",
          "region": "us-east"
        },
        "status": "online",
        "last_seen": "2025-11-01T10:30:00Z",
        "created_at": "2025-10-01T00:00:00Z",
        "updated_at": "2025-11-01T10:30:00Z"
      }
    ]
  }
}
```

### 3.2 获取设备详情
```http
GET /devices/{device_id}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "device_id": "dev-001",
    "name": "Core Switch 01",
    "device_type": "switch",
    "group_id": 10,
    "sentinel_id": "sentinel-001",
    "connection_config": {
      "host": "192.168.1.1",
      "port": 161,
      "community": "public"
    },
    "labels": {
      "env": "production"
    },
    "status": "online",
    "last_seen": "2025-11-01T10:30:00Z",
    "created_at": "2025-10-01T00:00:00Z",
    "updated_at": "2025-11-01T10:30:00Z",
    "metrics_summary": {
      "cpu_usage": 45.5,
      "memory_usage": 60.2,
      "uptime": 8640000
    },
    "collection_tasks": [
      {
        "task_id": "task-001",
        "plugin_name": "snmp",
        "interval": 60,
        "last_executed_at": "2025-11-01T10:29:00Z",
        "status": "success"
      }
    ]
  }
}
```

### 3.3 创建设备
```http
POST /devices
```

**请求体**:
```json
{
  "name": "Core Switch 01",
  "device_type": "switch",
  "group_id": 10,
  "sentinel_id": "sentinel-001",
  "connection_config": {
    "host": "192.168.1.1",
    "port": 161,
    "community": "public",
    "version": "2c"
  },
  "labels": {
    "env": "production",
    "region": "us-east"
  }
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "device_id": "dev-001"
  }
}
```

### 3.4 更新设备
```http
PUT /devices/{device_id}
```

**请求体**:
```json
{
  "name": "Core Switch 01 Updated",
  "group_id": 11,
  "connection_config": {
    "host": "192.168.1.2"
  },
  "labels": {
    "env": "production",
    "region": "us-west"
  }
}
```

### 3.5 删除设备
```http
DELETE /devices/{device_id}
```

### 3.6 批量导入设备
```http
POST /devices/batch-import
Content-Type: multipart/form-data
```

**表单数据**:
```
file: devices.csv
```

**CSV 格式**:
```csv
name,device_type,host,port,community,group_name,labels
Switch-01,switch,192.168.1.1,161,public,Core Network,"env=prod,region=us"
Switch-02,switch,192.168.1.2,161,public,Core Network,"env=prod,region=us"
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "total": 100,
    "success": 95,
    "failed": 5,
    "errors": [
      {
        "line": 10,
        "error": "重复的设备: 192.168.1.10"
      }
    ]
  }
}
```

### 3.7 测试设备连接
```http
POST /devices/{device_id}/test-connection
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "status": "success",
    "message": "连接成功",
    "latency_ms": 15,
    "details": {
      "device_name": "Cisco Catalyst 3750",
      "uptime": "30 days"
    }
  }
}
```

## 4. 设备分组 API

### 4.1 获取分组树
```http
GET /device-groups/tree
```

**响应**:
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "全部设备",
      "parent_id": null,
      "device_count": 100,
      "children": [
        {
          "id": 10,
          "name": "核心网络",
          "parent_id": 1,
          "device_count": 20,
          "children": []
        },
        {
          "id": 20,
          "name": "服务器",
          "parent_id": 1,
          "device_count": 50,
          "children": []
        }
      ]
    }
  ]
}
```

### 4.2 创建分组
```http
POST /device-groups
```

**请求体**:
```json
{
  "name": "边缘交换机",
  "parent_id": 10,
  "description": "边缘层交换机"
}
```

### 4.3 更新分组
```http
PUT /device-groups/{group_id}
```

### 4.4 删除分组
```http
DELETE /device-groups/{group_id}
```

## 5. 告警管理 API

### 5.1 获取告警规则列表
```http
GET /alert-rules
```

**查询参数**:
```yaml
page: 1
page_size: 20
enabled: true                    # 是否启用
severity: warning                # 告警级别
keyword: CPU                     # 关键字搜索
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "total": 50,
    "items": [
      {
        "id": 1,
        "rule_name": "CPU 使用率过高",
        "enabled": true,
        "severity": "warning",
        "condition": "cpu_usage > 80",
        "duration": 300,
        "filters": {
          "device_type": "server",
          "labels": {
            "env": "production"
          }
        },
        "notification_config": {
          "channels": ["email", "webhook"],
          "email": {
            "to": ["ops@example.com"]
          },
          "webhook": {
            "url": "https://hooks.slack.com/xxx"
          }
        },
        "created_at": "2025-10-01T00:00:00Z",
        "updated_at": "2025-10-15T00:00:00Z"
      }
    ]
  }
}
```

### 5.2 创建告警规则
```http
POST /alert-rules
```

**请求体**:
```json
{
  "rule_name": "CPU 使用率过高",
  "enabled": true,
  "severity": "warning",
  "condition": "cpu_usage > 80",
  "duration": 300,
  "filters": {
    "device_type": "server",
    "labels": {
      "env": "production"
    }
  },
  "notification_config": {
    "channels": ["email"],
    "email": {
      "to": ["ops@example.com"]
    }
  },
  "description": "当服务器 CPU 使用率持续 5 分钟超过 80% 时触发告警"
}
```

### 5.3 更新告警规则
```http
PUT /alert-rules/{rule_id}
```

### 5.4 删除告警规则
```http
DELETE /alert-rules/{rule_id}
```

### 5.5 启用/禁用告警规则
```http
POST /alert-rules/{rule_id}/toggle
```

**请求体**:
```json
{
  "enabled": false
}
```

### 5.6 获取告警事件列表
```http
GET /alert-events
```

**查询参数**:
```yaml
page: 1
page_size: 20
status: firing                   # firing, resolved, silenced
severity: critical               # critical, warning, info
device_id: dev-001
rule_id: 1
start_time: 2025-11-01T00:00:00Z
end_time: 2025-11-02T00:00:00Z
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "total": 200,
    "items": [
      {
        "id": 1,
        "alert_id": "alert-001",
        "rule_id": 1,
        "rule_name": "CPU 使用率过高",
        "device_id": "dev-001",
        "device_name": "Server-01",
        "metric_name": "cpu_usage",
        "severity": "warning",
        "message": "CPU 使用率 85.5% 超过阈值 80%",
        "current_value": 85.5,
        "threshold": 80,
        "labels": {
          "env": "production",
          "host": "server-01"
        },
        "triggered_at": "2025-11-01T10:00:00Z",
        "resolved_at": null,
        "status": "firing",
        "notification_sent": true,
        "acknowledged": false,
        "acknowledged_by": null,
        "acknowledged_at": null
      }
    ]
  }
}
```

### 5.7 确认告警
```http
POST /alert-events/{alert_id}/acknowledge
```

**请求体**:
```json
{
  "comment": "正在处理中"
}
```

### 5.8 解决告警
```http
POST /alert-events/{alert_id}/resolve
```

**请求体**:
```json
{
  "comment": "问题已解决"
}
```

### 5.9 静默告警
```http
POST /alert-events/{alert_id}/silence
```

**请求体**:
```json
{
  "duration": "1h",               # 1h, 2h, 1d, etc.
  "comment": "维护期间静默"
}
```

### 5.10 获取告警统计
```http
GET /alert-stats
```

**查询参数**:
```yaml
start_time: 2025-11-01T00:00:00Z
end_time: 2025-11-02T00:00:00Z
group_by: severity               # severity, device, rule
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "total": 500,
    "by_severity": {
      "critical": 50,
      "warning": 300,
      "info": 150
    },
    "by_status": {
      "firing": 100,
      "resolved": 350,
      "silenced": 50
    },
    "trend": [
      {
        "time": "2025-11-01T00:00:00Z",
        "count": 20
      },
      {
        "time": "2025-11-01T01:00:00Z",
        "count": 25
      }
    ]
  }
}
```

## 6. 采集任务 API

### 6.1 获取任务列表（Sentinel 调用）
```http
GET /tasks
X-Sentinel-ID: sentinel-001
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "tasks": [
      {
        "task_id": "task-001",
        "device_id": "dev-001",
        "plugin_name": "snmp",
        "config": {
          "oids": ["1.3.6.1.4.1.2021.11.9.0"]
        },
        "interval": 60,
        "timeout": 30,
        "priority": 5,
        "retry_count": 3,
        "device_config": {
          "host": "192.168.1.1",
          "port": 161,
          "community": "public"
        }
      }
    ],
    "config_version": 123
  }
}
```

### 6.2 上报任务执行结果
```http
POST /tasks/{task_id}/report
X-Sentinel-ID: sentinel-001
```

**请求体**:
```json
{
  "status": "success",            # success, failed, timeout
  "metrics_collected": 10,
  "error_message": null,
  "execution_time_ms": 150,
  "executed_at": "2025-11-01T10:00:00Z"
}
```

### 6.3 创建任务（管理端）
```http
POST /tasks
```

**请求体**:
```json
{
  "device_id": "dev-001",
  "sentinel_id": "sentinel-001",
  "plugin_name": "snmp",
  "config": {
    "oids": ["1.3.6.1.4.1.2021.11.9.0"]
  },
  "interval": 60,
  "priority": 5,
  "enabled": true
}
```

### 6.4 更新任务
```http
PUT /tasks/{task_id}
```

### 6.5 删除任务
```http
DELETE /tasks/{task_id}
```

### 6.6 手动触发任务
```http
POST /tasks/{task_id}/trigger
```

## 7. 数据采集 API

### 7.1 上报采集数据（Sentinel 调用）
```http
POST /data/ingest
X-Sentinel-ID: sentinel-001
Content-Encoding: gzip
```

**请求体**:
```json
{
  "metrics": [
    {
      "device_id": "dev-001",
      "name": "cpu_usage",
      "value": 75.5,
      "timestamp": 1698883200,
      "labels": {
        "host": "server-01",
        "region": "us-east"
      },
      "type": "gauge"
    },
    {
      "device_id": "dev-001",
      "name": "memory_usage",
      "value": 60.2,
      "timestamp": 1698883200,
      "labels": {
        "host": "server-01"
      },
      "type": "gauge"
    }
  ]
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "received": 1000,
    "accepted": 995,
    "rejected": 5
  }
}
```

### 7.2 查询指标数据
```http
GET /metrics/query
```

**查询参数**:
```yaml
query: cpu_usage{device_id="dev-001"}  # PromQL 风格查询
start: 1698883200                      # Unix 时间戳
end: 1698969600
step: 60                               # 步长（秒）
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "metric": {
      "name": "cpu_usage",
      "labels": {
        "device_id": "dev-001",
        "host": "server-01"
      }
    },
    "values": [
      [1698883200, "75.5"],
      [1698883260, "76.2"],
      [1698883320, "74.8"]
    ]
  }
}
```

### 7.3 查询指标列表
```http
GET /metrics/list
```

**查询参数**:
```yaml
device_id: dev-001
prefix: cpu_                     # 指标名前缀
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "metrics": [
      "cpu_usage",
      "cpu_system",
      "cpu_user",
      "cpu_idle"
    ]
  }
}
```

## 8. Sentinel 管理 API

### 8.1 Sentinel 注册
```http
POST /sentinels/register
```

**请求体**:
```json
{
  "name": "sentinel-office-01",
  "hostname": "sentinel-01.local",
  "ip_address": "192.168.1.100",
  "version": "1.0.0",
  "os": "linux",
  "arch": "amd64",
  "region": "office-beijing",
  "labels": {
    "env": "production"
  }
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "sentinel_id": "sentinel-001",
    "api_token": "sentinel_1a2b3c4d5e6f7g8h9i0j",
    "config": {
      "heartbeat_interval": 30,
      "task_fetch_interval": 60
    }
  }
}
```

### 8.2 心跳上报
```http
POST /sentinels/heartbeat
X-Sentinel-ID: sentinel-001
```

**请求体**:
```json
{
  "cpu_usage": 15.5,
  "memory_usage": 45.2,
  "disk_usage": 60.0,
  "task_count": 20,
  "plugin_count": 5,
  "uptime_seconds": 86400,
  "version": "1.0.0"
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "status": "ok",
    "config_version": 123,
    "commands": []                # 可能包含控制命令
  }
}
```

### 8.3 获取 Sentinel 列表
```http
GET /sentinels
```

**查询参数**:
```yaml
page: 1
page_size: 20
status: online                   # online, offline, error
region: office-beijing
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "total": 10,
    "items": [
      {
        "id": 1,
        "sentinel_id": "sentinel-001",
        "name": "sentinel-office-01",
        "hostname": "sentinel-01.local",
        "ip_address": "192.168.1.100",
        "version": "1.0.0",
        "os": "linux",
        "arch": "amd64",
        "region": "office-beijing",
        "status": "online",
        "cpu_usage": 15.5,
        "memory_usage": 45.2,
        "task_count": 20,
        "plugin_count": 5,
        "last_heartbeat": "2025-11-01T10:30:00Z",
        "registered_at": "2025-10-01T00:00:00Z"
      }
    ]
  }
}
```

### 8.4 获取 Sentinel 详情
```http
GET /sentinels/{sentinel_id}
```

### 8.5 远程控制 Sentinel
```http
POST /sentinels/{sentinel_id}/control
```

**请求体**:
```json
{
  "action": "reload_config",      # reload_config, restart, stop, update_plugin
  "params": {}
}
```

### 8.6 删除 Sentinel
```http
DELETE /sentinels/{sentinel_id}
```

## 9. 数据转发配置 API

### 9.1 获取转发器列表
```http
GET /forwarders
```

### 9.2 创建转发器
```http
POST /forwarders
```

**请求体**:
```json
{
  "name": "prometheus-prod",
  "type": "prometheus",
  "enabled": true,
  "endpoint": "http://prometheus:9090/api/v1/write",
  "auth_config": {
    "type": "basic",
    "username": "admin",
    "password": "password"
  },
  "batch_size": 1000,
  "flush_interval": 10,
  "retry_times": 3,
  "timeout": 30
}
```

### 9.3 更新转发器
```http
PUT /forwarders/{forwarder_id}
```

### 9.4 删除转发器
```http
DELETE /forwarders/{forwarder_id}
```

### 9.5 获取转发器统计
```http
GET /forwarders/{forwarder_id}/stats
```

**查询参数**:
```yaml
start_time: 2025-11-01T00:00:00Z
end_time: 2025-11-02T00:00:00Z
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "success_count": 95000,
    "failed_count": 500,
    "success_rate": 0.995,
    "total_bytes": 10485760,
    "avg_latency_ms": 25,
    "trend": [
      {
        "time": "2025-11-01T00:00:00Z",
        "success": 4000,
        "failed": 20
      }
    ]
  }
}
```

## 10. Dashboard API

### 10.1 获取 Dashboard 列表
```http
GET /dashboards
```

### 10.2 创建 Dashboard
```http
POST /dashboards
```

**请求体**:
```json
{
  "name": "网络设备监控",
  "description": "网络设备总览",
  "is_public": true,
  "layout": {
    "cols": 12,
    "rows": 6
  },
  "panels": [
    {
      "id": "panel-1",
      "type": "line-chart",
      "title": "CPU 使用率",
      "position": {"x": 0, "y": 0, "w": 6, "h": 3},
      "datasource": "prometheus",
      "query": "avg(cpu_usage{device_type='switch'})",
      "options": {
        "unit": "percent",
        "decimals": 2
      }
    }
  ],
  "variables": [
    {
      "name": "device_id",
      "type": "query",
      "query": "label_values(device_id)",
      "default": "dev-001"
    }
  ]
}
```

### 10.3 更新 Dashboard
```http
PUT /dashboards/{dashboard_id}
```

### 10.4 删除 Dashboard
```http
DELETE /dashboards/{dashboard_id}
```

### 10.5 获取 Dashboard 数据
```http
GET /dashboards/{dashboard_id}/data
```

**查询参数**:
```yaml
start_time: 2025-11-01T00:00:00Z
end_time: 2025-11-01T12:00:00Z
variables: {"device_id": "dev-001"}
```

## 11. 系统管理 API

### 11.1 获取系统信息
```http
GET /system/info
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "version": "1.0.0",
    "build_time": "2025-10-01T00:00:00Z",
    "go_version": "go1.21.0",
    "os": "linux",
    "arch": "amd64",
    "uptime_seconds": 864000,
    "statistics": {
      "total_devices": 1000,
      "online_devices": 950,
      "total_sentinels": 10,
      "online_sentinels": 9,
      "total_alert_rules": 50,
      "active_alerts": 20
    }
  }
}
```

### 11.2 健康检查
```http
GET /health
```

**响应**:
```json
{
  "status": "healthy",
  "components": {
    "database": "healthy",
    "redis": "healthy",
    "forwarders": {
      "prometheus": "healthy",
      "victoria-metrics": "degraded"
    }
  }
}
```

### 11.3 获取配置
```http
GET /system/config
```

### 11.4 更新配置
```http
PUT /system/config
```

**请求体**:
```json
{
  "config_key": "alert.evaluation_interval",
  "config_value": 15
}
```

## 12. WebSocket API

### 12.1 实时告警推送
```javascript
// WebSocket 连接
ws://gravital-core.example.com/ws/alerts?token={jwt_token}

// 接收消息格式
{
  "type": "alert",
  "data": {
    "alert_id": "alert-001",
    "rule_name": "CPU 使用率过高",
    "severity": "warning",
    "device_id": "dev-001",
    "message": "CPU 使用率 85.5% 超过阈值 80%",
    "triggered_at": "2025-11-01T10:00:00Z"
  }
}
```

### 12.2 实时指标推送
```javascript
// WebSocket 连接
ws://gravital-core.example.com/ws/metrics?device_id=dev-001&token={jwt_token}

// 接收消息格式
{
  "type": "metric",
  "data": {
    "device_id": "dev-001",
    "metrics": [
      {
        "name": "cpu_usage",
        "value": 75.5,
        "timestamp": 1698883200
      }
    ]
  }
}
```

## 13. 速率限制

### 13.1 限制规则
```yaml
用户 API:
  - 速率: 1000 req/min
  - 突发: 100 req/s

Sentinel API:
  - 速率: 10000 req/min
  - 突发: 1000 req/s

数据上报:
  - 速率: 100000 metrics/min
```

### 13.2 响应头
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1698883260
```

### 13.3 超限响应
```json
{
  "code": 42901,
  "message": "请求过于频繁，请稍后再试",
  "error": "RateLimitExceeded",
  "retry_after": 60
}
```

## 14. API 使用示例

### 14.1 Python 示例
```python
import requests

# 登录获取 Token
response = requests.post(
    "https://gravital-core.example.com/api/v1/auth/login",
    json={
        "username": "admin",
        "password": "password"
    }
)
token = response.json()["data"]["token"]

# 使用 Token 调用 API
headers = {"Authorization": f"Bearer {token}"}

# 获取设备列表
devices = requests.get(
    "https://gravital-core.example.com/api/v1/devices",
    headers=headers,
    params={"page": 1, "page_size": 20}
).json()

# 创建告警规则
alert_rule = requests.post(
    "https://gravital-core.example.com/api/v1/alert-rules",
    headers=headers,
    json={
        "rule_name": "CPU 告警",
        "condition": "cpu_usage > 80",
        "severity": "warning"
    }
).json()
```

### 14.2 Go 示例
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Code int `json:"code"`
    Data struct {
        Token string `json:"token"`
    } `json:"data"`
}

func main() {
    // 登录
    loginReq := LoginRequest{
        Username: "admin",
        Password: "password",
    }
    
    body, _ := json.Marshal(loginReq)
    resp, _ := http.Post(
        "https://gravital-core.example.com/api/v1/auth/login",
        "application/json",
        bytes.NewBuffer(body),
    )
    
    var loginResp LoginResponse
    json.NewDecoder(resp.Body).Decode(&loginResp)
    token := loginResp.Data.Token
    
    // 使用 Token
    req, _ := http.NewRequest(
        "GET",
        "https://gravital-core.example.com/api/v1/devices",
        nil,
    )
    req.Header.Set("Authorization", "Bearer "+token)
    
    client := &http.Client{}
    resp, _ = client.Do(req)
    // ...
}
```

### 14.3 cURL 示例
```bash
# 登录
TOKEN=$(curl -X POST https://gravital-core.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.data.token')

# 获取设备列表
curl -X GET https://gravital-core.example.com/api/v1/devices \
  -H "Authorization: Bearer $TOKEN"

# 创建设备
curl -X POST https://gravital-core.example.com/api/v1/devices \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Switch-01",
    "device_type": "switch",
    "connection_config": {
      "host": "192.168.1.1",
      "community": "public"
    }
  }'
```

## 15. 附录

### 15.1 时间格式
- **ISO 8601**: `2025-11-01T10:30:00Z`
- **Unix 时间戳**: `1698883200` (秒)

### 15.2 分页参数
- `page`: 页码，从 1 开始
- `page_size`: 每页数量，默认 20，最大 100

### 15.3 排序参数
- `sort`: 排序字段
- `order`: `asc` (升序) 或 `desc` (降序)

### 15.4 过滤语法
支持多种过滤方式：
```
# 精确匹配
?status=online

# 范围查询
?cpu_usage_gt=80&cpu_usage_lt=90

# 模糊搜索
?name_like=switch

# 多值查询
?device_type=switch,router

# 日期范围
?created_at_gte=2025-11-01&created_at_lte=2025-11-30
```

