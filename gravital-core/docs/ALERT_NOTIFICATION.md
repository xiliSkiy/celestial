# 告警通知功能实现文档

## 📋 概述

本文档详细说明告警通知功能的实现，包括多种通知渠道、通知去重、通知升级等功能。

---

## 🎯 功能特性

### 已实现功能

✅ **多种通知渠道**
- 邮件通知（Email）
- Webhook 通知
- 钉钉通知（DingTalk）
- 企业微信通知（WeChat Work）

✅ **智能去重机制**
- 相同告警在指定时间内只通知一次
- 可配置去重间隔（默认 5 分钟）

✅ **通知升级机制**
- 告警持续一定时间后升级通知
- 可配置升级时间和升级渠道

✅ **异步发送**
- 通知异步发送，不阻塞告警评估
- 支持批量发送

✅ **通知记录**
- 记录所有通知发送历史
- 支持查询通知状态

---

## 🏗️ 架构设计

### 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Alert Engine                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐                                           │
│  │ evaluateRule │                                           │
│  └──────┬───────┘                                           │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────┐                                           │
│  │ triggerAlert │                                           │
│  └──────┬───────┘                                           │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           Notification Service                       │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │                                                      │  │
│  │  1. 去重检查 (ShouldNotify)                          │  │
│  │  2. 升级检查 (shouldEscalate)                        │  │
│  │  3. 格式化内容 (formatAlertContent)                  │  │
│  │  4. 批量发送 (SendBatch)                             │  │
│  │  5. 记录结果 (RecordNotification)                    │  │
│  │                                                      │  │
│  └──────┬───────────────────────────────────────────────┘  │
│         │                                                   │
│         ├──────────┬──────────┬──────────┬─────────────┐   │
│         ▼          ▼          ▼          ▼             ▼   │
│  ┌──────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ │
│  │  Email   │ │Webhook │ │DingTalk│ │ WeChat │ │  SMS   │ │
│  │  Sender  │ │ Sender │ │ Sender │ │ Sender │ │ Sender │ │
│  └──────────┘ └────────┘ └────────┘ └────────┘ └────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

```
1. 告警触发
   ├─> Alert Engine 创建 AlertEvent
   └─> 调用 NotificationService.SendAlert()

2. 去重检查
   ├─> 查询缓存：是否在去重间隔内
   ├─> 是 ──> 跳过通知
   └─> 否 ──> 继续

3. 升级检查
   ├─> 查询缓存：首次通知时间
   ├─> 计算时间差
   ├─> 超过升级时间 ──> 使用升级渠道
   └─> 未超过 ──> 使用普通渠道

4. 准备通知
   ├─> 格式化主题和内容
   ├─> 遍历所有启用的渠道
   └─> 为每个接收人创建 Notification

5. 批量发送
   ├─> 并发发送到所有渠道
   ├─> 每个渠道独立处理
   └─> 返回发送结果

6. 记录结果
   ├─> 保存到 alert_notifications 表
   ├─> 更新去重缓存
   └─> 更新升级缓存
```

---

## 📦 核心组件

### 1. 通知服务 (NotificationService)

**文件**: `internal/notification/service.go`

**核心方法**:

```go
type Service interface {
    // 发送单个通知
    Send(ctx context.Context, notification *Notification) (*NotificationResult, error)
    
    // 批量发送通知
    SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error)
    
    // 发送告警通知（带去重和升级）
    SendAlert(ctx context.Context, event *model.AlertEvent, config *NotificationConfig) error
    
    // 注册通知渠道
    RegisterChannel(channel Channel, sender Sender) error
    
    // 去重检查
    ShouldNotify(ctx context.Context, alertID string, ruleID uint) (bool, error)
}
```

### 2. 通知渠道 (Sender)

**接口定义**:

```go
type Sender interface {
    // 发送通知
    Send(ctx context.Context, notification *Notification) error
    
    // 获取发送器名称
    Name() string
    
    // 验证配置
    Validate() error
}
```

**已实现的渠道**:

1. **EmailSender** - 邮件通知
   - 文件: `internal/notification/email.go`
   - 支持 SMTP/TLS
   - HTML 格式邮件
   - 优先级颜色标识

2. **WebhookSender** - Webhook 通知
   - 文件: `internal/notification/webhook.go`
   - 支持 GET/POST/PUT
   - 自定义请求头
   - 超时控制

3. **DingTalkSender** - 钉钉通知
   - 文件: `internal/notification/dingtalk.go`
   - Markdown 格式
   - 签名验证
   - @ 功能

