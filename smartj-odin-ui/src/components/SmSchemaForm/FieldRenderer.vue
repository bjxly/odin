<template>
  <el-form-item
    v-if="visible"
    :label="field.fieldLabel"
    :prop="field.fieldKey"
    :error="error"
    :required="field.isRequired"
    :span="field.span || 24"
  >
    <!-- 前置插槽 -->
    <slot :name="`before-field-${field.fieldKey}`" />

    <!-- 动态表格 -->
    <DynamicTableField
      v-if="field.widget === 'dynamic_table' || field.dataType === 'array'"
      :field="field"
      :model-value="modelValue"
      :readonly="readonly"
      @update:model-value="$emit('update:modelValue', $event)"
    />

    <!-- 嵌套对象 -->
    <NestedField
      v-else-if="field.widget === 'nested' || field.dataType === 'object'"
      :field="field"
      :model-value="modelValue"
      :readonly="readonly"
      @update:model-value="$emit('update:modelValue', $event)"
    />

    <!-- 标准控件 -->
    <component
      :is="widgetComponent"
      v-else
      :field="field"
      :model-value="modelValue"
      :readonly="readonly"
      @update:model-value="$emit('update:modelValue', $event)"
    />

    <!-- 后置插槽 -->
    <slot :name="`after-field-${field.fieldKey}`" />

    <!-- 帮助文本 -->
    <div v-if="field.helpText" class="sm-field__help">
      {{ field.helpText }}
    </div>
  </el-form-item>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FormField } from '@/types/schema'
import TextField from './widgets/TextField.vue'
import NumberField from './widgets/NumberField.vue'
import SelectField from './widgets/SelectField.vue'
import DateField from './widgets/DateField.vue'
import TextareaField from './widgets/TextareaField.vue'
import PhoneField from './widgets/PhoneField.vue'
import EmailField from './widgets/EmailField.vue'
import CheckboxField from './widgets/CheckboxField.vue'
import DynamicTableField from './DynamicTableField.vue'
import NestedField from './NestedField.vue'

const props = defineProps<{
  field: FormField
  modelValue: any
  readonly?: boolean
  visible?: boolean
  error?: string
}>()

defineEmits<{
  (e: 'update:modelValue', value: any): void
}>()

const widgetComponent = computed(() => {
  switch (props.field.widget) {
    case 'text': return TextField
    case 'number': return NumberField
    case 'select': return SelectField
    case 'date': return DateField
    case 'textarea': return TextareaField
    case 'phone': return PhoneField
    case 'email': return EmailField
    case 'checkbox': return CheckboxField
    default: return TextField
  }
})
</script>

<style scoped>
.sm-field__help {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
  margin-top: 4px;
}
</style>
