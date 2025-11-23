# 告警概览无数据排查指南

## 问题现象

告警概览页面显示"暂无活跃告警"，但实际上数据库中有告警事件。

## 原因分析

告警概览只显示**活跃告警**，即状态为以下两种的告警：
- `firing` - 告警中
- `acknowledged` - 已确认

**不会显示的告警**：
- `resolved` - 已解决的告警
- 其他状态的告警

## 排查步骤

### 1. 检查告警事件状态

```bash
# 连接数据库
psql -h localhost -U postgres -d celestial

# 执行以下 SQL
```

```sql
-- 查看告警状态分布
SELECT 
    status,
    COUNT(*) as count
FROM alert_events
GROUP BY status;
```

**预期结果**：
```
   status    | count
-------------+-------
 firing      |    15   ← 这些会显示在概览中
 acknowledged|     5   ← 这些会显示在概览中
 resolved    |    80   ← 这些不会显示
```

**如果所有告警都是 `resolved` 状态**，说明：
- 告警已经被解决了
- 或者告警引擎自动解决了（条件不再满足）

### 2. 检查告警引擎是否在运行

```bash
# 查看服务日志
tail -f logs/app.log | grep -i "alert"
```

**应该看到**：
```
INFO    Starting alert engine...
INFO    Alert engine started
INFO    Evaluating alert rules    count=3
```

**如果没有看到告警引擎日志**：
- 检查服务是否正常启动
- 检查配置文件是否正确

### 3. 检查告警规则是否启用

```sql
-- 查看告警规则状态
SELECT 
    id,
    rule_name,
    enabled,
    condition,
    severity
FROM alert_rules;
```

**确保**：
- `enabled = true` - 规则已启用
- 规则条件正确

### 4. 检查是否有新的告警触发

```sql
-- 查看最近 5 分钟的告警
SELECT 
    id,
    rule_id,
    device_id,
    status,
    triggered_at
FROM alert_events
WHERE triggered_at > NOW() - INTERVAL '5 minutes'
ORDER BY triggered_at DESC;
```

### 5. 手动触发告警测试

**方法 1：模拟设备离线**

```sql
-- 将设备状态设置为 offline
UPDATE devices 
SET status = 'offline' 
WHERE device_id = 'dev-25422c94';
```

**等待 30 秒**（告警引擎评估周期），然后检查：

```sql
-- 查看是否生成新告警
SELECT * FROM alert_events 
WHERE device_id = 'dev-25422c94' 
  AND status = 'firing'
ORDER BY triggered_at DESC 
LIMIT 1;
```

**方法 2：创建测试告警规则**

```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "rule_name": "测试告警",
    "severity": "warning",
    "condition": "device_status == 0",
    "duration": 60,
    "enabled": true,
    "filters": {},
    "notification_config": {}
  }'
```

## 常见问题及解决方案

### 问题 1：所有告警都是 `resolved` 状态

**原因**：
- 告警条件不再满足，被自动解决
- 设备已恢复正常

**解决方案**：
1. 检查设备实际状态是否正常
2. 如果设备确实有问题，检查告警规则条件是否正确
3. 查看告警引擎日志，确认评估逻辑

### 问题 2：告警引擎没有运行

**检查方法**：
```bash
# 查看进程
ps aux | grep server

# 查看日志
tail -f logs/app.log
```

**解决方案**：
```bash
# 重启服务
cd gravital-core
./server -c config/config.yaml
```

### 问题 3：告警规则未启用

**检查**：
```sql
SELECT rule_name, enabled FROM alert_rules;
```

**解决方案**：
```sql
-- 启用规则
UPDATE alert_rules SET enabled = true WHERE id = 1;
```

或通过 API：
```bash
curl -X POST http://localhost:8080/api/v1/alert-rules/1/toggle \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'
```

### 问题 4：告警事件没有关联规则

