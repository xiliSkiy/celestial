# LLDP 数据采集流程实现总结

## 完成的工作

### 1. LLDP 插件实现 ✅

**文件**: `orbital-sentinels/plugins/lldp/lldp.go`

- ✅ 实现了插件接口（Plugin）
- ✅ 支持 SNMP 和 SSH 两种采集协议（框架已实现，具体逻辑待实现）
- ✅ 将 LLDP 邻居信息编码到 metrics 中传递
- ✅ 插件配置 Schema 定义

**文件**: `orbital-sentinels/plugins/lldp/plugin.yaml`

- ✅ 插件配置文件
- ✅ 设备字段定义（host, protocol, snmp_community, ssh_username 等）

**文件**: `orbital-sentinels/plugins/lldp/README.md`

- ✅ 插件使用文档
- ✅ 配置说明
- ✅ 使用场景说明

### 2. Scheduler 扩展 ✅

**文件**: `orbital-sentinels/internal/scheduler/scheduler.go`

- ✅ 添加 `LLDPNeighbor` 结构体
- ✅ 添加 `onLLDP` 回调函数字段
- ✅ 添加 `SetLLDPHandler()` 方法
- ✅ 添加 `handleLLDPData()` 方法，从 metrics 中提取 LLDP 数据

### 3. CoreSender 扩展 ✅

**文件**: `orbital-sentinels/internal/sender/core_sender.go`

- ✅ 添加 `LLDPNeighborRequest` 结构体
- ✅ 添加 `SendLLDP()` 方法
- ✅ 实现 LLDP 数据上报到 `/api/v1/topology/lldp`

### 4. Sender 扩展 ✅

**文件**: `orbital-sentinels/internal/sender/sender.go`

- ✅ 添加 `GetCoreSender()` 方法，用于获取 CoreSender 实例

### 5. Agent 集成 ✅

**文件**: `orbital-sentinels/internal/agent/agent.go`

- ✅ 注册 LLDP 插件到插件管理器
- ✅ 设置 LLDP 数据处理器
- ✅ 将 LLDP 数据转换为请求格式并发送

## 完整数据流

```
1. 调度器执行 LLDP 采集任务
   │
   ├─> 调用 LLDP 插件的 Collect() 方法
   │   └─> 通过 SNMP/SSH 采集邻居信息（待实现具体逻辑）
   │       └─> 返回包含 LLDP 数据的 metrics
   │
   ├─> Scheduler.handleLLDPData()
   │   └─> 从 metrics 中提取 LLDP 邻居信息
   │       └─> 识别 metric.Name == "lldp_neighbor"
   │           └─> 从 labels 中提取邻居信息
   │               └─> 调用 onLLDP 处理器
   │
   ├─> Agent 的 LLDP 处理器
   │   └─> 转换为 sender.LLDPNeighborRequest 格式
   │       └─> 调用 CoreSender.SendLLDP()
   │
   ├─> CoreSender.SendLLDP()
   │   └─> POST /api/v1/topology/lldp
   │       Headers:
   │         X-Sentinel-ID: sentinel-001
   │         X-API-Token: <token>
   │       Body:
   │         {
   │           "device_id": "dev-001",
   │           "neighbors": [...]
   │         }
   │
   └─> 中心端接收
       ├─> TopologyHandler.IngestLLDP()
       ├─> TopologyService.UpsertLLDPNeighbor()
       └─> 写入 lldp_neighbors 表
```

## 关键实现点

### 1. 数据传递方式

由于插件接口只返回 `[]*plugin.Metric`，我们通过以下方式传递 LLDP 数据：

- 使用特殊的 metric 名称：`lldp_neighbor`
- 将 LLDP 邻居信息存储在 metric 的 `Labels` 中
- Scheduler 识别并提取这些数据

### 2. 异步处理

- LLDP 数据上报是异步的，不阻塞 metrics 的正常发送
- 使用独立的 HTTP 请求发送 LLDP 数据

### 3. 错误处理

- 如果 LLDP 数据上报失败，记录错误日志但不影响 metrics 发送
- 使用熔断器保护，避免频繁失败请求

## 待实现功能

### 1. SNMP 采集逻辑

需要在 `collectViaSNMP()` 方法中实现：

```go
func (p *LLDPPlugin) collectViaSNMP(ctx context.Context, task *plugin.CollectionTask) ([]LLDPNeighbor, error) {
    // 1. 使用 gosnmp 库连接设备
    // 2. 查询 LLDP MIB (1.0.8802.1.1.2)
    //    - lldpLocChassisId (1.0.8802.1.1.2.1.3.1.0)
    //    - lldpRemTable (1.0.8802.1.1.2.1.4.1)
    // 3. 解析 LLDP 邻居表
    // 4. 返回邻居列表
}
```

**需要的依赖**:
```go
import "github.com/gosnmp/gosnmp"
```

### 2. SSH 采集逻辑

需要在 `collectViaSSH()` 方法中实现：

