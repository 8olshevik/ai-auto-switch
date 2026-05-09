/**
 * 表单验证 Composable
 * @description 提供表单字段验证功能，支持错误高亮显示
 *
 * Requirements: 19.5
 */
import { reactive, ref, computed, type Ref } from 'vue'

export interface ValidationRule {
  required?: boolean
  minLength?: number
  maxLength?: number
  pattern?: RegExp
  validator?: (value: string) => boolean | string
  message?: string
}

export interface FieldValidation {
  touched: boolean
  error: string | null
}

export interface FormValidationState {
  [key: string]: FieldValidation
}

/**
 * 创建表单验证 Composable
 */
export function useFormValidation<T extends Record<string, string>>(formData: Ref<T>, rules: Record<string, ValidationRule[]>) {
  const errors = reactive<FormValidationState>({})

  // 初始化错误状态
  Object.keys(formData.value).forEach((key) => {
    errors[key] = { touched: false, error: null }
  })

  // 验证单个字段
  function validateField(fieldName: string): boolean {
    const value = formData.value[fieldName] ?? ''
    const fieldRules = rules[fieldName] || []

    for (const rule of fieldRules) {
      // Required check
      if (rule.required && !value.trim()) {
        errors[fieldName] = { touched: true, error: rule.message || '此字段为必填项' }
        return false
      }

      // Min length check
      if (rule.minLength !== undefined && value.length < rule.minLength) {
        errors[fieldName] = { touched: true, error: rule.message || `最少需要 ${rule.minLength} 个字符` }
        return false
      }

      // Max length check
      if (rule.maxLength !== undefined && value.length > rule.maxLength) {
        errors[fieldName] = { touched: true, error: rule.message || `最多允许 ${rule.maxLength} 个字符` }
        return false
      }

      // Pattern check
      if (rule.pattern && !rule.pattern.test(value)) {
        errors[fieldName] = { touched: true, error: rule.message || '格式不正确' }
        return false
      }

      // Custom validator
      if (rule.validator) {
        const result = rule.validator(value)
        if (result !== true) {
          errors[fieldName] = { 
            touched: true, 
            error: typeof result === 'string' ? result : (rule.message || '验证失败') 
          }
          return false
        }
      }
    }

    errors[fieldName] = { touched: true, error: null }
    return true
  }

  // 验证所有字段
  function validateAll(): boolean {
    let isValid = true
    Object.keys(formData.value).forEach((key) => {
      if (!validateField(key)) {
        isValid = false
      }
    })
    return isValid
  }

  // 标记字段为已触摸
  function touchField(fieldName: string) {
    if (errors[fieldName]) {
      errors[fieldName].touched = true
    }
  }

  // 清除字段错误
  function clearError(fieldName: string) {
    if (errors[fieldName]) {
      errors[fieldName].touched = false
      errors[fieldName].error = null
    }
  }

  // 清除所有错误
  function clearAllErrors() {
    Object.keys(errors).forEach((key) => {
      errors[key] = { touched: false, error: null }
    })
  }

  // 获取字段是否有错误（已触摸且有错误）
  const hasError = computed(() => (fieldName: string) => {
    return errors[fieldName]?.touched && !!errors[fieldName]?.error
  })

  // 获取错误消息
  const errorMessage = computed(() => (fieldName: string) => {
    return errors[fieldName]?.error || null
  })

  // 检查表单是否有效
  const isValid = computed(() => {
    return Object.values(errors).every((field) => !field.error)
  })

  return {
    errors,
    validateField,
    validateAll,
    touchField,
    clearError,
    clearAllErrors,
    hasError,
    errorMessage,
    isValid,
  }
}

/**
 * 常用验证规则工厂函数
 */
export const validators = {
  required: (message = '此字段为必填项'): ValidationRule => ({
    required: true,
    message,
  }),

  minLength: (length: number, message?: string): ValidationRule => ({
    minLength: length,
    message: message || `最少需要 ${length} 个字符`,
  }),

  maxLength: (length: number, message?: string): ValidationRule => ({
    maxLength: length,
    message: message || `最多允许 ${length} 个字符`,
  }),

  url: (message = '请输入有效的 URL'): ValidationRule => ({
    pattern: /^https?:\/\/.+/,
    message,
  }),

  email: (message = '请输入有效的邮箱地址'): ValidationRule => ({
    pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
    message,
  }),

  number: (message = '请输入数字'): ValidationRule => ({
    pattern: /^\d+$/,
    message,
  }),

  custom: (validator: (value: string) => boolean | string, message = '验证失败'): ValidationRule => ({
    validator,
    message,
  }),
}