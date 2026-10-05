<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { ElTag } from 'element-plus'
import type { BlockVO } from '@/types/assistant'
import { normalizeChartPayload } from '@/types/assistant'
import { formatNumber } from '@/utils/format'
import type { ChartOption, EChartsInstance } from '@/utils/echarts'
import ChartBlock from './ChartBlock.vue'

/**
 * chart 块（ECharts 实现）。
 *
 * 注册方式：在 `blocks/index.ts` 中以 `registerBlockRenderer('chart', EChartsBlock)`
 * **覆盖**内置的 CSS 版 ChartBlock，BlockRenderer 与消息流装配层零改动。
 *
 * 关键约定：
 * 1. **按需 + 异步**：echarts 通过 `import('@/utils/echarts')` 动态加载，
 *    只有真正出现 chart 块时才会拉取 echarts/zrender chunk，不进首屏；
 *    加载失败（弱网 / 资源不可达）自动降级为 CSS 版 ChartBlock。
 * 2. **列下标按 field 名反查**：payload 的 `columns` 给出与 `rows` 同序的列名，
 *    `dimensions[i].field` / `metrics[i].field` 在其中查下标，
 *    彻底摆脱「维度恒在第 0 列、度量为 i+1」的位置假设（聚合结果列序由 SQL 决定）。
 * 3. **chart_type 驱动形态**：bar / line / pie 走直角或极坐标图，
 *    metric（纯标量聚合）直接渲染 KPI 指标卡 —— 一个数字不需要坐标系。
 */
const props = defineProps<{ block: BlockVO }>()

const nc = computed(() => normalizeChartPayload(props.block?.payload))

/** echarts 是否已就绪 / 加载失败（失败即降级） */
const ready = ref(false)
const failed = ref(false)

const host = ref<HTMLDivElement | null>(null)
const chart = shallowRef<EChartsInstance | null>(null)
let observer: ResizeObserver | null = null

/** 图表画布高度（px） */
const CHART_HEIGHT = 280
/** 分类轴最多渲染多少个类目，超出部分在提示中说明 */
const MAX_CATEGORIES = 50

function toNum(v: unknown): number | null {
  if (typeof v === 'number') return Number.isFinite(v) ? v : null
  if (typeof v === 'string') {
    const s = v.trim()
    if (!s) return null
    const n = Number(s.replace(/,/g, ''))
    return Number.isFinite(n) ? n : null
  }
  return null
}

/** 按 field 名反查列下标；查不到时退化为「维度列在前、度量列在后」的推导下标 */
function colIndexOf(field: string, fallback: number): number {
  const i = nc.value.columns.indexOf(field)
  return i >= 0 ? i : fallback
}

/** 维度列下标（分类轴 / 饼图扇区名） */
const dimCol = computed(() => colIndexOf(nc.value.dimensions[0]?.field ?? '', 0))

/** 度量列下标（与 nc.metrics 同序） */
const metricCols = computed<number[]>(() =>
  nc.value.metrics.map((m, i) => colIndexOf(m.field, nc.value.dimensions.length + i)),
)

/**
 * 纯标量聚合 → KPI 指标卡。
 * 触发条件：后端显式建议 `metric`，或没有任何维度列（无从建立分类轴）。
 */
const isKpi = computed(() => {
  if (!nc.value.rows.length || !nc.value.metrics.length) return false
  if (nc.value.chartType === 'metric') return true
  return nc.value.dimensions.length === 0
})

const kpiItems = computed<Array<{ label: string; value: string; agg: string }>>(() => {
  if (!isKpi.value) return []
  const row = (nc.value.rows[0] ?? []) as unknown[]
  return nc.value.metrics.map((m, i) => ({
    label: m.label,
    value: formatNumber(row[metricCols.value[i]]),
    agg: String(m.agg ?? '').toUpperCase(),
  }))
})

