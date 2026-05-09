import api from './client'

export const skillsApi = {
  list: (platform?: string) =>
    api.get('/skills/', { params: platform ? { platform } : undefined }),

  install: (payload: any) =>
    api.post('/skills/install', payload),

  uninstall: (directory: string, platform?: string, location?: string) =>
    api.delete('/skills/', { data: { directory, platform, location } }),

  toggle: (directory: string, platform: string, location: string, enabled: boolean) =>
    api.put('/skills/toggle', { directory, platform, location, enabled }),

  getContent: (directory: string, platform: string, location: string) =>
    api.get('/skills/content', { params: { directory, platform, location } }),

  saveContent: (directory: string, platform: string, location: string, content: string) =>
    api.put('/skills/content', { directory, platform, location, content }),

  getRepos: () =>
    api.get('/skills/repos'),

  addRepo: (repo: { owner: string; name: string; branch: string; enabled?: boolean }) =>
    api.post('/skills/repos', repo),

  removeRepo: (owner: string, name: string) =>
    api.delete(`/skills/repos/${owner}/${name}`),
}
