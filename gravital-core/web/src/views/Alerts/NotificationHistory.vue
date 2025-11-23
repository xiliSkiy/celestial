<template>
  <el-dialog
    v-model="visible"
    title="通知历史"
    width="900px"
  >
    <el-table :data="notifications" v-loading="loading">
      <el-table-column prop="channel" label="渠道" width="100">
        <template #default="{ row }">
          <el-tag :type="getChannelType(row.channel)" size="small">
            {{ getChannelText(row.channel) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="recipient" label="接收人" width="200" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)" size="small">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sent_at" label="发送时间" width="180" />
      <el-table-column prop="error_message" label="错误信息" show-overflow-tooltip />
    </el-table>
    
    <template v-if="notifications.length === 0 && !loading">
      <el-empty description="暂无通知记录" />
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { alertApi } from '@/api/alert'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  modelValue: boolean
  eventId?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const visible = ref(props.modelValue)
const loading = ref(false)
const notifications = ref<any[]>([])

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
  if (!props.eventId) return
  
  loading.value = true
  try {
    const res = await alertApi.getNotificationHistory(props.eventId)
    notifications.value = res.data || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取通知历史失败')
    notifications.value = []
  } finally {
    loading.value = false
  }
}

const getChannelType = (channel: string) => {
  const types: Record<string, string> = {
    email: 'primary',
    webhook: 'success',
    dingtalk: 'warning',
    wechat: 'info',
    sms: 'danger'
  }
  return types[channel] || 'info'
}

const getChannelText = (channel: string) => {
  const texts: Record<string, string> = {
    email: '邮件',
    webhook: 'Webhook',
    dingtalk: '钉钉',
    wechat: '企业微信',
    sms: '短信'
  }
  return texts[channel] || channel
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    sent: 'success',
    failed: 'danger',
    pending: 'info',
    sending: 'warning'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    sent: '已发送',
    failed: '失败',
    pending: '待发送',
    sending: '发送中'
  }
  return texts[status] || status
}
</script>

<style scoped>
:deep(.el-dialog__body) {
  padding: 20px;
}
</style>

