# 端到端数据流与模块协作详解 - Part 3A: 模块职责与边界

> 本文档详细说明系统各模块的职责边界、交互方式和设计原则。

## 4. 模块职责与边界

### 4.1 中心端模块划分

```
gravital-core/
├── cmd/server/              # 启动入口
├── internal/
│   ├── api/                 # API 层
│   │   ├── handler/         # HTTP 请求处理器
│   │   ├── middleware/      # 中间件（认证、日志等）
│   │   └── router/          # 路由配置
│   │
│   ├── service/             # 业务逻辑层
│   │   ├── device_service.go
│   │   ├── task_service.go
│   │   ├── sentinel_service.go
│   │   ├── alert_service.go
│   │   ├── user_service.go
│   │   └── forwarder_service.go
│   │
│   ├── repository/          # 数据访问层
│   │   ├── device_repository.go
│   │   ├── task_repository.go
│   │   └── ...
│   │
│   ├── model/               # 数据模型
│   │   ├── device.go
│   │   ├── task.go
│   │   └── ...
│   │
│   ├── forwarder/           # 数据转发模块
│   │   ├── manager.go       # 转发管理器
│   │   ├── prometheus.go    # Prometheus 转发器
│   │   ├── victoria.go      # VictoriaMetrics 转发器
│   │   └── clickhouse.go    # ClickHouse 转发器
│   │
│   ├── alert/               # 告警模块
│   │   ├── engine.go        # 告警引擎
│   │   └── notifier.go      # 通知发送器
│   │
│   └── pkg/                 # 公共组件
│       ├── config/          # 配置管理
│       ├── logger/          # 日志
│       ├── database/        # 数据库连接
│       └── redis/           # Redis 连接
```

### 4.2 各层职责详解

#### 4.2.1 API 层 (Handler + Router)

**职责**：
- HTTP 请求解析和参数验证
- 调用 Service 层执行业务逻辑
- 统一响应格式封装
- 错误处理和日志记录

**不应该做**：
- ❌ 直接操作数据库
- ❌ 包含复杂的业务逻辑
- ❌ 直接调用第三方服务

**示例**：

```go
// ✅ 正确：Handler 只负责请求处理和响应
func (h *DeviceHandler) Create(c *gin.Context) {
    var req CreateDeviceRequest
    
    // 1. 参数验证
    if err := c.ShouldBindJSON(&req); err != nil {
        ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
        return
    }
    
    // 2. 调用 Service 层
    device, err := h.service.CreateDevice(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("Failed to create device", zap.Error(err))
        ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
        return
    }
    
    // 3. 返回响应
    SuccessResponse(c, device)
}

// ❌ 错误：Handler 中包含业务逻辑
func (h *DeviceHandler) Create(c *gin.Context) {
    var req CreateDeviceRequest
    c.ShouldBindJSON(&req)
    
    // ❌ 不应该在 Handler 中直接操作数据库
    device := &model.Device{
        DeviceID: generateID(),
        Name:     req.Name,
    }
    db.Create(device)
    
    // ❌ 不应该在 Handler 中包含复杂逻辑
    if device.DeviceType == "server" {
        // 创建默认任务
        task := &model.Task{...}
        db.Create(task)
    }
    
    SuccessResponse(c, device)
}
```

#### 4.2.2 Service 层 (Business Logic)

**职责**：
- 实现核心业务逻辑
- 协调多个 Repository 完成复杂操作
- 事务管理
- 业务规则验证
- 调用外部服务（如发送通知）

**不应该做**：
- ❌ 直接处理 HTTP 请求/响应
- ❌ 包含 SQL 语句
- ❌ 直接操作 ORM 模型

**示例**：

