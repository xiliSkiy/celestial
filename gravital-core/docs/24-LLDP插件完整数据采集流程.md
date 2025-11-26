# LLDP 插件完整数据采集流程

## 概述

本文档详细梳理 LLDP 插件从任务创建到数据存储的完整数据采集流程，包括各个组件的职责、数据流转和关键步骤。

## 完整流程图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 1: 任务创建                                │
└─────────────────────────────────────────────────────────────────────────┘

用户/系统
  │
  ├─> 在中心端创建设备
  │   └─> 配置 connection_config (host, protocol, auth 等)
  │
  └─> 创建 LLDP 采集任务
      └─> POST /api/v1/tasks
          {
            "device_id": "dev-001",
            "plugin_name": "lldp",
            "device_config": {
              "host": "192.168.1.1",
              "protocol": "snmp",
              "snmp_version": "2c",
              "auth": {
                "type": "snmp_v2c",
                "config": {
                  "community": "public"
                }
              }
            },
            "interval": 300,
            "timeout": 30,
            "enabled": true
          }

中心端 (Gravital Core)
  │
  ├─> TaskHandler.Create()
  │   └─> TaskService.CreateTask()
  │       └─> TaskRepository.Create()
  │           └─> 存储到 tasks 表
  │
  └─> 返回任务信息


┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 2: 任务获取                                │
└─────────────────────────────────────────────────────────────────────────┘

采集端 (Orbital Sentinel)
  │
  ├─> Agent 启动
  │   └─> 初始化 Scheduler
  │       └─> 设置 TaskClient (连接中心端)
  │
  └─> Scheduler.Start()
      │
      ├─> fetchTasksLoop() [每 5 分钟]
      │   └─> TaskClient.GetTasks()
      │       └─> GET /api/v1/tasks?sentinel_id=xxx
      │           │
      │           └─> 中心端返回任务列表
      │               [
      │                 {
      │                   "task_id": "task-001",
      │                   "device_id": "dev-001",
      │                   "plugin_name": "lldp",
      │                   "device_config": {...},
      │                   "interval": 300,
      │                   "timeout": 30
      │                 }
      │               ]
      │
      └─> UpdateTasksWithIntervals()
          └─> 更新本地任务列表
              └─> 创建 ScheduledTask
                  {
                    Task: CollectionTask{...},
                    NextRun: time.Now(),
                    Interval: 300s,
                    LastStatus: "pending"
                  }


┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 3: 任务调度                                │
└─────────────────────────────────────────────────────────────────────────┘

Scheduler
  │
  └─> scheduleLoop() [每秒检查]
      │
      └─> checkAndExecuteTasks()
          │
          └─> 检查所有任务
              │
              ├─> NextRun < now?
              ├─> LastStatus != "running"?
              │
              └─> executeTask()
                  │
                  └─> WorkerPool.Submit()
                      └─> runTask()


┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 4: 数据采集                                │
└─────────────────────────────────────────────────────────────────────────┘

