<template>
  <div class="sm-nested-field">
    <el-row :gutter="16">
      <el-col
        v-for="child in childFields"
        :key="child.fieldKey"
        :span="child.span || 24"
      >
        <el-form-item :label="child.fieldLabel" :required="child.isRequired">
          <el-input
            v-if="child.widget === 'text' || !child.widget"
            v-model="nestedData[child.fieldKey]"
            :placeholder="child.placeholder"
            :readonly="readonly"
          />
          <el-input-number
            v-else-if="child.widget === 'number'"
            v-model="nestedData[child.fieldKey]"
            :min="child.minValue"
            :max="child.maxValue"
            :disabled="readonly"
            controls-position="right"
            style="width: 100%"
          />
          <el-select
            v-else-if="child.widget === 'select'"
            v-model="nestedData[child.fieldKey]"
            :placeholder="child.placeholder"
            :disabled="readonly"
            style="width: 100%"
          >
            <el-option
              v-for="opt in (child.enumValues || [])"
              :key="String(opt)"
              :label="child.enumLabels?.[String(opt)] || String(opt)"
              :value="opt"
            />
          </el-select>
          <el-date-picker
            v-else-if="child.widget === 'date'"
            v-model="nestedData[child.fieldKey]"
            :placeholder="child.placeholder"
            :disabled="readonly"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
          <el-input
            v-else
            v-model="nestedData[child.fieldKey]"
            :placeholder="child.placeholder"
            :readonly="readonly"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FormField } from '@/types/schema'

const props = defineProps<{
  field: FormField
  modelValue: Record<string, any>
  readonly?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, any>): void
}>()

const childFields = computed(() => props.field.children || [])

const nestedData = computed({
  get: () => props.modelValue || {},
  set: (val) => emit('update:modelValue', val),
})
</script>

<style scoped>
.sm-nested-field {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  background: var(--el-fill-color-light);
}
</style>
