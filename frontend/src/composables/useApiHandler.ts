/**
 * API 处理器 Composable
 * @description 提供统一的 API 调用封装，整合错误处理、加载状态和 Toast 提示
 *
 * Requirements: 19.1, 19.2, 19.3, 19.4
 */
import { ref } from 'vue'
import api, { isOnline, onNetworkStatusChange, type AxiosError } from '@/api/client'
import { useToast } from './useToast'
import { useLoading } from './useLoading'

export interface ApiHandlerOptions<T> {
  /** 成功时显示 Toast */
  showSuccessToast?: boolean
  /** 成功 Toast 消息（支持函数） */
  successMessage?: string | ((data: T) => string)
  /** 失败时显示 Toast */
  showErrorToast?: boolean
  /** 失败时显示详细错误信息 */
  showDetailedError?: boolean
  /** 加载状态延迟（毫秒），默认 1000ms */
  loadingDelay?: number
  /** 忽略错误（不抛出异常） */
  ignoreError?: boolean
  /** 自定义错误处理 */
  onError?: (error: AxiosError) => void
}

export interface ApiHandlerResult<T> {
  /** 是否正在加载 */
  isLoading: ReturnType<typeof ref<boolean>>
  /** 加载消息 */
  loadingMessage: ReturnType<typeof ref<string>>
  /** 执行 API 调用 */
  execute: () => Promise<T | undefined>
  /** 网络是否在线 */
  isOnline: () => boolean
  /** 监听网络状态变化 */
  onNetworkStatusChange: (callback: (connected: boolean) => void) => () => void
}

/**
 * 创建 API 处理器
 * @param apiFn 返回 AxiosPromise 的 API 函数
 * @param options 配置选项
 */
export function useApiHandler<T>(
  apiFn: () => Promise<{ data: T }>,
  options: ApiHandlerOptions<T> = {}
): ApiHandlerResult<T> {
  const {
    showSuccessToast = false,
    successMessage,
    showErrorToast = true,
    showDetailedError = true,
    loadingDelay = 1000,
    ignoreError = false,
    onError,
  } = options

  const { showSuccess, showError } = useToast()
  const { isLoading, loadingMessage, withLoading } = useLoading()

  /**
   * 执行 API 调用
   */
  async function execute(): Promise<T | undefined> {
    const result = await withLoading(
      async () => {
        const response = await apiFn()
        
        // 成功处理
        if (showSuccessToast || successMessage) {
          const message = typeof successMessage === 'function' 
            ? successMessage(response.data) 
            : (successMessage || '操作成功')
          showSuccess(message)
        }
        
        return response.data
      },
      '',
      loadingDelay
    ).catch((error: AxiosError) => {
      // 错误处理
      if (onError) {
        onError(error)
      }

      const errorMessage = (error as any).parsedMessage || '操作失败'
      const errorDetails = (error as any).parsedDetails

      if (showErrorToast) {
        if (showDetailedError && errorDetails) {
          showError(`${errorMessage}: ${errorDetails}`)
        } else if (!navigator.onLine) {
          // 网络断开特殊处理
          showError('网络已断开，请检查网络连接后重试')
        } else {
          showError(errorMessage)
        }
      }

      if (!ignoreError) {
        throw error
      }
      
      return undefined
    })

    return result
  }

  return {
    isLoading,
    loadingMessage,
    execute,
    isOnline,
    onNetworkStatusChange,
  }
}

/**
 * 使用带有加载状态的 API 调用（快捷方式）
 */
export function useApi<T>(
  apiFn: () => Promise<{ data: T }>,
  loadingDelay = 1000
) {
  return useApiHandler(apiFn, { loadingDelay })
}

/**
 * 创建可重试的 API 处理器
 */
export function useRetryableApi<T>(
  apiFn: () => Promise<{ data: T }>,
  options: ApiHandlerOptions<T> & { maxRetries?: number; retryDelay?: number } = {}
): ApiHandlerResult<T> & { retry: () => Promise<T | undefined> } {
  const { maxRetries = 3, retryDelay = 2000, ...handlerOptions } = options
  const retryCount = ref(0)

  const handler = useApiHandler<T>(apiFn, {
    ...handlerOptions,
    onError: (error) => {
      // 自动重试逻辑
      if (retryCount.value < maxRetries && !navigator.onLine) {
        return // 网络错误，等待用户手动重试
      }
      handlerOptions.onError?.(error)
    },
  })

  async function retry(): Promise<T | undefined> {
    retryCount.value = 0
    return handler.execute()
  }

  return {
    ...handler,
    retry,
  }
}