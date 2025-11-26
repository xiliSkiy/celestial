# LLDP 数据采集流程实现说明

## 概述

本文档说明 LLDP (Link Layer Discovery Protocol) 数据采集流程的完整实现，包括采集端插件、数据上报和中心端接收处理。

## 实现架构

```
┌─────────────────────────────────────────────────────────────┐
│  采集端 (Orbital Sentinel)                                   │
│                                                              │
│  ┌──────────────┐                                            │
│  │ LLDP 插件    │                                            │
│  │ (lldp.go)   │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         │ Collect()                                         │
│         │ 返回 metrics (包含 LLDP 邻居信息)                  │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ Scheduler    │                                            │
│  │ - runTask()  │                                            │
│  │ - handleLLDP │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         │ onLLDP(deviceID, neighbors)                        │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ Agent        │                                            │
│  │ - SetLLDP    │                                            │
│  │   Handler    │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         │ SendLLDP()                                        │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ CoreSender   │                                            │
│  │ - SendLLDP() │                                            │
│  └──────┬───────┘                                            │
└─────────┼────────────────────────────────────────────────────┘
          │
          │ POST /api/v1/topology/lldp
          │
┌─────────▼────────────────────────────────────────────────────┐
│  中心端 (Gravital Core)                                      │
│                                                              │
│  ┌──────────────┐                                            │
│  │ Topology     │                                            │
│  │ Handler      │                                            │
│  │ - IngestLLDP │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ Topology     │                                            │
│  │ Service      │                                            │
│  │ - UpsertLLDP │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ Repository   │                                            │
│  │ - UpsertLLDP │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ 数据库       │                                            │
│  │ lldp_neighbors│                                           │
│  └──────────────┘                                            │
└──────────────────────────────────────────────────────────────┘
```

## 实现内容

### 1. LLDP 插件实现

**文件**: `orbital-sentinels/plugins/lldp/lldp.go`

#### 1.1 插件结构

```go
type LLDPPlugin struct {
    sdk.BasePlugin
    schema plugin.PluginSchema
}
```

#### 1.2 核心方法

- **Init()**: 初始化插件，加载配置 Schema
- **Collect()**: 采集 LLDP 邻居信息
  - 支持 SNMP 和 SSH 两种协议
  - 将邻居信息编码到 metrics 的 labels 中
  - 返回包含 LLDP 数据的 metrics

#### 1.3 数据格式

插件通过 metrics 传递 LLDP 数据：

```go
metric := &plugin.Metric{
    Name: "lldp_neighbor",
    Value: float64(i + 1),
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
}
```

### 2. Scheduler 扩展

**文件**: `orbital-sentinels/internal/scheduler/scheduler.go`

#### 2.1 新增字段

```go
type Scheduler struct {
    // ...
    onLLDP func(string, []LLDPNeighbor) // deviceID, neighbors
    // ...
}

type LLDPNeighbor struct {
    LocalInterface     string
    NeighborChassisID  string
    NeighborPortID     string
    NeighborSystemName string
    NeighborSystemDesc string
    NeighborPortDesc   string
    NeighborMgmtAddr   string
    TTL                int
}
```

#### 2.2 新增方法

- **SetLLDPHandler()**: 设置 LLDP 数据处理器
- **handleLLDPData()**: 从 metrics 中提取 LLDP 数据并调用处理器

#### 2.3 处理流程

```go
// 在 runTask 中
if err == nil && s.onLLDP != nil {
    s.handleLLDPData(metrics, st.Task)
}

// handleLLDPData 提取 LLDP 数据
func (s *Scheduler) handleLLDPData(metrics []*plugin.Metric, task *plugin.CollectionTask) {
    neighbors := make([]LLDPNeighbor, 0)
    
    for _, metric := range metrics {
        if metric.Name == "lldp_neighbor" {
            // 从 labels 中提取邻居信息
            neighbor := LLDPNeighbor{...}
            neighbors = append(neighbors, neighbor)
        }
    }
    
    if len(neighbors) > 0 {
        s.onLLDP(task.DeviceID, neighbors)
    }
}
```

### 3. CoreSender 扩展

**文件**: `orbital-sentinels/internal/sender/core_sender.go`

#### 3.1 新增方法

```go
// SendLLDP 发送 LLDP 数据
func (cs *CoreSender) SendLLDP(ctx context.Context, deviceID string, neighbors []LLDPNeighborRequest) error {
    // 构造请求体
    payload := map[string]interface{}{
        "device_id": deviceID,
        "neighbors": neighbors,
    }
    
    // 发送到 /api/v1/topology/lldp
    // ...
}
```

#### 3.2 请求格式

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

### 4. Agent 集成

**文件**: `orbital-sentinels/internal/agent/agent.go`

#### 4.1 设置 LLDP 处理器

```go
// 设置 LLDP 数据处理器（如果配置了中心端发送器）
if a.config.Sender.Mode == "core" || a.config.Sender.Mode == "hybrid" {
    if coreSender := a.sender.GetCoreSender(); coreSender != nil {
        a.scheduler.SetLLDPHandler(func(deviceID string, neighbors []scheduler.LLDPNeighbor) {
            // 转换为请求格式
            neighborRequests := make([]sender.LLDPNeighborRequest, 0, len(neighbors))
            for _, n := range neighbors {
                neighborRequests = append(neighborRequests, sender.LLDPNeighborRequest{...})
            }
            
            // 发送 LLDP 数据
            if err := coreSender.SendLLDP(ctx, deviceID, neighborRequests); err != nil {
                logger.Error("Failed to send LLDP data", ...)
            }
        })
    }
}
```

