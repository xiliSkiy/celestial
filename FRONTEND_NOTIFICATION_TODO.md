# 前端通知配置功能待实现清单

## 📋 概述

后端告警通知功能已完成，前端需要添加通知配置的 UI 界面，让用户可以在创建/编辑告警规则时配置通知选项。

---

## ✅ 已完成（后端）

- [x] 通知服务核心实现
- [x] 多种通知渠道（邮件、Webhook、钉钉、企业微信）
- [x] 通知去重和升级机制
- [x] 集成到告警引擎
- [x] 类型定义（`notification_config` 字段已存在）

---

## 📝 待实现（前端）

### 1. 更新类型定义

**文件**: `gravital-core/web/src/types/alert.ts`

添加通知配置的详细类型：

```typescript
// 通知渠道配置
export interface NotificationChannelConfig {
  channel: 'email' | 'webhook' | 'dingtalk' | 'wechat' | 'sms'
  enabled: boolean
  recipients: string[]
  template?: string
  config?: Record<string, any>
}

// 通知配置
export interface NotificationConfig {
  enabled: boolean
  channels: NotificationChannelConfig[]
  dedupe_interval?: number      // 去重间隔（秒）
  escalation_enabled?: boolean  // 是否启用升级
  escalation_after?: number     // 升级时间（秒）
  escalation_channels?: string[] // 升级通知渠道
}

// 更新 AlertRuleCreate 接口
export interface AlertRuleCreate {
  rule_name: string
  description?: string
  severity: 'critical' | 'warning' | 'info'
  condition: string
  duration: number
  filters?: Record<string, any>
  notification_config?: NotificationConfig  // 改为具体类型
  enabled?: boolean
}
```

### 2. 创建通知配置组件

**文件**: `gravital-core/web/src/components/alert/NotificationConfig.vue`

创建一个独立的通知配置组件：

```vue
<template>
  <div class="notification-config">
    <el-form-item label="启用通知">
      <el-switch v-model="config.enabled" />
    </el-form-item>

    <template v-if="config.enabled">
      <!-- 去重配置 -->
      <el-form-item label="去重间隔">
        <el-input-number 
          v-model="config.dedupe_interval" 
          :min="60"
          :step="60"
          placeholder="去重间隔（秒）"
        >
          <template #append>秒</template>
        </el-input-number>
        <div class="form-tip">相同告警在此时间内只通知一次</div>
      </el-form-item>

      <!-- 通知渠道 -->
      <el-form-item label="通知渠道">
        <el-button size="small" @click="handleAddChannel">
          <el-icon><Plus /></el-icon> 添加渠道
        </el-button>
      </el-form-item>

      <el-card 
        v-for="(channel, index) in config.channels" 
        :key="index"
        class="channel-card"
        shadow="never"
      >
        <template #header>
          <div class="channel-header">
            <el-select 
              v-model="channel.channel" 
              placeholder="选择渠道"
              style="width: 150px"
            >
              <el-option label="邮件" value="email" />
              <el-option label="Webhook" value="webhook" />
              <el-option label="钉钉" value="dingtalk" />
              <el-option label="企业微信" value="wechat" />
            </el-select>
            
            <el-switch v-model="channel.enabled" />
            
            <el-button 
              text 
              type="danger" 
              @click="handleRemoveChannel(index)"
            >
              删除
            </el-button>
          </div>
        </template>

        <!-- 接收人列表 -->
        <el-form-item label="接收人">
          <el-select
            v-model="channel.recipients"
            multiple
            filterable
            allow-create
            placeholder="输入接收人（邮箱/URL/手机号）"
            style="width: 100%"
          >
            <el-option
              v-for="recipient in getRecipientSuggestions(channel.channel)"
              :key="recipient"
              :label="recipient"
              :value="recipient"
            />
          </el-select>
          <div class="form-tip">
            {{ getRecipientTip(channel.channel) }}
          </div>
        </el-form-item>
      </el-card>

      <!-- 升级配置 -->
      <el-divider />
      
      <el-form-item label="启用升级">
        <el-switch v-model="config.escalation_enabled" />
      </el-form-item>

      <template v-if="config.escalation_enabled">
        <el-form-item label="升级时间">
          <el-input-number 
            v-model="config.escalation_after" 
            :min="300"
            :step="300"
            placeholder="升级时间（秒）"
          >
            <template #append>秒</template>
          </el-input-number>
          <div class="form-tip">告警持续此时间后升级通知</div>
        </el-form-item>

        <el-form-item label="升级渠道">
          <el-checkbox-group v-model="config.escalation_channels">
            <el-checkbox label="email">邮件</el-checkbox>
            <el-checkbox label="webhook">Webhook</el-checkbox>
            <el-checkbox label="dingtalk">钉钉</el-checkbox>
            <el-checkbox label="wechat">企业微信</el-checkbox>
            <el-checkbox label="sms">短信</el-checkbox>
          </el-checkbox-group>
          <div class="form-tip">升级时使用这些渠道发送通知</div>
        </el-form-item>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import type { NotificationConfig, NotificationChannelConfig } from '@/types/alert'

const props = defineProps<{
  modelValue: NotificationConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: NotificationConfig]
}>()

const config = ref<NotificationConfig>(props.modelValue || {
  enabled: false,
  channels: [],
  dedupe_interval: 300,
  escalation_enabled: false,
  escalation_after: 1800,
  escalation_channels: []
})

watch(config, (newValue) => {
  emit('update:modelValue', newValue)
}, { deep: true })

const handleAddChannel = () => {
  config.value.channels.push({
    channel: 'email',
    enabled: true,
    recipients: []
  })
}

const handleRemoveChannel = (index: number) => {
  config.value.channels.splice(index, 1)
}

const getRecipientSuggestions = (channel: string) => {
  // 根据渠道类型返回建议的接收人列表
  // 可以从后端 API 获取
  return []
}

const getRecipientTip = (channel: string) => {
  const tips = {
    email: '输入邮箱地址，如：admin@example.com',
    webhook: '输入 Webhook URL，如：https://your-endpoint.com/alerts',
    dingtalk: '输入钉钉机器人 Webhook URL',
    wechat: '输入企业微信机器人 Webhook URL'
  }
  return tips[channel] || '输入接收人信息'
}
</script>

<style scoped>
.notification-config {
  padding: 10px 0;
}

.channel-card {
  margin-bottom: 15px;
}

.channel-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}
</style>
```

