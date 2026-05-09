/**
 * 自动更新状态管理 Composable
 * @author sm
 * @description 管理应用自动更新的状态、事件监听和 API 调用
 */
import { ref, computed, reactive, onMounted, onUnmounted } from 'vue'
import { updateApi, wsClient } from '@/api'

// ==================== 类型定义 ====================

export type UpdateState = 'idle' | 'checking' | 'available' | 'downloading' | 'ready' | 'applying' | 'error'

export interface UpdateStateSnapshot {
  state: UpdateState
  current_version: string
  latest_version?: string
  notes?: string
  download_url?: string
  downloaded_bytes: number
  total_bytes: number
  progress: number // 0-100
  error?: string
  error_op?: 'check' | 'download' | 'apply'
  policy: string
}

export interface UpdateInfo {
  version: string
  pub_date: string
  notes: string
  download_url: string
  sha256: string
  size: number
}

export interface ProgressEvent {
  downloaded: number
  total: number
  percent: number
}

// ==================== 服务调用 ====================

export async function checkUpdate(): Promise<UpdateInfo | null> {
  try {
    const res = await updateApi.check()
    return res.data
  } catch (e) {
    console.error('Check update failed:', e)
    return null
  }
}

export async function downloadUpdate(): Promise<void> {
  // Download is handled via browser redirect or built-in updater
  const res = await updateApi.check()
  if (res.data?.download_url) {
    window.open(res.data.download_url, '_blank')
  }
}

export async function cancelDownload(): Promise<void> {
  // Cancel download - not supported via API, cleanup locally
  console.warn('[update] Cancel download not supported')
}

export async function requestRestart(): Promise<void> {
  try {
    await updateApi.install()
  } catch (e) {
    console.error('Install update failed:', e)
  }
}

export async function getUpdateState(): Promise<UpdateStateSnapshot> {
  // Not supported via API, return idle state
  return {
    state: 'idle',
    current_version: '',
    downloaded_bytes: 0,
    total_bytes: 0,
    progress: 0,
    policy: 'auto',
  }
}

export async function dismissUpdate(version: string): Promise<void> {
  // Store dismissed version locally
  localStorage.setItem('dismissed-update-version', version)
}

export async function getDismissedVersion(): Promise<string> {
  return localStorage.getItem('dismissed-update-version') || ''
}

// ==================== Composable ====================

// 全局状态（单例模式）
const globalState = reactive<UpdateStateSnapshot>({
  state: 'idle',
  current_version: '',
  latest_version: undefined,
  notes: undefined,
  download_url: undefined,
  downloaded_bytes: 0,
  total_bytes: 0,
  progress: 0,
  error: undefined,
  error_op: undefined,
  policy: 'auto',
})

const dismissedVersion = ref<string>('')
const isInitialized = ref(false)

/**
 * 自动更新状态管理 Composable
 */
