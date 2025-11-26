# LLDP 插件实现说明

## 概述

LLDP 插件已完整实现 SNMP 和 SSH 两种采集方式，支持多种网络设备厂商。

## 实现功能

### 1. SNMP 采集

#### 支持的 SNMP 版本
- SNMP v1
- SNMP v2c
- SNMP v3（支持 noAuthNoPriv、authNoPriv、authPriv）

#### 认证配置

**SNMP v1/v2c**:
```yaml
device_config:
  host: 192.168.1.1
  port: 161
  protocol: snmp
  snmp_version: 2c
  auth:
    type: snmp_v2c
    config:
      community: public
```

**SNMP v3**:
```yaml
device_config:
  host: 192.168.1.1
  port: 161
  protocol: snmp
  snmp_version: 3
  auth:
    type: snmp_v3
    config:
      username: admin
      security_level: authPriv
      auth_protocol: SHA
      auth_password: authpass123
      priv_protocol: AES
      priv_password: privpass123
```

#### LLDP MIB 查询

插件查询以下 LLDP MIB OIDs：
- `1.0.8802.1.1.2.1.3.7.1.3` - 本地接口名称
- `1.0.8802.1.1.2.1.4.1.1.4` - 邻居 Chassis ID
- `1.0.8802.1.1.2.1.4.1.1.6` - 邻居 Port ID
- `1.0.8802.1.1.2.1.4.1.1.9` - 邻居系统名称
- `1.0.8802.1.1.2.1.4.1.1.10` - 邻居系统描述
- `1.0.8802.1.1.2.1.4.1.1.8` - 邻居端口描述
- `1.0.8802.1.1.2.1.4.2.1.5` - 邻居管理地址

### 2. SSH 采集

#### 支持的认证方式
- 密码认证
- 密钥认证
- 混合认证（密码或密钥）

#### 认证配置

**密码认证**:
```yaml
device_config:
  host: 192.168.1.1
  port: 22
  protocol: ssh
  auth:
    type: ssh
    config:
      username: admin
      password: password123
      auth_method: password
```

**密钥认证**:
```yaml
device_config:
  host: 192.168.1.1
  port: 22
  protocol: ssh
  auth:
    type: ssh
    config:
      username: admin
      private_key: |
        -----BEGIN RSA PRIVATE KEY-----
        ...
        -----END RSA PRIVATE KEY-----
      passphrase: keypass123
      auth_method: key
```

#### 支持的命令

根据设备类型自动选择相应的 LLDP 命令：

| 设备类型 | 命令 |
|---------|------|
| Cisco IOS/IOS-XE | `show lldp neighbors detail` |
| Juniper JunOS | `show lldp neighbors` |
| Huawei VRP | `display lldp neighbor-information` |
| H3C Comware | `display lldp neighbor-information` |
| Arista EOS | `show lldp neighbors detail` |
| 默认 | `show lldp neighbors detail` |

### 3. 输出解析

插件实现了多厂商设备的输出解析：

#### Cisco 格式
```
Device ID           Local Intf     Hold-time  Capability      Port ID
Switch-02           Gi0/1         120        B               Gi0/2
```

#### Juniper 格式
```
Local Interface    Parent Interface    Chassis Id          Port info          System Name
ge-0/0/0.0         -                  00:11:22:33:44:55   ge-0/0/1.0         Switch-02
```

#### Huawei/H3C 格式
```
Local Interface: GigabitEthernet0/0/1
Chassis ID: 00-11-22-33-44-55
Port ID: GigabitEthernet0/0/2
System Name: Switch-02
```

### 4. 连接测试

插件实现了连接测试功能，用于验证设备配置是否正确：

- `testSNMPConnection()` - 测试 SNMP 连接
- `testSSHConnection()` - 测试 SSH 连接

## 数据格式

### Metrics 输出

LLDP 插件将邻居信息编码为 metrics：

