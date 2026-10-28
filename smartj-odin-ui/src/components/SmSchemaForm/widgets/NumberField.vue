<template>
  <el-input-number
    :model-value="modelValue"
    :placeholder="field.placeholder"
    :disabled="readonly"
    :min="field.minValue"
    :max="field.maxValue"
    controls-position="right"
    style="width: 100%"
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
