import api from './client'

export interface LogsListParams {
  page?: number
  pageSize?: number
  keyword?: string
  startDate?: string
  endDate?: string
  platform?: string
  provider?: string
  limit?: number
}

export interface LogsListResponse {
  data: any[]
  total: number
  page: number
  pageSize: number
}

export const logsApi = {
  list: (params?: LogsListParams) =>
    api.get('/logs/', { params }),

  stats: (platform?: string) =>
    api.get('/logs/stats', { params: platform ? { platform } : undefined }),

  heatmap: (days?: number) =>
    api.get('/logs/heatmap', { params: days ? { days } : undefined }),

  clear: () =>
    api.delete('/logs/'),

  providers: (platform?: string) =>
    api.get<string[]>('/logs/providers', { params: platform ? { platform } : undefined }),

  cost: (since: string, platform?: string) =>
    api.get<{ cost: number }>('/logs/cost', { params: { since, ...(platform ? { platform } : {}) } }),

  providerStats: (platform?: string) =>
    api.get('/logs/provider-stats', { params: platform ? { platform } : undefined }),
}
