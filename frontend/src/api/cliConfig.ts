import api from './client'

export type CLIPlatform = 'claude' | 'codex' | 'gemini'

export const cliConfigApi = {
  get: (platform: CLIPlatform) =>
    api.get(`/cli-config/${platform}`),

  update: (platform: CLIPlatform, editable: Record<string, any>) =>
    api.put(`/cli-config/${platform}`, editable),

  saveFileContent: (platform: CLIPlatform, filePath: string, content: string) =>
    api.put(`/cli-config/${platform}/file`, { filePath, content }),

  getTemplate: (platform: CLIPlatform) =>
    api.get(`/cli-config/${platform}/template`),

  setTemplate: (platform: CLIPlatform, template: Record<string, any>, isGlobalDefault: boolean) =>
    api.put(`/cli-config/${platform}/template`, { template, isGlobalDefault }),

  getLockedFields: (platform: CLIPlatform) =>
    api.get<string[]>(`/cli-config/${platform}/locked-fields`),

  restoreDefault: (platform: CLIPlatform) =>
    api.post(`/cli-config/${platform}/restore`),

  getSnapshots: (platform: CLIPlatform, params?: { apiUrl?: string; apiKey?: string; previewMode?: string }) =>
    api.get(`/cli-config/${platform}/snapshots`, { params }),

  // Custom CLI tools
  listCustomTools: () =>
    api.get('/cli-config/custom/tools'),

  getCustomTool: (id: string) =>
    api.get(`/cli-config/custom/tools/${id}`),

  createCustomTool: (tool: any) =>
    api.post('/cli-config/custom/tools', tool),

  updateCustomTool: (id: string, tool: any) =>
    api.put(`/cli-config/custom/tools/${id}`, tool),

  deleteCustomTool: (id: string) =>
    api.delete(`/cli-config/custom/tools/${id}`),

  getCustomToolProxy: (toolId: string) =>
    api.get(`/cli-config/custom/tools/${toolId}/proxy`),

  enableCustomToolProxy: (toolId: string) =>
    api.post(`/cli-config/custom/tools/${toolId}/proxy/enable`),

  disableCustomToolProxy: (toolId: string) =>
    api.post(`/cli-config/custom/tools/${toolId}/proxy/disable`),

  getCustomToolFileContent: (toolId: string, fileId: string) =>
    api.get(`/cli-config/custom/tools/${toolId}/files/${fileId}`),

  saveCustomToolFileContent: (toolId: string, fileId: string, content: string) =>
    api.put(`/cli-config/custom/tools/${toolId}/files/${fileId}`, { content }),

  getCustomToolLockedFields: (toolId: string) =>
    api.get<string[]>(`/cli-config/custom/tools/${toolId}/locked-fields`),
}
