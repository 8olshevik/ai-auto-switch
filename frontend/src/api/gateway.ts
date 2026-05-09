import api from './client'

export interface GatewayKey {
  id: number
  name: string
  keyPrefix: string
  rateLimit: number
  enabled: boolean
  createdAt: string
  lastUsedAt: string | null
}

export interface CreateKeyRequest {
  name: string
  rateLimit?: number
}

export interface CreateKeyResponse {
  key: GatewayKey
  rawKey: string // Only returned once after creation
}

export interface UpdateRateLimitRequest {
  keyId: number
  rateLimit: number
}

export interface GatewayStats {
  totalRequests: number
  bySourceApp: { sourceApp: string; count: number }[]
}

export const gatewayApi = {
  /**
   * 获取 API Key 列表
   */
  listKeys: () =>
    api.get<GatewayKey[]>('/gateway/keys'),

  /**
   * 创建新的 API Key
   * @returns 返回创建的 Key 信息，其中 rawKey 只在创建后返回一次
   */
  createKey: (data: CreateKeyRequest) =>
    api.post<CreateKeyResponse>('/gateway/keys', data),

  /**
   * 删除 API Key
   */
  deleteKey: (id: number) =>
    api.delete(`/gateway/keys/${id}`),

  /**
   * 启用/禁用 API Key
   */
  toggleKey: (id: number, enabled: boolean) =>
    api.put(`/gateway/keys/${id}/toggle`, { enabled }),

  /**
   * 更新 API Key 的速率限制
   */
  updateRateLimit: (data: UpdateRateLimitRequest) =>
    api.put('/gateway/rate-limit', data),

  /**
   * 获取网关使用统计
   * @param sourceApp 可选，按来源应用过滤
   */
  getStats: (sourceApp?: string) => {
    const params = sourceApp ? { source_app: sourceApp } : {}
    return api.get<GatewayStats>('/gateway/stats', { params })
  },
}