Scheduler.runTask()
  │
  ├─> 获取插件
  │   └─> PluginManager.GetPlugin("lldp")
  │       └─> 返回 LLDPPlugin 实例
  │
  ├─> 创建采集上下文
  │   └─> context.WithTimeout(ctx, 30s)
  │
  └─> 调用插件采集
      └─> LLDPPlugin.Collect(ctx, task)
          │
          ├─> 解析协议类型
          │   └─> protocol = "snmp" | "ssh"
          │
          ├─> [SNMP 采集路径]
          │   └─> collectViaSNMP(ctx, task)
          │       │
          │       ├─> 读取配置
          │       │   ├─> host, port, snmp_version
          │       │   └─> 从 auth.config 读取认证信息
          │       │
          │       ├─> 创建 SNMP 客户端
          │       │   └─> gosnmp.GoSNMP{
          │       │         Target: "192.168.1.1",
          │       │         Port: 161,
          │       │         Version: gosnmp.Version2c,
          │       │         Community: "public"
          │       │       }
          │       │
          │       ├─> 连接设备
          │       │   └─> snmp.Connect()
          │       │
          │       ├─> 查询 LLDP MIB
          │       │   └─> snmp.BulkWalk("1.0.8802.1.1.2.1.4.1.1.2")
          │       │       │
          │       │       └─> 遍历每个邻居
          │       │           └─> getLLDPNeighborViaSNMP()
          │       │               │
          │       │               ├─> 查询本地接口: 1.0.8802.1.1.2.1.3.7.1.3
          │       │               ├─> 查询 Chassis ID: 1.0.8802.1.1.2.1.4.1.1.4
          │       │               ├─> 查询 Port ID: 1.0.8802.1.1.2.1.4.1.1.6
          │       │               ├─> 查询系统名称: 1.0.8802.1.1.2.1.4.1.1.9
          │       │               ├─> 查询系统描述: 1.0.8802.1.1.2.1.4.1.1.10
          │       │               ├─> 查询端口描述: 1.0.8802.1.1.2.1.4.1.1.8
          │       │               └─> 查询管理地址: 1.0.8802.1.1.2.1.4.2.1.5
          │       │
          │       └─> 返回邻居列表
          │           [
          │             {
          │               LocalInterface: "GigabitEthernet0/1",
          │               NeighborChassisID: "00:11:22:33:44:55",
          │               NeighborPortID: "GigabitEthernet0/2",
          │               NeighborSystemName: "Switch-02",
              │               ...
          │             }
          │           ]
          │
          └─> [SSH 采集路径]
              └─> collectViaSSH(ctx, task)
                  │
                  ├─> 读取配置
                  │   ├─> host, port, device_type
                  │   └─> 从 auth.config 读取认证信息
                  │
                  ├─> 创建 SSH 客户端
                  │   └─> ssh.ClientConfig{
                  │         User: "admin",
                  │         Auth: []ssh.AuthMethod{ssh.Password("pass")}
                  │       }
                  │
                  ├─> 连接设备
                  │   └─> ssh.Dial("tcp", "192.168.1.1:22", config)
                  │
                  ├─> 创建会话
                  │   └─> client.NewSession()
                  │
                  ├─> 执行命令
                  │   └─> 根据 device_type 选择命令
                  │       ├─> Cisco: "show lldp neighbors detail"
                  │       ├─> Juniper: "show lldp neighbors"
                  │       └─> Huawei: "display lldp neighbor-information"
                  │
                  ├─> 解析输出
                  │   └─> parseLLDPOutput(output, deviceType)
                  │       ├─> parseCiscoLLDPOutput()
                  │       ├─> parseJuniperLLDPOutput()
                  │       └─> parseHuaweiLLDPOutput()
                  │
                  └─> 返回邻居列表

          │
          └─> 转换为 Metrics
              └─> 为每个邻居创建 metric
                  {
                    Name: "lldp_neighbor",
                    Value: 1.0,
                    Timestamp: 1234567890,
                    Labels: {
                      "device_id": "dev-001",
                      "local_interface": "GigabitEthernet0/1",
                      "neighbor_chassis_id": "00:11:22:33:44:55",
                      "neighbor_port_id": "GigabitEthernet0/2",
                      "neighbor_system_name": "Switch-02",
                      "neighbor_system_desc": "Cisco IOS",
                      "neighbor_port_desc": "GigabitEthernet0/2",
                      "neighbor_mgmt_addr": "192.168.1.2",
                      "ttl": "120"
                    },
                    Type: "gauge"
                  }


┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 5: 数据上报                                │
└─────────────────────────────────────────────────────────────────────────┘

