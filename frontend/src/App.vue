<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import { computed, onMounted, onUnmounted } from 'vue'
import Sidebar from './components/Sidebar.vue'
import UpdateNotification from './components/common/UpdateNotification.vue'
import NetworkStatusBanner from './components/common/NetworkStatusBanner.vue'
import ToastContainer from './components/Toast/ToastContainer.vue'
import { wsClient } from '@/api/websocket'

const applyTheme = () => {
  const userTheme = localStorage.getItem('theme')
  const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  const isDark = userTheme === 'dark' || (!userTheme && systemPrefersDark)

  document.documentElement.classList.toggle('dark', isDark)
}

onMounted(() => {
  applyTheme()

  // 可监听系统主题变化自动更新（可选）
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    applyTheme()
  })

  // Connect WebSocket if user is authenticated
  const token = localStorage.getItem('token')
  if (token) {
    wsClient.connect()
  }
})

onUnmounted(() => {
  wsClient.disconnect()
})

const route = useRoute()
const isTray = computed(() => route.path === '/tray')
const isLoginPage = computed(() => route.path === '/login')
const showSidebar = computed(() => !isTray.value && !isLoginPage.value)
</script>

<template>
  <div v-if="isTray" class="tray-layout">
    <RouterView v-slot="{ Component }">
      <component :is="Component" />
    </RouterView>
  </div>
  <div v-else-if="isLoginPage" class="login-layout">
    <RouterView v-slot="{ Component }">
      <component :is="Component" />
    </RouterView>
  </div>
  <div v-else class="app-layout">
    <Sidebar />
    <main class="main-content">
      <RouterView v-slot="{ Component }">
        <keep-alive>
          <component :is="Component" />
        </keep-alive>
      </RouterView>
    </main>
    <!-- 全局更新通知 -->
    <UpdateNotification />
    <!-- 全局网络状态横幅 (Requirements: 19.4) -->
    <NetworkStatusBanner />
    <!-- 全局 Toast 通知 -->
    <ToastContainer />
  </div>
</template>

<style scoped>
.tray-layout {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}

.login-layout {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}

.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  background: var(--mac-bg);
}
</style>
