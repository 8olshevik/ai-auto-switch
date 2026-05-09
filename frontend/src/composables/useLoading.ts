/**
 * 加载状态管理 Composable
 * @description 提供 isLoading 状态和 withLoading 包装器，长时间操作（>1s）自动显示加载指示器
 *
 * Requirements: 19.3
 */
import { ref } from 'vue'

/**
 * 加载状态 Composable
 */
export function useLoading() {
  const isLoading = ref(false)
  const loadingMessage = ref('')

  /**
   * 包装异步函数，自动管理加载状态
   * 如果操作超过 delay 毫秒，则显示加载指示器
   */
  async function withLoading<T>(
    asyncFn: () => Promise<T>,
    message = '',
    delay = 1000,
  ): Promise<T> {
    let showTimer: ReturnType<typeof setTimeout> | null = null

    // 延迟显示加载状态（仅在操作超过 delay 时显示）
    showTimer = setTimeout(() => {
      isLoading.value = true
      loadingMessage.value = message
    }, delay)

    try {
      const result = await asyncFn()
      return result
    } finally {
      if (showTimer) {
        clearTimeout(showTimer)
      }
      isLoading.value = false
      loadingMessage.value = ''
    }
  }

  return {
    isLoading,
    loadingMessage,
    withLoading,
  }
}
