/**
 * 表单引擎 composable
 * 管理表单状态、校验、值转换，独立于组件，可在任何场景使用
 */
import { reactive, watch, isRef, type Ref } from 'vue'
import type { FormSchema, FormField, TransformRule } from '@/types/schema'
import type { ValidationError } from './types'

export function useFormEngine(
  schema: Ref<FormSchema> | FormSchema,
  initialData?: Record<string, any>
) {
  const formData = reactive<Record<string, any>>({})
  const fieldErrors = reactive<Record<string, string>>({})
  const fieldReadonly = reactive<Record<string, boolean>>({})
  const fieldVisible = reactive<Record<string, boolean>>({})

  // 获取schema（支持Ref和直接值）
  function getSchema(): FormSchema {
    return isRef(schema) ? schema.value : schema
  }

  // 获取顶级字段
  function getTopFields(): FormField[] {
    const s = getSchema()
    return (s.fields || []).filter(f => (f.fieldLevel || 0) === 0)
  }

  // 获取所有字段（扁平化）
  function getAllFields(): FormField[] {
    const s = getSchema()
    return s.fields || []
  }

  // 根据fieldKey查找字段
  function findField(fieldKey: string): FormField | undefined {
    return getAllFields().find(f => f.fieldKey === fieldKey)
  }

  // 初始化默认值
  function initDefaults() {
    const fields = getTopFields()
    for (const field of fields) {
      if (field.defaultValue !== undefined && field.defaultValue !== null) {
        formData[field.fieldKey] = field.defaultValue
      } else if (field.dataType === 'array') {
        formData[field.fieldKey] = []
      } else if (field.dataType === 'object') {
        formData[field.fieldKey] = {}
      } else {
        formData[field.fieldKey] = undefined
      }
    }
  }

  // 设置初始数据
  function setInitialData(data?: Record<string, any>) {
    if (data) {
      Object.keys(data).forEach(key => {
        formData[key] = data[key]
      })
    }
  }

  // 初始化
  initDefaults()
  setInitialData(initialData)

  // 条件显示判断
  function evaluateCondition(rule: FormField['conditionalShow']): boolean {
    if (!rule) return true
    const results = (rule.rules || []).map(r => {
      const fieldValue = formData[r.field]
      switch (r.op) {
        case 'eq': return fieldValue === r.value
        case 'ne': return fieldValue !== r.value
        case 'gt': return Number(fieldValue) > Number(r.value)
        case 'lt': return Number(fieldValue) < Number(r.value)
        case 'contains': return String(fieldValue || '').includes(String(r.value))
        case 'in': return Array.isArray(r.value) && r.value.includes(fieldValue)
        default: return false
      }
    })
    if (results.length === 0) return true
    return rule.logic === 'OR' ? results.some(r => r) : results.every(r => r)
  }

  // 字段是否可见
  function isFieldVisible(field: FormField): boolean {
    if (field.isHidden) return false
    if (fieldReadonly[field.fieldKey] === false && fieldVisible[field.fieldKey] === false) return false
    if (field.conditionalShow) return evaluateCondition(field.conditionalShow)
    return true
  }

  // 值转换
  function applyTransform(value: any, rule?: TransformRule): any {
    if (!rule || value === undefined || value === null) return value
    switch (rule) {
      case 'trim': return String(value).trim()
      case 'uppercase': return String(value).toUpperCase()
      case 'lowercase': return String(value).toLowerCase()
      case 'phone_clean': return String(value).replace(/[\s\-()]/g, '')
      case 'date_normalize': {
        const d = new Date(value)
        if (isNaN(d.getTime())) return value
        return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
      }
      case 'amount_to_cents': {
        const num = Number(value)
        return isNaN(num) ? value : Math.round(num * 100)
      }
      default: return value
    }
  }

  // 校验单个字段
  function validateField(field: FormField): string | null {
    const value = formData[field.fieldKey]

    // 必填校验
    if (field.isRequired && (value === undefined || value === null || value === '')) {
      return `${field.fieldLabel}不能为空`
    }

    if (value === undefined || value === null || value === '') return null

    // 类型校验
    if (field.dataType === 'integer' || field.dataType === 'number') {
      if (isNaN(Number(value))) return `${field.fieldLabel}必须是数字`
    }

    // 格式校验
    if (field.formatConstraint) {
      const formatError = validateFormat(value, field.formatConstraint, field.fieldLabel)
      if (formatError) return formatError
    }

    // 正则校验
    if (field.fieldPattern) {
      try {
        const regex = new RegExp(field.fieldPattern)
        if (!regex.test(String(value))) return `${field.fieldLabel}格式不正确`
      } catch (e) {
        // 正则无效，跳过
      }
    }

    // 长度校验
    if (field.dataType === 'string') {
      const strVal = String(value)
      if (field.minLength !== undefined && strVal.length < field.minLength) {
        return `${field.fieldLabel}长度不能少于${field.minLength}个字符`
      }
      if (field.maxLength !== undefined && strVal.length > field.maxLength) {
        return `${field.fieldLabel}长度不能超过${field.maxLength}个字符`
      }
    }

    // 数值范围校验
    if (field.dataType === 'integer' || field.dataType === 'number') {
      const numVal = Number(value)
      if (field.minValue !== undefined && numVal < field.minValue) {
        return `${field.fieldLabel}不能小于${field.minValue}`
      }
      if (field.maxValue !== undefined && numVal > field.maxValue) {
        return `${field.fieldLabel}不能大于${field.maxValue}`
      }
    }

    // 枚举校验
    if (field.enumValues && field.enumValues.length > 0) {
      if (!field.enumValues.includes(value)) {
        return `${field.fieldLabel}的值无效`
      }
    }

    // 数组长度校验
    if (field.dataType === 'array' && Array.isArray(value)) {
      if (field.minItems !== undefined && value.length < field.minItems) {
        return `${field.fieldLabel}至少需要${field.minItems}项`
      }
      if (field.maxItems !== undefined && value.length > field.maxItems) {
        return `${field.fieldLabel}最多${field.maxItems}项`
      }
    }

    return null
  }

  // 格式校验
  function validateFormat(value: any, format: string, label: string): string | null {
    const strVal = String(value)
    switch (format) {
      case 'phone':
        if (!/^1[3-9]\d{9}$/.test(strVal.replace(/[\s\-]/g, ''))) return `${label}格式不正确`
        break
      case 'email':
        if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(strVal)) return `${label}格式不正确`
        break
      case 'url':
        try { new URL(strVal) } catch { return `${label}格式不正确` }
        break
      case 'date':
        if (isNaN(Date.parse(strVal))) return `${label}格式不正确`
        break
      case 'idcard':
        if (!/^\d{17}[\dXx]$/.test(strVal)) return `${label}格式不正确`
        break
    }
    return null
  }

  // 全量校验
  async function validate(): Promise<{ valid: boolean; errors: ValidationError[] }> {
    const errors: ValidationError[] = []
    // 清空之前的错误
    Object.keys(fieldErrors).forEach(k => delete fieldErrors[k])

    const fields = getTopFields()
    for (const field of fields) {
      if (!isFieldVisible(field)) continue
      const error = validateField(field)
      if (error) {
        fieldErrors[field.fieldKey] = error
        errors.push({
          fieldKey: field.fieldKey,
          fieldLabel: field.fieldLabel,
          errorMessage: error,
        })
      }
    }

    return { valid: errors.length === 0, errors }
  }

  // 获取表单数据（应用transform规则）
  function getFormData(): Record<string, any> {
    const result: Record<string, any> = {}
    const fields = getTopFields()
    for (const field of fields) {
      let value = formData[field.fieldKey]
      if (field.transformRule && value !== undefined && value !== null) {
        value = applyTransform(value, field.transformRule)
      }
      result[field.fieldKey] = value
    }
    return result
  }

  // 设置表单数据
  function setFormData(data: Record<string, any>) {
    Object.keys(data).forEach(key => {
      formData[key] = data[key]
    })
  }

  // 重置表单
  function resetForm() {
    Object.keys(formData).forEach(k => delete formData[k])
    Object.keys(fieldErrors).forEach(k => delete fieldErrors[k])
    initDefaults()
  }

  // 设置字段只读
  function setFieldReadonly(fieldKey: string, readonly: boolean) {
    fieldReadonly[fieldKey] = readonly
  }

  // 设置字段可见性
  function setFieldVisible(fieldKey: string, visible: boolean) {
    fieldVisible[fieldKey] = visible
  }

  // 动态表格操作
  function addArrayItem(fieldKey: string, item?: any) {
    if (!Array.isArray(formData[fieldKey])) {
      formData[fieldKey] = []
    }
    formData[fieldKey].push(item || {})
  }

  function removeArrayItem(fieldKey: string, index: number) {
    if (Array.isArray(formData[fieldKey])) {
      formData[fieldKey].splice(index, 1)
    }
  }

  // 监听formData变化
  watch(formData, () => {
    // 触发响应式更新
  }, { deep: true })

  return {
    formData,
    fieldErrors,
    fieldReadonly,
    fieldVisible,
    getTopFields,
    findField,
    isFieldVisible,
    evaluateCondition,
    validate,
    getFormData,
    setFormData,
    resetForm,
    setFieldReadonly,
    setFieldVisible,
    addArrayItem,
    removeArrayItem,
    applyTransform,
  }
}
