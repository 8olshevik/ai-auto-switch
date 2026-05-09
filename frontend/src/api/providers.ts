import api from './client'

export const providerApi = {
  load: (kind: string) =>
    api.get(`/providers/${kind}`),

  save: (kind: string, providers: any[]) =>
    api.post(`/providers/${kind}`, providers),

  duplicate: (kind: string, id: number) =>
    api.post(`/providers/${kind}/duplicate/${id}`),

  rename: (kind: string, id: number, newName: string) =>
    api.put(`/providers/${kind}/${id}/rename`, { newName }),

  reorder: (kind: string, ids: number[]) =>
    api.post(`/providers/${kind}/reorder`, { ids }),
}
