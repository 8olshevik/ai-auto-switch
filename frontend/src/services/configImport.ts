import { importApi } from '@/api'

export type ConfigImportStatus = {
  config_exists: boolean
  config_path?: string
  pending_providers: boolean
  pending_mcp: boolean
  pending_provider_count: number
  pending_mcp_count: number
}

export type ConfigImportResult = {
  status: ConfigImportStatus
  imported_providers: number
  imported_mcp: number
}

export const fetchConfigImportStatus = async (): Promise<ConfigImportStatus> => {
  const res = await importApi.getStatus()
  return res.data
}

export const importFromCcSwitch = async (): Promise<ConfigImportResult> => {
  const res = await importApi.importConfig()
  return res.data
}

// 从指定路径导入配置
export const importFromPath = async (path: string): Promise<ConfigImportResult> => {
  const res = await importApi.importConfig(path)
  return res.data
}

// 检查是否首次使用
export const isFirstRun = async (): Promise<boolean> => {
  const res = await importApi.isFirstRun()
  return res.data.firstRun
}

// 标记首次使用已完成
export const markFirstRunDone = async (): Promise<void> => {
  await importApi.markFirstRunDone()
}
