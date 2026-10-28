<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElOption, ElSelect, ElTag } from 'element-plus'
import type { BlockVO } from '@/types/assistant'
import { normalizeChartPayload } from '@/types/assistant'

/**
 * chart 块（CSS 降级渲染器）。
 *
 * 默认情况下 `chart` 类型已由 `EChartsBlock.vue` 覆盖注册（见 blocks/index.ts），
 * 本组件仅在 ECharts 异步 chunk 加载失败（弱网 / 资源不可达）时作为降级使用，
 * 因此必须与 EChartsBlock 共享同一套 payload 解析逻辑（`normalizeChartPayload`），
 * 保证对象形态（`dimensions:[{field,label}]`）与旧 `string[]` 形态都能渲染。
 *
 * 列下标统一按 field 名从 columns 反查，不再假设「维度恒在第 0 列、度量为 i+1」。
 */
const props = defineProps<{ block: BlockVO }>()

const MAX_BARS = 20

const nc = computed(() => normalizeChartPayload(props.block?.payload))
const chartType = computed(() => nc.value.chartType)
const dimensions = computed<string[]>(() => nc.value.dimensions.map((d) => d.label))
const metrics = computed<string[]>(() => nc.value.metrics.map((m) => m.label))
const rows = computed<unknown[][]>(() => nc.value.rows)
const plotRows = computed<unknown[][]>(() => rows.value.slice(0, MAX_BARS))

/**
 * 按 field 名反查列下标（摆脱 `i+1` 位置假设）；
 * columns 未下发或列名不在其中时，退化为「维度列在前、度量列在后」的推导下标。
 */
function colIndexOf(field: string, fallback: number): number {
  const i = nc.value.columns.indexOf(field)
  return i >= 0 ? i : fallback
}

/** 维度列下标（无维度时为 0，用于条形图标签） */
const dimCol = computed(() => colIndexOf(nc.value.dimensions[0]?.field ?? '', 0))

/** 宽松数值化：空字串 / null / 日期文本 / 非数字均视为不可绘图 */
function toNum(v: unknown): number | null {
  if (typeof v === 'number') return Number.isFinite(v) ? v : null
  if (typeof v === 'string') {
    const s = v.trim()
    if (!s) return null
    const n = Number(s)
    return Number.isFinite(n) ? n : null
  }
  return null
}

interface MetricSeries {
  name: string
  /** 在 rows 中的列下标 */
  col: number
  numeric: boolean
  values: (number | null)[]
}

const series = computed<MetricSeries[]>(() =>
  nc.value.metrics.map((m, i) => {
    const col = colIndexOf(m.field, nc.value.dimensions.length + i)
    const values = plotRows.value.map((r) => toNum(Array.isArray(r) ? r[col] : null))
    const hits = values.filter((v) => v !== null).length
    // 半数以上行能转成数字才当作数值指标（避免 phone 类长字符串被当成主指标时完全无法绘图）
    const numeric = plotRows.value.length > 0 && hits >= Math.ceil(plotRows.value.length / 2)
    return { name: m.label, col, numeric, values }
  }),
)

const numericSeries = computed<MetricSeries[]>(() => series.value.filter((s) => s.numeric))
const skippedMetrics = computed<string[]>(() =>
  series.value.filter((s) => !s.numeric).map((s) => s.name),
)

/**
 * 默认指标名命中度量词时优先选它。
 * 因为后端把 phone 这类“数值型标识列”也归入了 metrics，
 * 直接取首个数值列会画出无意义的电话号柱形图。
 */
const MEASURE_HINT =
  /amount|total|count|qty|quantity|price|cost|fee|sum|avg|score|rate|ratio|num|sales|revenue|profit|stock|inventory|weight|金额|数量|总数|单价|总价|库存|得分|合计|笔数|均值/i

const preferredSeries = computed<MetricSeries | null>(
  () =>
    numericSeries.value.find((s) => MEASURE_HINT.test(s.name)) ??
    numericSeries.value[0] ??
    null,
)

const activeName = ref('')

/** 默认选中首选数值指标；数据变化后若已选项失效则重新回落 */
watch(
  () => numericSeries.value.map((s) => s.name).join('|'),
  () => {
    if (!numericSeries.value.some((s) => s.name === activeName.value)) {
      activeName.value = preferredSeries.value?.name ?? ''
    }
  },
  { immediate: true },
)

const active = computed<MetricSeries | null>(
  () =>
    numericSeries.value.find((s) => s.name === activeName.value) ??
    numericSeries.value[0] ??
    null,
)

