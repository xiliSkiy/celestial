<template>
  <div class="sentinels-list">
    <!-- 操作栏 -->
    <el-card class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="query.keyword"
            placeholder="搜索 Sentinel"
            :prefix-icon="Search"
            style="width: 200px"
            clearable
            @change="fetchSentinels"
          />
          <el-select
            v-model="query.status"
            placeholder="状态"
            style="width: 120px"
            clearable
            @change="fetchSentinels"
          >
            <el-option label="全部" value="" />
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
          </el-select>
        </div>
      </div>
    </el-card>

    <!-- Sentinel 卡片网格 -->
    <div class="sentinel-grid">
      <el-card
        v-for="sentinel in sentinels"
        :key="sentinel.id"
        class="sentinel-card"
        shadow="hover"
      >
        <div class="sentinel-header">
          <h3>{{ sentinel.name }}</h3>
          <StatusBadge :status="sentinel.status" />
        </div>
        
        <div class="sentinel-body">
          <div class="sentinel-info">
            <div class="info-item">
              <span class="label">区域:</span>
              <span class="value">{{ sentinel.region }}</span>
            </div>
            <div class="info-item">
              <span class="label">IP:</span>
              <span class="value">{{ sentinel.ip_address }}</span>
            </div>
            <div class="info-item">
              <span class="label">版本:</span>
              <span class="value">{{ sentinel.version }}</span>
            </div>
          </div>
          
          <el-divider />
          
          <div class="sentinel-stats">
            <div class="stat-item">
              <div class="stat-label">设备数</div>
              <div class="stat-value">{{ sentinel.device_count || 0 }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">任务数</div>
              <div class="stat-value">{{ sentinel.task_count || 0 }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">CPU</div>
              <div class="stat-value">{{ sentinel.cpu_usage || 0 }}%</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">内存</div>
              <div class="stat-value">{{ formatMemory(sentinel.memory_usage) }}</div>
            </div>
          </div>
        </div>
        
        <div class="sentinel-footer">
          <el-button text type="primary" @click="handleView(sentinel)">
            详情
          </el-button>
          <el-button text type="danger" @click="handleDelete(sentinel)">
            删除
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- 分页 -->
    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.size"
      :total="total"
      :page-sizes="[12, 24, 48]"
      layout="total, sizes, prev, pager, next"
      @size-change="fetchSentinels"
      @current-change="fetchSentinels"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { sentinelApi } from '@/api/sentinel'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()

const query = reactive({
  page: 1,
  size: 12,
  keyword: '',
  status: ''
})

const sentinels = ref<any[]>([])
const total = ref(0)

const fetchSentinels = async () => {
  try {
    const res: any = await sentinelApi.getSentinels(query)
    sentinels.value = res.items
    total.value = res.total
  } catch (error) {
    ElMessage.error('获取 Sentinel 列表失败')
  }
}

const handleView = (sentinel: any) => {
  router.push(`/sentinels/${sentinel.id}`)
}

const handleDelete = (sentinel: any) => {
  ElMessageBox.confirm('确定要删除该 Sentinel 吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await sentinelApi.deleteSentinel(sentinel.id)
      ElMessage.success('删除成功')
      fetchSentinels()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

const formatMemory = (bytes: number) => {
  if (!bytes) return '0 MB'
  const mb = bytes / 1024 / 1024
  return `${mb.toFixed(0)} MB`
}

onMounted(() => {
  fetchSentinels()
})
</script>

<style scoped lang="scss">
.sentinels-list {
  .toolbar-card {
    margin-bottom: 20px;

    .toolbar {
      display: flex;
      gap: 10px;
    }
  }

  .sentinel-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 20px;
    margin-bottom: 20px;

    .sentinel-card {
      .sentinel-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 16px;

        h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 600;
        }
      }

      .sentinel-body {
        .sentinel-info {
          .info-item {
            display: flex;
            margin-bottom: 8px;

            .label {
              width: 60px;
              color: var(--text-secondary);
            }

            .value {
              flex: 1;
              color: var(--text-primary);
            }
          }
        }

        .sentinel-stats {
          display: grid;
          grid-template-columns: repeat(4, 1fr);
          gap: 16px;

          .stat-item {
            text-align: center;

            .stat-label {
              font-size: 12px;
              color: var(--text-secondary);
              margin-bottom: 4px;
            }

            .stat-value {
              font-size: 18px;
              font-weight: 600;
              color: var(--text-primary);
            }
          }
        }
      }

      .sentinel-footer {
        display: flex;
        justify-content: flex-end;
        gap: 8px;
        margin-top: 16px;
        padding-top: 16px;
        border-top: 1px solid var(--border-color);
      }
    }
  }

  :deep(.el-pagination) {
    justify-content: center;
  }
}
</style>

