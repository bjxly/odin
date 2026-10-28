<template>
  <el-input
    :model-value="modelValue"
    :placeholder="field.placeholder || t('cmp.schemaPhonePlaceholder')"
    :readonly="readonly"
    :maxlength="11"
    clearable
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template #prefix>
      <el-icon><Iphone /></el-icon>
    </template>
  </el-input>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Iphone } from '@element-plus/icons-vue'
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
