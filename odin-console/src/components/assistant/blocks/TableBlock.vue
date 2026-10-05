<script setup lang="ts">
import { computed } from 'vue'
import { ElTable, ElTableColumn, ElTag } from 'element-plus'
import { ROW_SOURCE } from '@/types/query'
import type { BlockVO, ChartField, TablePayload } from '@/types/assistant'
import { normalizeChartFields } from '@/types/assistant'
import { formatNumber } from '@/utils/format'

/**
 * table 块：查询结果表格。
 * 行级溯源列（_src_datasource / _src_table，键名恒为 ASCII）从数据列中剥离，
 * 合并为固定在右侧的「来源」列，直观展示每行数据来自哪个数据源的哪张表。
 *
 * 聚合结果（`aggregated: true`）额外处理：
 * - 表头「N 行」改为「N 个分组」（一行 = 一个分组，而非一条明细记录）；
 * - 维度列左对齐、度量列右对齐 + 千分位格式化（对齐 sm-frontend 列对齐规范）；
 * - 纯标量聚合（无维度 + 单度量，或后端建议 chart_type=metric）改渲染 KPI 指标卡，
 *   因为一个光秃秃的单行单列表格无法传达「这就是答案」的语义。
 */
const props = defineProps<{ block: BlockVO }>()

const p = computed<TablePayload>(() => (props.block?.payload ?? {}) as TablePayload)

/** 统一为「对象数组」：优先 records，其次由 columns × rows 拼装 */
const records = computed<Record<string, unknown>[]>(() => {
  const rec = p.value.records
  if (Array.isArray(rec) && rec.length) return rec as Record<string, unknown>[]
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
  const cols = (p.value.columns ?? []).map((c) => String(c))
  if (cols.length) return cols
  const acc: string[] = []
  for (const r of records.value) {
    for (const k of Object.keys(r)) if (!acc.includes(k)) acc.push(k)
  }
  return acc
})

const hasRowSource = computed(() =>
  records.value.some(
    (r) => r[ROW_SOURCE.datasource] != null || r[ROW_SOURCE.table] != null,
  ),
)

const dataCols = computed(() =>
  hasRowSource.value
    ? columns.value.filter((c) => c !== ROW_SOURCE.datasource && c !== ROW_SOURCE.table)
    : columns.value,
)

/** 行级唯一数据源名（用于表头汇总标签） */
const rowSources = computed<string[]>(() => {
  if (!hasRowSource.value) return []
  const acc = new Set<string>()
  for (const r of records.value) {
    const d = String(r[ROW_SOURCE.datasource] ?? '').trim()
    if (d) acc.add(d)
  }
  return [...acc]
})

const rowCount = computed(() => {
  const n = Number(p.value.row_count)
  return Number.isFinite(n) && n >= 0 ? n : records.value.length
})

const maxBodyHeight = 360

// --- 聚合结果：维度 / 度量列区分 ---
const aggregated = computed(() => !!p.value.aggregated)
/** 字段列表兼容 `{field,label}[]` 与旧 `string[]` 两种形态 */
const dimensions = computed<ChartField[]>(() => normalizeChartFields(p.value.dimensions))
const metrics = computed<ChartField[]>(() => normalizeChartFields(p.value.metrics))
const dimFields = computed<Set<string>>(() => new Set(dimensions.value.map((d) => d.field)))
const metricFields = computed<Set<string>>(() => new Set(metrics.value.map((m) => m.field)))

/**
 * 列角色判定：以后端下发的 dimensions/metrics 名单为准（按列名匹配，不假设位置）。
 * 未下发时退化为启发式：聚合结果里能数值化的列当度量。
 */
function isMeasure(col: string): boolean {
  if (metricFields.value.has(col)) return true
  if (dimFields.value.has(col)) return false
  if (!aggregated.value || metricFields.value.size || dimFields.value.size) return false
  return records.value.some((r) => {
    const v = r[col]
    if (v == null || v === '') return false
    return Number.isFinite(typeof v === 'number' ? v : Number(String(v).trim()))
  })
}

const measureCols = computed<Set<string>>(() => new Set(dataCols.value.filter(isMeasure)))

/** 列头文本：优先用后端下发的中文 label（本体 label），否则用列名本身 */
function colLabel(col: string): string {
  const m = metrics.value.find((x) => x.field === col)
  if (m?.label) return m.label
  const d = dimensions.value.find((x) => x.field === col)
  if (d?.label) return d.label
  return col
}

