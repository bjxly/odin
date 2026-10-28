<template>
  <div class="mapping-suggest">
    <div class="mapping-suggest__toolbar">
      <el-button
        size="small"
        :type="filterThreshold ? 'primary' : 'default'"
        @click="toggleFilter"
      >
        {{ t('odinMapping.filterGte05') }}
      </el-button>
      <el-button size="small" type="success" :disabled="!selectedIds.length" @click="handleApply">
        {{ t('odinMapping.applySelected') }}
      </el-button>
    </div>

    <el-table
      :data="filteredSuggestions"
      border
      size="small"
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="40" align="center" />
      <el-table-column prop="columnName" :label="t('odinMapping.columnName')" min-width="120" align="left" />
      <el-table-column prop="propertyName" :label="t('odinMapping.propertyName')" min-width="120" align="left" />
      <el-table-column :label="t('odinMapping.confidence')" width="160" align="center">
        <template #default="{ row }">
          <el-progress
            :percentage="Math.round(row.confidence * 100)"
            :color="confidenceColor(row.confidence)"
            :stroke-width="10"
            :text-inside="true"
          />
        </template>
      </el-table-column>
      <el-table-column prop="matchReason" :label="t('odinMapping.matchReason')" min-width="140" align="left" show-overflow-tooltip />
    </el-table>

    <el-empty v-if="filteredSuggestions.length === 0" :description="t('common.noData')" :image-size="50" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MappingSuggestVO } from '@/types/odin/mapping'

const props = defineProps<{
  suggestions: MappingSuggestVO[]
}>()

const emit = defineEmits<{
  (e: 'select', items: MappingSuggestVO[]): void
}>()

const { t } = useI18n()
const filterThreshold = ref(false)
const selectedRows = ref<MappingSuggestVO[]>([])
const selectedIds = computed(() => selectedRows.value.map((r) => r.columnName))

const filteredSuggestions = computed(() => {
  if (!filterThreshold.value) return props.suggestions
  return props.suggestions.filter((s) => s.confidence >= 0.5)
})

function confidenceColor(val: number): string {
  if (val >= 0.7) return 'var(--el-color-success)'
  if (val >= 0.4) return 'var(--el-color-warning)'
  return 'var(--el-color-danger)'
}

function toggleFilter() {
  filterThreshold.value = !filterThreshold.value
}

function handleSelectionChange(rows: MappingSuggestVO[]) {
  selectedRows.value = rows
}

function handleApply() {
  emit('select', selectedRows.value)
}
</script>

<style scoped>
.mapping-suggest {
  width: 100%;
}

.mapping-suggest__toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
</style>
