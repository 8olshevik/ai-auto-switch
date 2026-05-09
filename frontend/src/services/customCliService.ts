import { cliConfigApi } from '@/api'

// 类型定义
export interface ConfigFile {
  id: string
  label: string
  path: string
  format: 'json' | 'toml' | 'env'
  isPrimary?: boolean
}

export interface ProxyInjection {
  targetFileId: string
  baseUrlField: string
  authTokenField?: string
}

export interface CustomCliTool {
  id: string
  name: string
  configFiles: ConfigFile[]
  proxyInjection?: ProxyInjection[]
}

export interface CustomCliProxyStatus {
  enabled: boolean
  baseUrl: string
}

// ========== 工具 CRUD ==========

export const listCustomCliTools = async (): Promise<CustomCliTool[]> => {
  const res = await cliConfigApi.listCustomTools()
  return res.data
}

export const getCustomCliTool = async (id: string): Promise<CustomCliTool | null> => {
  const res = await cliConfigApi.getCustomTool(id)
  return res.data
}

export const createCustomCliTool = async (tool: Omit<CustomCliTool, 'id'>): Promise<CustomCliTool> => {
  const res = await cliConfigApi.createCustomTool(tool)
  return res.data
}

export const updateCustomCliTool = async (id: string, tool: CustomCliTool): Promise<void> => {
  await cliConfigApi.updateCustomTool(id, tool)
}

export const deleteCustomCliTool = async (id: string): Promise<void> => {
  await cliConfigApi.deleteCustomTool(id)
}

// ========== 代理管理 ==========

export const getCustomCliProxyStatus = async (toolId: string): Promise<CustomCliProxyStatus> => {
  const res = await cliConfigApi.getCustomToolProxy(toolId)
  return res.data
}

export const enableCustomCliProxy = async (toolId: string): Promise<void> => {
  await cliConfigApi.enableCustomToolProxy(toolId)
}

export const disableCustomCliProxy = async (toolId: string): Promise<void> => {
  await cliConfigApi.disableCustomToolProxy(toolId)
}

// ========== 配置文件读写 ==========

export const getCustomCliConfigContent = async (toolId: string, fileId: string): Promise<string> => {
  const res = await cliConfigApi.getCustomToolFileContent(toolId, fileId)
  return res.data
}

export const saveCustomCliConfigContent = async (toolId: string, fileId: string, content: string): Promise<void> => {
  await cliConfigApi.saveCustomToolFileContent(toolId, fileId, content)
}

export const getCustomCliLockedFields = async (toolId: string): Promise<string[]> => {
  const res = await cliConfigApi.getCustomToolLockedFields(toolId)
  return res.data
}
