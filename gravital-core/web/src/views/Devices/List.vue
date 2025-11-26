<template>
  <div class="devices-list">
    <!-- 操作栏 -->
    <el-card class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="handleCreate">
            添加设备
          </el-button>
          <el-button :icon="Upload" @click="handleImport">批量导入</el-button>
          <el-button :icon="Download" @click="handleExport">导出</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="query.keyword"
            placeholder="搜索设备"
            :prefix-icon="Search"
            style="width: 200px"
            clearable
            @change="fetchDevices"
          />
          <el-select
            v-model="query.status"
            placeholder="状态"
            style="width: 120px"
            clearable
            @change="fetchDevices"
          >
            <el-option label="全部" value="" />
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="错误" value="error" />
          </el-select>
          <el-select
            v-model="query.device_type"
            placeholder="类型"
            style="width: 120px"
            clearable
            @change="fetchDevices"
          >
            <el-option label="全部" value="" />
            <el-option label="服务器" value="server" />
            <el-option label="交换机" value="switch" />
            <el-option label="路由器" value="router" />
          </el-select>
          <el-select
            v-model="query.labels"
            placeholder="标签"
            style="width: 200px"
            multiple
            collapse-tags
            collapse-tags-tooltip
            clearable
            @change="fetchDevices"
          >
            <el-option
              v-for="tag in availableTags"
              :key="tag"
              :label="tag"
              :value="tag"
            />
          </el-select>
        </div>
      </div>
    </el-card>

    <!-- 设备列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="deviceStore.loading"
        :data="deviceStore.devices"
        style="width: 100%"
      >
        <el-table-column prop="device_id" label="设备ID" width="180" />
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column prop="device_type" label="类型" width="120" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <StatusBadge :status="row.status" />
          </template>
        </el-table-column>
        <el-table-column prop="sentinel_id" label="Sentinel" width="180" />
        <el-table-column label="标签">
          <template #default="{ row }">
            <el-tag
              v-for="(value, key) in row.labels"
              :key="key"
              size="small"
              style="margin-right: 5px"
            >
              {{ key }}:{{ value }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_seen" label="最后在线" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click="handleView(row)">
              详情
            </el-button>
            <el-button text type="primary" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button text type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.size"
        :total="deviceStore.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchDevices"
        @current-change="fetchDevices"
      />
    </el-card>

    <!-- 创建/编辑设备对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="设备名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入设备名称" />
        </el-form-item>
        
        <el-form-item label="设备类型" prop="device_type">
          <el-select v-model="form.device_type" placeholder="请选择设备类型">
            <el-option label="服务器" value="server" />
            <el-option label="交换机" value="switch" />
            <el-option label="路由器" value="router" />
            <el-option label="数据库" value="database" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="IP地址" prop="connection_config.host">
          <el-input v-model="form.connection_config.host" placeholder="请输入IP地址" />
        </el-form-item>
        
        <el-form-item label="端口" prop="connection_config.port">
          <el-input-number v-model="form.connection_config.port" :min="1" :max="65535" />
        </el-form-item>
        
        <el-form-item label="协议" prop="connection_config.protocol">
          <el-select 
            v-model="form.connection_config.protocol" 
            placeholder="请选择协议"
            @change="handleProtocolChange"
          >
            <el-option label="SSH" value="ssh" />
            <el-option label="SNMP" value="snmp" />
            <el-option label="HTTP" value="http" />
            <el-option label="HTTPS" value="https" />
          </el-select>
        </el-form-item>

        <!-- SNMP 版本选择 -->
        <el-form-item 
          v-if="form.connection_config.protocol === 'snmp'"
          label="SNMP 版本" 
          prop="connection_config.snmp_version"
        >
          <el-select 
            v-model="form.connection_config.snmp_version" 
            placeholder="请选择 SNMP 版本"
            @change="handleSNMPVersionChange"
          >
            <el-option label="SNMP v1" value="1" />
            <el-option label="SNMP v2c" value="2c" />
            <el-option label="SNMP v3" value="3" />
          </el-select>
        </el-form-item>
        
        <!-- 认证配置组件 -->
        <AuthConfig
          :protocol="form.connection_config.protocol"
          :snmp-version="form.connection_config.snmp_version || '2c'"
          v-model="form.connection_config"
        />
        
        <el-form-item label="标签">
          <TagInput v-model="form.labels" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDeviceStore } from '@/stores/device'
