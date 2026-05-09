<script setup lang="ts">
import { useToast } from '../../composables/useToast'
import ToastItem from './ToastItem.vue'

const { toasts, dismiss } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="toast-container" aria-live="polite" aria-atomic="false">
      <TransitionGroup name="toast-list">
        <ToastItem
          v-for="toast in toasts"
          :key="toast.id"
          :id="toast.id"
          :type="toast.type"
          :message="toast.message"
          @dismiss="dismiss"
        />
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 9999;
  pointer-events: none;
}

.toast-container > :deep(*) {
  pointer-events: auto;
}

.toast-list-enter-active {
  transition: all 0.25s ease;
}

.toast-list-leave-active {
  transition: all 0.2s ease;
}

.toast-list-enter-from {
  opacity: 0;
  transform: translateX(20px) scale(0.96);
}

.toast-list-leave-to {
  opacity: 0;
  transform: translateX(10px) scale(0.96);
}

.toast-list-move {
  transition: transform 0.25s ease;
}
</style>
