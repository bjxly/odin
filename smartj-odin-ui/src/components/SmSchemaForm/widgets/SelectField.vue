<template>
  <el-select
    :model-value="modelValue"
    :placeholder="field.placeholder"
    :disabled="readonly"
    clearable
    style="width: 100%"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <el-option
      v-for="opt in options"
      :key="String(opt.value)"
      :label="opt.label"
      :value="opt.value"
    />
  </el-select>
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

const options = computed(() => {
  const values = props.field.enumValues || []
  const labels = props.field.enumLabels || {}
  return values.map(v => ({
    value: v,
    label: labels[String(v)] || String(v),
  }))
})
</script>