4. **WeChatSender** - 企业微信通知
   - 文件: `internal/notification/wechat.go`
   - Markdown 格式
   - @ 功能

### 3. 通知配置 (NotificationConfig)

```go
type NotificationConfig struct {
    Enabled           bool             // 是否启用通知
    Channels          []ChannelConfig  // 通知渠道配置
    DedupeInterval    int              // 去重间隔（秒）
    EscalationEnabled bool             // 是否启用升级
    EscalationAfter   int              // 升级时间（秒）
    EscalationChannels []Channel       // 升级通知渠道
}

type ChannelConfig struct {
    Channel    Channel  // 渠道类型
    Enabled    bool     // 是否启用
    Recipients []string // 接收人列表
    Template   string   // 消息模板（可选）
    Config     map[string]interface{} // 渠道特定配置
}
```

---

## 🔧 配置说明

### 1. 告警规则通知配置

在创建告警规则时，可以配置通知选项：

```json
{
  "rule_name": "设备离线告警",
  "severity": "critical",
  "condition": "device_status != 1",
  "notification_config": {
    "enabled": true,
    "dedupe_interval": 300,
    "escalation_enabled": true,
    "escalation_after": 1800,
    "channels": [
      {
        "channel": "email",
        "enabled": true,
        "recipients": ["admin@example.com", "ops@example.com"]
      },
      {
        "channel": "dingtalk",
        "enabled": true,
        "recipients": ["webhook_url_1"]
      }
    ],
    "escalation_channels": ["email", "dingtalk", "wechat"]
  }
}
```

### 2. 邮件配置

**配置文件**: `config/config.yaml`

```yaml
notification:
  email:
    smtp_host: smtp.example.com
    smtp_port: 587
    smtp_user: noreply@example.com
    smtp_password: your_password
    from: Celestial Alert <noreply@example.com>
    use_tls: true
```

### 3. 钉钉配置

```yaml
notification:
  dingtalk:
    webhook_url: https://oapi.dingtalk.com/robot/send?access_token=xxx
    secret: SEC_xxx  # 可选，用于签名验证
    at_mobiles: []   # @ 的手机号列表
    at_all: false    # 是否 @ 所有人
```

### 4. 企业微信配置

```yaml
notification:
  wechat:
    webhook_url: https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx
    mentioned_list: []        # @ 的用户 ID 列表
    mentioned_mobile_list: [] # @ 的手机号列表
```

### 5. Webhook 配置

```yaml
notification:
  webhook:
    url: https://your-webhook-endpoint.com/alerts
    method: POST
    headers:
      Authorization: Bearer your_token
      X-Custom-Header: value
    timeout: 30
```

---

## 🚀 使用示例

### 1. 初始化通知服务

```go
// 创建通知服务
notificationSvc := notification.NewService(db, logger)

// 注册邮件渠道
emailConfig := &notification.EmailConfig{
    SMTPHost:     "smtp.example.com",
    SMTPPort:     587,
    SMTPUser:     "noreply@example.com",
    SMTPPassword: "password",
    From:         "Celestial Alert <noreply@example.com>",
    UseTLS:       true,
}
emailSender := notification.NewEmailSender(emailConfig, logger)
notificationSvc.RegisterChannel(notification.ChannelEmail, emailSender)

// 注册钉钉渠道
dingtalkConfig := &notification.DingTalkConfig{
    WebhookURL: "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    Secret:     "SEC_xxx",
}
dingtalkSender := notification.NewDingTalkSender(dingtalkConfig, logger)
notificationSvc.RegisterChannel(notification.ChannelDingTalk, dingtalkSender)

// 注册企业微信渠道
wechatConfig := &notification.WeChatConfig{
    WebhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
}
wechatSender := notification.NewWeChatSender(wechatConfig, logger)
notificationSvc.RegisterChannel(notification.ChannelWeChat, wechatSender)

// 注册 Webhook 渠道
webhookConfig := &notification.WebhookConfig{
    URL:    "https://your-webhook-endpoint.com/alerts",
    Method: "POST",
    Headers: map[string]string{
        "Authorization": "Bearer your_token",
    },
    Timeout: 30,
}
webhookSender := notification.NewWebhookSender(webhookConfig, logger)
notificationSvc.RegisterChannel(notification.ChannelWebhook, webhookSender)
```

### 2. 创建告警引擎（带通知）

```go
alertEngine := engine.NewAlertEngine(db, logger, &engine.Config{
    VMURL:           vmURL,
    CheckInterval:   30 * time.Second,
    NotificationSvc: notificationSvc,  // 传入通知服务
})
alertEngine.Start()
```

### 3. 创建带通知配置的告警规则

