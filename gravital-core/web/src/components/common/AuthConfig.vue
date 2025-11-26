<template>
  <div class="auth-config">
    <!-- SNMP v1/v2c 认证 -->
    <template v-if="protocol === 'snmp' && (snmpVersion === '1' || snmpVersion === '2c')">
      <el-form-item label="Community" prop="auth.config.community">
        <el-input
          v-model="authConfig.community"
          placeholder="请输入 SNMP Community (默认: public)"
          show-password
        />
      </el-form-item>
    </template>

    <!-- SNMP v3 认证 -->
    <template v-if="protocol === 'snmp' && snmpVersion === '3'">
      <el-form-item label="用户名" prop="auth.config.username">
        <el-input
          v-model="authConfig.username"
          placeholder="请输入 SNMP v3 用户名"
        />
      </el-form-item>

      <el-form-item label="安全级别" prop="auth.config.security_level">
        <el-select
          v-model="authConfig.security_level"
          placeholder="请选择安全级别"
          @change="handleSecurityLevelChange"
        >
          <el-option label="noAuthNoPriv (无认证无加密)" value="noAuthNoPriv" />
          <el-option label="authNoPriv (认证无加密)" value="authNoPriv" />
          <el-option label="authPriv (认证加密)" value="authPriv" />
        </el-select>
      </el-form-item>

      <template v-if="authConfig.security_level === 'authNoPriv' || authConfig.security_level === 'authPriv'">
        <el-form-item label="认证协议" prop="auth.config.auth_protocol">
          <el-select v-model="authConfig.auth_protocol" placeholder="请选择认证协议">
            <el-option label="MD5" value="MD5" />
            <el-option label="SHA" value="SHA" />
            <el-option label="SHA224" value="SHA224" />
            <el-option label="SHA256" value="SHA256" />
            <el-option label="SHA384" value="SHA384" />
            <el-option label="SHA512" value="SHA512" />
          </el-select>
        </el-form-item>

        <el-form-item label="认证密码" prop="auth.config.auth_password">
          <el-input
            v-model="authConfig.auth_password"
            type="password"
            placeholder="请输入认证密码"
            show-password
          />
        </el-form-item>
      </template>

      <template v-if="authConfig.security_level === 'authPriv'">
        <el-form-item label="隐私协议" prop="auth.config.priv_protocol">
          <el-select v-model="authConfig.priv_protocol" placeholder="请选择隐私协议">
            <el-option label="DES" value="DES" />
            <el-option label="AES" value="AES" />
            <el-option label="AES192" value="AES192" />
            <el-option label="AES256" value="AES256" />
          </el-select>
        </el-form-item>

        <el-form-item label="隐私密码" prop="auth.config.priv_password">
          <el-input
            v-model="authConfig.priv_password"
            type="password"
            placeholder="请输入隐私密码"
            show-password
          />
        </el-form-item>
      </template>

      <el-form-item label="Context Name" prop="auth.config.context_name">
        <el-input
          v-model="authConfig.context_name"
          placeholder="可选：Context Name"
        />
      </el-form-item>
    </template>

    <!-- SSH 认证 -->
    <template v-if="protocol === 'ssh'">
      <el-form-item label="认证方式" prop="auth.config.auth_method">
        <el-select
          v-model="authConfig.auth_method"
          placeholder="请选择认证方式"
          @change="handleSSHAuthMethodChange"
        >
          <el-option label="密码认证" value="password" />
          <el-option label="密钥认证" value="key" />
          <el-option label="密码或密钥" value="both" />
        </el-select>
      </el-form-item>

      <el-form-item label="用户名" prop="auth.config.username">
        <el-input
          v-model="authConfig.username"
          placeholder="请输入 SSH 用户名"
        />
      </el-form-item>

      <template v-if="authConfig.auth_method === 'password' || authConfig.auth_method === 'both'">
        <el-form-item label="密码" prop="auth.config.password">
          <el-input
            v-model="authConfig.password"
            type="password"
            placeholder="请输入 SSH 密码"
            show-password
          />
        </el-form-item>
      </template>

      <template v-if="authConfig.auth_method === 'key' || authConfig.auth_method === 'both'">
        <el-form-item label="私钥" prop="auth.config.private_key">
          <el-input
            v-model="authConfig.private_key"
            type="textarea"
            :rows="4"
            placeholder="请输入 SSH 私钥内容"
          />
        </el-form-item>

        <el-form-item label="私钥密码" prop="auth.config.passphrase">
          <el-input
            v-model="authConfig.passphrase"
            type="password"
            placeholder="可选：私钥密码"
            show-password
          />
        </el-form-item>
      </template>
    </template>

    <!-- HTTP/HTTPS 认证 -->
    <template v-if="protocol === 'http' || protocol === 'https'">
      <el-form-item label="认证类型" prop="auth.config.auth_type">
        <el-select
          v-model="authConfig.auth_type"
          placeholder="请选择认证类型"
          @change="handleHTTPAuthTypeChange"
        >
          <el-option label="无认证" value="none" />
          <el-option label="Basic Auth" value="basic" />
          <el-option label="API Key" value="apikey" />
          <el-option label="Bearer Token" value="bearer" />
          <el-option label="自定义 Header" value="custom" />
        </el-select>
      </el-form-item>

      <template v-if="authConfig.auth_type === 'basic'">
        <el-form-item label="用户名" prop="auth.config.username">
          <el-input
            v-model="authConfig.username"
            placeholder="请输入用户名"
          />
        </el-form-item>

        <el-form-item label="密码" prop="auth.config.password">
          <el-input
            v-model="authConfig.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
      </template>

      <template v-if="authConfig.auth_type === 'apikey'">
        <el-form-item label="API Key" prop="auth.config.api_key">
          <el-input
            v-model="authConfig.api_key"
            type="password"
            placeholder="请输入 API Key"
            show-password
          />
        </el-form-item>

        <el-form-item label="Header 名称" prop="auth.config.api_key_header">
          <el-input
            v-model="authConfig.api_key_header"
            placeholder="默认: X-API-Key"
          />
        </el-form-item>
      </template>

      <template v-if="authConfig.auth_type === 'bearer'">
        <el-form-item label="Bearer Token" prop="auth.config.bearer_token">
          <el-input
            v-model="authConfig.bearer_token"
            type="password"
            placeholder="请输入 Bearer Token"
            show-password
          />
        </el-form-item>
      </template>

      <template v-if="authConfig.auth_type === 'custom'">
        <el-form-item label="自定义 Header">
          <div
            v-for="(item, index) in customHeaders"
            :key="index"
            class="custom-header-item"
          >
            <el-input
              v-model="item.key"
              placeholder="Header 名称"
              style="width: 200px; margin-right: 10px"
            />
            <el-input
              v-model="item.value"
              type="password"
              placeholder="Header 值"
              show-password
              style="flex: 1"
            />
            <el-button
              type="danger"
              :icon="Delete"
              @click="removeCustomHeader(index)"
            />
          </div>
          <el-button
            type="primary"
            :icon="Plus"
            @click="addCustomHeader"
          >
            添加 Header
          </el-button>
        </el-form-item>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { SNMPv1v2cAuth, SNMPv3Auth, SSHAuth, HTTPAuth } from '@/types/auth'