```go
func (p *LLDPPlugin) collectViaSSH(ctx context.Context, task *plugin.CollectionTask) ([]LLDPNeighbor, error) {
    // 1. 使用 golang.org/x/crypto/ssh 连接设备
    // 2. 执行 LLDP 命令（根据设备类型）
    //    - Cisco: "show lldp neighbors detail"
    //    - Juniper: "show lldp neighbors"
    //    - Huawei: "display lldp neighbor-information"
    // 3. 解析命令输出
    // 4. 返回邻居列表
}
```

**需要的依赖**:
```go
import "golang.org/x/crypto/ssh"
```

### 3. 多厂商适配

不同厂商的设备输出格式不同，需要实现解析适配器：

- **Cisco IOS/IOS-XE**: `show lldp neighbors detail`
- **Juniper JunOS**: `show lldp neighbors`
- **Huawei VRP**: `display lldp neighbor-information`
- **H3C Comware**: `display lldp neighbor-information`
- **其他厂商**: 根据实际情况适配

## 测试建议

### 1. 单元测试

- 测试插件配置验证
- 测试数据提取逻辑
- 测试请求格式转换
- 测试错误处理

### 2. 集成测试

- 测试完整的采集和上报流程
- 测试 SNMP 采集（需要模拟 SNMP 设备）
- 测试 SSH 采集（需要模拟 SSH 设备）
- 测试错误处理和重试

### 3. 端到端测试

1. **配置 LLDP 采集任务**
   ```yaml
   tasks:
     - task_id: task-lldp-001
       device_id: dev-001
       plugin_name: lldp
       device_config:
         host: 192.168.1.1
         protocol: snmp
         snmp_community: public
       interval: 300s
   ```

2. **启动 Sentinel**
   - 验证插件加载
   - 验证任务执行
   - 验证数据上报

3. **验证中心端接收**
   - 检查 API 日志
   - 检查数据库 `lldp_neighbors` 表
   - 验证数据完整性

4. **验证拓扑自动发现**
   - 触发自动发现
   - 验证节点和链路创建

## 相关文件清单

### 采集端（Orbital Sentinel）

#### 新增文件
- `orbital-sentinels/plugins/lldp/lldp.go` - LLDP 插件实现
- `orbital-sentinels/plugins/lldp/plugin.yaml` - 插件配置
- `orbital-sentinels/plugins/lldp/README.md` - 插件文档

#### 修改文件
- `orbital-sentinels/internal/scheduler/scheduler.go` - 添加 LLDP 数据处理
- `orbital-sentinels/internal/sender/core_sender.go` - 添加 LLDP 数据上报
- `orbital-sentinels/internal/sender/sender.go` - 添加 GetCoreSender 方法
- `orbital-sentinels/internal/agent/agent.go` - 注册插件和设置处理器

### 中心端（Gravital Core）

#### 已有文件（之前已实现）
- `gravital-core/internal/api/handler/topology_handler.go` - LLDP 数据接收
- `gravital-core/internal/service/topology_service.go` - LLDP 数据处理
- `gravital-core/internal/repository/topology_repository.go` - LLDP 数据存储

#### 文档
- `gravital-core/docs/17-网络拓扑逻辑流程梳理.md` - 逻辑流程文档
- `gravital-core/docs/20-LLDP数据采集流程实现说明.md` - 实现说明
- `gravital-core/docs/21-LLDP数据采集流程实现总结.md` - 本文档

## 使用示例

### 1. 创建 LLDP 采集任务

在中心端创建任务：

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "dev-001",
    "plugin_name": "lldp",
    "device_config": {
      "host": "192.168.1.1",
      "protocol": "snmp",
      "snmp_community": "public",
      "snmp_version": "2c"
    },
    "interval": 300,
    "timeout": 30,
    "enabled": true
  }'
```

### 2. 查看上报的 LLDP 数据

```sql
SELECT 
    device_id,
    local_interface,
    neighbor_chassis_id,
    neighbor_system_name,
    neighbor_mgmt_addr,
    last_seen
FROM lldp_neighbors
ORDER BY last_seen DESC;
```

### 3. 触发拓扑自动发现

```bash
curl -X POST http://localhost:8080/api/v1/topologies/1/discover \
  -H "Authorization: Bearer <token>"
```

## 总结

✅ **已完成**:
- LLDP 插件框架
- Scheduler LLDP 数据处理
- CoreSender LLDP 数据上报
- Agent 集成
- 插件注册

⏳ **待实现**:
- SNMP 采集逻辑（需要 gosnmp 库）
- SSH 采集逻辑（需要 golang.org/x/crypto/ssh）
- 多厂商设备适配
- 单元测试和集成测试

现在系统已经具备了完整的 LLDP 数据采集和上报框架。一旦实现了 SNMP/SSH 采集逻辑，整个 LLDP 数据采集流程就可以正常工作了！

## 下一步

1. **实现 SNMP 采集**: 使用 gosnmp 库实现 SNMP 采集逻辑
2. **实现 SSH 采集**: 使用 golang.org/x/crypto/ssh 实现 SSH 采集逻辑
3. **设备适配**: 实现不同厂商设备的输出解析
4. **测试验证**: 编写单元测试和集成测试
5. **文档完善**: 补充使用示例和故障排查指南