Scheduler.runTask() [继续]
  │
  ├─> 生成设备状态指标
  │   └─> createDeviceStatusMetric()
  │       └─> {
  │             Name: "device_status",
  │             Value: 1.0,  // 1=online, 0=offline
  │             Labels: {
  │               "device_id": "dev-001",
  │               "device_type": "switch",
  │               "task_id": "task-001",
  │               "plugin": "lldp"
  │             }
  │           }
  │
  ├─> 合并指标
  │   └─> metrics = append(metrics, statusMetric)
  │
  └─> 调用指标处理器
      └─> onMetrics(metrics, task)
          │
          └─> Agent 设置的处理器
              └─> Buffer.Push(metrics)
                  └─> 添加到内存缓冲区

Buffer
  │
  └─> Sender.flushLoop() [定时刷新，默认 10 秒]
      │
      └─> Buffer.Pop(batchSize)
          └─> 获取一批 metrics
              │
              └─> CoreSender.Send(ctx, metrics)
                  │
                  ├─> 构造请求
                  │   └─> POST /api/v1/data/ingest
                  │       Headers:
                  │         Content-Type: application/json
                  │         Content-Encoding: gzip (可选)
                  │         X-Sentinel-ID: sentinel-001
                  │         X-API-Token: <token>
                  │       Body:
                  │         {
                  │           "metrics": [
                  │             {
                  │               "name": "lldp_neighbor",
                  │               "value": 1.0,
                  │               "timestamp": 1234567890,
                  │               "labels": {
                  │                 "device_id": "dev-001",
                  │                 "local_interface": "GigabitEthernet0/1",
                  │                 ...
                  │               },
                  │               "type": "gauge"
                  │             },
                  │             {
                  │               "name": "device_status",
                  │               "value": 1.0,
                  │               ...
                  │             }
                  │           ]
                  │         }
                  │
                  └─> 发送 HTTP 请求


┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 6: 数据接收与处理                           │
└─────────────────────────────────────────────────────────────────────────┘

中心端 (Gravital Core)
  │
  └─> ForwarderHandler.IngestMetrics()
      │
      ├─> 解析请求
      │   └─> 解压 gzip (如果使用)
      │       └─> 解析 JSON
      │
      ├─> 添加 Sentinel ID 标签
      │   └─> 为所有 metrics 添加 "sentinel_id" 标签
      │
      ├─> 提取设备状态
      │   └─> extractDeviceStatus(metrics)
      │       └─> 查找 "device_status" 指标
      │           └─> 更新 devices 表
      │               └─> UPDATE devices SET status='online', last_seen=... WHERE device_id=...
      │
      ├─> 提取并存储 LLDP 数据
      │   └─> extractAndStoreLLDPData(metrics)
      │       │
      │       ├─> 遍历 metrics
      │       │   └─> 查找 "lldp_neighbor" 指标
      │       │
      │       ├─> 提取邻居信息
      │       │   └─> 从 labels 中提取:
      │       │       ├─> device_id
      │       │       ├─> local_interface
      │       │       ├─> neighbor_chassis_id
      │       │       ├─> neighbor_port_id
      │       │       ├─> neighbor_system_name
      │       │       ├─> neighbor_system_desc
      │       │       ├─> neighbor_port_desc
      │       │       ├─> neighbor_mgmt_addr
      │       │       └─> ttl
      │       │
      │       ├─> 验证必要字段
      │       │   └─> 检查 local_interface, neighbor_chassis_id, neighbor_port_id
      │       │
      │       └─> 存储到数据库
      │           └─> TopologyService.UpsertLLDPNeighbor()
      │               └─> TopologyRepository.UpsertLLDPNeighbor()
      │                   └─> INSERT INTO lldp_neighbors ... ON CONFLICT ... UPDATE ...
      │
      └─> 转发到时序库
          └─> ForwarderService.IngestMetrics()
              └─> ForwarderManager.ForwardBatch()
                  └─> 转发到 VictoriaMetrics/Prometheus/ClickHouse


┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 7: 数据存储                                │
└─────────────────────────────────────────────────────────────────────────┘

