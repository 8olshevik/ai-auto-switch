<template>
  <div class="assistant-page">
    <div class="page-header">
      <h2>AI 助手</h2>
      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" @click="loadHistory" :disabled="loadingHistory">
          <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
            <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          {{ loadingHistory ? '加载中...' : '查看历史' }}
        </button>
        <button class="btn btn-danger btn-sm" @click="confirmClearHistory" :disabled="messages.length === 0">
          <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
            <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          清空对话
        </button>
      </div>
    </div>

    <!-- Chat Container -->
    <div class="chat-container">
      <!-- Message List -->
      <div class="message-list" ref="messageListRef">
        <div v-if="messages.length === 0" class="empty-state">
          <div class="empty-icon">
            <svg viewBox="0 0 24 24" width="48" height="48">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" fill="currentColor"/>
            </svg>
          </div>
          <p>与 AI 助手开始对话</p>
          <p class="hint">您可以询问关于 Code Switch 配置的问题</p>
        </div>
        
        <div
          v-for="message in messages"
          :key="message.id"
          class="message"
          :class="message.role"
        >
          <div class="message-avatar">
            <svg v-if="message.role === 'user'" viewBox="0 0 24 24" width="20" height="20">
              <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" fill="currentColor"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="20" height="20">
              <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z" fill="currentColor"/>
            </svg>
          </div>
          <div class="message-content">
            <div class="message-bubble" v-html="formatMessage(message.content)"></div>
            <div class="message-time">{{ formatTime(message.timestamp) }}</div>
          </div>
        </div>

        <!-- AI Typing Indicator -->
        <div v-if="isReceiving" class="message assistant">
          <div class="message-avatar">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z" fill="currentColor"/>
            </svg>
          </div>
          <div class="message-content">
            <div class="message-bubble typing">
              <span class="typing-indicator">
                <span></span>
                <span></span>
                <span></span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Input Area -->
      <div class="input-area">
        <textarea
          v-model="inputMessage"
          class="message-input"
          placeholder="输入消息..."
          rows="1"
          @keydown.enter.exact.prevent="sendMessage"
          @input="autoResize"
          ref="inputRef"
        ></textarea>
        <button
          class="send-btn"
          @click="sendMessage"
          :disabled="!inputMessage.trim() || isReceiving"
        >
          <svg viewBox="0 0 24 24" width="20" height="20">
            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" fill="currentColor"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Sensitive Operation Confirmation Dialog -->
    <div v-if="showSensitiveDialog" class="dialog-overlay" @click.self="cancelSensitiveOperation">
      <div class="dialog">
        <div class="dialog-header">
          <h3>敏感操作确认</h3>
          <button class="close-btn" @click="cancelSensitiveOperation">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="sensitive-warning">
            <svg viewBox="0 0 24 24" width="24" height="24">
              <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z" fill="currentColor"/>
            </svg>
            <span>AI 请求执行敏感操作</span>
          </div>
          <p class="sensitive-description">{{ pendingSensitiveOperation?.description }}</p>
          <div class="sensitive-details">
            <strong>操作类型:</strong> {{ pendingSensitiveOperation?.type }}
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="cancelSensitiveOperation">取消</button>
          <button class="btn btn-danger" @click="confirmSensitiveOperation">确认执行</button>
        </div>
      </div>
    </div>

    <!-- Clear History Confirmation Dialog -->
    <div v-if="showClearDialog" class="dialog-overlay" @click.self="showClearDialog = false">
      <div class="dialog dialog-sm">
        <div class="dialog-header">
          <h3>清空对话历史</h3>
        </div>
        <div class="dialog-body">
          <p>确定要清空所有对话历史吗？此操作无法撤销。</p>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="showClearDialog = false">取消</button>
          <button class="btn btn-danger" :disabled="clearing" @click="clearHistory">
            {{ clearing ? '清空中...' : '确认清空' }}
          </button>
        </div>
      </div>
    </div>

    <!-- History Dialog -->
    <div v-if="showHistoryDialog" class="dialog-overlay" @click.self="showHistoryDialog = false">
      <div class="dialog dialog-lg">
        <div class="dialog-header">
          <h3>对话历史</h3>
          <button class="close-btn" @click="showHistoryDialog = false">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </button>
        </div>
        <div class="dialog-body history-body">
          <div v-if="historyMessages.length === 0" class="empty-state">
            暂无对话历史
          </div>
          <div
            v-for="message in historyMessages"
            :key="message.id"
            class="message"
            :class="message.role"
          >
            <div class="message-avatar">
              <svg v-if="message.role === 'user'" viewBox="0 0 24 24" width="16" height="16">
                <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" fill="currentColor"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="16" height="16">
                <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z" fill="currentColor"/>
              </svg>
            </div>
            <div class="message-content">
              <div class="message-bubble" v-html="formatMessage(message.content)"></div>
              <div class="message-time">{{ formatTime(message.timestamp) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { wsClient, WS_EVENTS, type ChatMessage, type SensitiveOperation } from '../api'
import { assistantApi } from '../api'

// State
const messages = ref<ChatMessage[]>([])
const historyMessages = ref<ChatMessage[]>([])
const inputMessage = ref('')
const isReceiving = ref(false)
const loadingHistory = ref(false)
const clearing = ref(false)

// Dialogs
const showSensitiveDialog = ref(false)
const showClearDialog = ref(false)
const showHistoryDialog = ref(false)

// Pending sensitive operation
const pendingSensitiveOperation = ref<SensitiveOperation | null>(null)

// Refs
const messageListRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLTextAreaElement | null>(null)

// WebSocket event handlers
let unsubscribeReply: (() => void) | null = null
let unsubscribeSensitive: (() => void) | null = null

onMounted(() => {
  // Connect WebSocket
  wsClient.connect()

  // Listen for assistant replies (streaming)
  unsubscribeReply = wsClient.on(WS_EVENTS.ASSISTANT_REPLY, (data: { content: string; done?: boolean; conversationId?: string }) => {
    if (data.done) {
      isReceiving.value = false
      return
    }

    // Add or update the last assistant message
    const lastMsg = messages.value[messages.value.length - 1]
    if (lastMsg && lastMsg.role === 'assistant') {
      lastMsg.content += data.content
    } else {
      // Create new assistant message
      messages.value.push({
        id: Date.now().toString(),
        role: 'assistant',
        content: data.content,
        timestamp: new Date().toISOString(),
      })
    }

    scrollToBottom()
  })

  // Listen for sensitive operation requests
  unsubscribeSensitive = wsClient.on('assistant:sensitive', (data: SensitiveOperation) => {
    pendingSensitiveOperation.value = data
    showSensitiveDialog.value = true
  })
})

onUnmounted(() => {
  if (unsubscribeReply) unsubscribeReply()
  if (unsubscribeSensitive) unsubscribeSensitive()
})

// Auto-resize textarea
const autoResize = () => {
  if (inputRef.value) {
    inputRef.value.style.height = 'auto'
    inputRef.value.style.height = Math.min(inputRef.value.scrollHeight, 150) + 'px'
  }
}

// Scroll to bottom of message list
const scrollToBottom = () => {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

// Format message content (basic markdown-like formatting)
const formatMessage = (content: string): string => {
  if (!content) return ''
  
  // Escape HTML first
  let formatted = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  // Code blocks
  formatted = formatted.replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
  
  // Inline code
  formatted = formatted.replace(/`([^`]+)`/g, '<code>$1</code>')
  
  // Bold
  formatted = formatted.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  
  // Italic
  formatted = formatted.replace(/\*([^*]+)\*/g, '<em>$1</em>')
  
  // Line breaks
  formatted = formatted.replace(/\n/g, '<br>')

  return formatted
}

// Format timestamp
const formatTime = (timestamp: string): string => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Send message
const sendMessage = async () => {
  const content = inputMessage.value.trim()
  if (!content || isReceiving.value) return

  // Add user message
  const userMessage: ChatMessage = {
    id: Date.now().toString(),
    role: 'user',
    content,
    timestamp: new Date().toISOString(),
  }
  messages.value.push(userMessage)

  // Clear input
  inputMessage.value = ''
  if (inputRef.value) {
    inputRef.value.style.height = 'auto'
  }

  // Scroll to bottom
  scrollToBottom()

  // Start receiving
  isReceiving.value = true

  try {
    await assistantApi.sendMessage({ message: content })
    // Response will come via WebSocket
  } catch (error) {
    console.error('Failed to send message:', error)
    isReceiving.value = false
    
    // Add error message
    messages.value.push({
      id: Date.now().toString(),
      role: 'assistant',
      content: '发送消息失败，请稍后重试。',
      timestamp: new Date().toISOString(),
    })
    scrollToBottom()
  }
}

// Load history
const loadHistory = async () => {
  loadingHistory.value = true
  showHistoryDialog.value = true
  
  try {
    const data = await assistantApi.getHistory()
    historyMessages.value = data.data
  } catch (error) {
    console.error('Failed to load history:', error)
  } finally {
    loadingHistory.value = false
  }
}

// Confirm clear history
const confirmClearHistory = () => {
  showClearDialog.value = true
}

// Clear history
const clearHistory = async () => {
  clearing.value = true
  try {
    await assistantApi.clearHistory()
    messages.value = []
    historyMessages.value = []
    showClearDialog.value = false
  } catch (error) {
    console.error('Failed to clear history:', error)
    alert('清空失败：' + (error as Error).message)
  } finally {
    clearing.value = false
  }
}

// Confirm sensitive operation
const confirmSensitiveOperation = () => {
  if (pendingSensitiveOperation.value) {
    pendingSensitiveOperation.value.confirmed = true
    // Send confirmation via WebSocket or API
    console.log('Confirmed sensitive operation:', pendingSensitiveOperation.value)
  }
  showSensitiveDialog.value = false
  pendingSensitiveOperation.value = null
}

// Cancel sensitive operation
const cancelSensitiveOperation = () => {
  if (pendingSensitiveOperation.value) {
    pendingSensitiveOperation.value.confirmed = false
    console.log('Cancelled sensitive operation')
  }
  showSensitiveDialog.value = false
  pendingSensitiveOperation.value = null
}
</script>

<style scoped>
.assistant-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  padding: 1.5rem;
  color: var(--text-primary, #e0e0e0);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  flex-shrink: 0;
}

.page-header h2 {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

/* Chat Container */
.chat-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary, #2a2a2a);
  border-radius: 12px;
  border: 1px solid var(--border-color, #3a3a3a);
  overflow: hidden;
  min-height: 0;
}

/* Message List */
.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.message {
  display: flex;
  gap: 0.75rem;
  max-width: 85%;
}

.message.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message.assistant {
  align-self: flex-start;
}

.message-avatar {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary, #3a3a3a);
}

.message.user .message-avatar {
  background: #6366f1;
}

.message.assistant .message-avatar {
  background: #10b981;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.message-bubble {
  padding: 0.75rem 1rem;
  border-radius: 12px;
  background: var(--bg-tertiary, #3a3a3a);
  line-height: 1.5;
  word-break: break-word;
}

.message.user .message-bubble {
  background: #6366f1;
  color: white;
}

.message.assistant .message-bubble {
  background: var(--bg-tertiary, #3a3a3a);
}

.message-bubble :deep(code) {
  background: rgba(0, 0, 0, 0.2);
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  font-size: 0.875em;
}

.message-bubble :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  padding: 0.75rem;
  border-radius: 6px;
  overflow-x: auto;
  margin: 0.5rem 0;
}

.message-bubble :deep(pre code) {
  background: transparent;
  padding: 0;
}

.message-time {
  font-size: 0.75rem;
  color: var(--text-secondary, #a0a0a0);
}

/* Typing Indicator */
.typing-indicator {
  display: inline-flex;
  gap: 4px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: var(--text-secondary, #a0a0a0);
  border-radius: 50%;
  animation: typing 1.4s infinite;
}

.typing-indicator span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-indicator span:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% {
    transform: translateY(0);
    opacity: 0.4;
  }
  30% {
    transform: translateY(-4px);
    opacity: 1;
  }
}

/* Input Area */
.input-area {
  display: flex;
  gap: 0.75rem;
  padding: 1rem;
  border-top: 1px solid var(--border-color, #3a3a3a);
  background: var(--bg-tertiary, #333);
}

.message-input {
  flex: 1;
  padding: 0.75rem 1rem;
  background: var(--bg-secondary, #2a2a2a);
  border: 1px solid var(--border-color, #4a4a4a);
  border-radius: 8px;
  color: var(--text-primary, #e0e0e0);
  font-size: 0.9375rem;
  resize: none;
  font-family: inherit;
  max-height: 150px;
}

.message-input:focus {
  outline: none;
  border-color: #6366f1;
}

.message-input::placeholder {
  color: var(--text-secondary, #a0a0a0);
}

.send-btn {
  flex-shrink: 0;
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #6366f1;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  transition: all 0.2s;
}

.send-btn:hover:not(:disabled) {
  background: #4f46e5;
}

.send-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary, #a0a0a0);
  text-align: center;
}

.empty-icon {
  color: var(--text-secondary, #a0a0a0);
  opacity: 0.5;
  margin-bottom: 1rem;
}

.hint {
  font-size: 0.875rem;
  margin-top: 0.5rem;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #6366f1;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #4f46e5;
}

.btn-secondary {
  background: var(--bg-tertiary, #3a3a3a);
  color: var(--text-primary, #e0e0e0);
}

.btn-secondary:hover:not(:disabled) {
  background: #4a4a4a;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

/* Dialog */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: var(--bg-secondary, #2a2a2a);
  border-radius: 12px;
  width: 100%;
  max-width: 480px;
  border: 1px solid var(--border-color, #3a3a3a);
}

.dialog-lg {
  max-width: 640px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.dialog-sm {
  max-width: 360px;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  border-bottom: 1px solid var(--border-color, #3a3a3a);
  flex-shrink: 0;
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-secondary, #a0a0a0);
  cursor: pointer;
  padding: 0.25rem;
}

.close-btn:hover {
  color: var(--text-primary, #e0e0e0);
}

.dialog-body {
  padding: 1.25rem;
}

.history-body {
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-height: 400px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem;
  border-top: 1px solid var(--border-color, #3a3a3a);
  flex-shrink: 0;
}

/* Sensitive Operation Dialog */
.sensitive-warning {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: #f59e0b;
  font-weight: 500;
  margin-bottom: 1rem;
}

.sensitive-description {
  margin: 0 0 1rem;
  color: var(--text-secondary, #a0a0a0);
}

.sensitive-details {
  font-size: 0.875rem;
  color: var(--text-secondary, #a0a0a0);
}

.sensitive-details strong {
  color: var(--text-primary, #e0e0e0);
}
</style>