<template>
  <div class="notification-config">
    <el-form-item label="启用通知">
      <el-switch v-model="localConfig.enabled" />
    </el-form-item>

    <template v-if="localConfig.enabled">
      <!-- 去重配置 -->
      <el-form-item label="去重间隔">
        <el-input-number 
          v-model="localConfig.dedupe_interval" 
          :min="60"
          :step="60"
          placeholder="去重间隔（秒）"
          style="width: 200px"
        />
        <span style="margin-left: 10px">秒</span>
        <div class="form-tip">相同告警在此时间内只通知一次</div>
      </el-form-item>

      <!-- 通知渠道 -->
      <el-form-item label="通知渠道">
        <el-button size="small" @click="handleAddChannel">
          <el-icon><Plus /></el-icon> 添加渠道
        </el-button>
      </el-form-item>

      <el-card 
        v-for="(channel, index) in localConfig.channels" 
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
            default-first-option
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
        <el-switch v-model="localConfig.escalation_enabled" />
      </el-form-item>

      <template v-if="localConfig.escalation_enabled">
        <el-form-item label="升级时间">
          <el-input-number 
            v-model="localConfig.escalation_after" 
            :min="300"
            :step="300"
            placeholder="升级时间（秒）"
            style="width: 200px"
          />
          <span style="margin-left: 10px">秒</span>
          <div class="form-tip">告警持续此时间后升级通知</div>
        </el-form-item>

        <el-form-item label="升级渠道">
          <el-checkbox-group v-model="localConfig.escalation_channels">
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
import { computed } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import type { NotificationConfig, NotificationChannelConfig } from '@/types/alert'

const props = defineProps<{
  modelValue?: NotificationConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: NotificationConfig]
}>()

// 使用 computed 直接操作父组件的数据，避免数据同步问题
const localConfig = computed({
  get: () => props.modelValue || {
    enabled: false,
    channels: [],
    dedupe_interval: 300,
    escalation_enabled: false,
    escalation_after: 1800,
    escalation_channels: []
  },
  set: (value) => {
    emit('update:modelValue', value)
  }
})

const handleAddChannel = () => {
  const newConfig = { ...localConfig.value }
  if (!newConfig.channels) {
    newConfig.channels = []
  }
  newConfig.channels = [...newConfig.channels, {
    channel: 'email',
    enabled: true,
    recipients: []
  }]
  emit('update:modelValue', newConfig)
}

const handleRemoveChannel = (index: number) => {
  const newConfig = { ...localConfig.value }
  newConfig.channels = [...newConfig.channels]
  newConfig.channels.splice(index, 1)
  emit('update:modelValue', newConfig)
}

const getRecipientSuggestions = (channel: string) => {
  // 根据渠道类型返回建议的接收人列表
  // 可以从后端 API 获取或使用本地存储的常用接收人
  const suggestions: Record<string, string[]> = {
    email: ['admin@example.com', 'ops@example.com'],
    webhook: ['https://your-endpoint.com/alerts'],
    dingtalk: ['https://oapi.dingtalk.com/robot/send?access_token=xxx'],
    wechat: ['https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx']
  }
  return suggestions[channel] || []
}

const getRecipientTip = (channel: string) => {
  const tips: Record<string, string> = {
    email: '输入邮箱地址，如：admin@example.com',
    webhook: '输入 Webhook URL，如：https://your-endpoint.com/alerts',
    dingtalk: '输入钉钉机器人 Webhook URL',
    wechat: '输入企业微信机器人 Webhook URL',
    sms: '输入手机号码，如：13800138000'
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
  border: 1px solid #e4e7ed;
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

:deep(.el-form-item) {
  margin-bottom: 18px;
}

:deep(.el-card__header) {
  padding: 12px 15px;
  background-color: #f5f7fa;
}

:deep(.el-card__body) {
  padding: 15px;
}
</style>

