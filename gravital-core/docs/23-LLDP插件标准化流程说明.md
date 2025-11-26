# LLDP 插件标准化流程说明

## 概述

将 LLDP 插件改为通过标准的任务机制和 metrics 上报流程，与其他插件保持一致，简化架构并提高可维护性。

## 改动原因

### 原有问题

1. **特殊处理流程**: LLDP 插件有独立的处理流程，与其他插件不一致
2. **双重上报**: LLDP 数据通过专门的接口上报，metrics 通过标准接口上报
3. **代码冗余**: 需要维护两套数据上报逻辑
4. **架构复杂**: 增加了 Scheduler、Agent、CoreSender 的特殊处理代码

### 改进目标

1. **统一流程**: LLDP 插件与其他插件使用相同的任务获取和上报流程
2. **简化架构**: 移除特殊处理代码，统一通过 metrics 上报
3. **易于维护**: 减少代码重复，降低维护成本

## 实现方案

### 数据流

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
│         │ 返回 metrics (包含 lldp_neighbor 指标)            │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ Scheduler    │                                            │
│  │ - runTask()  │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         │ onMetrics(metrics, task)                           │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ Buffer       │                                            │
│  └──────┬───────┘                                            │
│         │                                                    │
│         │ Send()                                            │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ CoreSender   │                                            │
│  │ - Send()     │                                            │
│  └──────┬───────┘                                            │
└─────────┼────────────────────────────────────────────────────┘
          │
          │ POST /api/v1/data/ingest
          │ (包含所有 metrics，包括 lldp_neighbor)
          │
┌─────────▼────────────────────────────────────────────────────┐
│  中心端 (Gravital Core)                                      │
│                                                              │
│  ┌──────────────┐                                            │
│  │ Forwarder    │                                            │
│  │ Handler      │                                            │
│  │ - IngestMetrics │                                         │
│  └──────┬───────┘                                            │
│         │                                                    │
│         ├─> extractDeviceStatus()                            │
│         │   更新设备状态                                      │
│         │                                                    │
│         ├─> extractAndStoreLLDPData()                        │
│         │   提取 LLDP 数据并存储                              │
│         │                                                    │
│         └─> service.IngestMetrics()                          │
│            转发 metrics 到时序库                             │
│                                                              │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │ Topology     │                                            │
│  │ Service      │                                            │
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

## 代码改动

### 1. 中心端改动

#### ForwarderHandler 扩展

**文件**: `gravital-core/internal/api/handler/forwarder_handler.go`

**新增方法**:
```go
// extractAndStoreLLDPData 从指标中提取并存储 LLDP 邻居数据
func (h *ForwarderHandler) extractAndStoreLLDPData(ctx context.Context, metrics []*forwarder.Metric) error {
    // 1. 遍历 metrics，查找 lldp_neighbor 指标
    // 2. 提取 LLDP 邻居信息（从 labels 中）
    // 3. 调用 topologyService.UpsertLLDPNeighbor() 存储
}
```

**修改**:
- 在 `IngestMetrics` 中调用 `extractAndStoreLLDPData`
- 添加 `topologyService` 依赖

#### Router 更新

**文件**: `gravital-core/internal/api/router/router.go`

**修改**:
```go
// 传递 topologyService 给 ForwarderHandler
forwarderHandler := handler.NewForwarderHandler(forwarderService, topologyService, db, log)
```

### 2. 采集端改动

#### Scheduler 简化

**文件**: `orbital-sentinels/internal/scheduler/scheduler.go`

**删除**:
- `onLLDP` 回调字段
- `LLDPNeighbor` 结构体
- `SetLLDPHandler()` 方法
- `handleLLDPData()` 方法

**保留**:
- LLDP 插件返回的 metrics 正常通过 `onMetrics` 处理器上报

#### Agent 简化

**文件**: `orbital-sentinels/internal/agent/agent.go`

**删除**:
- LLDP 处理器设置代码（`SetLLDPHandler` 调用）

#### CoreSender 简化

**文件**: `orbital-sentinels/internal/sender/core_sender.go`

**删除**:
- `LLDPNeighborRequest` 结构体
- `SendLLDP()` 方法

## LLDP 数据格式

### Metrics 格式

LLDP 插件返回的 metrics 格式：

