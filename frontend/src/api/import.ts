import api from './client'

export const importApi = {
  importConfig: (path?: string) =>
    api.post('/import/config', path ? { path } : {}),

  exportConfig: () =>
    api.get('/import/export'),

  getStatus: () =>
    api.get('/import/status'),

  isFirstRun: () =>
    api.get<{ firstRun: boolean }>('/import/first-run'),

  markFirstRunDone: () =>
    api.post('/import/first-run-done'),
}