```go
// ✅ 正确：Service 包含业务逻辑
func (s *deviceService) CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*model.Device, error) {
    // 1. 业务规则验证
    if req.GroupID != nil {
        group, err := s.groupRepo.GetByID(ctx, *req.GroupID)
        if err != nil {
            return nil, fmt.Errorf("group not found: %w", err)
        }
    }
    
    // 2. 生成业务数据
    deviceID := fmt.Sprintf("dev-%s", generateShortID())
    
    // 3. 构造数据模型
    device := &model.Device{
        DeviceID:         deviceID,
        Name:             req.Name,
        DeviceType:       req.DeviceType,
        ConnectionConfig: req.ConnectionConfig,
        Status:           "unknown",
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }
    
    // 4. 调用 Repository 持久化
    if err := s.deviceRepo.Create(ctx, device); err != nil {
        return nil, fmt.Errorf("failed to create device: %w", err)
    }
    
    // 5. 执行关联操作（如果需要事务，使用 DB Transaction）
    if req.CreateDefaultTasks {
        s.createDefaultTasks(ctx, device)
    }
    
    return device, nil
}

// ❌ 错误：Service 中包含 SQL
func (s *deviceService) CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*model.Device, error) {
    // ❌ 不应该在 Service 中写 SQL
    result := s.db.Exec("INSERT INTO devices (device_id, name) VALUES (?, ?)", 
        generateID(), req.Name)
    
    return nil, nil
}
```

#### 4.2.3 Repository 层 (Data Access)

**职责**：
- 封装数据库操作
- 提供 CRUD 接口
- 构造查询条件
- 处理数据库错误

**不应该做**：
- ❌ 包含业务逻辑
- ❌ 调用其他 Repository
- ❌ 直接返回 ORM 错误（应该包装）

**示例**：

```go
// ✅ 正确：Repository 只负责数据访问
type DeviceRepository interface {
    Create(ctx context.Context, device *model.Device) error
    GetByID(ctx context.Context, id uint) (*model.Device, error)
    GetByDeviceID(ctx context.Context, deviceID string) (*model.Device, error)
    List(ctx context.Context, req *ListDevicesRequest) ([]*model.Device, int64, error)
    Update(ctx context.Context, device *model.Device) error
    Delete(ctx context.Context, id uint) error
}

type deviceRepository struct {
    db *gorm.DB
}

func (r *deviceRepository) Create(ctx context.Context, device *model.Device) error {
    if err := r.db.WithContext(ctx).Create(device).Error; err != nil {
        return fmt.Errorf("failed to create device: %w", err)
    }
    return nil
}

func (r *deviceRepository) List(ctx context.Context, req *ListDevicesRequest) ([]*model.Device, int64, error) {
    var devices []*model.Device
    var total int64
    
    query := r.db.WithContext(ctx).Model(&model.Device{})
    
    // 构造查询条件
    if req.Keyword != "" {
        query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
    }
    if req.DeviceType != "" {
        query = query.Where("device_type = ?", req.DeviceType)
    }
    
    // 计数
    query.Count(&total)
    
    // 分页
    offset := (req.Page - 1) * req.Size
    if err := query.Offset(offset).Limit(req.Size).Find(&devices).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to list devices: %w", err)
    }
    
    return devices, total, nil
}

// ❌ 错误：Repository 包含业务逻辑
func (r *deviceRepository) Create(ctx context.Context, device *model.Device) error {
    // ❌ 不应该在 Repository 中验证业务规则
    if device.DeviceType == "server" && device.ConnectionConfig["port"] == nil {
        return errors.New("server must have port")
    }
    
    // ❌ 不应该调用其他 Repository
    group, _ := r.groupRepo.GetByID(ctx, device.GroupID)
    
    r.db.Create(device)
    return nil
}
```

### 4.3 采集端模块划分

```
orbital-sentinels/
├── cmd/sentinel/            # 启动入口
├── internal/
│   ├── agent/               # Agent 控制器
│   │   └── agent.go         # 协调各组件
│   │
│   ├── register/            # 注册模块
│   │   ├── manager.go       # 注册管理器
│   │   └── types.go         # 注册相关类型
│   │
│   ├── heartbeat/           # 心跳模块
│   │   └── heartbeat.go     # 心跳发送器
│   │
│   ├── client/              # API 客户端
│   │   └── task_client.go   # 任务拉取客户端
│   │
│   ├── scheduler/           # 任务调度模块
│   │   ├── scheduler.go     # 调度器
│   │   ├── worker_pool.go   # 工作池
│   │   └── priority_queue.go # 优先级队列
│   │
│   ├── plugin/              # 插件管理
│   │   ├── manager.go       # 插件管理器
│   │   ├── interface.go     # 插件接口定义
│   │   └── loader.go        # 插件加载器
│   │
│   ├── sender/              # 数据发送模块
│   │   ├── sender.go        # 发送器
│   │   ├── core_sender.go   # 中心端发送器
│   │   ├── direct_sender.go # 直连发送器
│   │   └── buffer.go        # 缓冲区
│   │
│   ├── credentials/         # 凭证管理
│   │   ├── credentials.go   # 凭证结构
│   │   └── manager.go       # 凭证管理器
│   │
│   └── pkg/                 # 公共组件
│       ├── config/          # 配置管理
│       └── logger/          # 日志
│
└── plugins/                 # 插件目录
    ├── ping/                # Ping 插件
    ├── snmp/                # SNMP 插件
    └── ssh/                 # SSH 插件
```

