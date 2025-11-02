<template>
  <div class="sentinel-detail">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="page-title">Sentinel 详情</span>
      </template>
    </el-page-header>

    <el-card v-loading="loading" class="sentinel-info-card">
      <template #header>
        <div class="card-header">
          <span>Sentinel 信息</span>
          <div>
            <StatusBadge v-if="sentinel" :status="sentinel.status" />
          </div>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="Sentinel ID">
          {{ sentinel?.sentinel_id }}
        </el-descriptions-item>
        <el-descriptions-item label="名称">
          {{ sentinel?.name }}
        </el-descriptions-item>
        <el-descriptions-item label="区域">
          {{ sentinel?.region }}
        </el-descriptions-item>
        <el-descriptions-item label="IP地址">
          {{ sentinel?.ip_address }}
        </el-descriptions-item>
        <el-descriptions-item label="主机名">
          {{ sentinel?.hostname }}
        </el-descriptions-item>
        <el-descriptions-item label="版本">
          {{ sentinel?.version }}
        </el-descriptions-item>
        <el-descriptions-item label="操作系统">
          {{ sentinel?.os_type }} {{ sentinel?.os_version }}
        </el-descriptions-item>
        <el-descriptions-item label="架构">
          {{ sentinel?.arch }}
        </el-descriptions-item>
        <el-descriptions-item label="最后心跳">
          {{ sentinel?.last_heartbeat }}
        </el-descriptions-item>
        <el-descriptions-item label="注册时间">
          {{ sentinel?.created_at }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card>
          <el-statistic title="CPU 使用率" :value="sentinel?.cpu_usage || 0" suffix="%" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <el-statistic title="内存使用" :value="sentinel?.memory_usage || 0" suffix="MB" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <el-statistic title="管理设备数" :value="sentinel?.device_count || 0" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <el-statistic title="活跃任务" :value="sentinel?.task_count || 0" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { sentinelApi } from '@/api/sentinel'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const sentinel = ref<any>(null)

const fetchSentinel = async () => {
  loading.value = true
  try {
    const id = route.params.id as string
    const res: any = await sentinelApi.getSentinel(id)
    sentinel.value = res
  } catch (error) {
    ElMessage.error('获取 Sentinel 详情失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchSentinel()
})
</script>

<style scoped lang="scss">
.sentinel-detail {
  .sentinel-info-card {
    margin: 20px 0;

    .card-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  .stats-row {
    margin-top: 20px;
  }
}
</style>