### 5. Sender 扩展

**文件**: `orbital-sentinels/internal/sender/sender.go`

#### 5.1 新增方法

```go
// GetCoreSender 获取中心端发送器
func (s *Sender) GetCoreSender() *CoreSender {
    return s.coreSender
}
```

## 数据流

### 完整流程

```
1. 调度器执行 LLDP 采集任务
   │
   ├─> 调用 LLDP 插件的 Collect() 方法
   │   └─> 通过 SNMP/SSH 采集邻居信息
   │       └─> 返回包含 LLDP 数据的 metrics
   │
   ├─> Scheduler.handleLLDPData()
   │   └─> 从 metrics 中提取 LLDP 邻居信息
   │       └─> 调用 onLLDP 处理器
   │
   ├─> Agent 的 LLDP 处理器
   │   └─> 转换为请求格式
   │       └─> 调用 CoreSender.SendLLDP()
   │
   ├─> CoreSender.SendLLDP()
   │   └─> POST /api/v1/topology/lldp
   │       └─> 发送 LLDP 数据到中心端
   │
   └─> 中心端接收
       ├─> TopologyHandler.IngestLLDP()
       ├─> TopologyService.UpsertLLDPNeighbor()
       └─> 写入 lldp_neighbors 表
```

## 配置说明

### 插件配置

**文件**: `orbital-sentinels/plugins/lldp/plugin.yaml`

```yaml
meta:
  name: lldp
  version: 1.0.0
  description: LLDP 邻居发现插件
  device_types:
    - switch
    - router
    - network_device

device_fields:
  - name: host
    type: string
    required: true
  - name: protocol
    type: string
    default: snmp
  - name: snmp_community
    type: string
    default: public
  # ...
```

### 任务配置示例

```yaml
tasks:
  - task_id: task-lldp-001
    device_id: dev-001
    plugin_name: lldp
    device_config:
      host: 192.168.1.1
      protocol: snmp
      snmp_community: public
      snmp_version: 2c
    interval: 300s  # 每 5 分钟采集一次
```

## 待实现功能

### 1. SNMP 采集实现

需要在 `collectViaSNMP()` 方法中实现：

```go
func (p *LLDPPlugin) collectViaSNMP(ctx context.Context, task *plugin.CollectionTask) ([]LLDPNeighbor, error) {
    // 1. 连接 SNMP 设备
    // 2. 查询 LLDP MIB (1.0.8802.1.1.2)
    // 3. 解析 LLDP 邻居表
    // 4. 返回邻居列表
}
```

**需要的库**: `github.com/gosnmp/gosnmp`

### 2. SSH 采集实现

需要在 `collectViaSSH()` 方法中实现：

```go
func (p *LLDPPlugin) collectViaSSH(ctx context.Context, task *plugin.CollectionTask) ([]LLDPNeighbor, error) {
    // 1. SSH 连接到设备
    // 2. 执行 LLDP 相关命令
    //    - Cisco: "show lldp neighbors"
    //    - Juniper: "show lldp neighbors"
    //    - Huawei: "display lldp neighbor-information"
    // 3. 解析命令输出
    // 4. 返回邻居列表
}
```

**需要的库**: `golang.org/x/crypto/ssh`

### 3. 输出解析

不同厂商的设备输出格式不同，需要实现适配器：

- Cisco IOS/IOS-XE
- Juniper JunOS
- Huawei VRP
- H3C Comware
- 其他厂商

## 测试建议

### 1. 单元测试

- 测试插件配置验证
- 测试数据提取逻辑
- 测试请求格式转换

### 2. 集成测试

- 测试完整的采集和上报流程
- 测试错误处理
- 测试重试机制

### 3. 端到端测试

- 配置 LLDP 采集任务
- 验证数据上报到中心端
- 验证数据写入数据库
- 验证拓扑自动发现功能

## 相关文件

### 采集端

- `orbital-sentinels/plugins/lldp/lldp.go` - LLDP 插件实现
- `orbital-sentinels/plugins/lldp/plugin.yaml` - 插件配置
- `orbital-sentinels/plugins/lldp/README.md` - 插件文档
- `orbital-sentinels/internal/scheduler/scheduler.go` - 调度器扩展
- `orbital-sentinels/internal/sender/core_sender.go` - 发送器扩展
- `orbital-sentinels/internal/agent/agent.go` - Agent 集成

### 中心端

- `gravital-core/internal/api/handler/topology_handler.go` - LLDP 数据接收
- `gravital-core/internal/service/topology_service.go` - LLDP 数据处理
- `gravital-core/internal/repository/topology_repository.go` - LLDP 数据存储

## 总结

✅ **已实现**:
- LLDP 插件框架
- Scheduler LLDP 数据处理
- CoreSender LLDP 数据上报
- Agent 集成
- 插件配置文件

⏳ **待实现**:
- SNMP 采集逻辑
- SSH 采集逻辑
- 多厂商设备适配
- 单元测试和集成测试

现在系统已经具备了完整的 LLDP 数据采集和上报框架，一旦实现了 SNMP/SSH 采集逻辑，整个流程就可以正常工作了！