### 4.4 模块交互规则

#### 4.4.1 分层调用规则

```
┌─────────────────────────────────────────┐
│           API Layer (Handler)            │  ← 只能调用 Service
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│        Service Layer (Business)          │  ← 可以调用 Repository、外部服务
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│      Repository Layer (Data Access)      │  ← 只能操作数据库
└─────────────────────────────────────────┘
```

**规则**：
- ✅ 上层可以调用下层
- ❌ 下层不能调用上层
- ❌ 同层之间不能直接调用（通过依赖注入）

#### 4.4.2 跨模块通信

**场景 1：Service 需要调用另一个 Service**

```go
// ✅ 正确：通过依赖注入
type taskService struct {
    taskRepo     repository.TaskRepository
    deviceRepo   repository.DeviceRepository
    sentinelRepo repository.SentinelRepository
    // 注入其他 Service（如果需要）
    deviceService DeviceService  // 通过接口依赖
}

func (s *taskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*model.CollectionTask, error) {
    // 验证设备存在（调用 deviceService）
    device, err := s.deviceService.GetDevice(ctx, req.DeviceID)
    if err != nil {
        return nil, fmt.Errorf("device not found: %w", err)
    }
    
    // ... 创建任务逻辑
}

// ❌ 错误：直接 new 一个 Service
func (s *taskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*model.CollectionTask, error) {
    // ❌ 不应该直接创建 Service 实例
    deviceService := service.NewDeviceService(...)
    device, _ := deviceService.GetDevice(ctx, req.DeviceID)
    
    // ...
}
```

**场景 2：异步事件通知**

```go
// 使用事件总线（Event Bus）
type EventBus interface {
    Publish(event Event)
    Subscribe(eventType string, handler EventHandler)
}

// 设备服务发布事件
func (s *deviceService) CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*model.Device, error) {
    device, err := s.deviceRepo.Create(ctx, device)
    if err != nil {
        return nil, err
    }
    
    // 发布设备创建事件
    s.eventBus.Publish(Event{
        Type: "device.created",
        Data: device,
    })
    
    return device, nil
}

// 任务服务订阅事件
func (s *taskService) Start() {
    s.eventBus.Subscribe("device.created", func(event Event) {
        device := event.Data.(*model.Device)
        // 自动创建默认任务
        s.createDefaultTasks(context.Background(), device)
    })
}
```

### 4.5 模块边界案例分析

#### 案例 1：设备删除时的级联处理

**需求**：删除设备时，需要同时删除关联的采集任务和告警规则。

**错误实现**（违反单一职责）：

```go
// ❌ 在 DeviceService 中处理所有逻辑
func (s *deviceService) DeleteDevice(ctx context.Context, id uint) error {
    // 删除设备
    s.deviceRepo.Delete(ctx, id)
    
    // ❌ 跨越边界：直接操作任务
    s.db.Where("device_id = ?", id).Delete(&model.Task{})
    
    // ❌ 跨越边界：直接操作告警规则
    s.db.Where("device_id = ?", id).Delete(&model.AlertRule{})
    
    return nil
}
```

**正确实现**（通过事件解耦）：

