<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MessageBlock, TableBlockPayload } from '@/types/odin'

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()

const p = computed<TableBlockPayload>(() => (props.block?.payload ?? {}) as TableBlockPayload)

/** 统一为「对象数组」：优先 records，其次由 columns × rows 拼装 */
const records = computed<Record<string, unknown>[]>(() => {
  const rec = p.value.records
  if (Array.isArray(rec) && rec.length) return rec
  const cols = p.value.columns ?? []
  const rows = p.value.rows ?? []
  if (!cols.length || !rows.length) return []
  return rows.map((r) => {
    const o: Record<string, unknown> = {}
    cols.forEach((c, i) => { o[String(c)] = (r as unknown[])?.[i] })
    return o
  })
})

const columns = computed<string[]>(() => {
  const cols = (p.value.columns ?? []).map(String)
  if (cols.length) return cols
  const acc: string[] = []
  for (const r of records.value) {
    for (const k of Object.keys(r)) if (!acc.includes(k)) acc.push(k)
  }
  return acc
})

const aggregated = computed(() => !!p.value.aggregated)

const rowCount = computed(() => {
  const n = Number(p.value.rowCount)
  return Number.isFinite(n) && n >= 0 ? n : records.value.length
})

/** 判断列是否为数值列（右对齐 + 千分位） */
function isNumericCol(col: string): boolean {
  return records.value.some((r) => {
    const v = r[col]
    if (v == null || v === '') return false
    return Number.isFinite(typeof v === 'number' ? v : Number(String(v).trim()))
  })
}

const numericCols = computed<Set<string>>(() => new Set(columns.value.filter(isNumericCol)))

function formatCell(col: string, v: unknown): string {
  if (v == null) return ''
  if (numericCols.value.has(col)) return formatNumber(v)
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

function formatNumber(v: unknown): string {
  const n = typeof v === 'number' ? v : Number(String(v).trim())
  if (!Number.isFinite(n)) return String(v)
  if (Number.isInteger(n)) return n.toLocaleString('zh-CN')
  return n.toLocaleString('zh-CN', { maximumFractionDigits: 2 })
}
</script>

<template>
  <div class="table-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || t('odinAssistant.blockTable') }}</span>
      <el-tag size="small" type="primary" effect="plain">
        {{ aggregated ? t('odinAssistant.tableGroups', { n: rowCount }) : t('odinAssistant.tableRows', { n: rowCount }) }}
      </el-tag>
      <el-tag v-if="aggregated" size="small" type="warning" effect="plain">{{ t('odinAssistant.tableAgg') }}</el-tag>
    </div>

    <div v-if="!records.length" class="empty">{{ t('odinAssistant.tableEmpty') }}</div>

    <el-table
      v-else
      :data="records"
      size="small"
      border
      stripe
      :max-height="360"
      class="result-table"
    >
      <el-table-column type="index" label="#" width="52" align="center" fixed="left" />
      <el-table-column
        v-for="c in columns"
        :key="c"
        :prop="c"
        :label="c"
        :align="numericCols.has(c) ? 'right' : 'left'"
        :min-width="Math.max(100, Math.min(220, c.length * 12 + 40))"
        show-overflow-tooltip
      >
        <template #default="{ row }">
          <span class="cell" :class="{ numeric: numericCols.has(c) }">{{ formatCell(c, row[c]) }}</span>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.table-block {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  overflow: hidden;
  background: var(--el-fill-color-blank);
}
.block-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.block-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.result-table {
  width: 100%;
}
.cell {
  font-size: 12px;
  color: var(--el-text-color-regular);
}
.cell.numeric {
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-weight: 600;
  color: var(--el-color-primary);
}
.empty {
  padding: 16px 12px;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  text-align: center;
}
</style>
