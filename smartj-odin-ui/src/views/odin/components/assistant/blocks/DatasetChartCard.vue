<script setup lang="ts">
import { computed, defineAsyncComponent, inject, nextTick, onBeforeUnmount, onMounted, ref, watch, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { init, use, type EChartsType } from 'echarts/core'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { DataZoomComponent, GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import {
  Close,
  DataAnalysis,
  DataLine,
  Download,
  FullScreen,
  Grid,
  PieChart as PieChartIcon,
  Refresh,
  RefreshLeft,
  Sort,
  TrendCharts,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { BlockContextKey } from './blockContext'
import { createFromChart } from '@/api/odin/insight'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import type { ChartBlockPayload, ChartHint, QueryAggMethod, ChartSelection, QueryDataset, QueryChartType } from '@/types/odin'
import {
  buildChartOption,
  chartIssue,
  computePivot,
  defaultSelection,
  dimensionColumns,
  filterRows,
  hasTotalRow,
  numericColumns,
  selectHint,
  transposeDataset,
  withoutTotalRows,
} from '@/utils/queryChart'

use([BarChart, LineChart, PieChart, DataZoomComponent, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

// 旧版 chart 块（含 columns/rows 的整块 payload）兼容层：仅当无法解析 dataset 时才异步加载，避免拖入完整 echarts。
const LegacyChart = defineAsyncComponent(() => import('./EChartsBlock.vue'))

const props = defineProps<{ block: { type: string; title?: string; payload?: unknown } }>()
const { t } = useI18n()
const ctx = inject(BlockContextKey, undefined)
const workspace = useOdinWorkspaceStore()

const chartTypes: QueryChartType[] = ['bar', 'line', 'pie']
const chartIcons: Record<QueryChartType, Component> = { bar: DataLine, line: TrendCharts, pie: PieChartIcon }
const aggMethods: QueryAggMethod[] = ['sum', 'avg', 'max', 'min', 'count']
// 「无」不能直接用 null：EP 默认把 null 当空值显示 placeholder 而非选项名。
const NO_SERIES = -1

const datasetId = computed(() => (props.block?.payload as ChartBlockPayload | undefined)?.datasetId ?? '')
const raw = computed<QueryDataset | null>(() => ctx?.datasets?.find((d) => d.id === datasetId.value) ?? null)
const hint = computed<ChartHint | undefined>(() => (datasetId.value ? ctx?.chartHints?.[datasetId.value] : undefined))

/** 分析表视图轴向：初始由模型转置提示决定，用户可切换（仅 source=table 可用） */
const transposed = ref(false)
/** 分析表默认剔除合计行（表格视图同步剔除），用户可勾选包含 */
const includeTotals = ref(false)

/** 卡片内所有消费方统一走归一化视图：转置 → 剔合计，避免两套坐标语义 */
const data = computed<QueryDataset | null>(() => {
  let d = raw.value
  if (!d) return null
  if (transposed.value && d.source === 'table') d = transposeDataset(d)
  return includeTotals.value ? d : withoutTotalRows(d)
})
const canTranspose = computed(() => raw.value?.source === 'table' && (raw.value.columns.length ?? 0) >= 2)
const showTotalToggle = computed(() => raw.value?.source === 'table' && hasTotalRow(raw.value))
/** 外部视图变更计数（推荐迟到/转置提示初始化）：命中时覆盖选择；用户自身切换不计数，保留已选轴 */
const viewSeq = ref(0)

const EMPTY_SELECTION: ChartSelection = { type: 'bar', x: 0, ys: [], horizontal: false, seriesDim: null, agg: 'sum' }
const selection = ref<ChartSelection>(data.value ? defaultSelection(data.value) : EMPTY_SELECTION)
const touched = ref(false)
const showTable = ref(false)
const page = ref(1)
const pageSize = ref(10)

const title = computed(() => data.value?.title || props.block?.title || t('odinAssistant.chart.title'))
const metrics = computed(() => (data.value ? numericColumns(data.value) : []))
const dimensions = computed(() => (data.value ? dimensionColumns(data.value) : []))
const allCategories = computed(() => (data.value ? [...new Set(data.value.rows.map((row) => row[selection.value.x]))] : []))
const categories = ref<string[]>([])
const rows = computed(() => (data.value ? filterRows(data.value, selection.value.x, categories.value) : []))
const pageRows = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return rows.value.slice(start, start + pageSize.value)
})
const seriesModel = computed({
  get: () => selection.value.seriesDim ?? NO_SERIES,
  set: (value: number) => {
    markTouched()
    selection.value.seriesDim = value === NO_SERIES ? null : value
  },
})
const pivot = computed(() => computePivot(rows.value, selection.value.x, selection.value.seriesDim))
const aggNoticeKey = computed(() => (selection.value.seriesDim === null ? 'odinAssistant.chart.aggNoticeX' : 'odinAssistant.chart.aggNoticeXY'))
const issue = computed(() => (data.value ? chartIssue(data.value, rows.value, selection.value) : 'empty'))
const chartFailed = ref(false)

// 内嵌表格视图：把字符串二维数组转为按列索引取值的对象数组
const tableColumns = computed(() =>
  (data.value?.columns ?? []).map((name, i) => ({
    prop: String(i),
    label: name,
    minWidth: 110,
    align: metrics.value.includes(i) ? 'right' : 'left',
  })),
)
const tableData = computed(() =>
  pageRows.value.map((r) => {
    const obj: Record<string, string> = {}
    ;(data.value?.columns ?? []).forEach((_, i) => {
      obj[String(i)] = r[i] ?? ''
    })
    return obj
  }),
)

const chartHost = ref<HTMLDivElement>()
const chartStage = ref<HTMLDivElement>()
const isFullscreen = ref(false)
let chart: EChartsType | undefined
let resizeObserver: ResizeObserver | undefined
let themeObserver: MutationObserver | undefined
let frame = 0
let mounted = false

function columnLabel(name: string, index: number) {
  return (data.value?.columns ?? []).filter((n) => n === name).length > 1 ? `${name} (${index + 1})` : name
}
// 仅给纯数值单元格加千分位；编号列与年份列已被 numericColumns 排除。
function formatNumberCell(value: unknown): string {
  if (typeof value !== 'string') return String(value ?? '')
  const negative = value.startsWith('-')
  const unsigned = negative ? value.slice(1) : value
  const dot = unsigned.indexOf('.')
  const int = dot < 0 ? unsigned : unsigned.slice(0, dot)
  const frac = dot < 0 ? '' : unsigned.slice(dot)
  if (!/^\d+$/.test(int)) return value
  return `${negative ? '-' : ''}${int.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}${frac}`
}
function aggLabel(method: QueryAggMethod) {
  return t(`odinAssistant.chart.agg_${method}`)
}
function selectType(type: QueryChartType) {
  selection.value.type = type
  onTypeChange()
}
function markTouched() {
  touched.value = true
}
function onXChange() {
  markTouched()
  categories.value = [...allCategories.value]
  selection.value.ys = selection.value.ys.filter((i) => i !== selection.value.x)
  if (selection.value.seriesDim === selection.value.x) selection.value.seriesDim = null
}
function onTypeChange() {
  markTouched()
  if (selection.value.type === 'pie') {
    selection.value.ys = selection.value.ys.slice(0, 1)
    selection.value.seriesDim = null
  }
}
function swapAxes() {
  markTouched()
  selection.value.horizontal = !selection.value.horizontal
}
function setCategories(all: boolean) {
  markTouched()
  categories.value = all ? [...allCategories.value] : []
}
function resetSelection() {
  markTouched()
  if (!data.value) return
  applyHint()
}

function toggleTranspose() {
  markTouched()
  transposed.value = !transposed.value
}

/** 按当前视图（含轴向/合计开关）重新应用推荐，推荐不过校验则回退启发式。 */
function applyHint() {
  if (!data.value) return
  selection.value = selectHint(data.value, hint.value) || defaultSelection(data.value)
  categories.value = [...new Set(data.value.rows.map((row) => row[selection.value.x]))]
  page.value = 1
}

// 视图更换后旧列索引失效必须重建：推荐迟到（未交互）或转置提示初始化时应用推荐；
// 用户已交互且是自己切换的视图（转置/合计）则回退启发式，不覆盖其手动选择的语义。
watch(data, (d) => {
  if (!d) return
  if (!touched.value || viewSeq.value > 0) applyHint()
  else selection.value = defaultSelection(d)
  categories.value = [...new Set(d.rows.map((row) => row[selection.value.x]))]
  page.value = 1
}, { immediate: true })
// 推荐晚于图表块到达：覆盖启发式初选；携带转置提示时同步切换轴向
watch(hint, (h) => {
  if (!h || !raw.value) return
  if (h.transposed && raw.value.source === 'table' && !transposed.value) transposed.value = true
  viewSeq.value++
}, { immediate: true })
watch(rows, () => {
  page.value = 1
})

function cssValue(name: string) {
  return window.getComputedStyle(chartHost.value!).getPropertyValue(name).trim()
}
function renderChart() {
  const d = data.value
  if (!mounted || !d || !chartHost.value || issue.value || !chartHost.value.clientWidth) return
  try {
    chart ??= init(chartHost.value)
    chart.setOption(
      buildChartOption(
        d,
        rows.value,
        selection.value,
        {
          text: cssValue('--el-text-color-regular'),
          border: cssValue('--el-border-color-lighter'),
          background: cssValue('--el-bg-color'),
          palette: ['primary', 'success', 'warning', 'danger', 'info'].map((name) => cssValue(`--el-color-${name}`)),
        },
        { noValue: t('odinAssistant.chart.noValue'), emptyGroup: t('odinAssistant.chart.emptyGroup') },
      ),
      { notMerge: true },
    )
    chart.resize()
    chartFailed.value = false
  } catch {
    chartFailed.value = true
  }
}
function scheduleRender() {
  if (!mounted) return
  cancelAnimationFrame(frame)
  frame = requestAnimationFrame(renderChart)
}
function exportPNG() {
  if (!chart || issue.value) return
  try {
    const a = document.createElement('a')
    a.download = `${datasetId.value.replace(/[^a-zA-Z0-9_-]/g, '_')}.png`
    a.href = chart.getDataURL({ type: 'png', pixelRatio: 2, backgroundColor: cssValue('--el-bg-color') })
    a.click()
  } catch {
    chartFailed.value = true
  }
}

// 全屏：元素被 UA 提到顶层，流内少了高度会夹到别处，需捕获/还原祖先滚动。
let savedScroll: { el: HTMLElement; top: number }[] = []
function captureScroll() {
  savedScroll = []
  for (let el = chartStage.value?.parentElement ?? null; el; el = el.parentElement) {
    if (el.scrollHeight > el.clientHeight) savedScroll.push({ el, top: el.scrollTop })
  }
}
function restoreScroll() {
  const list = savedScroll
  savedScroll = []
  const apply = () => list.forEach(({ el, top }) => {
    el.scrollTop = top
  })
  apply()
  requestAnimationFrame(apply)
}
function toggleFullscreen() {
  if (!chartStage.value) return
  if (document.fullscreenElement) void document.exitFullscreen()
  else {
    captureScroll()
    void chartStage.value.requestFullscreen()
  }
}
function exitFullscreen() {
  if (document.fullscreenElement) void document.exitFullscreen()
}
function onFullscreenChange() {
  isFullscreen.value = document.fullscreenElement === chartStage.value
  if (!isFullscreen.value) restoreScroll()
}

// ===== 保存为洞察看板 =====
const insightVisible = ref(false)
const insightSaving = ref(false)
const insightName = ref('')
const insightOntologyId = ref('')
function openInsight() {
  insightName.value = title.value
  insightOntologyId.value = workspace.activeOntologyId || ''
  if (!workspace.ontologies.length) void workspace.loadOntologies()
  insightVisible.value = true
}
async function submitInsight() {
  const d = data.value
  if (!d || !insightName.value.trim()) return
  insightSaving.value = true
  try {
    await createFromChart({
      name: insightName.value.trim(),
      ontologyId: insightOntologyId.value || undefined,
      title: title.value,
      chartType: selection.value.type,
      columns: d.columns,
      rows: d.rows,
      chartOption: JSON.parse(JSON.stringify(selection.value)),
    })
    ElMessage.success(t('odinAssistant.chart.insightSaveOk'))
    insightVisible.value = false
  } catch {
    ElMessage.error(t('odinAssistant.chart.insightSaveFail'))
  } finally {
    insightSaving.value = false
  }
}

watch([rows, selection, issue], scheduleRender, { deep: true, flush: 'post' })
onMounted(async () => {
  mounted = true
  document.addEventListener('fullscreenchange', onFullscreenChange)
  await nextTick()
  if (!mounted || !chartHost.value || !data.value) return
  resizeObserver = new ResizeObserver(scheduleRender)
  resizeObserver.observe(chartHost.value)
  themeObserver = new MutationObserver(scheduleRender)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme', 'class', 'style'] })
  scheduleRender()
})
onBeforeUnmount(() => {
  mounted = false
  cancelAnimationFrame(frame)
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  resizeObserver?.disconnect()
  themeObserver?.disconnect()
  chart?.dispose()
  chart = undefined
})
</script>

<template>
  <!-- 无法解析数据集（旧版整块 payload）→ 交给兼容层渲染，避免历史消息白屏 -->
  <LegacyChart v-if="!data" :block="(props as any)" />

  <section v-else class="dataset-chart-card">
    <header class="card-header">
      <div class="card-heading">
        <strong>{{ title }}</strong>
        <span class="row-count">{{ t('odinAssistant.chart.rows', { n: data.rows.length }) }}</span>
      </div>
      <div class="card-actions">
        <el-tooltip :content="t('odinAssistant.chart.table')" placement="top">
          <el-button size="small" text bg :type="showTable ? 'primary' : ''" @click="showTable = !showTable">
            <el-icon><Grid /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip v-for="type in chartTypes" :key="type" :content="t(`odinAssistant.chart.${type}`)" placement="top">
          <el-button size="small" text bg :type="selection.type === type ? 'primary' : ''" @click="selectType(type)">
            <el-icon><component :is="chartIcons[type]" /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip :content="t('odinAssistant.chart.export')" placement="top">
          <el-button size="small" text bg :disabled="!!issue || chartFailed" @click="exportPNG">
            <el-icon><Download /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip :content="t('odinAssistant.chart.saveInsight')" placement="top">
          <el-button size="small" text bg @click="openInsight">
            <el-icon><DataAnalysis /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip :content="t(isFullscreen ? 'odinAssistant.chart.exitFullscreen' : 'odinAssistant.chart.fullscreen')" placement="top">
          <el-button size="small" text bg :disabled="!!issue || chartFailed" @click="toggleFullscreen">
            <el-icon><FullScreen /></el-icon>
          </el-button>
        </el-tooltip>
      </div>
    </header>

    <p v-if="data.truncated" class="truncated-note">
      {{ t('odinAssistant.chart.truncated', { n: data.rows.length }) }}
    </p>

    <div class="chart-controls">
      <el-select v-model="selection.x" class="control-x" size="small" @change="onXChange">
        <template #prefix>{{ t('odinAssistant.chart.xShort') }}</template>
        <el-option v-for="(name, index) in data.columns" :key="index" :label="columnLabel(name, index)" :value="index" />
      </el-select>
      <el-select
        v-model="categories"
        class="control-categories"
        size="small"
        multiple
        filterable
        clearable
        collapse-tags
        collapse-tags-tooltip
        @change="markTouched"
      >
        <template #prefix>{{ t('odinAssistant.chart.categoriesShort') }}</template>
        <template #header>
          <div class="category-actions">
            <el-button link size="small" @click="setCategories(true)">{{ t('odinAssistant.chart.selectAll') }}</el-button>
            <el-button link size="small" @click="setCategories(false)">{{ t('odinAssistant.chart.clear') }}</el-button>
          </div>
        </template>
        <el-option v-for="value in allCategories" :key="value" :label="value || t('odinAssistant.chart.emptyValue')" :value="value" />
      </el-select>
      <el-select v-model="seriesModel" class="control-series" size="small" :disabled="selection.type === 'pie'">
        <template #prefix>{{ t('odinAssistant.chart.seriesShort') }}</template>
        <el-option :label="t('odinAssistant.chart.seriesNone')" :value="NO_SERIES" />
        <el-option v-for="index in dimensions" :key="index" :label="columnLabel(data.columns[index], index)" :value="index" :disabled="index === selection.x" />
      </el-select>
      <el-select
        v-model="selection.ys"
        class="control-y"
        size="small"
        multiple
        clearable
        collapse-tags
        collapse-tags-tooltip
        :multiple-limit="selection.type === 'pie' ? 1 : 0"
        @change="markTouched"
      >
        <template #prefix>{{ t('odinAssistant.chart.yShort') }}</template>
        <el-option v-for="index in metrics" :key="index" :label="columnLabel(data.columns[index], index)" :value="index" :disabled="index === selection.x" />
      </el-select>
      <el-select v-model="selection.agg" class="control-agg" size="small" :disabled="!pivot.multi" @change="markTouched">
        <template #prefix>{{ t('odinAssistant.chart.aggShort') }}</template>
        <el-option v-for="method in aggMethods" :key="method" :label="aggLabel(method)" :value="method" />
      </el-select>
      <el-tooltip v-if="canTranspose" :content="t('odinAssistant.chart.transpose')" placement="top">
        <el-button size="small" text bg :type="transposed ? 'primary' : ''" @click="toggleTranspose">
          <el-icon><Refresh /></el-icon>
        </el-button>
      </el-tooltip>
      <el-checkbox v-if="showTotalToggle" v-model="includeTotals" size="small" class="control-totals">
        {{ t('odinAssistant.chart.includeTotals') }}
      </el-checkbox>
      <el-tooltip :content="t('odinAssistant.chart.swap')" placement="top">
        <el-button size="small" text bg :type="selection.horizontal ? 'primary' : ''" :disabled="selection.type === 'pie'" @click="swapAxes">
          <el-icon><Sort /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip :content="t('odinAssistant.chart.reset')" placement="top">
        <el-button size="small" text bg @click="resetSelection">
          <el-icon><RefreshLeft /></el-icon>
        </el-button>
      </el-tooltip>
    </div>

    <p v-if="pivot.multi && !issue && !chartFailed" class="agg-notice" role="status">
      {{ t(aggNoticeKey, { agg: aggLabel(selection.agg) }) }}
    </p>

    <div v-if="issue || chartFailed" class="chart-empty" role="status">
      {{ t(`odinAssistant.chart.${issue || 'renderFailed'}`) }}
    </div>
    <div ref="chartStage" class="chart-stage">
      <div v-show="!issue && !chartFailed" ref="chartHost" class="chart-host" role="img" :aria-label="title" />
      <el-button v-if="isFullscreen" class="stage-exit" size="small" text bg :title="t('odinAssistant.chart.exitFullscreen')" @click="exitFullscreen">
        <el-icon><Close /></el-icon>
      </el-button>
    </div>

    <div v-if="showTable" class="data-table">
      <el-table :data="tableData" size="small" border max-height="320">
        <el-table-column
          v-for="col in tableColumns"
          :key="col.prop"
          :prop="col.prop"
          :label="col.label"
          :min-width="col.minWidth"
          :align="(col.align as any)"
        >
          <template #default="scope">{{ col.align === 'right' ? formatNumberCell(scope.row[col.prop]) : scope.row[col.prop] }}</template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-if="rows.length > pageSize"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        class="table-pager"
        layout="prev, pager, next"
        :total="rows.length"
        background
      />
    </div>

    <el-dialog v-model="insightVisible" :title="t('odinAssistant.chart.saveInsight')" width="360px" append-to-body>
      <el-form label-width="84px">
        <el-form-item :label="t('odinAssistant.chart.insightName')">
          <el-input v-model="insightName" :placeholder="t('odinAssistant.chart.insightNamePh')" maxlength="128" />
        </el-form-item>
        <el-form-item :label="t('odinAssistant.chart.insightOntology')">
          <el-select v-model="insightOntologyId" clearable size="small" style="width: 100%">
            <el-option v-for="o in workspace.ontologies" :key="o.id" :label="o.name" :value="o.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="insightVisible = false">{{ t('odinAssistant.chart.insightCancel') }}</el-button>
        <el-button size="small" type="primary" :loading="insightSaving" :disabled="!insightName.trim()" @click="submitInsight">
          {{ t('odinAssistant.chart.insightSave') }}
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.dataset-chart-card {
  box-sizing: border-box;
  min-width: 0;
  width: 100%;
  margin: 6px 0;
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
}
.card-header,
.card-heading,
.card-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.card-header {
  justify-content: space-between;
}
.card-heading {
  min-width: 0;
}
.card-heading strong {
  overflow-wrap: anywhere;
}
.card-actions .el-button + .el-button {
  margin-left: 0;
}
.row-count {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.truncated-note {
  margin: 4px 0 0;
  color: var(--el-color-warning);
  font-size: 12px;
}
/* 控件单行；空间不足时整行横向滚动而非换行。标签移入各 select 的 prefix 插槽以腾出宽度。 */
.chart-controls {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
  font-size: 13px;
  overflow-x: auto;
  padding-bottom: 2px;
}
.chart-controls .el-select {
  min-width: 0;
}
.chart-controls .el-button {
  flex-shrink: 0;
}
.chart-controls .el-button + .el-button {
  margin-left: 0;
}
.chart-controls :deep(.el-select__prefix) {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.control-x {
  flex: 1 0 104px;
}
.control-categories {
  flex: 1.6 0 132px;
}
.control-series {
  flex: 1 0 100px;
}
.control-y {
  flex: 1.2 0 112px;
}
.control-agg {
  flex: 1 0 96px;
}
.control-totals {
  flex-shrink: 0;
}
.category-actions {
  display: flex;
  justify-content: space-between;
}
.agg-notice {
  margin: 8px 0 0;
  padding: 4px 10px;
  border: 1px solid var(--el-border-color-lighter);
  border-left: 3px solid var(--el-color-warning);
  border-radius: 6px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
  font-size: 12px;
  line-height: 1.6;
}
.chart-stage {
  position: relative;
}
.chart-stage:fullscreen {
  box-sizing: border-box;
  padding: 16px;
  background: var(--el-bg-color);
}
.chart-stage:fullscreen .chart-host {
  height: 100%;
  margin-top: 0;
}
.stage-exit {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 1;
}
.chart-host {
  width: 100%;
  height: 360px;
  margin-top: 10px;
}
.chart-empty {
  padding: 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.data-table {
  min-width: 0;
  margin-top: 10px;
}
.table-pager {
  margin-top: 8px;
  justify-content: flex-end;
}
</style>
