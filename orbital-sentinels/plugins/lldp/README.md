# LLDP 插件

## 概述

LLDP (Link Layer Discovery Protocol) 邻居发现插件用于采集网络设备的 LLDP 邻居信息，用于网络拓扑自动发现。

## 功能特性

- 支持 SNMP 协议采集
- 支持 SSH 协议采集
- 自动解析 LLDP 邻居表
- 上报邻居信息到中心端

## 配置说明

### 设备配置字段

| 字段名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| host | string | 是 | - | 设备 IP 地址或主机名 |
| protocol | string | 否 | snmp | 采集协议 (snmp/ssh) |
| snmp_community | string | 否 | public | SNMP Community |
| snmp_version | string | 否 | 2c | SNMP 版本 (1/2c/3) |
| ssh_username | string | 否 | - | SSH 用户名 |
| ssh_password | password | 否 | - | SSH 密码 |

### 配置示例

#### SNMP 方式

```yaml
device_config:
  host: 192.168.1.1
  protocol: snmp
  snmp_community: public
  snmp_version: 2c
```

#### SSH 方式

```yaml
device_config:
  host: 192.168.1.1
  protocol: ssh
  ssh_username: admin
  ssh_password: password123
```

## 采集的数据

插件会采集以下 LLDP 邻居信息：

- 本地接口 (Local Interface)
- 邻居 Chassis ID
- 邻居端口 ID
- 邻居系统名称
- 邻居系统描述
- 邻居端口描述
- 邻居管理地址
- TTL

## 数据上报

采集到的 LLDP 邻居信息会通过专门的 API 上报到中心端：

```
POST /api/v1/topology/lldp
```

数据格式：

```json
{
  "device_id": "dev-001",
  "neighbors": [
    {
      "local_interface": "GigabitEthernet0/1",
      "neighbor_chassis_id": "00:11:22:33:44:55",
      "neighbor_port_id": "GigabitEthernet0/2",
      "neighbor_system_name": "Switch-02",
      "neighbor_system_desc": "Cisco IOS Software",
      "neighbor_port_desc": "GigabitEthernet0/2",
      "neighbor_mgmt_addr": "192.168.1.2",
      "ttl": 120
    }
  ]
}
```

## 使用场景

1. **网络拓扑自动发现**: 通过 LLDP 协议自动发现网络设备之间的连接关系
2. **网络监控**: 监控网络设备的邻居关系变化
3. **故障排查**: 快速定位网络连接问题

## 注意事项

1. 设备必须启用 LLDP 协议
2. SNMP 方式需要设备支持 LLDP MIB (1.0.8802.1.1.2)
3. SSH 方式需要设备支持 LLDP 相关命令（如 "show lldp neighbors"）
4. 不同厂商的设备输出格式可能不同，需要适配

## 开发状态

- ✅ 插件框架已实现
- ⏳ SNMP 采集逻辑待实现
- ⏳ SSH 采集逻辑待实现
- ✅ 数据上报逻辑已实现

## 后续计划

1. 实现 SNMP 采集逻辑
2. 实现 SSH 采集逻辑
3. 支持更多厂商的设备
4. 添加缓存机制，减少重复采集