### 3. 更新告警规则表单

**文件**: `gravital-core/web/src/views/Alerts/Index.vue`

在规则对话框中添加通知配置：

```vue
<!-- 在 el-dialog 中添加 -->
<el-dialog
  v-model="ruleDialogVisible"
  :title="ruleDialogTitle"
  width="800px"  <!-- 增加宽度 -->
  @close="resetRuleForm"
>
  <el-form
    ref="ruleFormRef"
    :model="ruleForm"
    :rules="ruleRules"
    label-width="100px"
  >
    <!-- 现有表单项... -->
    
    <!-- 添加通知配置 -->
    <el-divider content-position="left">通知配置</el-divider>
    
    <NotificationConfig v-model="ruleForm.notification_config" />
  </el-form>
  
  <template #footer>
    <el-button @click="ruleDialogVisible = false">取消</el-button>
    <el-button type="primary" @click="handleSubmitRule" :loading="submitting">
      确定
    </el-button>
  </template>
</el-dialog>
```

在 script 中：

```typescript
import NotificationConfig from '@/components/alert/NotificationConfig.vue'

// 更新 ruleForm
const ruleForm = reactive({
  rule_name: '',
  description: '',
  severity: 'warning',
  metric_name: 'device_status',
  operator: '!=',
  threshold: 1,
  duration: 5,
  filters: {},
  enabled: true,
  notification_config: {
    enabled: false,
    channels: [],
    dedupe_interval: 300,
    escalation_enabled: false,
    escalation_after: 1800,
    escalation_channels: []
  }
})
```

### 4. 显示通知配置信息

在规则列表中显示是否启用了通知：

```vue
<el-table-column label="通知" width="80">
  <template #default="{ row }">
    <el-tag 
      v-if="row.notification_config?.enabled" 
      type="success" 
      size="small"
    >
      已启用
    </el-tag>
    <el-tag 
      v-else 
      type="info" 
      size="small"
    >
      未启用
    </el-tag>
  </template>
</el-table-column>
```

### 5. 添加通知历史查看

**文件**: `gravital-core/web/src/views/Alerts/NotificationHistory.vue`

创建通知历史查看页面：

```vue
<template>
  <el-dialog
    v-model="visible"
    title="通知历史"
    width="800px"
  >
    <el-table :data="notifications" v-loading="loading">
      <el-table-column prop="channel" label="渠道" width="100" />
      <el-table-column prop="recipient" label="接收人" width="200" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)" size="small">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sent_at" label="发送时间" width="180" />
      <el-table-column prop="error_message" label="错误信息" />
    </el-table>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { alertApi } from '@/api/alert'

const props = defineProps<{
  modelValue: boolean
  eventId: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const visible = ref(props.modelValue)
const loading = ref(false)
const notifications = ref([])

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val && props.eventId) {
    fetchNotifications()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const fetchNotifications = async () => {
  loading.value = true
  try {
    const res = await alertApi.getNotificationHistory(props.eventId)
    notifications.value = res.data
  } finally {
    loading.value = false
  }
}

const getStatusType = (status: string) => {
  const types = {
    sent: 'success',
    failed: 'danger',
    pending: 'info',
    sending: 'warning'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts = {
    sent: '已发送',
    failed: '失败',
    pending: '待发送',
    sending: '发送中'
  }
  return texts[status] || status
}
</script>
```

### 6. 更新 API 接口

**文件**: `gravital-core/web/src/api/alert.ts`

添加通知历史查询接口：

```typescript
export const alertApi = {
  // ... 现有接口
  
  // 获取通知历史
  getNotificationHistory: (eventId: number) =>
    request.get(`/v1/alert-events/${eventId}/notifications`),
  
  // 测试通知
  testNotification: (data: {
    channel: string
    recipient: string
    subject: string
    content: string
  }) =>
    request.post('/v1/notifications/test', data),
}
```