const categories = computed<string[]>(() =>
  nc.value.rows.slice(0, MAX_CATEGORIES).map((r) => String(((r as unknown[]) ?? [])[dimCol.value] ?? '—')),
)

/** 实际参与绘图的行（受 MAX_CATEGORIES 限制） */
const plotRows = computed<unknown[][]>(() => nc.value.rows.slice(0, MAX_CATEGORIES) as unknown[][])

const hasData = computed(() => nc.value.rows.length > 0 && nc.value.metrics.length > 0)

/** 渲染形态：kpi / chart / empty / fallback（echarts 加载失败） */
const mode = computed<'kpi' | 'chart' | 'empty' | 'fallback'>(() => {
  if (failed.value) return 'fallback'
  if (!hasData.value) return 'empty'
  if (isKpi.value) return 'kpi'
  return 'chart'
})

/** 构造 ECharts option；仅在 chart 形态下调用 */
function buildOption(): ChartOption {
  const type = nc.value.chartType

  if (type === 'pie') {
    // 饼图只表达单个度量的构成比，多度量时取第一个（其余在下方标签中说明）
    const idx = metricCols.value[0] ?? nc.value.dimensions.length
    const data = plotRows.value.map((r, i) => ({
      name: categories.value[i] ?? '—',
      value: toNum(r[idx]) ?? 0,
    }))
    return {
      tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
      legend: { type: 'scroll', bottom: 0, textStyle: { fontSize: 11 } },
      series: [
        {
          name: nc.value.metrics[0]?.label ?? '指标',
          type: 'pie',
          radius: ['42%', '68%'],
          center: ['50%', '44%'],
          avoidLabelOverlap: true,
          itemStyle: { borderColor: '#ffffff', borderWidth: 2 },
          label: { formatter: '{b}\n{d}%', fontSize: 11 },
          data,
        },
      ],
    }
  }

  // bar / line 共用直角坐标系；series 必须分支构造，
  // 否则 `type: 'bar' | 'line'` 无法归入 ComposeOption 的联合成员（且会触发多余属性检查）。
  const axis = {
    tooltip: {
      trigger: 'axis' as const,
      axisPointer: { type: type === 'line' ? ('line' as const) : ('shadow' as const) },
      valueFormatter: (v: unknown) => formatNumber(v),
    },
    legend: { type: 'scroll' as const, bottom: 0, textStyle: { fontSize: 11 } },
    grid: { left: 8, right: 16, top: 20, bottom: 44, containLabel: true },
    xAxis: {
      type: 'category' as const,
      data: categories.value,
      axisLabel: { fontSize: 11, interval: 0, rotate: categories.value.length > 6 ? 30 : 0 },
    },
    yAxis: { type: 'value' as const, axisLabel: { fontSize: 11 } },
  }

  if (type === 'line') {
    return {
      ...axis,
      series: nc.value.metrics.map((m, i) => ({
        name: m.label,
        type: 'line' as const,
        smooth: true,
        showSymbol: plotRows.value.length <= 30,
        data: plotRows.value.map((r) => toNum(r[metricCols.value[i] ?? 0])),
      })),
    }
  }

  return {
    ...axis,
    series: nc.value.metrics.map((m, i) => ({
      name: m.label,
      type: 'bar' as const,
      barMaxWidth: 40,
      data: plotRows.value.map((r) => toNum(r[metricCols.value[i] ?? 0])),
    })),
  }
}

function render() {
  if (!chart.value || mode.value !== 'chart') return
  chart.value.setOption(buildOption(), true)
}

async function ensureChart() {
  if (mode.value !== 'chart') return
  if (!chart.value) {
    if (!ready.value) {
      try {
        const ec = (await import('@/utils/echarts')).default
        ready.value = true
        // 异步 import 期间组件可能已卸载或形态已变化
        if (!host.value || mode.value !== 'chart') return
        chart.value = ec.init(host.value)
      } catch {
        failed.value = true
        return
      }
    } else if (host.value) {
      const ec = (await import('@/utils/echarts')).default
      chart.value = ec.init(host.value)
    }
    if (!chart.value || !host.value) return
    if (!observer && typeof ResizeObserver !== 'undefined') {
      observer = new ResizeObserver(() => chart.value?.resize())
      observer.observe(host.value)
    }
  }
  render()
}

