import api from './client'

export interface ProxyStatus {
  running: boolean
  addr: string
  platform?: string
}

export const proxyApi = {
  getStatus: (platform?: string) =>
    api.get<ProxyStatus>('/proxy/status', { params: platform ? { platform } : undefined }),

  start: (platform?: string) =>
    api.post('/proxy/start', platform ? { platform } : {}),

  stop: (platform?: string) =>
    api.post('/proxy/stop', platform ? { platform } : {}),

  getLastUsed: () =>
    api.get('/proxy/last-used'),
}
