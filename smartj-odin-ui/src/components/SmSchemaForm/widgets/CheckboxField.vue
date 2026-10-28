<template>
  <el-checkbox
    :model-value="!!modelValue"
    :disabled="readonly"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    {{ field.fieldLabel }}
  </el-checkbox>
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
