// 认证配置类型定义

// SNMP v1/v2c 认证配置
export interface SNMPv1v2cAuth {
  community: string
}

// SNMP v3 认证配置
export interface SNMPv3Auth {
  username: string
  security_level: 'noAuthNoPriv' | 'authNoPriv' | 'authPriv'
  auth_protocol?: 'MD5' | 'SHA' | 'SHA224' | 'SHA256' | 'SHA384' | 'SHA512'
  auth_password?: string
  priv_protocol?: 'DES' | 'AES' | 'AES192' | 'AES256'
  priv_password?: string
  context_name?: string
  context_engine_id?: string
}

// SSH 认证配置
export interface SSHAuth {
  username: string
  // 密码认证
  password?: string
  // 密钥认证
  private_key?: string
  passphrase?: string // 私钥密码
  // 认证方式
  auth_method?: 'password' | 'key' | 'both'
}

// HTTP/HTTPS 认证配置
export interface HTTPAuth {
  // Basic Auth
  username?: string
  password?: string
  // API Key
  api_key?: string
  api_key_header?: string // 默认 X-API-Key
  // Bearer Token
  bearer_token?: string
  // 自定义 Header
  custom_headers?: Record<string, string>
}

// 通用认证配置（根据协议类型使用不同的配置）
export type AuthConfig = 
  | { type: 'snmp_v1' | 'snmp_v2c'; config: SNMPv1v2cAuth }
  | { type: 'snmp_v3'; config: SNMPv3Auth }
  | { type: 'ssh'; config: SSHAuth }
  | { type: 'http' | 'https'; config: HTTPAuth }
  | { type: 'none'; config: Record<string, never> }

// 设备连接配置（包含认证信息）
export interface DeviceConnectionConfig {
  host: string
  port?: number
  protocol: 'snmp' | 'ssh' | 'http' | 'https' | 'telnet' | 'modbus' | 'opcua'
  // SNMP 特定配置
  snmp_version?: '1' | '2c' | '3'
  // 认证配置（根据协议动态使用）
  auth?: AuthConfig
  // 其他连接配置
  timeout?: number
  retries?: number
  [key: string]: any
}

