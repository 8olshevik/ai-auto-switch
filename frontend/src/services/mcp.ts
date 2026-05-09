import { mcpApi } from '@/api'

export type McpPlatform = 'claude-code' | 'codex' | 'gemini'
export type McpServerType = 'stdio' | 'http'

export type McpServer = {
  name: string
  type: McpServerType
  command?: string
  args: string[]
  env: Record<string, string>
  url?: string
  website?: string
  tips?: string
  enable_platform: McpPlatform[]
  enabled_in_claude: boolean
  enabled_in_codex: boolean
  enabled_in_gemini: boolean
  missing_placeholders: string[]
}

export const fetchMcpServers = async (): Promise<McpServer[]> => {
  const res = await mcpApi.list()
  return res.data
}

export const saveMcpServers = async (servers: McpServer[]): Promise<void> => {
  await mcpApi.saveAll(servers)
}

export type McpParseResult = {
  servers: McpServer[]
  conflicts: string[]
  needName: boolean
}

export type ConflictStrategy = 'skip' | 'overwrite'

export const parseMcpJSON = async (jsonStr: string): Promise<McpParseResult | null> => {
  const res = await mcpApi.parse(jsonStr)
  return res.data
}

export const importMcpServers = async (
  servers: McpServer[],
  strategy: ConflictStrategy
): Promise<number> => {
  const res = await mcpApi.import(servers, strategy)
  return res.data.count ?? res.data
}