interface Props {
  protocol: string
  modelValue: any
  snmpVersion?: string
}

const props = withDefaults(defineProps<Props>(), {
  snmpVersion: '2c'
})

const emit = defineEmits<{
  'update:modelValue': [value: any]
}>()

// 认证配置（根据协议类型动态使用）
const authConfig = computed({
  get: () => {
    if (!props.modelValue?.auth) {
      return getDefaultAuthConfig()
    }
    return props.modelValue.auth.config || getDefaultAuthConfig()
  },
  set: (value) => {
    emit('update:modelValue', {
      ...props.modelValue,
      auth: {
        type: getAuthType(),
        config: value
      }
    })
  }
})

// 自定义 Header 的键值对
const customHeaders = ref<Array<{ key: string; value: string }>>([])

// 获取默认认证配置
function getDefaultAuthConfig() {
  if (props.protocol === 'snmp') {
    if (props.snmpVersion === '3') {
      return {
        username: '',
        security_level: 'noAuthNoPriv' as const,
        auth_protocol: 'MD5' as const,
        auth_password: '',
        priv_protocol: 'DES' as const,
        priv_password: '',
        context_name: ''
      } as SNMPv3Auth
    } else {
      return {
        community: 'public'
      } as SNMPv1v2cAuth
    }
  } else if (props.protocol === 'ssh') {
    return {
      username: '',
      password: '',
      private_key: '',
      passphrase: '',
      auth_method: 'password' as const
    } as SSHAuth
  } else if (props.protocol === 'http' || props.protocol === 'https') {
    return {
      auth_type: 'none' as const,
      username: '',
      password: '',
      api_key: '',
      api_key_header: 'X-API-Key',
      bearer_token: '',
      custom_headers: {} as Record<string, string>
    } as HTTPAuth & { auth_type: string }
  }
  return {}
}

