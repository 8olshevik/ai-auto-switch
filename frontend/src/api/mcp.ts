import api from './client'

export interface MCPServer {
  name: string
  type: string
  url?: string
  command?: string
  args?: string[]
  env?: Record<string, string>
}

export const mcpApi = {
  list: () =>
    api.get('/mcp/'),

  create: (server: MCPServer) =>
    api.post('/mcp/', server),

  update: (id: string, server: MCPServer) =>
    api.put(`/mcp/${id}`, server),

  delete: (id: string) =>
    api.delete(`/mcp/${id}`),

  /** Save all MCP servers (bulk update) */
  saveAll: (servers: any[]) =>
    api.post('/mcp/bulk', servers),

  /** Parse MCP JSON string */
  parse: (jsonStr: string) =>
    api.post('/mcp/parse', { json: jsonStr }),

  /** Import MCP servers with conflict strategy */
  import: (servers: any[], strategy: string) =>
    api.post('/mcp/import', { servers, strategy }),
}
