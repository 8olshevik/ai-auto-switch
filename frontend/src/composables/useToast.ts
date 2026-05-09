/**
 * Toast 通知系统 Composable
 * @description 提供全局 Toast 通知功能，支持 success/error/warning 类型
 *
 * Requirements: 19.1, 19.2, 19.4
 */
import { reactive, readonly } from 'vue'

export type ToastType = 'success' | 'error' | 'warning'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
  duration: number
}

// 全局状态（单例）
const toasts = reactive<ToastItem[]>([])
let nextId = 1

/** 默认持续时间（毫秒） */
const DEFAULT_DURATION: Record<ToastType, number> = {
  success: 4000,
  error: 6000,
  warning: 5000,
}

function addToast(type: ToastType, message: string, duration?: number) {
  const id = nextId++
  const toast: ToastItem = {
    id,
    type,
    message,
    duration: duration ?? DEFAULT_DURATION[type],
  }
  toasts.push(toast)

  // 自动移除
  if (toast.duration > 0) {
    setTimeout(() => {
      removeToast(id)
    }, toast.duration)
  }
}

function removeToast(id: number) {
  const index = toasts.findIndex((t) => t.id === id)
  if (index !== -1) {
    toasts.splice(index, 1)
  }
}

/**
 * Toast 通知 Composable
 */
export function useToast() {
  function showSuccess(message: string, duration?: number) {
    addToast('success', message, duration)
  }

  function showError(message: string, duration?: number) {
    addToast('error', message, duration)
  }

  function showWarning(message: string, duration?: number) {
    addToast('warning', message, duration)
  }

  function dismiss(id: number) {
    removeToast(id)
  }

  return {
    toasts: readonly(toasts),
    showSuccess,
    showError,
    showWarning,
    dismiss,
  }
}