const bars = computed<Array<{ label: string; value: number | null }>>(() =>
  plotRows.value.map((r, i) => ({
    label: String((Array.isArray(r) ? r[dimCol.value] : null) ?? '—'),
    value: active.value?.values[i] ?? null,
  })),
)

const maxValue = computed(() => {
  let max = 0
  for (const v of active.value?.values ?? []) {
    const a = Math.abs(v ?? 0)
    if (a > max) max = a
  }
  return max
})

function widthOf(v: number | null): string {
  if (v == null || !maxValue.value) return '2%'
  const pct = (Math.abs(v) / maxValue.value) * 100
  return `${Math.max(2, Math.min(100, pct)).toFixed(1)}%`
}

function num(v: number | null): string {
  if (v == null || !Number.isFinite(v)) return '—'
  const a = Math.abs(v)
  if (a >= 1e6) return v.toExponential(2)
  return Number.isInteger(v) ? String(v) : v.toFixed(2)
}
</script>

<template>
  <div class="chart-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || '可视化建议' }}</span>
      <ElTag size="small" type="primary" effect="plain">{{ chartType }}</ElTag>
      <ElTag v-if="nc.aggregated" size="small" type="warning" effect="plain">聚合结果</ElTag>
      <ElTag v-if="nc.topN" size="small" type="info" effect="plain">TOP-{{ nc.topN }}</ElTag>
      <ElTag v-if="dimensions.length" size="small" effect="plain">
        维度：{{ dimensions.join(', ') }}
      </ElTag>
      <ElTag v-if="metrics.length" size="small" type="warning" effect="plain">
        指标 {{ metrics.length }} 列 · 可绘图 {{ numericSeries.length }}
      </ElTag>
    </div>

    <div v-if="!plotRows.length" class="empty">无可视化数据（结果集为空）</div>
    <div v-else-if="!active" class="empty">
      结果集中没有可绘图的数值指标列
      <span v-if="skippedMetrics.length">（{{ skippedMetrics.join(', ') }} 均为文本）</span>
    </div>

    <template v-else>
      <div v-if="numericSeries.length > 1" class="metric-picker">
        <span class="picker-label">绘图指标</span>
        <ElSelect v-model="activeName" size="small" class="picker-select">
          <ElOption v-for="s in numericSeries" :key="s.name" :value="s.name" :label="s.name" />
        </ElSelect>
      </div>

      <div class="bars">
        <div v-for="(b, i) in bars" :key="i" class="bar-row">
          <span class="bar-label" :title="b.label">{{ b.label }}</span>
          <div class="bar-track">
            <div class="bar-fill" :style="{ width: widthOf(b.value) }" />
          </div>
          <span class="bar-value">{{ num(b.value) }}</span>
        </div>
      </div>

      <div class="chart-note">
        <span v-if="rows.length > MAX_BARS">
          仅展示前 {{ MAX_BARS }} 行（共 {{ rows.length }} 行）。
        </span>
        <span v-if="skippedMetrics.length">已跳过非数值列：{{ skippedMetrics.join(', ') }}。</span>
        <span v-if="numericSeries.length > 1">
          当前绘制「{{ active?.name }}」，共 {{ numericSeries.length }} 个可绘图指标，可在上方切换。
        </span>
        <span class="hint">图表引擎未加载，已降级为轻量条形图。</span>
      </div>
    </template>
  </div>
</template>

<style scoped>
.chart-block {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
  padding: 10px 12px;
}

.block-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.block-title {
  font-size: 13px;
  font-weight: 700;
  color: #1a365d;
}

.bars {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.metric-picker {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.picker-label {
  font-size: 11px;
  color: #718096;
}

.picker-select {
  width: 180px;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bar-label {
  width: 120px;
  flex-shrink: 0;
  font-size: 11px;
  color: #4a5568;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bar-track {
  flex: 1;
  height: 14px;
  background: #edf2f7;
  border-radius: 7px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #4c8bf5, #2b6cb0);
  border-radius: 7px;
  transition: width 0.3s ease;
}

.bar-value {
  width: 76px;
  flex-shrink: 0;
  font-size: 11px;
  color: #2b6cb0;
  font-weight: 600;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.chart-note {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #e2e8f0;
  font-size: 11px;
  color: #718096;
  display: flex;
  flex-wrap: wrap;
  gap: 4px 10px;
}

.chart-note .hint {
  color: #a0aec0;
}

.empty {
  padding: 12px 0;
  font-size: 12px;
  color: #a0aec0;
  text-align: center;
}
</style>