function cellTextOf(col: string, v: unknown): string {
  if (v == null) return ''
  if (measureCols.value.has(col)) return formatNumber(v)
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

/**
 * 纯标量聚合 → KPI 指标卡。
 * 条件：聚合结果 + 后端建议 metric，或无维度列且 ≤1 行。
 */
const isKpi = computed(() => {
  if (!aggregated.value || !records.value.length) return false
  if (String(p.value.chart_type ?? '').trim().toLowerCase() === 'metric') return true
  return dimFields.value.size === 0 && !dataCols.value.some((c) => !measureCols.value.has(c))
})

/** KPI 卡数据：取首行（全局聚合只有一行）的各度量列 */
const kpiItems = computed<Array<{ label: string; value: string; agg: string }>>(() => {
  const row = records.value[0] ?? {}
  const cols = dataCols.value.filter((c) => measureCols.value.has(c))
  const list = cols.length ? cols : dataCols.value
  return list.map((c) => ({
    label: colLabel(c),
    value: formatNumber(row[c]),
    agg: String(metrics.value.find((m) => m.field === c)?.agg ?? '').toUpperCase(),
  }))
})

function rowSource(row: Record<string, unknown>): string {
  const d = String(row?.[ROW_SOURCE.datasource] ?? '').trim()
  const t = String(row?.[ROW_SOURCE.table] ?? '').trim()
  if (d && t) return `${d}.${t}`
  return d || t || '—'
}
</script>

<template>
  <div class="table-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || '查询结果' }}</span>
      <ElTag size="small" type="primary" effect="plain">
        {{ aggregated ? `${rowCount} 个分组` : `${rowCount} 行` }}
      </ElTag>
      <ElTag v-if="aggregated" size="small" type="warning" effect="plain">
        聚合 · {{ measureCols.size }} 度量
      </ElTag>
      <ElTag v-if="hasRowSource && !isKpi" size="small" type="success" effect="plain">行级溯源</ElTag>
      <span v-if="rowSources.length && !isKpi" class="sources">{{ rowSources.join(' / ') }}</span>
    </div>

    <div v-if="!records.length" class="empty">结果为空（0 行）</div>

    <!-- 纯标量聚合：KPI 指标卡（一个数字就是一个答案，无需表格） -->
    <div v-else-if="isKpi" class="kpi-wrap">
      <div v-for="(k, i) in kpiItems" :key="i" class="kpi-card">
        <div class="kpi-value">{{ k.value }}</div>
        <div class="kpi-label">
          {{ k.label }}
          <span v-if="k.agg" class="kpi-agg">{{ k.agg }}</span>
        </div>
      </div>
      <div v-if="rowSources.length" class="kpi-source">数据来源：{{ rowSources.join(' / ') }}</div>
    </div>

    <ElTable
      v-else
      :data="records"
      size="small"
      border
      stripe
      :max-height="maxBodyHeight"
      class="result-table"
    >
      <ElTableColumn type="index" label="#" width="52" align="center" fixed="left" />
      <ElTableColumn
        v-for="c in dataCols"
        :key="c"
        :prop="c"
        :label="colLabel(c)"
        :align="measureCols.has(c) ? 'right' : 'left'"
        :min-width="Math.max(110, Math.min(240, colLabel(c).length * 12 + 40))"
        show-overflow-tooltip
      >
        <template #header>
          <span>{{ colLabel(c) }}</span>
          <span v-if="measureCols.has(c)" class="col-id">度量</span>
        </template>
        <template #default="{ row }">
          <span class="cell" :class="{ numeric: measureCols.has(c) }">{{ cellTextOf(c, row[c]) }}</span>
        </template>
      </ElTableColumn>
      <ElTableColumn v-if="hasRowSource" label="来源" min-width="180" fixed="right">
        <template #header>
          <span>来源</span>
          <span class="col-id">数据源.来源表</span>
        </template>
        <template #default="{ row }">
          <ElTag size="small" effect="plain" type="primary">{{ rowSource(row) }}</ElTag>
        </template>
      </ElTableColumn>
    </ElTable>
  </div>
</template>

<style scoped>
.table-block {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  overflow: hidden;
  background: #ffffff;
}

.block-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #f0f7ff, #e8f4fd);
  border-bottom: 1px solid #e2e8f0;
}

.block-title {
  font-size: 13px;
  font-weight: 700;
  color: #1a365d;
}

.sources {
  margin-left: auto;
  font-size: 11px;
  color: #718096;
}

.result-table {
  width: 100%;
}

.cell {
  font-size: 12px;
  color: #2d3748;
}

.cell.numeric {
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-weight: 600;
  color: #2b6cb0;
}

.kpi-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 14px 12px;
}

.kpi-card {
  flex: 1 1 140px;
  min-width: 140px;
  padding: 12px 14px;
  border: 1px solid #bee3f8;
  border-radius: 10px;
  background: linear-gradient(135deg, #f0f7ff, #ffffff);
}

.kpi-value {
  font-size: 26px;
  font-weight: 700;
  line-height: 1.2;
  color: #2b6cb0;
  font-family: 'JetBrains Mono', Consolas, monospace;
  word-break: break-all;
}

.kpi-label {
  margin-top: 4px;
  font-size: 12px;
  color: #4a5568;
  display: flex;
  align-items: center;
  gap: 6px;
}

.kpi-agg {
  font-size: 10px;
  color: #a0aec0;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.kpi-source {
  flex-basis: 100%;
  font-size: 11px;
  color: #718096;
}

.col-id {
  display: block;
  font-size: 10px;
  color: #a0aec0;
  font-weight: 400;
}

.empty {
  padding: 16px 12px;
  font-size: 12px;
  color: #a0aec0;
  text-align: center;
}
</style>