function dispose() {
  observer?.disconnect()
  observer = null
  chart.value?.dispose()
  chart.value = null
}

// 形态切换（kpi ↔ chart）会重建/销毁 DOM，需要重新 init
watch(
  mode,
  async (m, prev) => {
    if (m !== 'chart') {
      dispose()
      return
    }
    if (prev !== 'chart') await nextTick()
    await ensureChart()
  },
  { immediate: true },
)

// payload 变化（SSE 增量下发）→ 只重绘，不重建实例
watch(
  () => [nc.value.chartType, nc.value.rows.length, nc.value.columns.join('|')].join('#'),
  () => {
    if (mode.value === 'chart') render()
  },
)

onBeforeUnmount(dispose)

const dimLabels = computed(() => nc.value.dimensions.map((d) => d.label))
const metricLabels = computed(() =>
  nc.value.metrics.map((m) => (m.agg ? `${m.label}（${String(m.agg).toUpperCase()}）` : m.label)),
)
const omitted = computed(() => Math.max(0, nc.value.rows.length - MAX_CATEGORIES))
</script>

<template>
  <!-- echarts 加载失败：降级为 CSS 条形图（BlockRenderer 无感知） -->
  <ChartBlock v-if="mode === 'fallback'" :block="block" />

  <div v-else class="echarts-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || '可视化建议' }}</span>
      <ElTag size="small" type="primary" effect="plain">{{ nc.chartType }}</ElTag>
      <ElTag v-if="nc.aggregated" size="small" type="warning" effect="plain">聚合结果</ElTag>
      <ElTag v-if="nc.topN" size="small" type="info" effect="plain">TOP-{{ nc.topN }}</ElTag>
      <ElTag v-if="dimLabels.length" size="small" effect="plain">
        维度：{{ dimLabels.join('、') }}
      </ElTag>
      <ElTag v-if="metricLabels.length" size="small" type="success" effect="plain">
        度量：{{ metricLabels.join('、') }}
      </ElTag>
    </div>

    <div v-if="mode === 'empty'" class="empty">无可视化数据（结果集为空或缺少度量列）</div>

    <!-- 纯标量聚合：KPI 指标卡 -->
    <div v-else-if="mode === 'kpi'" class="kpi-wrap">
      <div v-for="(k, i) in kpiItems" :key="i" class="kpi-card">
        <div class="kpi-value">{{ k.value }}</div>
        <div class="kpi-label">
          {{ k.label }}
          <span v-if="k.agg" class="kpi-agg">{{ k.agg }}</span>
        </div>
      </div>
    </div>

    <template v-else>
      <div ref="host" class="chart-host" :style="{ height: `${CHART_HEIGHT}px` }" />
      <div v-if="omitted > 0" class="chart-note">
        仅绘制前 {{ MAX_CATEGORIES }} 个分组（共 {{ nc.rows.length }} 个），可在结果表中查看全部。
      </div>
    </template>
  </div>
</template>

<style scoped>
.echarts-block {
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
  margin-bottom: 8px;
}

.block-title {
  font-size: 13px;
  font-weight: 700;
  color: #1a365d;
}

.chart-host {
  width: 100%;
}

.chart-note {
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px dashed #e2e8f0;
  font-size: 11px;
  color: #718096;
}

.kpi-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 6px 0;
}

.kpi-card {
  flex: 1 1 150px;
  min-width: 150px;
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

.empty {
  padding: 12px 0;
  font-size: 12px;
  color: #a0aec0;
  text-align: center;
}
</style>
