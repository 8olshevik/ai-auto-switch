<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../api/auth'

const router = useRouter()

const username = ref('')
const password = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

async function handleLogin() {
  errorMessage.value = ''

  if (!username.value.trim() || !password.value.trim()) {
    errorMessage.value = 'Please enter both username and password'
    return
  }

  isLoading.value = true

  try {
    const response = await authApi.login(username.value, password.value)
    const { access_token } = response.data
    localStorage.setItem('token', access_token)
    router.push('/')
  } catch (error: any) {
    if (error.response?.status === 401) {
      errorMessage.value = error.response?.data?.message || 'Invalid username or password'
    } else if (error.response?.data?.message) {
      errorMessage.value = error.response.data.message
    } else if (error.message) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = 'Login failed. Please try again.'
    }
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <!-- App Title -->
      <div class="login-header">
        <h1 class="app-title">Code Switch</h1>
        <p class="app-subtitle">AI API Gateway & Provider Manager</p>
      </div>

      <!-- Login Form -->
      <form class="login-form" @submit.prevent="handleLogin">
        <!-- Error Message -->
        <div v-if="errorMessage" class="error-banner" role="alert">
          <svg class="error-icon" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
          </svg>
          <span>{{ errorMessage }}</span>
        </div>

        <!-- Username Field -->
        <div class="form-group">
          <label for="username" class="form-label">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            autocomplete="username"
            placeholder="Enter your username"
            class="form-input"
            :disabled="isLoading"
          />
        </div>

        <!-- Password Field -->
        <div class="form-group">
          <label for="password" class="form-label">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="Enter your password"
            class="form-input"
            :disabled="isLoading"
            @keyup.enter="handleLogin"
          />
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          class="login-button"
          :disabled="isLoading"
        >
          <svg
            v-if="isLoading"
            class="spinner"
            viewBox="0 0 24 24"
            fill="none"
            aria-hidden="true"
          >
            <circle class="spinner-track" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" />
            <path class="spinner-head" d="M12 2a10 10 0 019.95 9" stroke="currentColor" stroke-width="3" stroke-linecap="round" />
          </svg>
          <span>{{ isLoading ? 'Signing in...' : 'Sign In' }}</span>
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--app-background);
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: var(--mac-surface);
  border: 1px solid var(--mac-border);
  border-radius: 24px;
  padding: 40px 32px;
  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.25);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.app-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--mac-text);
  margin: 0 0 8px;
  letter-spacing: -0.02em;
}

.app-subtitle {
  font-size: 0.9rem;
  color: var(--mac-text-secondary);
  margin: 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  border-radius: 12px;
  background: rgba(255, 59, 48, 0.1);
  border: 1px solid rgba(255, 59, 48, 0.3);
  color: #ff453a;
  font-size: 0.85rem;
  line-height: 1.4;
}

.error-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--mac-text-secondary);
}

.form-input {
  width: 100%;
  border-radius: 12px;
  border: 1px solid var(--mac-border);
  background: var(--mac-surface-strong);
  padding: 12px 14px;
  font: inherit;
  font-size: 0.95rem;
  color: var(--mac-text);
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--mac-accent);
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.25);
  background: var(--mac-surface);
}

.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-input::placeholder {
  color: var(--mac-text-secondary);
  opacity: 0.6;
}

.login-button {
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  border: none;
  border-radius: 12px;
  background: var(--mac-accent);
  color: #fff;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s ease, box-shadow 0.2s ease, opacity 0.2s ease;
  box-shadow: 0 4px 12px rgba(10, 132, 255, 0.3);
  margin-top: 4px;
}

.login-button:hover:not(:disabled) {
  background: #0070e0;
  box-shadow: 0 6px 18px rgba(10, 132, 255, 0.4);
}

.login-button:active:not(:disabled) {
  transform: scale(0.98);
}

.login-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.spinner {
  width: 18px;
  height: 18px;
  animation: spin 1s linear infinite;
}

.spinner-track {
  opacity: 0.25;
}

.spinner-head {
  opacity: 0.85;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
