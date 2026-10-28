<template>
  <el-date-picker
    :model-value="modelValue"
    :placeholder="field.placeholder"
    :disabled="readonly"
    :type="dateType"
    value-format="YYYY-MM-DD"
    style="width: 100%"
    @update:model-value="$emit('update:modelValue', $event)"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FormField } from '@/types/schema'

const props = defineProps<{
  field: FormField
  modelValue: any
  readonly?: boolean
}>()

defineEmits<{
  (e: 'update:modelValue', value: any): void
}>()

const dateType = computed(() => {
  if (props.field.formatConstraint === 'datetime') return 'datetime'
  return 'date'
})
</script>