```go
metric := &plugin.Metric{
    Name: "lldp_neighbor",
    Value: float64(i + 1), // 序号，用于区分多个邻居
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

### 数据提取

中心端从 metrics 中提取 LLDP 数据：

```go
for _, m := range metrics {
    if m.Name != "lldp_neighbor" {
        continue
    }
    
    neighbor := &model.LLDPNeighbor{
        DeviceID:           m.Labels["device_id"],
        LocalInterface:     m.Labels["local_interface"],
        NeighborChassisID:  m.Labels["neighbor_chassis_id"],
        NeighborPortID:     m.Labels["neighbor_port_id"],
        NeighborSystemName: m.Labels["neighbor_system_name"],
        NeighborSystemDesc: m.Labels["neighbor_system_desc"],
        NeighborPortDesc:   m.Labels["neighbor_port_desc"],
        NeighborMgmtAddr:   m.Labels["neighbor_mgmt_addr"],
    }
    
    // 解析 TTL
    if ttlStr, ok := m.Labels["ttl"]; ok {
        ttl, _ := strconv.Atoi(ttlStr)
        neighbor.TTL = ttl
    }
    
    // 存储到数据库
    topologyService.UpsertLLDPNeighbor(ctx, neighbor)
}
```

## 优势

### 1. 统一流程

- ✅ 所有插件使用相同的任务获取机制
- ✅ 所有插件使用相同的 metrics 上报流程
- ✅ 代码结构更清晰，易于理解

### 2. 简化架构

- ✅ 移除了 Scheduler 中的特殊处理
- ✅ 移除了 Agent 中的特殊处理器
- ✅ 移除了 CoreSender 中的特殊方法
- ✅ 减少了代码量，降低了维护成本

### 3. 易于扩展

- ✅ 新增插件时无需考虑特殊处理
- ✅ 所有插件数据统一通过 metrics 上报
- ✅ 中心端可以统一处理所有数据

### 4. 性能优化

- ✅ 减少了 HTTP 请求次数（LLDP 数据随 metrics 一起上报）
- ✅ 减少了网络开销
- ✅ 提高了数据上报效率

## 兼容性

### 保留的接口

为了向后兼容，保留了 `POST /api/v1/topology/lldp` 接口，但不再被采集端使用。如果将来有其他系统需要直接上报 LLDP 数据，可以使用此接口。

### 数据迁移

无需数据迁移，因为：
1. LLDP 数据仍然存储在同一张表中
2. 数据格式没有变化
3. 只是数据来源从专门接口改为 metrics 提取

## 测试建议

### 1. 功能测试

- [ ] LLDP 插件正常采集数据
- [ ] LLDP 数据通过 metrics 上报
- [ ] 中心端正确提取 LLDP 数据
- [ ] LLDP 数据正确存储到数据库
- [ ] 拓扑自动发现正常工作

### 2. 性能测试

- [ ] 大量 LLDP 数据上报性能
- [ ] metrics 上报延迟
- [ ] 数据库写入性能

### 3. 集成测试

- [ ] 完整的采集和上报流程
- [ ] 错误处理和重试
- [ ] 数据一致性验证

## 相关文件

### 修改的文件

**中心端**:
- `gravital-core/internal/api/handler/forwarder_handler.go` - 添加 LLDP 数据提取
- `gravital-core/internal/api/router/router.go` - 传递 topologyService

**采集端**:
- `orbital-sentinels/internal/scheduler/scheduler.go` - 移除 LLDP 特殊处理
- `orbital-sentinels/internal/agent/agent.go` - 移除 LLDP 处理器
- `orbital-sentinels/internal/sender/core_sender.go` - 移除 SendLLDP 方法

### 保留的文件

- `gravital-core/internal/api/handler/topology_handler.go` - `IngestLLDP` 方法保留（备用）
- `gravital-core/internal/api/router/router.go` - `/api/v1/topology/lldp` 路由保留（备用）

## 总结

✅ **已完成**:
- 移除采集端的 LLDP 特殊处理
- 在中心端添加 LLDP 数据提取逻辑
- 统一通过 metrics 上报流程
- 简化代码架构

现在 LLDP 插件已经完全标准化，与其他插件使用相同的流程，代码更简洁，维护更容易！

