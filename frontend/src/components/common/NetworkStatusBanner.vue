<script setup lang="ts">
/**
 * 网络状态横幅组件
 * @description 网络断开时显示错误提示和重试选项
 *
 * Requirements: 19.4
 */
import { ref, onMounted, onUnmounted } from 'vue'
import { isOnline, onNetworkStatusChange } from '@/api/client'

const emit = defineEmits<{
  retry: []
}>()

const isConnected = ref(true)
const showBanner = ref(false)
const retryTimer: ReturnType<typeof setTimeout> | null = null

// 显示/隐藏横幅的过渡状态
const HIDE_DELAY = 3000 // 网络恢复后延迟 3 秒隐藏

let removeListener: (() => void) | null = null

onMounted(() => {
  // 初始状态
  isConnected.value = navigator.onLine
  showBanner.value = !isConnected.value

  // 监听网络状态变化
  removeListener = onNetworkStatusChange((connected) => {
    isConnected.value = connected
    
    if (connected) {
      // 网络恢复后延迟隐藏
      if (!retryTimer) {
        // @ts-ignore
        // eslint-disable-next-line no-undef
        timer = setTimeout(() => {
          showBanner.value = false
        }, HIDE_DELAY)
      }
    } else {
      // 网络断开，立即显示
      if (retryTimer) {
        clearTimeout(retryTimer)
      }
      showBanner.value = true
    }
  })
})

onUnmounted(() => {
  if (removeListener) {
    removeListener()
  }
  if (retryTimer) {
    clearTimeout(retryTimer)
  }
})

function handleRetry() {
  emit('retry')
}
</script>

<template>
  <Transition name="network-banner">
    <div v-if="showBanner && !isConnected" class="network-banner" role="alert">
      <div class="banner-content">
        <span class="banner-icon">📡</span>
        <div class="banner-text">
          <span class="banner-title">网络连接已断开</span>
          <span class="banner-description">请检查您的网络连接后重试</span>
        </div>
        <button class="banner-retry" @click="handleRetry">
          重试
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.network-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10000;
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a5a 100%);
  color: white;
  padding: 12px 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.banner-content {
  display: flex;
  align-items: center;
  gap: 12px;
  max-width: 1200px;
  margin: 0 auto;
}

.banner-icon {
  font-size: 1.5rem;
}

.banner-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.banner-title {
  font-weight: 600;
  font-size: 0.95rem;
}

.banner-description {
  font-size: 0.85rem;
  opacity: 0.9;
}

.banner-retry {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.4);
  border-radius: 8px;
  color: white;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.banner-retry:hover {
  background: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.6);
}

.banner-retry:active {
  transform: scale(0.98);
}

/* Transition animations */
.network-banner-enter-active {
  transition: all 0.3s ease;
}

.network-banner-leave-active {
  transition: all 0.3s ease;
}

.network-banner-enter-from,
.network-banner-leave-to {
  transform: translateY(-100%);
  opacity: 0;
}
</style>