export function useUpdateStore() {
  // ==================== 计算属性 ====================

  const hasUpdate = computed(() =>
    globalState.state === 'available' ||
    globalState.state === 'downloading' ||
    globalState.state === 'ready'
  )

  const isChecking = computed(() => globalState.state === 'checking')
  const isDownloading = computed(() => globalState.state === 'downloading')
  const isReady = computed(() => globalState.state === 'ready')
  const isApplying = computed(() => globalState.state === 'applying')
  const hasError = computed(() => globalState.state === 'error')

  const isDismissed = computed(() =>
    dismissedVersion.value !== '' &&
    globalState.latest_version === dismissedVersion.value
  )

  const progressPercent = computed(() => {
    if (globalState.total_bytes === 0) return 0
    return Math.round((globalState.downloaded_bytes / globalState.total_bytes) * 100)
  })

  const downloadedMB = computed(() =>
    (globalState.downloaded_bytes / 1024 / 1024).toFixed(1)
  )

  const totalMB = computed(() =>
    (globalState.total_bytes / 1024 / 1024).toFixed(1)
  )

  // ==================== 方法 ====================

  // 存储 WebSocket 取消订阅函数
  let unsubState: (() => void) | null = null
  let unsubProgress: (() => void) | null = null

  /**
   * 初始化更新服务（启动时调用一次）
   */
  async function init() {
    if (isInitialized.value) return

    // 加载已忽略的版本
    const dismissed = localStorage.getItem('dismissed-update-version')
    if (dismissed) {
      dismissedVersion.value = dismissed
    }

    // Setup WebSocket event listeners for update:state and update:progress
    unsubState = wsClient.on('update:state', (data: UpdateStateSnapshot) => {
      Object.assign(globalState, data)
    })

    unsubProgress = wsClient.on('update:progress', (data: ProgressEvent) => {
      globalState.downloaded_bytes = data.downloaded
      globalState.total_bytes = data.total
      globalState.progress = data.percent
    })

    // 获取初始状态
    try {
      const state = await getUpdateState()
      Object.assign(globalState, state)
    } catch (e) {
      console.error('Failed to get initial update state:', e)
    }

    isInitialized.value = true

    // 延迟 1 秒后自动检查更新
    setTimeout(() => {
      doCheckUpdate()
    }, 1000)
  }

  /**
   * 清理事件监听
   */
  function cleanup() {
    if (unsubState) {
      unsubState()
      unsubState = null
    }
    if (unsubProgress) {
      unsubProgress()
      unsubProgress = null
    }
  }

  /**
   * 检查更新
   */
  async function doCheckUpdate() {
    if (globalState.state === 'downloading' ||
        globalState.state === 'ready' ||
        globalState.state === 'applying') {
      return
    }

    globalState.state = 'checking'
    try {
      await checkUpdate()
    } catch (e) {
      console.error('Check update failed:', e)
    }
  }

  /**
   * 开始下载
   */
  async function doDownload() {
    if (globalState.state !== 'available' &&
        !(globalState.state === 'error' && globalState.error_op === 'download')) {
      return
    }

    try {
      await downloadUpdate()
    } catch (e) {
      console.error('Download failed:', e)
    }
  }

  /**
   * 取消下载
   */
  async function doCancel() {
    if (globalState.state !== 'downloading') return

    try {
      await cancelDownload()
    } catch (e) {
      console.error('Cancel download failed:', e)
    }
  }

  /**
   * 请求重启并更新
   */
  async function doRestart() {
    if (globalState.state !== 'ready') return

    try {
      await requestRestart()
    } catch (e) {
      console.error('Restart failed:', e)
    }
  }

  /**
   * 忽略当前版本
   */
  async function doDismiss() {
    if (!globalState.latest_version) return

    const version = globalState.latest_version
    dismissedVersion.value = version
    localStorage.setItem('dismissed-update-version', version)

    try {
      await dismissUpdate(version)
    } catch (e) {
      console.error('Dismiss failed:', e)
    }
  }

  /**
   * 刷新状态
   */
  async function refresh() {
    try {
      const state = await getUpdateState()
      Object.assign(globalState, state)
    } catch (e) {
      console.error('Refresh state failed:', e)
    }
  }

  // ==================== 生命周期 ====================

  onMounted(() => {
    init()
  })

  onUnmounted(() => {
    // 注意：不在 unmount 时清理事件监听，因为其他组件可能还需要
  })

  // ==================== 返回 ====================

  return {
    // 状态
    state: globalState,
    dismissedVersion,

    // 计算属性
    hasUpdate,
    isChecking,
    isDownloading,
    isReady,
    isApplying,
    hasError,
    isDismissed,
    progressPercent,
    downloadedMB,
    totalMB,

    // 方法
    init,
    cleanup,
    checkUpdate: doCheckUpdate,
    download: doDownload,
    cancel: doCancel,
    restart: doRestart,
    dismiss: doDismiss,
    refresh,
  }
}