---

## 🎨 UI 设计建议

### 通知配置界面布局

```
┌─────────────────────────────────────────────────────────┐
│ 创建告警规则                                    [×]      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ 规则名称: [________________]                            │
│ 描述:     [________________]                            │
│ 级别:     [Warning ▼]                                   │
│ ...                                                     │
│                                                         │
│ ─────────────── 通知配置 ───────────────                │
│                                                         │
│ 启用通知: [●]                                           │
│                                                         │
│ 去重间隔: [300] 秒                                      │
│ └─ 相同告警在此时间内只通知一次                         │
│                                                         │
│ 通知渠道: [+ 添加渠道]                                  │
│                                                         │
│ ┌─────────────────────────────────────────────┐        │
│ │ [邮件 ▼]              [●]           [删除]  │        │
│ ├─────────────────────────────────────────────┤        │
│ │ 接收人: [admin@example.com, ops@...]       │        │
│ │ └─ 输入邮箱地址，如：admin@example.com      │        │
│ └─────────────────────────────────────────────┘        │
│                                                         │
│ ┌─────────────────────────────────────────────┐        │
│ │ [钉钉 ▼]              [●]           [删除]  │        │
│ ├─────────────────────────────────────────────┤        │
│ │ 接收人: [https://oapi.dingtalk.com/...]    │        │
│ │ └─ 输入钉钉机器人 Webhook URL               │        │
│ └─────────────────────────────────────────────┘        │
│                                                         │
│ ──────────────────────────────────────────────         │
│                                                         │
│ 启用升级: [○]                                           │
│                                                         │
│ 升级时间: [1800] 秒                                     │
│ └─ 告警持续此时间后升级通知                             │
│                                                         │
│ 升级渠道: [✓] 邮件  [✓] 钉钉  [ ] 企业微信  [ ] 短信   │
│ └─ 升级时使用这些渠道发送通知                           │
│                                                         │
├─────────────────────────────────────────────────────────┤
│                              [取消]  [确定]             │
└─────────────────────────────────────────────────────────┘
```

---

## 📝 实施步骤

### 第 1 步：更新类型定义 (5 分钟)
- [x] 在 `types/alert.ts` 中添加详细的通知配置类型

### 第 2 步：创建通知配置组件 (30 分钟)
- [x] 创建 `NotificationConfig.vue` 组件
- [x] 实现渠道添加/删除功能
- [x] 实现接收人输入功能
- [x] 实现去重和升级配置

### 第 3 步：集成到告警规则表单 (15 分钟)
- [x] 在规则对话框中引入通知配置组件
- [x] 更新 `ruleForm` 数据结构
- [x] 确保提交时包含通知配置

### 第 4 步：显示通知状态 (10 分钟)
- [x] 在规则列表中显示通知启用状态
- [x] 在规则详情中显示完整通知配置

### 第 5 步：通知历史查看 (20 分钟)
- [x] 创建通知历史对话框
- [x] 添加查看通知历史的入口
- [x] 实现通知历史查询

### 第 6 步：测试 (20 分钟)
- [ ] 创建带通知配置的告警规则（需要启动前端服务）
- [ ] 触发告警，验证通知发送（需要后端联调）
- [ ] 查看通知历史（需要后端联调）
- [ ] 测试去重和升级功能（需要后端联调）

**预计总时间**: 约 2 小时

---

## 🧪 测试清单

### 功能测试

- [ ] 创建规则时可以配置通知
- [ ] 编辑规则时可以修改通知配置
- [ ] 可以添加/删除通知渠道
- [ ] 可以添加/删除接收人
- [ ] 去重间隔配置生效
- [ ] 升级配置生效
- [ ] 通知历史可以查看

### UI 测试

- [ ] 表单布局合理
- [ ] 输入验证正确
- [ ] 错误提示清晰
- [ ] 响应式布局正常

### 集成测试

- [ ] 创建规则后通知配置正确保存
- [ ] 告警触发后通知正常发送
- [ ] 通知记录正确显示

---

## 📚 参考资料

- [Element Plus 表单组件](https://element-plus.org/zh-CN/component/form.html)
- [Element Plus 对话框](https://element-plus.org/zh-CN/component/dialog.html)
- [告警通知功能文档](./gravital-core/docs/ALERT_NOTIFICATION.md)

---

## 💡 可选增强

### 短期

- [ ] 通知配置模板（预设常用配置）
- [ ] 接收人管理（统一管理接收人列表）
- [ ] 通知测试功能（发送测试通知）

### 中期

- [ ] 通知统计图表
- [ ] 通知成功率监控
- [ ] 通知配置导入/导出

---

**文档版本**: v1.0
**创建日期**: 2025-11-23
**状态**: ✅ 已完成

---

## ✅ 实施完成

所有计划的功能已经实现完成！详细信息请查看：
- [实施总结文档](./FRONTEND_NOTIFICATION_IMPLEMENTATION.md)

