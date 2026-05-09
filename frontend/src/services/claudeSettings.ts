import { proxyApi } from '@/api'

export interface ClaudeProxyStatus {
  enabled: boolean
  base_url: string
}

type Platform = 'claude' | 'codex'

export const fetchProxyStatus = async (platform: Platform): Promise<ClaudeProxyStatus> => {
  const res = await proxyApi.getStatus(platform)
  return {
    enabled: res.data.running,
    base_url: res.data.addr || '',
  }
}

export const enableProxy = async (platform: Platform): Promise<void> => {
  await proxyApi.start(platform)
}

export const disableProxy = async (platform: Platform): Promise<void> => {
  await proxyApi.stop(platform)
}
