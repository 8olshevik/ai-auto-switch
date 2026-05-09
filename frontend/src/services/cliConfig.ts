import { cliConfigApi } from '@/api'

// CLI 平台类型
export type CLIPlatform = 'claude' | 'codex' | 'gemini'

// 配置字段信息
export interface CLIConfigField {
  key: string
  value: string
  locked: boolean
  hint?: string
  type: 'string' | 'boolean' | 'object'
  required?: boolean
}

// 配置文件预览
export interface CLIConfigFile {
  path: string
  format?: 'json' | 'toml' | 'env' | string
  content: string
}

// CLI 配置数据
export interface CLIConfig {
  platform: CLIPlatform
  fields: CLIConfigField[]
  rawContent?: string
  rawFiles?: CLIConfigFile[]
  configFormat?: 'json' | 'toml' | 'env'
  envContent?: Record<string, string>
  filePath?: string
  editable?: Record<string, any>
}

// CLI 配置模板
export interface CLITemplate {
  template: Record<string, any>
  isGlobalDefault: boolean
}

// CLI 配置快照
export interface CLIConfigSnapshots {
  currentFiles: CLIConfigFile[]
  previewFiles: CLIConfigFile[]
  mode: 'proxy' | 'direct'
}

// 获取指定平台的 CLI 配置
export async function fetchCLIConfig(platform: CLIPlatform): Promise<CLIConfig> {
  const res = await cliConfigApi.get(platform)
  return res.data
}

// 保存 CLI 配置
export async function saveCLIConfig(platform: CLIPlatform, editable: Record<string, any>): Promise<void> {
  await cliConfigApi.update(platform, editable)
}

// 保存指定配置文件内容
export async function saveCLIConfigFileContent(
  platform: CLIPlatform,
  filePath: string,
  content: string
): Promise<void> {
  await cliConfigApi.saveFileContent(platform, filePath, content)
}

// 获取指定平台的全局模板
export async function fetchCLITemplate(platform: CLIPlatform): Promise<CLITemplate> {
  const res = await cliConfigApi.getTemplate(platform)
  return res.data
}

// 设置指定平台的全局模板
export async function setCLITemplate(
  platform: CLIPlatform,
  template: Record<string, any>,
  isGlobalDefault: boolean
): Promise<void> {
  await cliConfigApi.setTemplate(platform, template, isGlobalDefault)
}

// 获取指定平台的锁定字段列表
export async function fetchLockedFields(platform: CLIPlatform): Promise<string[]> {
  const res = await cliConfigApi.getLockedFields(platform)
  return res.data
}

// 恢复默认配置
export async function restoreDefaultConfig(platform: CLIPlatform): Promise<void> {
  await cliConfigApi.restoreDefault(platform)
}

// 获取配置快照
export async function fetchCLIConfigSnapshots(
  platform: CLIPlatform,
  apiUrl: string = '',
  apiKey: string = '',
  previewMode: 'current' | 'direct' | 'proxy' | '' = ''
): Promise<CLIConfigSnapshots> {
  const params: Record<string, string> = {}
  if (apiUrl) params.apiUrl = apiUrl
  if (apiKey) params.apiKey = apiKey
  if (previewMode) params.previewMode = previewMode
  const res = await cliConfigApi.getSnapshots(platform, params)
  return res.data
}
