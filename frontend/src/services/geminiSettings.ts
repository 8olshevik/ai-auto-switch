import { proxyApi } from '@/api'

export interface GeminiProxyStatus {
  enabled: boolean
  base_url: string
}

export const fetchGeminiProxyStatus = async (): Promise<GeminiProxyStatus> => {
  const res = await proxyApi.getStatus('gemini')
  return {
    enabled: res.data.running,
    base_url: res.data.addr || '',
  }
}

export const enableGeminiProxy = async (): Promise<void> => {
  await proxyApi.start('gemini')
}

export const disableGeminiProxy = async (): Promise<void> => {
  await proxyApi.stop('gemini')
}
