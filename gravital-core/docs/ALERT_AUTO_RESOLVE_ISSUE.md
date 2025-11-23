# 告警自动解决问题排查

## 问题现象

设备设置了告警规则，一直在产生告警，但告警立即被自动解决，无法在告警概览中看到。

## 根本原因

告警规则的**条件设置反了**！

## device_status 指标说明

告警引擎从数据库查询设备状态，转换规则：

```go
if device.Status == "online" {
    value = 1.0  // 在线
} else {
    value = 0.0  // 离线
}
```

**值的含义**：
- `device_status = 1.0` → 设备**在线**
- `device_status = 0.0` → 设备**离线**

## 常见错误

### ❌ 错误的规则

```json
{
  "rule_name": "设备离线告警",
  "condition": "device_status != 0",  // ❌ 错误！
  "severity": "critical"
}
```

**问题分析**：
- 设备离线时：`device_status = 0`
- 条件检查：`0 != 0` → **false**（不满足）
- 结果：不会触发告警 ❌

- 设备在线时：`device_status = 1`
- 条件检查：`1 != 0` → **true**（满足）
- 结果：设备在线时反而告警 ❌❌

### ✅ 正确的规则

```json
{
  "rule_name": "设备离线告警",
  "condition": "device_status == 0",  // ✅ 正确！
  "severity": "critical"
}
```

**正确逻辑**：
- 设备离线时：`device_status = 0`
- 条件检查：`0 == 0` → **true**（满足）
- 结果：触发告警 ✅

- 设备在线时：`device_status = 1`
- 条件检查：`1 == 0` → **false**（不满足）
- 结果：不告警或自动解决 ✅

## 告警规则对照表

| 告警场景 | 正确条件 | 错误条件 | 说明 |
|---------|---------|---------|------|
| 设备离线告警 | `device_status == 0` | `device_status != 0` | 离线时值为 0 |
| 设备在线告警 | `device_status == 1` | `device_status != 1` | 在线时值为 1 |
| 设备离线告警（另一种写法）| `device_status < 1` | `device_status > 0` | 小于 1 表示离线 |
| 设备离线告警（另一种写法）| `device_status != 1` | `device_status == 1` | 不等于 1 表示离线 |

## 检查你的告警规则

### 1. 查看当前规则

```sql
SELECT 
    id,
    rule_name,
    condition,
    enabled
FROM alert_rules;
```

### 2. 分析规则逻辑

假设结果是：
```
 id |    rule_name     |     condition      | enabled
----+------------------+--------------------+---------
  1 | 设备离线告警      | device_status != 0 | t
```

**这个规则是错误的！** 应该改为 `device_status == 0`

### 3. 修复规则

**方法 1：通过 SQL 修复**

```sql
-- 修复设备离线告警规则
UPDATE alert_rules 
SET condition = 'device_status == 0'
WHERE rule_name = '设备离线告警' 
  AND condition = 'device_status != 0';
```

**方法 2：通过 API 修复**

```bash
curl -X PUT http://localhost:8080/api/v1/alert-rules/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "rule_name": "设备离线告警",
    "severity": "critical",
    "condition": "device_status == 0",
    "duration": 300,
    "enabled": true,
    "filters": {},
    "notification_config": {}
  }'
```

**方法 3：通过前端修复**

1. 打开告警规则页面
2. 点击规则的"编辑"按钮
3. 修改条件：
   - 指标名称：`device_status`
   - 运算符：`==`
   - 阈值：`0`
4. 保存

## 验证修复

### 1. 模拟设备离线

```sql
-- 将一个设备设置为离线
UPDATE devices 
SET status = 'offline' 
WHERE device_id = 'dev-25422c94';
```

### 2. 等待告警引擎评估

```bash
# 等待 30 秒（告警引擎评估周期）
sleep 30
```

### 3. 检查告警事件

```sql
-- 查看最新的告警
SELECT 
    id,
    rule_id,
    device_id,
    status,
    message,
    triggered_at
FROM alert_events
WHERE device_id = 'dev-25422c94'
ORDER BY triggered_at DESC
LIMIT 1;
```

**预期结果**：
```
 id | rule_id |   device_id    | status |           message            |     triggered_at
----+---------+----------------+--------+------------------------------+----------------------
  1 |       1 | dev-25422c94   | firing | 设备离线告警: 当前值 0.00... | 2025-11-23 10:30:00
```

状态应该是 `firing`，不应该立即变成 `resolved`。

### 4. 查看告警概览

