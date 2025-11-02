<template>
  <div class="settings">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 基本设置 -->
      <el-tab-pane label="基本设置" name="basic">
        <el-card>
          <template #header>
            <span>系统配置</span>
          </template>
          <el-form
            ref="configFormRef"
            :model="configForm"
            label-width="150px"
          >
            <el-form-item label="站点名称">
              <el-input v-model="configForm.site_name" placeholder="Gravital Core" />
            </el-form-item>
            <el-form-item label="站点 URL">
              <el-input v-model="configForm.site_url" placeholder="http://localhost:8080" />
            </el-form-item>
            <el-form-item label="告警邮箱">
              <el-input v-model="configForm.alert_email" placeholder="admin@example.com" />
            </el-form-item>
            <el-form-item label="告警 Webhook">
              <el-input v-model="configForm.alert_webhook" placeholder="https://hooks.example.com/alert" />
            </el-form-item>
            <el-form-item label="数据保留天数">
              <el-input-number v-model="configForm.retention_days" :min="1" :max="365" />
            </el-form-item>
            <el-form-item label="最大设备数">
              <el-input-number v-model="configForm.max_devices" :min="1" :max="10000" />
            </el-form-item>
            <el-form-item label="最大 Sentinel 数">
              <el-input-number v-model="configForm.max_sentinels" :min="1" :max="1000" />
            </el-form-item>
            <el-form-item>
              <el-button 
                type="primary" 
                @click="handleSaveConfig"
                :loading="savingConfig"
              >
                保存配置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 用户管理 -->
      <el-tab-pane label="用户管理" name="users">
        <el-card>
          <div class="toolbar">
            <el-button type="primary" :icon="Plus" @click="handleCreateUser">
              添加用户
            </el-button>
          </div>

          <el-table :data="users" v-loading="loadingUsers" class="mt-20">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="username" label="用户名" width="150" />
            <el-table-column prop="email" label="邮箱" width="200" />
            <el-table-column label="角色" width="120">
              <template #default="{ row }">
                <el-tag>{{ row.role?.name || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <StatusBadge :status="row.enabled ? 'online' : 'offline'" />
              </template>
            </el-table-column>
            <el-table-column prop="last_login" label="最后登录" width="180" />
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" width="250" fixed="right">
              <template #default="{ row }">
                <el-button 
                  text 
                  type="primary" 
                  @click="handleEditUser(row)"
                  :disabled="row.id === 1"
                >
                  编辑
                </el-button>
                <el-button 
                  text 
                  :type="row.enabled ? 'warning' : 'success'" 
                  @click="handleToggleUser(row)"
                  :disabled="row.id === 1"
                >
                  {{ row.enabled ? '禁用' : '启用' }}
                </el-button>
                <el-button 
                  text 
                  type="info" 
                  @click="handleResetPassword(row)"
                >
                  重置密码
                </el-button>
                <el-button 
                  text 
                  type="danger" 
                  @click="handleDeleteUser(row)"
                  :disabled="row.id === 1"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="userQuery.page"
            v-model:page-size="userQuery.size"
            :total="userTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchUsers"
            @current-change="fetchUsers"
          />
        </el-card>
      </el-tab-pane>

      <!-- 角色管理 -->
      <el-tab-pane label="角色管理" name="roles">
        <el-card>
          <div class="toolbar">
            <el-button type="primary" :icon="Plus" @click="handleCreateRole">
              添加角色
            </el-button>
          </div>

          <el-table :data="roles" v-loading="loadingRoles" class="mt-20">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="角色名称" width="150" />
            <el-table-column prop="description" label="描述" />
            <el-table-column label="权限" width="300">
              <template #default="{ row }">
                <el-tag 
                  v-for="(perm, index) in row.permissions?.slice(0, 3)" 
                  :key="index"
                  size="small"
                  style="margin-right: 5px"
                >
                  {{ perm }}
                </el-tag>
                <el-tag v-if="row.permissions?.length > 3" size="small">
                  +{{ row.permissions.length - 3 }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button 
                  text 
                  type="primary" 
                  @click="handleEditRole(row)"
                  :disabled="row.id === 1"
                >
                  编辑
                </el-button>
                <el-button 
                  text 
                  type="danger" 
                  @click="handleDeleteRole(row)"
                  :disabled="row.id === 1"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 用户对话框 -->
    <el-dialog
      v-model="userDialogVisible"
      :title="userDialogTitle"
      width="600px"
      @close="resetUserForm"
    >
      <el-form
        ref="userFormRef"
        :model="userForm"
        :rules="userRules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        
        <el-form-item label="密码" prop="password" v-if="!currentUser">
          <el-input 
            v-model="userForm.password" 
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="角色" prop="role_id">
          <el-select v-model="userForm.role_id" placeholder="请选择角色">
            <el-option
              v-for="role in roles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="启用">
          <el-switch v-model="userForm.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="userDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitUser" :loading="submittingUser">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 角色对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      :title="roleDialogTitle"
      width="700px"
      @close="resetRoleForm"
    >
      <el-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="roleRules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="roleForm.description" 
            type="textarea"
            :rows="2"
            placeholder="请输入描述"
          />
        </el-form-item>
        
        <el-form-item label="权限" prop="permissions">
          <el-checkbox-group v-model="roleForm.permissions">
            <el-checkbox label="*">所有权限</el-checkbox>
            <el-divider />
            <div class="permission-group">
              <h4>设备管理</h4>
              <el-checkbox label="devices.read">查看设备</el-checkbox>
              <el-checkbox label="devices.write">编辑设备</el-checkbox>
              <el-checkbox label="devices.delete">删除设备</el-checkbox>
            </div>
            <el-divider />
            <div class="permission-group">
              <h4>Sentinel 管理</h4>
              <el-checkbox label="sentinels.read">查看 Sentinel</el-checkbox>
              <el-checkbox label="sentinels.write">编辑 Sentinel</el-checkbox>
              <el-checkbox label="sentinels.delete">删除 Sentinel</el-checkbox>
            </div>
            <el-divider />
            <div class="permission-group">
              <h4>任务管理</h4>
              <el-checkbox label="tasks.read">查看任务</el-checkbox>
              <el-checkbox label="tasks.write">编辑任务</el-checkbox>
              <el-checkbox label="tasks.delete">删除任务</el-checkbox>
            </div>
            <el-divider />
            <div class="permission-group">
              <h4>告警管理</h4>
              <el-checkbox label="alerts.read">查看告警</el-checkbox>
              <el-checkbox label="alerts.write">编辑告警</el-checkbox>
              <el-checkbox label="alerts.delete">删除告警</el-checkbox>
            </div>
            <el-divider />
            <div class="permission-group">
              <h4>数据转发</h4>
              <el-checkbox label="forwarders.read">查看转发器</el-checkbox>
              <el-checkbox label="forwarders.write">编辑转发器</el-checkbox>
              <el-checkbox label="forwarders.delete">删除转发器</el-checkbox>
            </div>
            <el-divider />
            <div class="permission-group">
              <h4>系统设置</h4>
              <el-checkbox label="settings.read">查看设置</el-checkbox>
              <el-checkbox label="settings.write">编辑设置</el-checkbox>
            </div>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitRole" :loading="submittingRole">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="重置密码"
      width="500px"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="100px"
      >
        <el-form-item label="新密码" prop="password">
          <el-input 
            v-model="passwordForm.password" 
            type="password"
            placeholder="请输入新密码"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="passwordForm.confirmPassword" 
            type="password"
            placeholder="请再次输入密码"
            show-password
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPassword" :loading="submittingPassword">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { userApi, type UserForm, type RoleForm, type SystemConfig } from '@/api/user'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'

const activeTab = ref('basic')

// 系统配置
const configFormRef = ref<FormInstance>()
const savingConfig = ref(false)
const configForm = reactive<SystemConfig>({
  site_name: 'Gravital Core',
  site_url: 'http://localhost:8080',
  alert_email: '',
  alert_webhook: '',
  retention_days: 30,
  max_devices: 1000,
  max_sentinels: 100
})

// 用户管理
const loadingUsers = ref(false)
const users = ref<any[]>([])
const userTotal = ref(0)
const userQuery = reactive({
  page: 1,
  size: 20,
  keyword: ''
})

const userDialogVisible = ref(false)
const userDialogTitle = ref('添加用户')
const userFormRef = ref<FormInstance>()
const currentUser = ref<any>(null)
const submittingUser = ref(false)

const userForm = reactive<UserForm>({
  username: '',
  email: '',
  password: '',
  role_id: 2,
  enabled: true
})

const userRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  role_id: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

// 角色管理
const loadingRoles = ref(false)
const roles = ref<any[]>([])

const roleDialogVisible = ref(false)
const roleDialogTitle = ref('添加角色')
const roleFormRef = ref<FormInstance>()
const currentRole = ref<any>(null)
const submittingRole = ref(false)

const roleForm = reactive<RoleForm>({
  name: '',
  permissions: [],
  description: ''
})

const roleRules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  permissions: [{ required: true, message: '请选择权限', trigger: 'change' }]
}

// 重置密码
const passwordDialogVisible = ref(false)
const passwordFormRef = ref<FormInstance>()
const submittingPassword = ref(false)
const currentPasswordUser = ref<any>(null)

const passwordForm = reactive({
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== passwordForm.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules: FormRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

// Tab 切换
const handleTabChange = (tab: string) => {
  if (tab === 'users') {
    fetchUsers()
  } else if (tab === 'roles') {
    fetchRoles()
  }
}

// 获取系统配置
const fetchSystemConfig = async () => {
  try {
    const res: any = await userApi.getSystemConfig()
    Object.assign(configForm, res)
  } catch (error) {
    console.error('获取系统配置失败:', error)
  }
}

// 保存系统配置
const handleSaveConfig = async () => {
  savingConfig.value = true
  try {
    await userApi.updateSystemConfig(configForm)
    ElMessage.success('保存成功')
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    savingConfig.value = false
  }
}

// 获取用户列表
const fetchUsers = async () => {
  loadingUsers.value = true
  try {
    const res: any = await userApi.getUsers(userQuery)
    users.value = res.items || []
    userTotal.value = res.total || 0
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loadingUsers.value = false
  }
}

// 获取角色列表
const fetchRoles = async () => {
  loadingRoles.value = true
  try {
    const res: any = await userApi.getRoles()
    roles.value = res.items || res || []
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  } finally {
    loadingRoles.value = false
  }
}

// 创建用户
const handleCreateUser = () => {
  userDialogTitle.value = '添加用户'
  currentUser.value = null
  userDialogVisible.value = true
}

// 编辑用户
const handleEditUser = (row: any) => {
  userDialogTitle.value = '编辑用户'
  currentUser.value = row
  Object.assign(userForm, {
    username: row.username,
    email: row.email,
    role_id: row.role_id,
    enabled: row.enabled
  })
  delete userForm.password
  userDialogVisible.value = true
}

// 删除用户
const handleDeleteUser = (row: any) => {
  ElMessageBox.confirm('确定要删除该用户吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await userApi.deleteUser(row.id)
      ElMessage.success('删除成功')
      fetchUsers()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 启用/禁用用户
const handleToggleUser = async (row: any) => {
  try {
    await userApi.toggleUser(row.id, !row.enabled)
    ElMessage.success(row.enabled ? '已禁用' : '已启用')
    fetchUsers()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 重置密码
const handleResetPassword = (row: any) => {
  currentPasswordUser.value = row
  passwordForm.password = ''
  passwordForm.confirmPassword = ''
  passwordDialogVisible.value = true
}

// 提交用户
const handleSubmitUser = async () => {
  if (!userFormRef.value) return
  
  await userFormRef.value.validate(async (valid) => {
    if (valid) {
      submittingUser.value = true
      try {
        if (currentUser.value) {
          await userApi.updateUser(currentUser.value.id, userForm)
          ElMessage.success('更新成功')
        } else {
          await userApi.createUser(userForm)
          ElMessage.success('创建成功')
        }
        userDialogVisible.value = false
        fetchUsers()
      } catch (error: any) {
        ElMessage.error(error.message || '操作失败')
      } finally {
        submittingUser.value = false
      }
    }
  })
}

// 提交密码
const handleSubmitPassword = async () => {
  if (!passwordFormRef.value) return
  
  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      submittingPassword.value = true
      try {
        await userApi.resetPassword(currentPasswordUser.value.id, passwordForm.password)
        ElMessage.success('密码重置成功')
        passwordDialogVisible.value = false
      } catch (error: any) {
        ElMessage.error(error.message || '重置失败')
      } finally {
        submittingPassword.value = false
      }
    }
  })
}

// 重置用户表单
const resetUserForm = () => {
  userFormRef.value?.resetFields()
  Object.assign(userForm, {
    username: '',
    email: '',
    password: '',
    role_id: 2,
    enabled: true
  })
}

// 创建角色
const handleCreateRole = () => {
  roleDialogTitle.value = '添加角色'
  currentRole.value = null
  roleDialogVisible.value = true
}

// 编辑角色
const handleEditRole = (row: any) => {
  roleDialogTitle.value = '编辑角色'
  currentRole.value = row
  Object.assign(roleForm, {
    name: row.name,
    permissions: row.permissions || [],
    description: row.description || ''
  })
  roleDialogVisible.value = true
}

// 删除角色
const handleDeleteRole = (row: any) => {
  ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await userApi.deleteRole(row.id)
      ElMessage.success('删除成功')
      fetchRoles()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 提交角色
const handleSubmitRole = async () => {
  if (!roleFormRef.value) return
  
  await roleFormRef.value.validate(async (valid) => {
    if (valid) {
      submittingRole.value = true
      try {
        if (currentRole.value) {
          await userApi.updateRole(currentRole.value.id, roleForm)
          ElMessage.success('更新成功')
        } else {
          await userApi.createRole(roleForm)
          ElMessage.success('创建成功')
        }
        roleDialogVisible.value = false
        fetchRoles()
      } catch (error: any) {
        ElMessage.error(error.message || '操作失败')
      } finally {
        submittingRole.value = false
      }
    }
  })
}

// 重置角色表单
const resetRoleForm = () => {
  roleFormRef.value?.resetFields()
  Object.assign(roleForm, {
    name: '',
    permissions: [],
    description: ''
  })
}

onMounted(() => {
  fetchSystemConfig()
  fetchRoles()
})
</script>

<style scoped lang="scss">
.settings {
  .toolbar {
    display: flex;
    gap: 10px;
  }

  .mt-20 {
    margin-top: 20px;
  }

  :deep(.el-pagination) {
    margin-top: 20px;
    justify-content: flex-end;
  }

  .permission-group {
    margin: 10px 0;

    h4 {
      margin: 0 0 10px 0;
      font-size: 14px;
      color: var(--text-primary);
    }

    .el-checkbox {
      margin-right: 15px;
      margin-bottom: 10px;
    }
  }
}
</style>