**检查**：
```sql
SELECT 
    ae.id,
    ae.rule_id,
    ae.status,
    ar.rule_name
FROM alert_events ae
LEFT JOIN alert_rules ar ON ae.rule_id = ar.id
WHERE ae.status IN ('firing', 'acknowledged')
  AND ar.id IS NULL;
```

**如果有记录**，说明告警事件的 `rule_id` 对应的规则已被删除。

**解决方案**：
```sql
-- 清理无效的告警事件
UPDATE alert_events 
SET status = 'resolved', resolved_at = NOW()
WHERE rule_id NOT IN (SELECT id FROM alert_rules);
```

## 验证修复

### 1. 创建测试告警

```bash
# 1. 创建告警规则
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "rule_name": "设备离线测试",
    "severity": "critical",
    "condition": "device_status == 0",
    "duration": 60,
    "enabled": true,
    "filters": {},
    "notification_config": {}
  }'

# 2. 模拟设备离线
psql -d celestial -c "UPDATE devices SET status = 'offline' WHERE device_id = 'dev-25422c94';"

# 3. 等待 30 秒（告警引擎评估周期）
sleep 30

# 4. 检查告警概览
curl http://localhost:8080/api/v1/alert-aggregations \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期响应**：
```json
{
  "code": 0,
  "data": [
    {
      "rule_id": 1,
      "rule_name": "设备离线测试",
      "severity": "critical",
      "total_count": 1,
      "firing_count": 1,
      "acked_count": 0,
      "devices": [
        {
          "device_id": "dev-25422c94",
          "status": "firing"
        }
      ]
    }
  ]
}
```

### 2. 前端验证

1. 打开浏览器访问告警页面
2. 点击"告警概览"标签
3. 应该看到聚合卡片显示

**如果仍然没有数据**：
- 打开浏览器开发者工具（F12）
- 查看 Network 标签
- 检查 `/api/v1/alert-aggregations` 请求
- 查看响应内容

## 快速诊断脚本

创建一个诊断脚本：

```bash
#!/bin/bash
# diagnose_alerts.sh

echo "=== 告警系统诊断 ==="
echo ""

echo "1. 检查告警引擎状态..."
ps aux | grep -i "server" | grep -v grep
echo ""

echo "2. 检查告警规则..."
psql -d celestial -c "SELECT id, rule_name, enabled FROM alert_rules;"
echo ""

echo "3. 检查告警事件状态分布..."
psql -d celestial -c "SELECT status, COUNT(*) FROM alert_events GROUP BY status;"
echo ""

echo "4. 检查活跃告警数量..."
psql -d celestial -c "SELECT COUNT(*) as active_alerts FROM alert_events WHERE status IN ('firing', 'acknowledged');"
echo ""

echo "5. 查看最近的告警..."
psql -d celestial -c "SELECT id, rule_id, device_id, status, triggered_at FROM alert_events ORDER BY triggered_at DESC LIMIT 5;"
echo ""

echo "6. 测试 API 接口..."
curl -s http://localhost:8080/api/v1/alert-aggregations \
  -H "Authorization: Bearer YOUR_TOKEN" | jq '.'
echo ""

echo "=== 诊断完成 ==="
```

使用方法：
```bash
chmod +x diagnose_alerts.sh
./diagnose_alerts.sh
```

## 总结

告警概览显示数据的**必要条件**：

1. ✅ 告警引擎正在运行
2. ✅ 至少有一个告警规则启用
3. ✅ 存在状态为 `firing` 或 `acknowledged` 的告警事件
4. ✅ 告警事件正确关联到告警规则
5. ✅ 前端能正常访问 API 接口

**最常见的原因**：
- 所有告警都已经被解决（status = 'resolved'）
- 告警引擎还没有评估到满足条件的情况
- 告警规则条件设置不正确

**快速解决**：
1. 手动模拟一个告警场景（如设备离线）
2. 等待 30 秒让告警引擎评估
3. 刷新告警概览页面

