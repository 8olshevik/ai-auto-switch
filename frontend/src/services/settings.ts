import { settingsApi } from '@/api'

export interface BlacklistSettings {
  failureThreshold: number
  durationMinutes: number
}

/**
 * 获取拉黑配置
 */
export const getBlacklistSettings = async (): Promise<BlacklistSettings> => {
  const res = await settingsApi.getBlacklist()
  return res.data
}

/**
 * 更新拉黑配置
 */
export const updateBlacklistSettings = async (
  threshold: number,
  duration: number
): Promise<void> => {
  await settingsApi.updateBlacklist(threshold, duration)
}

/**
 * 获取等级拉黑开关状态
 */
export const getLevelBlacklistEnabled = async (): Promise<boolean> => {
  const res = await settingsApi.getLevelBlacklistEnabled()
  return res.data.enabled
}

/**
 * 设置等级拉黑开关状态
 */
export const setLevelBlacklistEnabled = async (enabled: boolean): Promise<void> => {
  await settingsApi.setLevelBlacklistEnabled(enabled)
}

/**
 * 获取拉黑功能总开关状态
 */
export const getBlacklistEnabled = async (): Promise<boolean> => {
  const res = await settingsApi.getBlacklistEnabled()
  return res.data.enabled
}

/**
 * 设置拉黑功能总开关状态
 */
export const setBlacklistEnabled = async (enabled: boolean): Promise<void> => {
  await settingsApi.setBlacklistEnabled(enabled)
}