数据库 (PostgreSQL)
  │
  └─> lldp_neighbors 表
      │
      └─> 存储结构:
          {
            device_id: "dev-001",
            local_interface: "GigabitEthernet0/1",
            neighbor_chassis_id: "00:11:22:33:44:55",
            neighbor_port_id: "GigabitEthernet0/2",
            neighbor_system_name: "Switch-02",
            neighbor_system_desc: "Cisco IOS Software",
            neighbor_port_desc: "GigabitEthernet0/2",
            neighbor_mgmt_addr: "192.168.1.2",
            ttl: 120,
            last_seen: "2024-01-01 12:00:00",
            created_at: "2024-01-01 12:00:00",
            updated_at: "2024-01-01 12:00:00"
          }


┌─────────────────────────────────────────────────────────────────────────┐
│                          阶段 8: 拓扑自动发现                             │
└─────────────────────────────────────────────────────────────────────────┘

拓扑发现调度器 (TopologyDiscoveryScheduler)
  │
  └─> discoveryLoop() [每 5 分钟检查]
      │
      └─> 检查所有启用自动发现的拓扑
          │
          ├─> 检查发现间隔
          │   └─> LastDiscoveryAt + DiscoveryInterval < now?
          │
          └─> 触发自动发现
              └─> TopologyDiscoveryService.DiscoverTopology()
                  │
                  ├─> 获取所有 LLDP 邻居数据
                  │   └─> TopologyRepository.GetAllLLDPNeighbors()
                  │
                  ├─> 构建设备映射
                  │   └─> 根据 Chassis ID 和 Port ID 匹配设备
                  │
                  ├─> 创建拓扑节点
                  │   └─> 为每个设备创建节点
                  │
                  ├─> 创建拓扑链路
                  │   └─> 根据 LLDP 邻居关系创建链路
                  │
                  └─> 更新拓扑
                      └─> 保存节点和链路到数据库
```

## 详细步骤说明

### 步骤 1: 任务创建

**位置**: 中心端 API

**接口**: `POST /api/v1/tasks`

**请求示例**:
```json
{
  "device_id": "dev-001",
  "plugin_name": "lldp",
  "device_config": {
    "host": "192.168.1.1",
    "port": 161,
    "protocol": "snmp",
    "snmp_version": "2c",
    "auth": {
      "type": "snmp_v2c",
      "config": {
        "community": "public"
      }
    }
  },
  "interval": 300,
  "timeout": 30,
  "enabled": true
}
```

**处理流程**:
1. `TaskHandler.Create()` 接收请求
2. `TaskService.CreateTask()` 验证和创建任务
3. `TaskRepository.Create()` 存储到 `tasks` 表
4. 返回任务信息

### 步骤 2: 任务获取

**位置**: 采集端 Scheduler

**流程**:
1. `Scheduler.fetchTasksLoop()` 每 5 分钟执行一次
2. `TaskClient.GetTasks()` 调用 `GET /api/v1/tasks?sentinel_id=xxx`
3. 中心端返回分配给该 Sentinel 的任务列表
4. `UpdateTasksWithIntervals()` 更新本地任务列表

**关键代码**:
```go
// 从中心端获取任务
tasksWithIntervals, err := s.taskClient.GetTasks(ctx)

