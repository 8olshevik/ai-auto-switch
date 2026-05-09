import api from './client'

export interface HealthCheckRequest {
  platform: string
  provider_id?: number
}

export interface HealthHistoryParams {
  platform: string
  provider: string
  limit?: number
}

export const healthApi = {
  /** Get latest results for all providers (grouped by platform) */
  getLatest: () =>
    api.get('/health/'),

  /** Run health check for a platform or single provider */
  check: (req: HealthCheckRequest) =>
    api.post('/health/check', req),

  /** Run single provider check */
  checkSingle: (platform: string, providerId: number) =>
    api.post(`/health/check/${providerId}`, { platform }),

  /** Get health check history */
  history: (params: HealthHistoryParams) =>
    api.get('/health/history', { params }),

  /** Start background polling */
  startPolling: () =>
    api.post('/health/polling/start'),

  /** Stop background polling */
  stopPolling: () =>
    api.post('/health/polling/stop'),

  /** Get polling status */
  getPollingStatus: () =>
    api.get<{ running: boolean }>('/health/polling/status'),

  /** Set availability monitor enabled for a provider */
  setMonitorEnabled: (platform: string, providerId: number, enabled: boolean) =>
    api.put(`/health/providers/${providerId}/monitor`, { platform, enabled }),

  /** Set connectivity auto-blacklist for a provider */
  setAutoBlacklist: (platform: string, providerId: number, enabled: boolean) =>
    api.put(`/health/providers/${providerId}/auto-blacklist`, { platform, enabled }),

  /** Save availability config for a provider */
  saveConfig: (platform: string, providerId: number, config: any) =>
    api.put(`/health/providers/${providerId}/config`, { platform, ...config }),

  /** Cleanup old records */
  cleanup: (daysToKeep: number) =>
    api.delete('/health/history', { params: { days: daysToKeep } }),

  /** Get/set auto-test enabled */
  getAutoTest: () =>
    api.get<{ enabled: boolean }>('/health/auto-test'),

  setAutoTest: (enabled: boolean) =>
    api.put('/health/auto-test', { enabled }),
}
