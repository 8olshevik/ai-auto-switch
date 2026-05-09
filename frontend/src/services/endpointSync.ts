/**
 * 端点同步服务
 * 从 Claude、Codex、Gemini 三个平台获取供应商 API 端点
 */

import { providerApi } from '@/api'

/**
 * 同步的端点数据结构
 */
export interface SyncedEndpoint {
  url: string                              // 标准化的基础 URL
  source: 'claude' | 'codex' | 'gemini'   // 来源平台
  providerName: string                     // 供应商名称
}

/**
 * 提取 API URL 的基础地址（去除路径部分）
 */
function extractBaseUrl(apiUrl: string): string {
  if (!apiUrl || !apiUrl.trim()) {
    return ''
  }

  try {
    const url = new URL(apiUrl)
    return `${url.protocol}//${url.host}`
  } catch {
    const trimmed = apiUrl.trim()
    const versionIndex = trimmed.indexOf('/v1')
    if (versionIndex > 0) {
      return trimmed.substring(0, versionIndex)
    }
    console.warn('无效的 API URL:', apiUrl)
    return ''
  }
}

/**
 * 从所有供应商服务获取端点列表
 */
export async function fetchAllProviderEndpoints(): Promise<SyncedEndpoint[]> {
  const endpoints: SyncedEndpoint[] = []
  const platforms: Array<'claude' | 'codex' | 'gemini'> = ['claude', 'codex', 'gemini']

  const results = await Promise.allSettled(
    platforms.map(platform => providerApi.load(platform))
  )

  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      const providers = result.value.data
      if (Array.isArray(providers)) {
        for (const provider of providers) {
          const baseUrl = extractBaseUrl(provider.api_url || provider.apiUrl || '')
          if (baseUrl) {
            endpoints.push({
              url: baseUrl,
              source: platforms[index],
              providerName: provider.name || '',
            })
          }
        }
      }
    }
  })

  return endpoints
}
