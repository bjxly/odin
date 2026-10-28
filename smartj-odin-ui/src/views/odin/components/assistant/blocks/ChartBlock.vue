<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MessageBlock, ChartBlockPayload } from '@/types/odin'

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()

const MAX_BARS = 20

const p = computed<ChartBlockPayload>(() => (props.block?.payload ?? {}) as ChartBlockPayload)

const chartType = computed(() => p.value.chartType || 'bar')
const dimensions = computed(() => (p.value.dimensions ?? []).map((d) => typeof d === 'string' ? { field: d, label: d } : d))
const metrics = computed(() => (p.value.metrics ?? []).map((m) => typeof m === 'string' ? { field: m, label: m } : m))
const columns = computed<string[]>(() => (p.value.columns ?? []).map(String))
const rows = computed<unknown[][]>(() => (p.value.rows ?? []).map((r) => Array.isArray(r) ? r : []))
const plotRows = computed(() => rows.value.slice(0, MAX_BARS))

function colIndexOf(field: string, fallback: number): number {
  const i = columns.value.indexOf(field)
  return i >= 0 ? i : fallback
}

const dimCol = computed(() => colIndexOf(dimensions.value[0]?.field ?? '', 0))

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
  col: number
  numeric: boolean
  values: (number | null)[]
}

const series = computed<MetricSeries[]>(() =>
  metrics.value.map((m, i) => {
    const col = colIndexOf(m.field, dimensions.value.length + i)
    const values = plotRows.value.map((r) => toNum(r[col]))
    const hits = values.filter((v) => v !== null).length
    const numeric = plotRows.value.length > 0 && hits >= Math.ceil(plotRows.value.length / 2)
    return { name: m.label || m.field, col, numeric, values }
  }),
)

const numericSeries = computed(() => series.value.filter((s) => s.numeric))
const activeName = ref('')

watch(
  () => numericSeries.value.map((s) => s.name).join('|'),
  () => {
    if (!numericSeries.value.some((s) => s.name === activeName.value)) {
      activeName.value = numericSeries.value[0]?.name ?? ''
    }
  },
  { immediate: true },
)

const active = computed(() => numericSeries.value.find((s) => s.name === activeName.value) ?? numericSeries.value[0] ?? null)

const bars = computed(() =>
  plotRows.value.map((r, i) => ({
    label: String(r[dimCol.value] ?? '—'),
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
  return `${Math.max(2, Math.min(100, (Math.abs(v) / maxValue.value) * 100)).toFixed(1)}%`
}

function num(v: number | null): string {
  if (v == null || !Number.isFinite(v)) return '—'
  if (Math.abs(v) >= 1e6) return v.toExponential(2)
  return Number.isInteger(v) ? v.toLocaleString() : v.toFixed(2)
}
</script>

<template>
  <div class="chart-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || t('odinAssistant.blockChart') }}</span>
      <el-tag size="small" type="primary" effect="plain">{{ chartType }}</el-tag>
      <el-tag v-if="dimensions.length" size="small" effect="plain">
        {{ dimensions.map((d) => d.label).join(', ') }}
      </el-tag>
    </div>

    <div v-if="!plotRows.length" class="empty">{{ t('odinAssistant.chartEmpty') }}</div>
    <div v-else-if="!active" class="empty">{{ t('odinAssistant.chartNoNumeric') }}</div>

    <template v-else>
      <div v-if="numericSeries.length > 1" class="metric-picker">
        <span class="picker-label">{{ t('odinAssistant.chartMetricPicker') }}</span>
        <el-select v-model="activeName" size="small" style="width: 160px">
          <el-option v-for="s in numericSeries" :key="s.name" :value="s.name" :label="s.name" />
        </el-select>
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
        <span v-if="rows.length > MAX_BARS">仅展示前 {{ MAX_BARS }} 行（共 {{ rows.length }} 行）</span>
        <span class="hint">{{ t('odinAssistant.chartFallback') }}</span>
      </div>
    </template>
  </div>
</template>

<style scoped>
.chart-block {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
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
  color: var(--el-text-color-primary);
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
  color: var(--el-text-color-secondary);
}
.bar-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.bar-label {
  width: 110px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--el-text-color-regular);
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bar-track {
  flex: 1;
  height: 14px;
  background: var(--el-fill-color-light);
  border-radius: 7px;
  overflow: hidden;
}
.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--el-color-primary-light-3), var(--el-color-primary));
  border-radius: 7px;
  transition: width 0.3s ease;
}
.bar-value {
  width: 72px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--el-color-primary);
  font-weight: 600;
  font-family: 'JetBrains Mono', Consolas, monospace;
}
.chart-note {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed var(--el-border-color-lighter);
  font-size: 11px;
  color: var(--el-text-color-secondary);
  display: flex;
  flex-wrap: wrap;
  gap: 4px 10px;
}
.chart-note .hint {
  color: var(--el-text-color-placeholder);
}
.empty {
  padding: 12px 0;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  text-align: center;
}
</style>
