import api from './client'

export interface Prompt {
  name: string
  content: string
  enabled?: boolean
}

export const promptsApi = {
  list: (platform?: string) =>
    api.get(`/prompts/${platform || 'claude'}`),

  create: (platform: string, id: string, prompt: Prompt) =>
    api.post(`/prompts/${platform}`, { id, prompt }),

  update: (platform: string, id: string, prompt: Prompt) =>
    api.put(`/prompts/${platform}/${id}`, prompt),

  delete: (platform: string, id: string) =>
    api.delete(`/prompts/${platform}/${id}`),
}
