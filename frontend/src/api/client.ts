/**
 * API 客户端
 * @description 增强的错误处理，支持网络断开检测和具体错误信息
 *
 * Requirements: 19.2, 19.4
 */
import axios, { type AxiosError, type AxiosResponse } from 'axios'
import { useToast } from '@/composables/useToast'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
})

// 网络状态追踪
let isNetworkConnected = true
const networkListeners: Set<(connected: boolean) => void> = new Set()

/**
 * 获取当前网络连接状态
 */
export function isOnline(): boolean {
  return isNetworkConnected
}

/**
 * 监听网络状态变化
 */
export function onNetworkStatusChange(callback: (connected: boolean) => void): () => void {
  networkListeners.add(callback)
  return () => networkListeners.delete(callback)
}

/**
 * 更新网络状态并通知所有监听器
 */
function updateNetworkStatus(connected: boolean) {
  if (isNetworkConnected !== connected) {
    isNetworkConnected = connected
    networkListeners.forEach((cb) => cb(connected))
  }
}

// 监听浏览器网络状态变化
if (typeof window !== 'undefined') {
  window.addEventListener('online', () => updateNetworkStatus(true))
  window.addEventListener('offline', () => updateNetworkStatus(false))
}

// Request interceptor: attach JWT token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
}, (error) => {
  return Promise.reject(error)
})

// 解析错误响应，获取具体错误信息
function parseErrorMessage(error: AxiosError): { message: string; details?: string } {
  // 网络错误
  if (!error.response) {
    if (error.code === 'ECONNABORTED') {
      return { message: '请求超时，请稍后重试' }
    }
    if (!navigator.onLine) {
      return { message: '网络已断开，请检查您的网络连接' }
    }
    return { message: '网络错误，请检查网络连接后重试' }
  }

  const response = error.response as AxiosResponse<{ error?: string; message?: string }>
  
  // 服务器返回的错误
  switch (error.response.status) {
    case 400:
      return { 
        message: response.data?.error || '请求参数错误',
        details: response.data?.message 
      }
    case 401:
      return { message: '登录已过期，请重新登录' }
    case 403:
      return { message: '没有权限执行此操作' }
    case 404:
      return { message: '请求的资源不存在' }
    case 409:
      return { 
        message: response.data?.error || '操作冲突',
        details: response.data?.message 
      }
    case 422:
      return { 
        message: response.data?.error || '数据验证失败',
        details: response.data?.message 
      }
    case 429:
      return { message: '请求过于频繁，请稍后重试' }
    case 500:
      return { message: '服务器内部错误，请稍后重试' }
    case 502:
      return { message: '服务器网关错误，请稍后重试' }
    case 503:
      return { message: '服务暂不可用，请稍后重试' }
    default:
      return { 
        message: response.data?.error || '操作失败',
        details: response.data?.message 
      }
  }
}

// Response interceptor: handle errors
api.interceptors.response.use(
  (response) => {
    // 成功响应，确保网络状态正常
    updateNetworkStatus(true)
    return response
  },
  (error: AxiosError) => {
    const { message, details } = parseErrorMessage(error)
    
    // 401 特殊处理：清除 token 并跳转登录
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
      return Promise.reject(error)
    }

    // 更新网络状态
    if (!error.response && !navigator.onLine) {
      updateNetworkStatus(false)
    }

    // 为错误对象添加解析后的消息
    const enhancedError = error as AxiosError & { 
      parsedMessage?: string
      parsedDetails?: string
    }
    enhancedError.parsedMessage = message
    enhancedError.parsedDetails = details

    return Promise.reject(enhancedError)
  }
)

export default api
export type { AxiosError, AxiosResponse }
