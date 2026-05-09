import api from './client'

export interface LoginResponse {
  access_token: string
  refresh_token: string
  token_type: string
}

export interface UserInfo {
  user_id: number
  username: string
  role: string
}

export const authApi = {
  login: (username: string, password: string) =>
    api.post<LoginResponse>('/auth/login', { username, password }),

  logout: () =>
    api.post('/auth/logout'),

  getMe: () =>
    api.get<UserInfo>('/auth/me'),
}
