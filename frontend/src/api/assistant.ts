/**
 * AI Assistant API
 * @description API service for AI Assistant chat functionality
 */
import api from './client'

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

export interface SendMessageRequest {
  message: string
}

export interface SendMessageResponse {
  conversationId: string
}

export interface SensitiveOperation {
  type: string
  description: string
  confirmed: boolean
}

export const assistantApi = {
  /**
   * Send a message to the AI assistant
   */
  sendMessage: (data: SendMessageRequest) =>
    api.post<SendMessageResponse>('/assistant/chat', data),

  /**
   * Get conversation history
   */
  getHistory: (conversationId?: string) =>
    api.get<ChatMessage[]>('/assistant/history', {
      params: conversationId ? { conversationId } : undefined,
    }),

  /**
   * Clear conversation history
   */
  clearHistory: (conversationId?: string) =>
    api.delete('/assistant/history', {
      params: conversationId ? { conversationId } : undefined,
    }),
}