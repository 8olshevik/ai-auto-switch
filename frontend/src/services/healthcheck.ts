import { healthApi } from '@/api'

// 健康状态常量
export const HealthStatus = {
  OPERATIONAL: 'operational',
  DEGRADED: 'degraded',
  FAILED: 'failed',
  VALIDATION_ERROR: 'validation_failed',
} as const

// 健康检查结果类型
export interface HealthCheckResult {
  id: number
  providerId: number
  providerName: string
  platform: string
  model?: string
  endpoint?: string
  status: string
  latencyMs: number
  errorMessage: string
  checkedAt: string
}

// Provider 时间线类型
export interface ProviderTimeline {
  providerId: number
  providerName: string
  platform: string
  availabilityMonitorEnabled: boolean
  connectivityAutoBlacklist: boolean
  availabilityConfig?: AvailabilityConfig | null
  items: HealthCheckResult[]
  latest: HealthCheckResult | null
  uptime: number
  avgLatencyMs: number
}

// 可用性高级配置
export interface AvailabilityConfig {
  testModel?: string
  testEndpoint?: string
  timeout?: number
}

/**
 * 获取所有 Provider 的最新状态（按平台分组）
 */
export async function getLatestResults(): Promise<Record<string, ProviderTimeline[]>> {
  const res = await healthApi.getLatest()
  return res.data
}

/**
 * 获取单个 Provider 的历史记录
 */
export async function getHistory(platform: string, providerName: string, limit: number = 20): Promise<any> {
  const res = await healthApi.history({ platform, provider: providerName, limit })
  return res.data
}

/**
 * 手动触发单个 Provider 检测
 */
export async function runSingleCheck(platform: string, providerId: number): Promise<HealthCheckResult> {
  const res = await healthApi.checkSingle(platform, providerId)
  return res.data
}

/**
 * 手动触发全部检测
 */
export async function runAllChecks(): Promise<Record<string, HealthCheckResult[]>> {
  const res = await healthApi.check({ platform: '' })
  return res.data
}

/**
 * 启动后台定时巡检
 */
export async function startBackgroundPolling(): Promise<void> {
  await healthApi.startPolling()
}

/**
 * 停止后台巡检
 */
export async function stopBackgroundPolling(): Promise<void> {
  await healthApi.stopPolling()
}

/**
 * 检查后台巡检是否运行中
 */
export async function isPollingRunning(): Promise<boolean> {
  const res = await healthApi.getPollingStatus()
  return res.data.running
}

/**
 * 启用/禁用指定 Provider 的可用性监控
 */
export async function setAvailabilityMonitorEnabled(
  platform: string,
  providerId: number,
  enabled: boolean
): Promise<void> {
  await healthApi.setMonitorEnabled(platform, providerId, enabled)
}

/**
 * 启用/禁用指定 Provider 的连通性自动拉黑
 */
export async function setConnectivityAutoBlacklist(
  platform: string,
  providerId: number,
  enabled: boolean
): Promise<void> {
  await healthApi.setAutoBlacklist(platform, providerId, enabled)
}

/**
 * 保存 Provider 的可用性高级配置
 */
export async function saveAvailabilityConfig(
  platform: string,
  providerId: number,
  config: AvailabilityConfig
): Promise<void> {
  await healthApi.saveConfig(platform, providerId, config)
}

/**
 * 清理过期的历史记录
 */
export async function cleanupOldRecords(daysToKeep: number = 7): Promise<number> {
  const res = await healthApi.cleanup(daysToKeep)
  return res.data
}

/**
 * 格式化状态为中文
 */
export function formatStatus(status: string): string {
  switch (status) {
    case HealthStatus.OPERATIONAL:
      return '正常'
    case HealthStatus.DEGRADED:
      return '延迟'
    case HealthStatus.FAILED:
      return '故障'
    case HealthStatus.VALIDATION_ERROR:
      return '验证失败'
    default:
      return status
  }
}

/**
 * 获取状态对应的颜色类
 */
export function getStatusColor(status: string): string {
  switch (status) {
    case HealthStatus.OPERATIONAL:
      return 'text-green-500'
    case HealthStatus.DEGRADED:
      return 'text-yellow-500'
    case HealthStatus.FAILED:
      return 'text-red-500'
    case HealthStatus.VALIDATION_ERROR:
      return 'text-red-500'
    default:
      return 'text-gray-500'
  }
}

/**
 * 获取状态对应的图标
 */
export function getStatusIcon(status: string): string {
  switch (status) {
    case HealthStatus.OPERATIONAL:
      return '\u{1F7E2}'
    case HealthStatus.DEGRADED:
      return '\u{1F7E1}'
    case HealthStatus.FAILED:
      return '\u{1F534}'
    case HealthStatus.VALIDATION_ERROR:
      return '\u{1F534}'
    default:
      return '\u{26AB}'
  }
}
