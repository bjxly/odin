<template>
  <div class="sm-dynamic-table">
    <el-table
      :data="tableData"
      border
      style="width: 100%"
    >
      <el-table-column
        type="index"
        label="#"
        width="50"
      />

      <!-- 动态列 -->
      <el-table-column
        v-for="child in childFields"
        :key="child.fieldKey"
        :label="child.fieldLabel"
        :min-width="120"
      >
        <template #default="{ row }">
          <el-input
            v-if="child.widget === 'text' || !child.widget"
            v-model="row[child.fieldKey]"
            :placeholder="child.placeholder"
            :readonly="readonly"
            size="small"
          />
          <el-input-number
            v-else-if="child.widget === 'number'"
            v-model="row[child.fieldKey]"
            :min="child.minValue"
            :max="child.maxValue"
            :disabled="readonly"
            size="small"
            controls-position="right"
            style="width: 100%"
          />
          <el-select
            v-else-if="child.widget === 'select'"
            v-model="row[child.fieldKey]"
            :placeholder="child.placeholder"
            :disabled="readonly"
            size="small"
          >
            <el-option
              v-for="opt in child.enumValues || []"
              :key="String(opt)"
              :label="child.enumLabels?.[String(opt)] || String(opt)"
              :value="opt"
            />
          </el-select>
          <el-date-picker
            v-else-if="child.widget === 'date'"
            v-model="row[child.fieldKey]"
            :placeholder="child.placeholder"
            :disabled="readonly"
            size="small"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
          <el-input
            v-else
            v-model="row[child.fieldKey]"
            :placeholder="child.placeholder"
            :readonly="readonly"
            size="small"
          />
        </template>
      </el-table-column>

      <!-- 操作列 -->
      <el-table-column
        v-if="!readonly"
        :label="$t('common.operation')"
        width="80"
        fixed="right"
      >
        <template #default="{ $index }">
          <el-button
            type="danger"
            link
            size="small"
            @click="removeRow($index)"
          >
            {{ $t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-button
      v-if="!readonly"
      type="primary"
      plain
      size="small"
      class="sm-dynamic-table__add"
      @click="addRow"
    >
      {{ field.addButtonText || t('cmp.schemaAddRow') }}
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormField } from '@/types/schema'

const { t } = useI18n()

const props = defineProps<{
  field: FormField
  modelValue: any[]
  readonly?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: any[]): void
}>()

const childFields = computed(() => props.field.children || [])

const tableData = computed({
  get: () => (Array.isArray(props.modelValue) ? props.modelValue : []),
  set: (val) => emit('update:modelValue', val),
})

function addRow() {
  const newRow: Record<string, any> = {}
  childFields.value.forEach((f) => {
    newRow[f.fieldKey] = f.defaultValue || undefined
  })
  tableData.value = [...tableData.value, newRow]
}

function removeRow(index: number) {
  const data = [...tableData.value]
  data.splice(index, 1)
  tableData.value = data
}
</script>

<style scoped>
.sm-dynamic-table {
  width: 100%;
}
.sm-dynamic-table__add {
  margin-top: 8px;
  width: 100%;
  border-style: dashed;
}
</style>
