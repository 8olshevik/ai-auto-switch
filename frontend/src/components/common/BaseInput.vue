<template>
  <div class="input-wrapper" :class="{ 'has-error': error }">
    <input
      v-bind="$attrs"
      :type="type"
      class="base-input"
      :class="{ 'input-error': error }"
      :value="modelValue"
      autocorrect="off"
      autocapitalize="none"
      spellcheck="false"
      autocomplete="off"
      @input="onInput"
      @blur="onBlur"
    />
    <p v-if="error && showError" class="input-error-message">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    modelValue?: string
    type?: string
    error?: string
    showError?: boolean
  }>(),
  {
    modelValue: '',
    type: 'text',
    error: '',
    showError: true,
  },
)

const emit = defineEmits<{ 
  (e: 'update:modelValue', value: string): void
  (e: 'blur'): void
}>()

useAttrs()

const onInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

const onBlur = () => {
  emit('blur')
}
</script>

<style scoped>
.input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.base-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--mac-border);
  border-radius: 8px;
  background: var(--mac-surface);
  color: var(--mac-text);
  font-size: 0.95rem;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.base-input:focus {
  outline: none;
  border-color: var(--mac-primary);
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.15);
}

.base-input::placeholder {
  color: var(--mac-text-tertiary);
}

/* Error state */
.input-error {
  border-color: rgba(255, 59, 48, 0.6) !important;
  background: rgba(255, 59, 48, 0.05);
}

.input-error:focus {
  border-color: rgba(255, 59, 48, 0.8) !important;
  box-shadow: 0 0 0 3px rgba(255, 59, 48, 0.15) !important;
}

.input-error-message {
  color: #ff3b30;
  font-size: 0.8rem;
  margin: 0;
  padding-left: 2px;
}
</style>