```go
rule := &model.AlertRule{
    RuleName:  "设备离线告警",
    Severity:  "critical",
    Condition: "device_status != 1",
    Enabled:   true,
    NotificationConfig: map[string]interface{}{
        "enabled":            true,
        "dedupe_interval":    300,
        "escalation_enabled": true,
        "escalation_after":   1800,
        "channels": []map[string]interface{}{
            {
                "channel":    "email",
                "enabled":    true,
                "recipients": []string{"admin@example.com"},
            },
            {
                "channel":    "dingtalk",
                "enabled":    true,
                "recipients": []string{"webhook_url"},
            },
        },
        "escalation_channels": []string{"email", "dingtalk", "wechat"},
    },
}
```

---

## 📊 通知内容格式

### 邮件通知

**主题**: `[critical] 设备离线告警: 当前值 0.00 != 阈值 1.00`

**内容** (HTML):
```html
<div style="background-color: #d32f2f; color: white; padding: 10px;">
  <h2>[critical] 设备离线告警</h2>
</div>

<div style="padding: 20px; background-color: #f5f5f5;">
  <pre>
告警详情：
- 告警ID: alert-xxx
- 设备ID: dev-001
- 指标名称: device_status
- 严重级别: critical
- 告警消息: 设备离线告警: 当前值 0.00 != 阈值 1.00
- 触发时间: 2025-11-23 10:00:00
- 当前状态: firing
  </pre>
</div>

<div style="margin-top: 20px;">
  <h4>详细信息：</h4>
  <ul>
    <li><strong>event_id:</strong> 123</li>
    <li><strong>device_id:</strong> dev-001</li>
    <li><strong>metric_name:</strong> device_status</li>
  </ul>
</div>
```

### 钉钉通知

**格式**: Markdown

```markdown
### [critical] 设备离线告警

**优先级**: 🔴 紧急

\`\`\`
告警详情：
- 告警ID: alert-xxx
- 设备ID: dev-001
- 指标名称: device_status
- 严重级别: critical
- 告警消息: 设备离线告警: 当前值 0.00 != 阈值 1.00
- 触发时间: 2025-11-23 10:00:00
- 当前状态: firing
\`\`\`

**详细信息**:
- event_id: 123
- device_id: dev-001
- metric_name: device_status

> 发送时间: 2025-11-23 10:00:00
```

### Webhook 通知

**格式**: JSON

```json
{
  "id": "notif-alert-xxx-email-1732348800",
  "channel": "webhook",
  "recipient": "https://your-endpoint.com/alerts",
  "subject": "[critical] 设备离线告警",
  "content": "告警详情：\n- 告警ID: alert-xxx\n...",
  "priority": "critical",
  "alert_id": "alert-xxx",
  "metadata": {
    "event_id": 123,
    "device_id": "dev-001",
    "metric_name": "device_status",
    "escalated": false
  },
  "created_at": "2025-11-23T10:00:00Z"
}
```

---

## 🔍 通知去重机制

### 工作原理

1. **缓存结构**:
   ```go
   dedupeCache map[string]time.Time  // alertID -> lastNotifyTime
   ```

2. **去重检查**:
   ```go
   func (s *service) ShouldNotify(alertID string, ruleID uint) (bool, error) {
       lastNotifyTime, exists := s.dedupeCache[alertID]
       if !exists {
           return true, nil  // 首次通知
       }
       
       elapsed := time.Since(lastNotifyTime).Seconds()
       return elapsed >= float64(dedupeInterval), nil
   }
   ```

3. **更新缓存**:
   ```go
   func (s *service) updateDedupeCache(alertID string) {
       s.dedupeCache[alertID] = time.Now()
   }
   ```

### 配置示例

```json
{
  "notification_config": {
    "dedupe_interval": 300  // 5 分钟内相同告警只通知一次
  }
}
```

---

## 📈 通知升级机制

### 工作原理

1. **缓存结构**:
   ```go
   escalationCache map[string]time.Time  // alertID -> firstNotifyTime
   ```

2. **升级检查**:
   ```go
   func (s *service) shouldEscalate(alertID string, config *NotificationConfig) bool {
       if !config.EscalationEnabled {
           return false
       }
       
       firstNotifyTime, exists := s.escalationCache[alertID]
       if !exists {
           return false
       }
       
       elapsed := time.Since(firstNotifyTime).Seconds()
       return elapsed >= float64(config.EscalationAfter)
   }
   ```