import StatusBadge from '@/components/common/StatusBadge.vue'
import TagInput from '@/components/common/TagInput.vue'
import AuthConfig from '@/components/common/AuthConfig.vue'
import { Plus, Upload, Download, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const deviceStore = useDeviceStore()

const query = reactive({
  page: 1,
  size: 20,
  keyword: '',
  status: '',
  device_type: '',
  labels: [] as string[]
})

const dialogVisible = ref(false)
const dialogTitle = ref('添加设备')
const formRef = ref<FormInstance>()
const currentDevice = ref<any>(null)
const availableTags = ref<string[]>([])

const form = reactive({
  name: '',
  device_type: '',
  connection_config: {
    host: '',
    port: 22,
    protocol: 'ssh',
    snmp_version: '2c',
    auth: {
      type: 'ssh',
      config: {
        username: '',
        password: '',
        auth_method: 'password'
      }
    }
  },
  labels: {}
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  device_type: [{ required: true, message: '请选择设备类型', trigger: 'change' }],
  'connection_config.host': [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  'connection_config.port': [{ required: true, message: '请输入端口', trigger: 'blur' }]
}

const fetchDevices = () => {
  deviceStore.fetchDevices(query)
}

const fetchAvailableTags = async () => {
  try {
    const tags = await deviceStore.fetchDeviceTags()
    availableTags.value = tags
  } catch (error) {
    console.error('获取标签列表失败', error)
  }
}

const handleCreate = () => {
  dialogTitle.value = '添加设备'
  currentDevice.value = null
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑设备'
  currentDevice.value = row
  
  // 处理连接配置，确保 auth 结构正确
  const connectionConfig = { ...row.connection_config }
  
  // 如果旧数据没有 auth 结构，进行兼容性转换
  if (!connectionConfig.auth) {
    connectionConfig.auth = convertLegacyAuth(connectionConfig)
  }
  
  Object.assign(form, {
    name: row.name,
    device_type: row.device_type,
    connection_config: connectionConfig,
    labels: { ...row.labels }
  })
  dialogVisible.value = true
}

// 兼容旧数据格式，转换为新的 auth 结构
function convertLegacyAuth(config: any) {
  if (config.protocol === 'snmp') {
    const version = config.snmp_version || '2c'
    if (version === '3') {
      return {
        type: 'snmp_v3',
        config: {
          username: config.username || '',
          security_level: config.security_level || 'noAuthNoPriv',
          auth_protocol: config.auth_protocol || 'MD5',
          auth_password: config.auth_password || '',
          priv_protocol: config.priv_protocol || 'DES',
          priv_password: config.priv_password || '',
          context_name: config.context_name || ''
        }
      }
    } else {
      return {
        type: `snmp_${version}`,
        config: {
          community: config.community || config.password || 'public'
        }
      }
    }
  } else if (config.protocol === 'ssh') {
    return {
      type: 'ssh',
      config: {
        username: config.username || '',
        password: config.password || '',
        private_key: config.private_key || '',
        passphrase: config.passphrase || '',
        auth_method: config.private_key ? 'key' : 'password'
      }
    }
  } else if (config.protocol === 'http' || config.protocol === 'https') {
    return {
      type: config.protocol,
      config: {
        auth_type: config.api_key ? 'apikey' : config.bearer_token ? 'bearer' : config.username ? 'basic' : 'none',
        username: config.username || '',
        password: config.password || '',
        api_key: config.api_key || '',
        api_key_header: config.api_key_header || 'X-API-Key',
        bearer_token: config.bearer_token || '',
        custom_headers: config.custom_headers || {}
      }
    }
  }
  
  return {
    type: 'none',
    config: {}
  }
}

const handleView = (row: any) => {
  router.push(`/devices/${row.id}`)
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm('确定要删除该设备吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deviceStore.deleteDevice(row.id)
    fetchDevices()
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (currentDevice.value) {
          await deviceStore.updateDevice(currentDevice.value.id, form)
        } else {
          await deviceStore.createDevice(form)
        }
        dialogVisible.value = false
        fetchDevices()
      } catch (error) {
        console.error(error)
      }
    }
  })
}

const handleProtocolChange = () => {
  // 协议变化时，重置端口和认证配置
  const defaultPorts: Record<string, number> = {
    ssh: 22,
    snmp: 161,
    http: 80,
    https: 443
  }
  
  form.connection_config.port = defaultPorts[form.connection_config.protocol] || 22
  
  // 重置 SNMP 版本（如果不是 SNMP 协议）
  if (form.connection_config.protocol !== 'snmp') {
    form.connection_config.snmp_version = undefined
  } else if (!form.connection_config.snmp_version) {
    form.connection_config.snmp_version = '2c'
  }
}

const handleSNMPVersionChange = () => {
  // SNMP 版本变化时，认证配置会在 AuthConfig 组件中自动处理
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, {
    name: '',
    device_type: '',
    connection_config: {
      host: '',
      port: 22,
      protocol: 'ssh',
      snmp_version: '2c',
      auth: {
        type: 'ssh',
        config: {
          username: '',
          password: '',
          auth_method: 'password'
        }
      }
    },
    labels: {}
  })
}

const handleImport = () => {
  ElMessage.info('批量导入功能开发中...')
}

const handleExport = () => {
  ElMessage.info('导出功能开发中...')
}

onMounted(() => {
  fetchDevices()
  fetchAvailableTags()
})
</script>

<style scoped lang="scss">
.devices-list {
  .toolbar-card {
    margin-bottom: 20px;

    .toolbar {
      display: flex;
      align-items: center;
      justify-content: space-between;

      .toolbar-left,
      .toolbar-right {
        display: flex;
        gap: 10px;
      }
    }
  }

  .table-card {
    :deep(.el-pagination) {
      margin-top: 20px;
      justify-content: flex-end;
    }
  }
}
</style>

