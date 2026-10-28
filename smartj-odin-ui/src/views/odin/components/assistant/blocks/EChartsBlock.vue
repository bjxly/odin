<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MessageBlock, ChartBlockPayload } from '@/types/odin'

/**
 * ECharts 图表块：使用 echarts 渲染真实图表。
 * 通过 defineAsyncComponent 懒加载（见 blocks/index.ts），
 * 仅在消息流出现 chart 块时才下载 echarts chunk。
 */
const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()

const chartRef = ref<HTMLDivElement | null>(null)
const loadError = ref(false)
let chartInstance: any = null

const p = computed<ChartBlockPayload>(() => (props.block?.payload ?? {}) as ChartBlockPayload)
const chartType = computed(() => p.value.chartType || 'bar')
const dimensions = computed(() => (p.value.dimensions ?? []).map((d) => typeof d === 'string' ? { field: d, label: d } : d))
const metrics = computed(() => (p.value.metrics ?? []).map((m) => typeof m === 'string' ? { field: m, label: m } : m))
const columns = computed<string[]>(() => (p.value.columns ?? []).map(String))
const rows = computed<unknown[][]>(() => (p.value.rows ?? []).map((r) => Array.isArray(r) ? r : []))

function colIndexOf(field: string, fallback: number): number {
  const i = columns.value.indexOf(field)
  return i >= 0 ? i : fallback
}

const dimCol = computed(() => colIndexOf(dimensions.value[0]?.field ?? '', 0))

function buildOption() {
  const xData = rows.value.map((r) => String(r[dimCol.value] ?? ''))
  const seriesList = metrics.value.map((m, i) => {
    const col = colIndexOf(m.field, dimensions.value.length + i)
    return {
      name: m.label || m.field,
      type: chartType.value === 'line' ? 'line' : chartType.value === 'pie' ? 'pie' : 'bar',
      data: chartType.value === 'pie'
        ? rows.value.map((r, ri) => ({ name: xData[ri], value: Number(r[col]) || 0 }))
        : rows.value.map((r) => Number(r[col]) || 0),
      smooth: true,
    }
  })

  if (chartType.value === 'pie') {
    return {
      tooltip: { trigger: 'item' },
      legend: { bottom: 0, type: 'scroll' },
      series: seriesList.slice(0, 1),
    }
  }

  return {
    tooltip: { trigger: 'axis' },
    legend: { bottom: 0, data: metrics.value.map((m) => m.label || m.field) },
    grid: { left: 50, right: 20, top: 20, bottom: 40 },
    xAxis: { type: 'category', data: xData, axisLabel: { rotate: xData.length > 8 ? 30 : 0 } },
    yAxis: { type: 'value' },
    series: seriesList,
  }
}

async function initChart() {
  if (!chartRef.value || !rows.value.length) return
  try {
    const echarts = await import('echarts')
    if (!chartRef.value) return
    chartInstance = echarts.init(chartRef.value)
    chartInstance.setOption(buildOption())
  } catch {
    loadError.value = true
  }
}

function resizeChart() {
  chartInstance?.resize()
}

onMounted(() => {
  void nextTick(initChart)
  window.addEventListener('resize', resizeChart)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeChart)
  chartInstance?.dispose()
  chartInstance = null
})

watch(() => [rows.value.length, chartType.value], () => {
  chartInstance?.dispose()
  chartInstance = null
  void nextTick(initChart)
})
</script>

<template>
  <div class="echarts-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || t('odinAssistant.blockChart') }}</span>
      <el-tag size="small" type="primary" effect="plain">{{ chartType }}</el-tag>
    </div>

    <div v-if="!rows.length" class="empty">{{ t('odinAssistant.chartEmpty') }}</div>

    <!-- ECharts 加载失败时降级到 CSS ChartBlock -->
    <ChartBlockFallback v-else-if="loadError" :block="block" />

    <div v-else ref="chartRef" class="chart-container" />
  </div>
</template>

<script lang="ts">
import ChartBlockFallback from './ChartBlock.vue'
export default { components: { ChartBlockFallback } }
</script>

<style scoped>
.echarts-block {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
  padding: 10px 12px;
}
.block-head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 10px;
}
.block-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.chart-container {
  width: 100%;
  height: 280px;
}
.empty {
  padding: 12px 0;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  text-align: center;
}
</style>
