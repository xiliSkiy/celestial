/**
 * 格式化工具函数
 */

/**
 * 格式化日期时间
 * @param dateString 日期字符串
 * @param format 格式化模板，默认 'YYYY-MM-DD HH:mm:ss'
 * @returns 格式化后的日期时间字符串
 */
export function formatDateTime(dateString: string | Date, format = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!dateString) return '-'
  
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString
  
  if (isNaN(date.getTime())) return '-'
  
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  
  return format
    .replace('YYYY', String(year))
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

/**
 * 格式化日期
 * @param dateString 日期字符串
 * @returns 格式化后的日期字符串
 */
export function formatDate(dateString: string | Date): string {
  return formatDateTime(dateString, 'YYYY-MM-DD')
}

/**
 * 格式化时间
 * @param dateString 日期字符串
 * @returns 格式化后的时间字符串
 */
export function formatTime(dateString: string | Date): string {
  return formatDateTime(dateString, 'HH:mm:ss')
}

/**
 * 格式化相对时间
 * @param dateString 日期字符串
 * @returns 相对时间字符串，如 "3分钟前"
 */
export function formatRelativeTime(dateString: string | Date): string {
  if (!dateString) return '-'
  
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString
  
  if (isNaN(date.getTime())) return '-'
  
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  if (days > 7) {
    return formatDate(date)
  } else if (days > 0) {
    return `${days}天前`
  } else if (hours > 0) {
    return `${hours}小时前`
  } else if (minutes > 0) {
    return `${minutes}分钟前`
  } else if (seconds > 0) {
    return `${seconds}秒前`
  } else {
    return '刚刚'
  }
}

/**
 * 格式化文件大小
 * @param bytes 字节数
 * @returns 格式化后的文件大小字符串
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

/**
 * 格式化数字
 * @param num 数字
 * @param decimals 小数位数
 * @returns 格式化后的数字字符串
 */
export function formatNumber(num: number, decimals = 2): string {
  if (num === null || num === undefined) return '-'
  return num.toFixed(decimals)
}

/**
 * 格式化百分比
 * @param value 数值（0-100）
 * @param decimals 小数位数
 * @returns 格式化后的百分比字符串
 */
export function formatPercent(value: number, decimals = 2): string {
  if (value === null || value === undefined) return '-'
  return `${value.toFixed(decimals)}%`
}

/**
 * 格式化带宽
 * @param bps 比特每秒
 * @returns 格式化后的带宽字符串
 */
export function formatBandwidth(bps: number): string {
  if (!bps) return '-'
  
  const k = 1000
  const sizes = ['bps', 'Kbps', 'Mbps', 'Gbps', 'Tbps']
  const i = Math.floor(Math.log(bps) / Math.log(k))
  
  return `${(bps / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

/**
 * 格式化持续时间
 * @param seconds 秒数
 * @returns 格式化后的持续时间字符串
 */
export function formatDuration(seconds: number): string {
  if (!seconds) return '-'
  
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  
  const parts: string[] = []
  if (days > 0) parts.push(`${days}天`)
  if (hours > 0) parts.push(`${hours}小时`)
  if (minutes > 0) parts.push(`${minutes}分钟`)
  if (secs > 0 || parts.length === 0) parts.push(`${secs}秒`)
  
  return parts.join(' ')
}

/**
 * 格式化 IP 地址
 * @param ip IP 地址
 * @returns 格式化后的 IP 地址字符串
 */
export function formatIP(ip: string): string {
  if (!ip) return '-'
  return ip
}

/**
 * 截断字符串
 * @param str 字符串
 * @param maxLength 最大长度
 * @param suffix 后缀
 * @returns 截断后的字符串
 */
export function truncate(str: string, maxLength: number, suffix = '...'): string {
  if (!str) return ''
  if (str.length <= maxLength) return str
  return str.substring(0, maxLength - suffix.length) + suffix
}