```go
// ✅ DeviceService 只负责删除设备
func (s *deviceService) DeleteDevice(ctx context.Context, id uint) error {
    device, err := s.deviceRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    // 删除设备
    if err := s.deviceRepo.Delete(ctx, id); err != nil {
        return err
    }
    
    // 发布删除事件
    s.eventBus.Publish(Event{
        Type: "device.deleted",
        Data: device,
    })
    
    return nil
}

// ✅ TaskService 监听事件并处理自己的逻辑
func (s *taskService) Start() {
    s.eventBus.Subscribe("device.deleted", func(event Event) {
        device := event.Data.(*model.Device)
        s.taskRepo.DeleteByDeviceID(context.Background(), device.DeviceID)
    })
}

// ✅ AlertService 监听事件并处理自己的逻辑
func (s *alertService) Start() {
    s.eventBus.Subscribe("device.deleted", func(event Event) {
        device := event.Data.(*model.Device)
        s.alertRepo.DeleteByDeviceID(context.Background(), device.DeviceID)
    })
}
```

#### 案例 2：数据转发的职责划分

**需求**：采集数据需要转发到多个时序库。

**职责划分**：

```
┌──────────────────────────────────────────────────────────┐
│  Ingest Handler                                          │
│  职责：接收 HTTP 请求，解压数据，调用 Service            │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│  Forwarder Service                                       │
│  职责：业务逻辑，决定转发策略                             │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│  Forwarder Manager                                       │
│  职责：管理转发器实例，批量缓冲，并发转发                 │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ├──────────┬──────────┬───────────┐
                         ▼          ▼          ▼           ▼
                    ┌─────────┐┌─────────┐┌─────────┐┌─────────┐
                    │Prometheus││Victoria ││ClickHouse││Webhook │
                    │Forwarder││ Metrics ││Forwarder││Forwarder│
                    │         ││Forwarder││         ││         │
                    └─────────┘└─────────┘└─────────┘└─────────┘
                    职责：实现具体的转发协议和格式转换
```

**实现**：

```go
// Handler：只负责 HTTP 处理
func (h *ForwarderHandler) IngestMetrics(c *gin.Context) {
    // 1. 解压数据
    reader := decompressIfNeeded(c.Request.Body, c.GetHeader("Content-Encoding"))
    
    // 2. 解析 JSON
    var req struct {
        Metrics []*forwarder.Metric `json:"metrics"`
    }
    json.NewDecoder(reader).Decode(&req)
    
    // 3. 调用 Service
    if err := h.service.IngestMetrics(c.Request.Context(), req.Metrics); err != nil {
        ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
        return
    }
    
    SuccessResponse(c, gin.H{"received": len(req.Metrics)})
}

// Service：业务逻辑（可以在这里添加过滤、采样等）
func (s *forwarderService) IngestMetrics(ctx context.Context, metrics []*forwarder.Metric) error {
    // 可以添加业务逻辑：
    // - 数据验证
    // - 采样（只转发部分数据）
    // - 标签增强
    
    // 转发到 Manager
    return s.manager.ForwardBatch(metrics)
}

// Manager：转发管理（缓冲、并发）
func (m *Manager) ForwardBatch(metrics []*Metric) error {
    // 写入缓冲区
    for _, metric := range metrics {
        m.buffer <- metric
    }
    return nil
}

// Forwarder：具体实现（协议转换）
func (f *VictoriaMetricsForwarder) Write(metrics []*Metric) error {
    // 转换为 Prometheus Remote Write 格式
    writeRequest := convertToPrometheusFormat(metrics)
    
    // 发送 HTTP 请求
    return f.sendHTTP(writeRequest)
}
```

---

## 4.6 依赖注入和组件初始化

### 4.6.1 中心端初始化流程

