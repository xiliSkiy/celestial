# Ping 插件

ICMP Ping 连通性检测插件，用于检测目标主机的网络可达性和延迟。

## 功能特性

- ✅ ICMP Ping 连通性检测
- ✅ RTT（往返时延）测量
- ✅ 丢包率统计
- ✅ 跨平台支持（Linux、macOS、Windows）

## 配置说明

### 设备字段

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|-----|------|------|--------|------|
| host | string | 是 | - | 目标主机 IP 地址或域名 |
| count | int | 否 | 4 | Ping 次数 |
| timeout | int | 否 | 5 | 超时时间（秒） |

### 插件配置

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|-----|------|------|--------|------|
| interval | int | 否 | 60 | 采集间隔（秒） |

## 采集指标

| 指标名 | 类型 | 单位 | 说明 |
|--------|------|------|------|
| ping_reachable | gauge | - | 是否可达 (1=可达, 0=不可达) |
| ping_rtt_ms | gauge | milliseconds | 往返时延（毫秒） |
| ping_packet_loss | gauge | percent | 丢包率 |

## 使用示例

### 配置示例

```yaml
device:
  host: "192.168.1.1"
  count: 4
  timeout: 5

plugin_config:
  interval: 60
```

### 采集结果示例

```json
[
  {
    "name": "ping_reachable",
    "value": 1,
    "timestamp": 1698883200,
    "labels": {
      "device_id": "device-001",
      "host": "192.168.1.1"
    },
    "type": "gauge"
  },
  {
    "name": "ping_rtt_ms",
    "value": 12.5,
    "timestamp": 1698883200,
    "labels": {
      "device_id": "device-001",
      "host": "192.168.1.1"
    },
    "type": "gauge"
  },
  {
    "name": "ping_packet_loss",
    "value": 0,
    "timestamp": 1698883200,
    "labels": {
      "device_id": "device-001",
      "host": "192.168.1.1"
    },
    "type": "gauge"
  }
]
```

## 注意事项

1. **权限要求**: 在某些系统上，ping 命令可能需要特殊权限
2. **防火墙**: 确保目标主机允许 ICMP 流量
3. **超时设置**: 根据网络环境合理设置超时时间
4. **采集频率**: 不建议设置过高的采集频率，避免对网络造成压力

## 故障排查

### 问题：Ping 一直失败

**可能原因**:
- 目标主机不可达
- 防火墙阻止 ICMP 流量
- 网络配置错误

**解决方案**:
1. 手动执行 `ping <host>` 验证连通性
2. 检查防火墙规则
3. 检查网络配置

### 问题：RTT 值异常高

**可能原因**:
- 网络拥塞
- 目标主机负载过高
- 路由路径问题

**解决方案**:
1. 使用 `traceroute` 检查路由路径
2. 检查网络带宽使用情况
3. 检查目标主机负载

## 开发说明

### 构建插件

```bash
cd plugins/ping
go build -o ping.so -buildmode=plugin ping.go
```

### 测试插件

```bash
go test -v
```

## 许可证

Apache 2.0