```bash
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
      "rule_name": "设备离线告警",
      "severity": "critical",
      "firing_count": 1,
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

### 5. 恢复设备在线

```sql
-- 将设备恢复在线
UPDATE devices 
SET status = 'online' 
WHERE device_id = 'dev-25422c94';

-- 等待 30 秒

-- 检查告警是否自动解决
SELECT status FROM alert_events WHERE device_id = 'dev-25422c94' ORDER BY id DESC LIMIT 1;
```

**预期结果**：状态应该变为 `resolved`

## 其他可能的原因

### 1. 设备状态频繁变化

如果设备状态在 `online` 和 `offline` 之间频繁切换：

```
10:00:00 - 设备离线 → 触发告警 (firing)
10:00:15 - 设备在线 → 自动解决 (resolved)
10:00:30 - 设备离线 → 再次触发告警 (firing)
10:00:45 - 设备在线 → 再次自动解决 (resolved)
```

**解决方案**：增加 `duration`（持续时间）

```json
{
  "condition": "device_status == 0",
  "duration": 300  // 持续 5 分钟才告警
}
```

**注意**：当前版本的告警引擎**尚未实现** duration 检查，这是一个待实现的功能。

### 2. 设备监控服务频繁更新状态

检查设备监控服务的日志：

```bash
tail -f logs/app.log | grep -i "device monitor"
```

如果看到频繁的状态更新，可能需要调整设备监控的检查间隔。

### 3. 多个告警引擎实例

如果运行了多个服务实例，可能会互相干扰：

```bash
# 检查是否有多个实例
ps aux | grep server | grep -v grep
```

**解决方案**：确保只运行一个实例。

## 最佳实践

### 1. 告警规则命名规范

```
✅ 设备离线告警 (condition: device_status == 0)
✅ CPU 使用率过高 (condition: cpu_usage > 80)
✅ 内存使用率告警 (condition: memory_usage > 90)
```

### 2. 条件设置原则

- **明确告警触发条件**：什么情况下需要告警？
- **使用正向逻辑**：直接表达触发条件，不要用否定
- **添加描述**：在 description 中说明告警含义

### 3. 测试流程

创建新规则后：
1. 先禁用规则
2. 手动模拟告警场景
3. 启用规则并观察
4. 确认告警正常触发
5. 确认告警正常解决

## 快速诊断脚本

```bash
#!/bin/bash
# diagnose_auto_resolve.sh

echo "=== 告警自动解决诊断 ==="
echo ""

echo "1. 检查告警规则条件..."
psql -d celestial -c "
SELECT 
    id,
    rule_name,
    condition,
    CASE 
        WHEN condition LIKE '%!= 0%' THEN '❌ 可能有问题：离线告警应该用 == 0'
        WHEN condition LIKE '%== 0%' THEN '✅ 正确：离线告警'
        WHEN condition LIKE '%== 1%' THEN '✅ 正确：在线告警'
        ELSE '⚠️  需要人工检查'
    END as status
FROM alert_rules
WHERE enabled = true;
"
echo ""

echo "2. 检查最近的告警事件..."
psql -d celestial -c "
SELECT 
    id,
    device_id,
    status,
    triggered_at,
    resolved_at,
    CASE 
        WHEN resolved_at IS NOT NULL 
             AND resolved_at - triggered_at < INTERVAL '1 minute'
        THEN '⚠️  快速解决（可能是条件错误）'
        ELSE '✅ 正常'
    END as analysis
FROM alert_events
ORDER BY triggered_at DESC
LIMIT 5;
"
echo ""

echo "3. 检查设备状态..."
psql -d celestial -c "
SELECT 
    device_id,
    status,
    CASE 
        WHEN status = 'online' THEN 'device_status = 1'
        ELSE 'device_status = 0'
    END as metric_value
FROM devices
LIMIT 5;
"
echo ""

echo "=== 诊断完成 ==="
```

## 总结

**告警自动解决的最常见原因**：

1. ❌ **告警条件设置反了**（90% 的情况）
   - 错误：`device_status != 0`
   - 正确：`device_status == 0`

2. ⚠️  设备状态频繁变化
   - 需要增加 duration 或调整监控间隔

3. ⚠️  多个服务实例冲突
   - 确保只运行一个实例

**修复步骤**：
1. 检查告警规则条件
2. 修正条件设置
3. 测试验证
4. 观察告警概览

执行诊断脚本快速定位问题：
```bash
chmod +x diagnose_auto_resolve.sh
./diagnose_auto_resolve.sh
```