```go
metric := &plugin.Metric{
    Name: "lldp_neighbor",
    Value: float64(i + 1), // 序号
    Timestamp: time.Now().Unix(),
    Labels: map[string]string{
        "device_id":            task.DeviceID,
        "local_interface":      neighbor.LocalInterface,
        "neighbor_chassis_id":  neighbor.NeighborChassisID,
        "neighbor_port_id":     neighbor.NeighborPortID,
        "neighbor_system_name": neighbor.NeighborSystemName,
        "neighbor_system_desc": neighbor.NeighborSystemDesc,
        "neighbor_port_desc":   neighbor.NeighborPortDesc,
        "neighbor_mgmt_addr":   neighbor.NeighborMgmtAddr,
        "ttl":                  strconv.Itoa(neighbor.TTL),
    },
    Type: plugin.MetricTypeGauge,
}
```

## 依赖库

### 已添加的依赖

1. **github.com/gosnmp/gosnmp** - SNMP 客户端库
2. **golang.org/x/crypto/ssh** - SSH 客户端库

### 安装依赖

```bash
go get github.com/gosnmp/gosnmp
go get golang.org/x/crypto/ssh
```

## 使用示例

### 1. SNMP v2c 采集

```yaml
tasks:
  - task_id: task-lldp-001
    device_id: dev-001
    plugin_name: lldp
    device_config:
      host: 192.168.1.1
      port: 161
      protocol: snmp
      snmp_version: 2c
      auth:
        type: snmp_v2c
        config:
          community: public
    interval: 300
```

### 2. SNMP v3 采集

```yaml
tasks:
  - task_id: task-lldp-002
    device_id: dev-002
    plugin_name: lldp
    device_config:
      host: 192.168.1.2
      port: 161
      protocol: snmp
      snmp_version: 3
      auth:
        type: snmp_v3
        config:
          username: admin
          security_level: authPriv
          auth_protocol: SHA
          auth_password: authpass123
          priv_protocol: AES
          priv_password: privpass123
    interval: 300
```

### 3. SSH 采集

```yaml
tasks:
  - task_id: task-lldp-003
    device_id: dev-003
    plugin_name: lldp
    device_config:
      host: 192.168.1.3
      port: 22
      protocol: ssh
      device_type: cisco
      auth:
        type: ssh
        config:
          username: admin
          password: password123
          auth_method: password
    interval: 300
```

## 错误处理

### SNMP 错误

- 连接失败：返回连接错误
- 认证失败：返回认证错误
- MIB 查询失败：记录警告，继续处理其他邻居

### SSH 错误

- 连接失败：返回连接错误
- 认证失败：返回认证错误
- 命令执行失败：返回执行错误
- 输出解析失败：记录警告，返回已解析的邻居

## 性能优化

1. **批量查询**: SNMP 使用 BulkWalk 批量查询邻居表
2. **超时控制**: 设置合理的超时时间（默认 10 秒）
3. **连接复用**: SSH 连接在采集完成后立即关闭
4. **错误恢复**: 单个邻居解析失败不影响其他邻居

## 限制和注意事项

1. **SNMP v3 配置**: 需要完整的认证配置，测试连接功能暂不支持
2. **主机密钥验证**: SSH 连接使用 `InsecureIgnoreHostKey`，生产环境应配置主机密钥验证
3. **设备类型检测**: 需要正确设置 `device_type` 以使用正确的命令和解析器
4. **输出格式**: 不同厂商和版本的设备输出格式可能略有差异，解析器可能需要调整

## 测试建议

### 单元测试

- 测试 SNMP 连接和查询
- 测试 SSH 连接和命令执行
- 测试各种输出格式的解析
- 测试错误处理

### 集成测试

- 测试完整的采集流程
- 测试不同厂商设备
- 测试不同 SNMP 版本
- 测试不同 SSH 认证方式

### 端到端测试

- 配置采集任务
- 验证数据上报
- 验证数据存储
- 验证拓扑自动发现

## 后续改进

1. **更多设备支持**: 添加更多厂商设备的命令和解析器
2. **主机密钥验证**: 实现 SSH 主机密钥验证
3. **连接池**: 实现 SSH 连接池以提高性能
4. **缓存机制**: 实现 LLDP 数据缓存，减少重复采集
5. **指标扩展**: 添加更多 LLDP 相关指标（如邻居数量、接口状态等）

