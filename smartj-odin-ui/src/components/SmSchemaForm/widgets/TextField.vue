<template>
  <el-input
    :model-value="modelValue"
    :placeholder="field.placeholder"
    :readonly="readonly"
    :maxlength="field.maxLength"
    clearable
    @update:model-value="$emit('update:modelValue', $event)"
  />
</template>

<script setup lang="ts">
import type { FormField } from '@/types/schema'

defineProps<{
  field: FormField
  modelValue: any
  readonly?: boolean
}>()

defineEmits<{
  (e: 'update:modelValue', value: any): void
}>()
</script>
