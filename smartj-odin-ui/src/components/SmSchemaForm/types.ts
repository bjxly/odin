/**
 * SmSchemaForm 组件内部类型定义
 * 
 * 核心类型 FormSchema / FormField 从 @/types/schema 导入
 * 此文件定义组件特有的辅助类型
 */
import type { FormSchema, FormField, ConditionalShowRule } from '@/types/schema'

/** 重新导出核心类型，方便组件内部使用 */
export type { FormSchema, FormField, ConditionalShowRule }

/** 校验错误 */
export interface ValidationError {
  fieldKey: string
  fieldLabel: string
  errorMessage: string
}

/** 组件 Props */
export interface SchemaFormRendererProps {
  schema: FormSchema
  modelValue?: Record<string, any>
  readonly?: boolean
  showSubmit?: boolean
  submitText?: string
  showReset?: boolean
  labelPosition?: 'left' | 'right' | 'top'
  gutter?: number
}

/** 组件 Emits */
export interface SchemaFormRendererEmits {
  (e: 'update:modelValue', data: Record<string, any>): void
  (e: 'submit', data: Record<string, any>): void
  (e: 'reset'): void
  (e: 'validate-fail', errors: ValidationError[]): void
}

/** 组件暴露方法 */
export interface SchemaFormRendererExpose {
  validate: () => Promise<{ valid: boolean; errors: ValidationError[] }>
  getFormData: () => Record<string, any>
  setFormData: (data: Record<string, any>) => void
  resetForm: () => void
  setFieldReadonly: (fieldKey: string, readonly: boolean) => void
  setFieldVisible: (fieldKey: string, visible: boolean) => void
}
