import { blacklistApi } from '@/api'

// 黑名单状态接口
export interface BlacklistStatus {
  platform: string
  providerName: string
  failureCount: number
  blacklistedAt?: string
  blacklistedUntil?: string
  lastFailureAt?: string
  isBlacklisted: boolean
  remainingSeconds: number
  blacklistLevel: number
  lastRecoveredAt?: string
  forgivenessRemaining: number
}

// 黑名单配置接口
export interface BlacklistSettings {
  failureThreshold: number
  durationMinutes: number
}

/**
 * 获取指定平台的黑名单状态列表
 */
export const getBlacklistStatus = async (platform: string): Promise<BlacklistStatus[]> => {
  const res = await blacklistApi.getStatus(platform)
  return res.data
}

/**
 * 手动解除拉黑
 */
export const manualUnblock = async (platform: string, providerName: string): Promise<void> => {
  await blacklistApi.recover(platform, providerName)
}

/**
 * 获取黑名单配置
 */
export const getBlacklistSettings = async (): Promise<BlacklistSettings> => {
  const res = await blacklistApi.getSettings()
  return res.data
}

/**
 * 更新黑名单配置
 */
export const updateBlacklistSettings = async (threshold: number, duration: number): Promise<void> => {
  await blacklistApi.updateSettings(threshold, duration)
}