3. **升级流程**:
   ```
   告警触发
   ├─> 首次通知：使用普通渠道
   ├─> 记录首次通知时间
   ├─> 30 分钟后（可配置）
   ├─> 告警仍未解决
   └─> 升级通知：使用升级渠道（如：短信、电话）
   ```

### 配置示例

```json
{
  "notification_config": {
    "escalation_enabled": true,
    "escalation_after": 1800,  // 30 分钟后升级
    "channels": [
      {"channel": "email", "enabled": true, "recipients": ["ops@example.com"]},
      {"channel": "dingtalk", "enabled": true, "recipients": ["webhook_url"]}
    ],
    "escalation_channels": ["email", "dingtalk", "wechat", "sms"]
  }
}
```

---

## 📝 数据库记录

### alert_notifications 表

所有通知发送记录都会保存到数据库：

```sql
CREATE TABLE alert_notifications (
    id              SERIAL PRIMARY KEY,
    alert_event_id  INTEGER NOT NULL,
    channel         VARCHAR(64),
    recipient       VARCHAR(255),
    status          VARCHAR(32),  -- pending, sending, sent, failed
    sent_at         TIMESTAMP,
    error_message   TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (alert_event_id) REFERENCES alert_events(id) ON DELETE CASCADE
);
```

### 查询通知历史

```sql
-- 查询某个告警事件的所有通知记录
SELECT * FROM alert_notifications 
WHERE alert_event_id = 123 
ORDER BY created_at DESC;

-- 统计通知成功率
SELECT 
    channel,
    COUNT(*) as total,
    SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as success,
    ROUND(100.0 * SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM alert_notifications
GROUP BY channel;
```

---

## 🧪 测试

### 单元测试

```bash
cd gravital-core
go test ./internal/notification/... -v
```

### 集成测试

```bash
# 测试邮件通知
curl -X POST http://localhost:8080/api/v1/notifications/test \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "email",
    "recipient": "test@example.com",
    "subject": "测试通知",
    "content": "这是一条测试通知"
  }'

# 测试钉钉通知
curl -X POST http://localhost:8080/api/v1/notifications/test \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "dingtalk",
    "recipient": "webhook_url",
    "subject": "测试通知",
    "content": "这是一条测试通知"
  }'
```

---

## 🔧 故障排查

### 1. 通知未发送

**检查清单**:
- [ ] 通知配置是否启用 (`enabled: true`)
- [ ] 通知渠道是否注册
- [ ] 接收人列表是否正确
- [ ] 是否被去重机制过滤

**查看日志**:
```bash
docker-compose logs -f gravital-core | grep -i notification
```

### 2. 邮件发送失败

**常见问题**:
- SMTP 服务器地址或端口错误
- 认证信息错误
- TLS 配置错误
- 防火墙阻止

**解决方案**:
```bash
# 测试 SMTP 连接
telnet smtp.example.com 587

# 检查配置
grep -A 10 "email:" config/config.yaml
```

### 3. 钉钉通知失败

**常见问题**:
- Webhook URL 错误
- 签名验证失败
- IP 白名单限制

**解决方案**:
```bash
# 测试 Webhook
curl -X POST "https://oapi.dingtalk.com/robot/send?access_token=xxx" \
  -H "Content-Type: application/json" \
  -d '{"msgtype":"text","text":{"content":"test"}}'
```

---

## 📚 最佳实践

### 1. 通知渠道选择

- **critical 级别**: 邮件 + 钉钉 + 短信
- **warning 级别**: 邮件 + 钉钉
- **info 级别**: 钉钉

### 2. 去重间隔设置

- **频繁波动的指标**: 10-15 分钟
- **稳定的指标**: 5 分钟
- **关键告警**: 1 分钟

### 3. 升级时间设置

- **critical 级别**: 15-30 分钟
- **warning 级别**: 30-60 分钟
- **info 级别**: 不启用升级

### 4. 接收人配置

- 设置多个接收人，避免单点故障
- 使用分组邮箱（如 ops@example.com）
- 定期更新接收人列表

---

## 🚀 未来增强

### 短期（1-2 周）

- [ ] 短信通知渠道
- [ ] Slack 通知渠道
- [ ] Telegram 通知渠道
- [ ] 通知模板自定义

### 中期（1-2 月）

- [ ] 通知静默时段
- [ ] 通知优先级路由
- [ ] 通知统计报表
- [ ] 通知失败重试

### 长期（3-6 月）

- [ ] 智能通知（基于 AI 的通知优化）
- [ ] 通知聚合（多个告警合并为一条通知）
- [ ] 通知确认机制
- [ ] 通知审计日志

---

**文档版本**: v1.0
**最后更新**: 2025-11-23
**维护者**: Celestial Team