// 获取认证类型
function getAuthType(): string {
  if (props.protocol === 'snmp') {
    if (props.snmpVersion === '3') {
      return 'snmp_v3'
    } else {
      return `snmp_${props.snmpVersion}`
    }
  } else if (props.protocol === 'ssh') {
    return 'ssh'
  } else if (props.protocol === 'http' || props.protocol === 'https') {
    return props.protocol
  }
  return 'none'
}

// SNMP v3 安全级别变化
function handleSecurityLevelChange() {
  // 当安全级别降低时，清除不需要的字段
  if (authConfig.value.security_level === 'noAuthNoPriv') {
    authConfig.value.auth_protocol = undefined
    authConfig.value.auth_password = undefined
    authConfig.value.priv_protocol = undefined
    authConfig.value.priv_password = undefined
  } else if (authConfig.value.security_level === 'authNoPriv') {
    authConfig.value.priv_protocol = undefined
    authConfig.value.priv_password = undefined
  }
}

// SSH 认证方式变化
function handleSSHAuthMethodChange() {
  if (authConfig.value.auth_method === 'password') {
    authConfig.value.private_key = undefined
    authConfig.value.passphrase = undefined
  } else if (authConfig.value.auth_method === 'key') {
    authConfig.value.password = undefined
  }
}

// HTTP 认证类型变化
function handleHTTPAuthTypeChange() {
  // 清除其他认证类型的字段
  if (authConfig.value.auth_type !== 'basic') {
    authConfig.value.username = undefined
    authConfig.value.password = undefined
  }
  if (authConfig.value.auth_type !== 'apikey') {
    authConfig.value.api_key = undefined
    authConfig.value.api_key_header = undefined
  }
  if (authConfig.value.auth_type !== 'bearer') {
    authConfig.value.bearer_token = undefined
  }
  if (authConfig.value.auth_type !== 'custom') {
    authConfig.value.custom_headers = {}
    customHeaders.value = []
  }
}

// 添加自定义 Header
function addCustomHeader() {
  customHeaders.value.push({ key: '', value: '' })
}

// 删除自定义 Header
function removeCustomHeader(index: number) {
  customHeaders.value.splice(index, 1)
  updateCustomHeaders()
}

// 更新 custom_headers
function updateCustomHeaders() {
  const headers: Record<string, string> = {}
  customHeaders.value.forEach((item) => {
    if (item.key && item.value) {
      headers[item.key] = item.value
    }
  })
  authConfig.value.custom_headers = headers
}

// 监听自定义 Header 变化
watch(customHeaders, () => {
  updateCustomHeaders()
}, { deep: true })

// 初始化自定义 Header（如果已有数据）
watch(() => authConfig.value.custom_headers, (headers) => {
  if (authConfig.value.auth_type === 'custom' && headers) {
    customHeaders.value = Object.entries(headers).map(([key, value]) => ({
      key,
      value: value as string
    }))
  }
}, { immediate: true })

// 监听协议变化，重置认证配置
watch(() => props.protocol, () => {
  customHeaders.value = []
  emit('update:modelValue', {
    ...props.modelValue,
    auth: {
      type: getAuthType(),
      config: getDefaultAuthConfig()
    }
  })
})

// 监听 SNMP 版本变化
watch(() => props.snmpVersion, () => {
  if (props.protocol === 'snmp') {
    emit('update:modelValue', {
      ...props.modelValue,
      auth: {
        type: getAuthType(),
        config: getDefaultAuthConfig()
      }
    })
  }
})
</script>

<style scoped lang="scss">
.auth-config {
  .custom-header-item {
    display: flex;
    align-items: center;
    margin-bottom: 10px;
  }
}
</style>

