<template>
  <el-input
    :model-value="modelValue"
    :placeholder="field.placeholder || t('cmp.schemaEmailPlaceholder')"
    :readonly="readonly"
    :maxlength="field.maxLength"
    clearable
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template #prefix>
      <el-icon><Message /></el-icon>
    </template>
  </el-input>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Message } from '@element-plus/icons-vue'
import type { FormField } from '@/types/schema'

const { t } = useI18n()

defineProps<{
  field: FormField
  modelValue: any
  readonly?: boolean
}>()

defineEmits<{
  (e: 'update:modelValue', value: any): void
}>()
</script>
