<script setup lang="ts">
import type { ToastType } from '../../composables/useToast'

defineProps<{
  id: number
  type: ToastType
  message: string
}>()

const emit = defineEmits<{
  dismiss: [id: number]
}>()

const icons: Record<ToastType, string> = {
  success: '✓',
  error: '✕',
  warning: '⚠',
}
</script>

<template>
  <div class="toast-item" :class="`toast-${type}`" role="alert" aria-live="polite">
    <span class="toast-icon" :class="`toast-icon-${type}`">{{ icons[type] }}</span>
    <span class="toast-message">{{ message }}</span>
    <button
      class="toast-close"
      @click="emit('dismiss', id)"
      aria-label="关闭通知"
    >
      ×
    </button>
  </div>
</template>

<style scoped>
.toast-item {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 280px;
  max-width: 420px;
  padding: 12px 16px;
  border-radius: 14px;
  border: 1px solid var(--mac-border);
  background: var(--mac-surface-strong);
  color: var(--mac-text);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.2);
  font-size: 0.9rem;
  animation: toast-slide-in 0.25s ease forwards;
}

.toast-success {
  border-color: rgba(52, 199, 89, 0.4);
}

.toast-error {
  border-color: rgba(255, 59, 48, 0.5);
}

.toast-warning {
  border-color: rgba(245, 158, 11, 0.5);
}

.toast-icon {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
}

.toast-icon-success {
  background: rgba(52, 199, 89, 0.2);
  color: #34c759;
}

.toast-icon-error {
  background: rgba(255, 59, 48, 0.2);
  color: #ff3b30;
}

.toast-icon-warning {
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.toast-message {
  flex: 1;
  line-height: 1.4;
  word-break: break-word;
}

.toast-close {
  flex-shrink: 0;
  border: none;
  background: transparent;
  color: var(--mac-text-secondary);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
  border-radius: 4px;
  transition: color 0.15s ease;
}

.toast-close:hover {
  color: var(--mac-text);
}

@keyframes toast-slide-in {
  from {
    opacity: 0;
    transform: translateX(20px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translateX(0) scale(1);
  }
}
</style>