// 更新任务列表
s.UpdateTasksWithIntervals(tasksWithIntervals)
```

### 步骤 3: 任务调度

**位置**: 采集端 Scheduler

**流程**:
1. `scheduleLoop()` 每秒检查一次
2. `checkAndExecuteTasks()` 检查哪些任务需要执行
3. 条件: `NextRun < now` 且 `LastStatus != "running"`
4. `executeTask()` 提交到工作池执行

**关键代码**:
```go
// 检查并执行任务
for _, st := range s.tasks {
    if st.NextRun.Before(now) && st.LastStatus != TaskStatusRunning {
        s.executeTask(st)
    }
}
```

### 步骤 4: 数据采集

#### 4.1 SNMP 采集

**位置**: `LLDPPlugin.collectViaSNMP()`

**详细流程**:

1. **读取配置**
   ```go
   host := p.getString(task.DeviceConfig, "host", "")
   port := p.getInt(task.DeviceConfig, "port", 161)
   snmpVersion := p.getString(task.DeviceConfig, "snmp_version", "2c")
   
   // 从 auth 结构读取认证信息
   if authRaw, ok := task.DeviceConfig["auth"]; ok {
       // 解析 auth.config
   }
   ```

2. **创建 SNMP 客户端**
   ```go
   snmp := &gosnmp.GoSNMP{
       Target:    host,
       Port:      uint16(port),
       Version:   gosnmp.Version2c,
       Community: community,
       Timeout:   10 * time.Second,
       Retries:   3,
   }
   ```

3. **连接设备**
   ```go
   if err := snmp.Connect(); err != nil {
       return nil, fmt.Errorf("failed to connect: %w", err)
   }
   ```

4. **查询 LLDP MIB**
   ```go
   // 遍历本地端口号索引
   portNumOID := "1.0.8802.1.1.2.1.4.1.1.2"
   err := snmp.BulkWalk(portNumOID, func(pdu gosnmp.SnmpPDU) error {
       // 提取索引
       // 查询每个邻居的详细信息
       neighbor, err := p.getLLDPNeighborViaSNMP(snmp, remTimeMark, remLocalPortNum)
       neighbors = append(neighbors, *neighbor)
       return nil
   })
   ```

5. **查询邻居详细信息**
   - 本地接口: `1.0.8802.1.1.2.1.3.7.1.3.{portNum}`
   - Chassis ID: `1.0.8802.1.1.2.1.4.1.1.4.{timeMark}.{portNum}`
   - Port ID: `1.0.8802.1.1.2.1.4.1.1.6.{timeMark}.{portNum}`
   - 系统名称: `1.0.8802.1.1.2.1.4.1.1.9.{timeMark}.{portNum}`
   - 系统描述: `1.0.8802.1.1.2.1.4.1.1.10.{timeMark}.{portNum}`
   - 端口描述: `1.0.8802.1.1.2.1.4.1.1.8.{timeMark}.{portNum}`
   - 管理地址: `1.0.8802.1.1.2.1.4.2.1.5.{timeMark}.{portNum}`

#### 4.2 SSH 采集

**位置**: `LLDPPlugin.collectViaSSH()`

**详细流程**:

1. **读取配置**
   ```go
   host := p.getString(task.DeviceConfig, "host", "")
   port := p.getInt(task.DeviceConfig, "port", 22)
   deviceType := p.getString(task.DeviceConfig, "device_type", "")
   
   // 从 auth 结构读取认证信息
   username := p.getString(configMap, "username", "")
   password := p.getString(configMap, "password", "")
   privateKey := p.getString(configMap, "private_key", "")
   ```

2. **创建 SSH 客户端**
   ```go
   config := &ssh.ClientConfig{
       User:            username,
       HostKeyCallback: ssh.InsecureIgnoreHostKey(),
       Timeout:         30 * time.Second,
   }
   
   if authMethod == "password" {
       config.Auth = []ssh.AuthMethod{ssh.Password(password)}
   } else if authMethod == "key" {
       signer, _ := ssh.ParsePrivateKey([]byte(privateKey))
       config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
   }
   ```

3. **连接设备**
   ```go
   client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
   ```

4. **执行命令**
   ```go
   command := p.getLLDPCommand(deviceType)
   // Cisco: "show lldp neighbors detail"
   // Juniper: "show lldp neighbors"
   // Huawei: "display lldp neighbor-information"
   
   session, _ := client.NewSession()
   var stdout bytes.Buffer
   session.Stdout = &stdout
   session.Run(command)
   ```

5. **解析输出**
   ```go
   output := stdout.String()
   neighbors := p.parseLLDPOutput(output, deviceType)
   // 根据设备类型选择解析器
   ```

### 步骤 5: 数据转换

**位置**: `LLDPPlugin.Collect()`

**流程**:
1. 将邻居列表转换为 metrics
2. 每个邻居创建一个 `lldp_neighbor` metric
3. 邻居信息存储在 metric 的 `Labels` 中

**Metrics 格式**:
```go
metric := &plugin.Metric{
    Name:      "lldp_neighbor",
    Value:     float64(i + 1),
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

### 步骤 6: 数据上报

**位置**: 采集端 Sender

**流程**:
1. `Scheduler.runTask()` 调用 `onMetrics(metrics, task)`
2. `Buffer.Push(metrics)` 添加到缓冲区
3. `Sender.flushLoop()` 定时刷新（默认 10 秒）
4. `Buffer.Pop(batchSize)` 获取一批 metrics
5. `CoreSender.Send()` 发送到中心端

**HTTP 请求**:
```
POST /api/v1/data/ingest
Headers:
  Content-Type: application/json
  Content-Encoding: gzip (可选)
  X-Sentinel-ID: sentinel-001
  X-API-Token: <token>
Body:
  {
    "metrics": [
      {
        "name": "lldp_neighbor",
        "value": 1.0,
        "timestamp": 1234567890,
        "labels": {...},
        "type": "gauge"
      },
      {
        "name": "device_status",
        "value": 1.0,
        ...
      }
    ]
  }
```

### 步骤 7: 数据接收与处理

**位置**: 中心端 ForwarderHandler

**流程**:

1. **接收请求**
   ```go
   func (h *ForwarderHandler) IngestMetrics(c *gin.Context) {
       var req struct {
           Metrics []*forwarder.Metric `json:"metrics"`
       }
       // 解析请求
   }
   ```

2. **提取设备状态**
   ```go
   deviceStatusMap := h.extractDeviceStatus(req.Metrics)
   // 更新 devices 表
   h.updateDeviceStatusInDB(ctx, deviceStatusMap)
   ```

3. **提取 LLDP 数据**
   ```go
   h.extractAndStoreLLDPData(ctx, req.Metrics)
   ```

4. **存储 LLDP 数据**
   ```go
   for _, m := range metrics {
       if m.Name == "lldp_neighbor" {
           neighbor := &model.LLDPNeighbor{
               DeviceID:           m.Labels["device_id"],
               LocalInterface:     m.Labels["local_interface"],
               NeighborChassisID:  m.Labels["neighbor_chassis_id"],
               // ...
           }
           h.topologyService.UpsertLLDPNeighbor(ctx, neighbor)
       }
   }
   ```

5. **转发到时序库**
   ```go
   h.service.IngestMetrics(ctx, req.Metrics)
   // 转发到 VictoriaMetrics/Prometheus/ClickHouse
   ```

### 步骤 8: 数据存储

**位置**: 数据库

**表结构**: `lldp_neighbors`

**存储逻辑**:
```sql
INSERT INTO lldp_neighbors (
    device_id,
    local_interface,
    neighbor_chassis_id,
    neighbor_port_id,
    neighbor_system_name,
    neighbor_system_desc,
    neighbor_port_desc,
    neighbor_mgmt_addr,
    ttl,
    last_seen
) VALUES (...)
ON CONFLICT (device_id, local_interface, neighbor_chassis_id, neighbor_port_id)
DO UPDATE SET
    neighbor_system_name = EXCLUDED.neighbor_system_name,
    neighbor_system_desc = EXCLUDED.neighbor_system_desc,
    neighbor_port_desc = EXCLUDED.neighbor_port_desc,
    neighbor_mgmt_addr = EXCLUDED.neighbor_mgmt_addr,
    ttl = EXCLUDED.ttl,
    last_seen = EXCLUDED.last_seen,
    updated_at = NOW();
```

### 步骤 9: 拓扑自动发现

**位置**: 拓扑发现调度器

**流程**:
1. `TopologyDiscoveryScheduler.discoveryLoop()` 每 5 分钟检查
2. 查找启用自动发现的拓扑
3. 检查发现间隔是否到期
4. 触发 `TopologyDiscoveryService.DiscoverTopology()`
5. 从 `lldp_neighbors` 表读取数据
6. 构建设备映射和连接关系
7. 创建拓扑节点和链路

## 关键组件

### 采集端组件

1. **Scheduler** - 任务调度器
   - 从中心端获取任务
   - 定时执行任务
   - 管理任务状态

2. **LLDPPlugin** - LLDP 采集插件
   - SNMP 采集
   - SSH 采集
   - 数据转换

3. **Buffer** - 数据缓冲区
   - 临时存储 metrics
   - 批量处理

4. **CoreSender** - 数据发送器
   - HTTP 请求发送
   - 压缩和重试

### 中心端组件

1. **TaskHandler** - 任务管理
   - 创建、更新、删除任务
   - 任务分配

2. **ForwarderHandler** - 数据接收
   - 接收 metrics
   - 提取设备状态
   - 提取 LLDP 数据

3. **TopologyService** - 拓扑服务
   - LLDP 数据存储
   - 拓扑自动发现

4. **TopologyDiscoveryScheduler** - 拓扑发现调度器
   - 定时触发自动发现
   - 清理过期数据

## 数据格式

### CollectionTask

```go
type CollectionTask struct {
    TaskID       string
    DeviceID     string
    PluginName   string
    DeviceConfig map[string]interface{}
    Timeout      time.Duration
}
```

### LLDPNeighbor

```go
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

### Metric

```go
type Metric struct {
    Name      string
    Value     float64
    Timestamp int64
    Labels    map[string]string
    Type      MetricType
}
```

## 错误处理

### 采集阶段错误

1. **连接失败**
   - SNMP: 返回连接错误，任务标记为失败
   - SSH: 返回连接错误，任务标记为失败

2. **认证失败**
   - 记录错误日志
   - 任务标记为失败

3. **查询失败**
   - 记录警告日志
   - 返回已成功采集的数据

### 上报阶段错误

1. **网络错误**
   - 重试机制（可配置）
   - 数据保留在缓冲区

2. **中心端错误**
   - 记录错误日志
   - 重试或丢弃（根据配置）

### 存储阶段错误

1. **数据库错误**
   - 记录错误日志
   - 不中断其他数据的处理

2. **数据验证失败**
   - 记录警告日志
   - 跳过无效数据

## 性能优化

1. **批量处理**: 批量查询 SNMP、批量上报 metrics
2. **并发控制**: 工作池限制并发数
3. **缓存机制**: 任务列表缓存，减少查询
4. **压缩传输**: 使用 gzip 压缩减少网络开销
5. **超时控制**: 设置合理的超时时间

## 监控指标

### 采集端指标

- 任务执行次数
- 任务成功/失败次数
- 采集耗时
- 上报延迟
- 缓冲区大小

### 中心端指标

- 接收 metrics 数量
- LLDP 数据存储数量
- 存储耗时
- 错误数量

## 总结

LLDP 插件完整的数据采集流程包括：

1. **任务创建** - 在中心端创建采集任务
2. **任务获取** - 采集端定期从中心端获取任务
3. **任务调度** - 根据间隔定时执行任务
4. **数据采集** - 通过 SNMP 或 SSH 采集 LLDP 邻居信息
5. **数据转换** - 将邻居信息转换为 metrics
6. **数据上报** - 通过 HTTP 批量上报到中心端
7. **数据接收** - 中心端接收并解析 metrics
8. **数据存储** - 存储 LLDP 数据到数据库
9. **拓扑发现** - 基于 LLDP 数据自动发现网络拓扑

整个流程实现了从设备采集到拓扑发现的完整闭环，支持多种采集方式和设备类型，具有良好的扩展性和容错能力。

