import api from './client'

export interface SystemSettings {
  proxyPort: number
  proxyListenAddr: string
  logLevel: string
  databasePath: string
  port: number
}

export interface UpdateSettingsRequest {
  proxyPort?: number
  proxyListenAddr?: string
  logLevel?: string
  databasePath?: string
  port?: number
}

export interface BlacklistSettingsResponse {
  failureThreshold: number
  durationMinutes: number
}

export const settingsApi = {
  get: () =>
    api.get<SystemSettings>('/settings/'),

  update: (settings: UpdateSettingsRequest) =>
    api.put<SystemSettings>('/settings/', settings),

  getApp: () =>
    api.get('/settings/app'),

  updateApp: (settings: any) =>
    api.put('/settings/app', settings),

  getBlacklist: () =>
    api.get<BlacklistSettingsResponse>('/settings/blacklist'),

  updateBlacklist: (threshold: number, duration: number) =>
    api.put('/settings/blacklist', { failureThreshold: threshold, durationMinutes: duration }),

  getLevelBlacklistEnabled: () =>
    api.get<{ enabled: boolean }>('/settings/level-blacklist'),

  setLevelBlacklistEnabled: (enabled: boolean) =>
    api.put('/settings/level-blacklist', { enabled }),

  getBlacklistEnabled: () =>
    api.get<{ enabled: boolean }>('/settings/blacklist-enabled'),

  setBlacklistEnabled: (enabled: boolean) =>
    api.put('/settings/blacklist-enabled', { enabled }),
}