```go
// cmd/server/main.go
func main() {
    // 1. 加载配置
    cfg := config.Load()
    
    // 2. 初始化基础组件
    logger := logger.Init(cfg.Logging)
    db := database.Connect(cfg.Database)
    redis := redis.Connect(cfg.Redis)
    
    // 3. 初始化 Repository 层
    deviceRepo := repository.NewDeviceRepository(db)
    taskRepo := repository.NewTaskRepository(db)
    sentinelRepo := repository.NewSentinelRepository(db)
    alertRepo := repository.NewAlertRepository(db)
    userRepo := repository.NewUserRepository(db)
    forwarderRepo := repository.NewForwarderRepository(db)
    
    // 4. 初始化 Service 层（注入依赖）
    deviceService := service.NewDeviceService(deviceRepo, logger)
    taskService := service.NewTaskService(taskRepo, deviceRepo, sentinelRepo, logger)
    sentinelService := service.NewSentinelService(sentinelRepo, redis, logger)
    alertService := service.NewAlertService(alertRepo, logger)
    userService := service.NewUserService(userRepo, logger)
    forwarderService := service.NewForwarderService(forwarderRepo, cfg, logger)
    
    // 5. 初始化 Handler 层
    deviceHandler := handler.NewDeviceHandler(deviceService, logger)
    taskHandler := handler.NewTaskHandler(taskService, logger)
    sentinelHandler := handler.NewSentinelHandler(sentinelService, logger)
    alertHandler := handler.NewAlertHandler(alertService, logger)
    userHandler := handler.NewUserHandler(userService, logger)
    forwarderHandler := handler.NewForwarderHandler(forwarderService, logger)
    
    // 6. 初始化路由
    r := router.Setup(
        deviceHandler,
        taskHandler,
        sentinelHandler,
        alertHandler,
        userHandler,
        forwarderHandler,
        logger,
    )
    
    // 7. 启动后台服务
    forwarderService.Start()
    alertService.Start()
    
    // 8. 启动 HTTP 服务器
    server := &http.Server{
        Addr:    cfg.Server.Addr,
        Handler: r,
    }
    server.ListenAndServe()
}
```

### 4.6.2 采集端初始化流程

```go
// cmd/sentinel/main.go
func main() {
    // 1. 加载配置
    cfg := config.Load()
    
    // 2. 初始化日志
    logger := logger.Init(cfg.Logging)
    
    // 3. 创建 Agent
    agent := agent.NewAgent(cfg, logger)
    
    // 4. 启动 Agent（内部会初始化所有组件）
    if err := agent.Start(); err != nil {
        logger.Fatal("Failed to start agent", zap.Error(err))
    }
    
    // 5. 等待信号
    waitForShutdown()
    
    // 6. 优雅关闭
    agent.Stop()
}

// internal/agent/agent.go
func (a *Agent) Start() error {
    // 1. 注册到中心端（获取凭证）
    if err := a.handleRegistration(); err != nil {
        return err
    }
    
    // 2. 初始化组件
    a.pluginMgr = plugin.NewManager(a.config.Plugin, a.logger)
    a.scheduler = scheduler.NewScheduler(a.pluginMgr, a.config.Scheduler, a.logger)
    a.sender = sender.NewSender(a.config.Sender, buffer.NewRingBuffer(10000))
    a.heartbeat = heartbeat.NewHeartbeat(a.config.Core, a.logger)
    a.taskClient = client.NewTaskClient(a.config.Core, a.logger)
    
    // 3. 加载插件
    a.pluginMgr.LoadBuiltinPlugins()
    
    // 4. 配置调度器回调
    a.scheduler.SetMetricsHandler(a.sender.Send)
    a.scheduler.SetTaskClient(a.taskClient)
    
    // 5. 启动各组件
    a.sender.Start()
    a.scheduler.Start()
    a.heartbeat.Start()
    
    return nil
}
```

---

## 总结

### 模块职责清单

| 模块 | 核心职责 | 不应该做 |
|-----|---------|---------|
| **Handler** | HTTP 请求处理、参数验证、响应封装 | 业务逻辑、数据库操作 |
| **Service** | 业务逻辑、事务管理、规则验证 | HTTP 处理、SQL 语句 |
| **Repository** | 数据库 CRUD、查询构造 | 业务逻辑、调用其他 Repository |
| **Forwarder Manager** | 缓冲管理、并发转发、统计 | 协议实现、格式转换 |
| **Forwarder** | 协议实现、格式转换、HTTP 发送 | 缓冲管理、业务逻辑 |
| **Scheduler** | 任务调度、优先级管理、并发控制 | 数据采集、数据发送 |
| **Plugin** | 数据采集、指标生成 | 任务调度、数据发送 |
| **Sender** | 数据缓冲、批量发送、重试 | 数据采集、任务调度 |

### 设计原则

1. **单一职责原则**：每个模块只负责一件事
2. **依赖倒置原则**：依赖接口而不是具体实现
3. **开闭原则**：对扩展开放，对修改关闭
4. **接口隔离原则**：接口应该小而专注
5. **最少知识原则**：模块之间尽量少了解对方的内部实现


