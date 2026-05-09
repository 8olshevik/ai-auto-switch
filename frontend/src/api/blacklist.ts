import api from './client'

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

export interface BlacklistSettings {
  failureThreshold: number
  durationMinutes: number
}

export const blacklistApi = {
  getStatus: (platform: string) =>
    api.get<BlacklistStatus[]>('/blacklist/', { params: { platform } }),

  recover: (platform: string, providerName: string) =>
    api.post('/blacklist/recover', { platform, providerName }),

  getSettings: () =>
    api.get<BlacklistSettings>('/blacklist/settings'),

  updateSettings: (threshold: number, duration: number) =>
    api.put('/blacklist/settings', { failureThreshold: threshold, durationMinutes: duration }),
}
