import api from './client'

export const updateApi = {
  getVersion: () =>
    api.get<{ version: string }>('/update/version'),

  check: () =>
    api.get('/update/check'),

  install: () =>
    api.post('/update/install'),
}
