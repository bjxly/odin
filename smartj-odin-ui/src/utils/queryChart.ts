import type { EChartsOption, SeriesOption } from 'echarts'
import type { ChartHint, ChartSelection, QueryAggMethod, QueryDataset, QueryChartType } from '@/types/odin'

/**
 * 数据集 → ECharts 图表的纯函数工具集（移植自参考项目 queryChart.ts，
 * 字段对齐 ODIN QueryDataset：columns/rows/rowCount，ChartHint：dataId/xIndex/yIndices）。
 *
 * 设计原则：数据走侧信道（dataset 快照），模型只推荐列与图型；模型给出的任何令牌与列索引
 * 必须通过 selectHint 的严格校验，不合法即回退前端启发式，永不由模型直接生成 ECharts option。
 */

/** 空值不是 0，前导零编号与超安全整数不参与图形运算。表格始终使用原始字符串。 */
export function parseMetric(value: string): number | null {
  const s = (value ?? '').trim()
  if (!/^[+-]?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?$/.test(s)) return null
  const n = Number(s)
  if (!Number.isFinite(n) || Math.abs(n) > Number.MAX_SAFE_INTEGER) return null
  return n
}

/** 年份是维度不是指标：进了指标列就会被当 Y 轴数值选中，还会被千分位格式化成 2,025。 */
const yearColumn = /^(?:年|年份|年度|year)$/i

/** 可识别的时间/序时列（用于趋势图启发式）。 */
const timeColumn = /^(?:年|月份|月|日期|时间|日|季度|年份|年度|date|datetime|time|month|year|quarter|day)$/i

export function numericColumns(data: QueryDataset): number[] {
  const result: number[] = []
  data.columns.forEach((name, i) => {
    if (/(?:^id$|_id$|编号|编码|手机号|邮编)/i.test(name)) return
    if (yearColumn.test(name.trim())) return
    const values = data.rows.map((r) => r[i]).filter((v) => (v ?? '').trim() !== '' && (v ?? '').trim() !== 'NULL')
    if (values.length && values.every((v) => parseMetric(v) !== null)) result.push(i)
  })
  return result
}

/** 系列维度候选：非指标列（文本、年份、编号列都算维度）。 */
export function dimensionColumns(data: QueryDataset): number[] {
  const metrics = numericColumns(data)
  return data.columns.map((_, i) => i).filter((i) => !metrics.includes(i))
}

export function defaultSelection(data: QueryDataset): ChartSelection {
  const metrics = numericColumns(data)
  // 优先选可识别时间列作为 X 轴并画趋势线；否则首个非指标列 + 柱状图。
  const timeIdx = data.columns.findIndex((name, i) => timeColumn.test(name.trim()) && !metrics.includes(i))
  const dimIdx = data.columns.findIndex((_, i) => !metrics.includes(i))
  const hasTime = timeIdx >= 0
  const x = hasTime ? timeIdx : dimIdx < 0 ? 0 : dimIdx
  const type: QueryChartType = hasTime ? 'line' : 'bar'
  return { type, x, ys: metrics.filter((i) => i !== x).slice(0, 1), horizontal: false, seriesDim: null, agg: 'sum' }
}

/**
 * 模型推荐的图表选择：逐条严格校验，任一不满足即返回 null（前端回退启发式）。
 * 校验链：dataId 命中本轮数据集、type∈{bar,line,pie}、x 索引合法、y 全部是指标列且不含 x、无重复。
 */
export function selectHint(data: QueryDataset, value: unknown): ChartSelection | null {
  if (!value || typeof value !== 'object') return null
  const h = value as ChartHint
  const metrics = numericColumns(data)
  if (
    h.dataId !== data.id ||
    !['bar', 'line', 'pie'].includes(h.type) ||
    !Number.isInteger(h.xIndex) ||
    h.xIndex < 0 ||
    h.xIndex >= data.columns.length ||
    !Array.isArray(h.yIndices) ||
    !h.yIndices.length ||
    !h.yIndices.every((i) => Number.isInteger(i) && metrics.includes(i) && i !== h.xIndex) ||
    new Set(h.yIndices).size !== h.yIndices.length
  )
    return null
  const s: ChartSelection = { type: h.type, x: h.xIndex, ys: [...h.yIndices], horizontal: false, seriesDim: null, agg: 'sum' }
  return chartIssue(data, data.rows, s) ? null : s
}

export function filterRows(data: QueryDataset, x: number, categories: string[]): string[][] {
  const allowed = new Set(categories)
  return data.rows.filter((row) => allowed.has(row[x]))
}

/** 空单元格返回 null（柱缺失、折线断点），不按 0 处理；count 统计的是行数而非数值。 */
export function aggregate(values: string[], method: QueryAggMethod): number | null {
  if (!values.length) return null
  if (method === 'count') return values.length
  const nums: number[] = []
  for (const v of values) {
    const n = parseMetric(v)
    if (n !== null) nums.push(n)
  }
  if (!nums.length) return null
  if (method === 'avg') return nums.reduce((a, b) => a + b, 0) / nums.length
  if (method === 'max') return nums.reduce((a, b) => (b > a ? b : a))
  if (method === 'min') return nums.reduce((a, b) => (b < a ? b : a))
  return nums.reduce((a, b) => a + b, 0)
}

export interface Pivot {
  cats: string[]
  groups: string[]
  /** 键为 `${类别序}|${系列序}`，值为落在该格的原始行；缺失的组合没有键。 */
  cells: Map<string, string[][]>
  /** 是否存在多行单元格：决定是否披露聚合、是否允许切换聚合方式。 */
  multi: boolean
}

function pivotKey(cat: number, group: number) {
  return `${cat}|${group}`
}

/** 类别与系列都按首次出现顺序收集，与表格行序一致，避免图例顺序随字典序跳变。 */
export function computePivot(rows: string[][], x: number, seriesDim: number | null): Pivot {
  const cats: string[] = []
  const groups: string[] = []
  const catAt = new Map<string, number>()
  const groupAt = new Map<string, number>()
  const cells = new Map<string, string[][]>()
  let multi = false
  for (const row of rows) {
    const cat = row[x]
    const group = seriesDim === null ? '' : row[seriesDim]
    let ci = catAt.get(cat)
    if (ci === undefined) {
      ci = cats.length
      catAt.set(cat, ci)
      cats.push(cat)
    }
    let gi = groupAt.get(group)
    if (gi === undefined) {
      gi = groups.length
      groupAt.set(group, gi)
      groups.push(group)
    }
    const cell = cells.get(pivotKey(ci, gi))
    if (cell) {
      cell.push(row)
      multi = true
    } else {
      cells.set(pivotKey(ci, gi), [row])
    }
  }
  return { cats, groups, cells, multi }
}

/** 返回空态文案 key（'' 表示可正常绘图）。 */
export function chartIssue(data: QueryDataset, rows: string[][], s: ChartSelection): string {
  if (!rows.length) return 'empty'
  if (!s.ys.length || s.ys.some((i) => i === s.x || !numericColumns(data).includes(i))) return 'noMetrics'
  if (s.type === 'pie') {
    if (s.ys.length !== 1) return 'pieSingle'
    const values = rows.map((r) => parseMetric(r[s.ys[0]]))
    if (values.some((n) => n === null || n < 0) || !values.some((n) => n !== null && n > 0)) return 'pieValues'
  }
  return ''
}

/**
 * 前端启发式闸门：决定这条数据集是否值得出图（保证「不需要图时不出图」）。
 * 无指标列 / 行数 > 30 / 列数超上限（查询 8 列、分析表 10 列）/ 指标全为 0 或全空 → 只出表格不出图。
 */
export function shouldShowChart(data: QueryDataset): boolean {
  const metrics = numericColumns(data)
  if (!metrics.length) return false
  if (data.rows.length > 30) return false
  if (data.columns.length > (data.source === 'table' ? 10 : 8)) return false
  const anyNonZero = data.rows.some((r) => metrics.some((m) => {
    const n = parseMetric(r[m])
    return n !== null && n !== 0
  }))
  return anyNonZero
}

/** 分析表转置视图：首列作为指标名列，其余列的列名成为类别（如「本月/上月」在表头上的多期对比）。 */
export function transposeDataset(data: QueryDataset): QueryDataset {
  if (data.columns.length < 2 || !data.rows.length) return data
  const columns = [data.columns[0], ...data.rows.map((r) => r[0] ?? '')]
  const rows: string[][] = []
  for (let c = 1; c < data.columns.length; c++) {
    const row: string[] = [data.columns[c]]
    for (const r of data.rows) row.push(r[c] ?? '')
    rows.push(row)
  }
  return { ...data, columns, rows, rowCount: rows.length }
}

/** 合计/小计行：首列文本命中即视为汇总行（成图默认剔除，避免汇总柱压扁正常类别）。 */
const totalRow = /^\s*(?:合计|总计|小计|总合|汇总)\s*$/

export function hasTotalRow(data: QueryDataset): boolean {
  return data.source === 'table' && data.rows.some((r) => totalRow.test(r[0] ?? ''))
}

/** 剔除合计行：仅作用于分析表视图（查询结果的「合计」可能是真实数据）。 */
export function withoutTotalRows(data: QueryDataset): QueryDataset {
  if (data.source !== 'table') return data
  const rows = data.rows.filter((r) => !totalRow.test(r[0] ?? ''))
  return rows.length === data.rows.length ? data : { ...data, rows, rowCount: rows.length }
}

/**
 * 出图数据集选择规则（实流与历史还原共用）：
 * ① 有模型推荐的优先成图；② 无推荐时分析表优于原始查询（表是结论口径）；
 * ③ 原始查询结果最多一图，分析表无推荐时取最后一张（结论表通常在正文末尾）；④ 都要过启发式闸门。
 */
export function pickChartDatasets(datasets: QueryDataset[] | undefined, hints: Record<string, ChartHint> | undefined): QueryDataset[] {
  if (!datasets?.length) return []
  const hinted = datasets.filter((d) => !!hints?.[d.id])
  if (hinted.length) return hinted
  const tables = datasets.filter((d) => d.source === 'table' && shouldShowChart(d))
  if (tables.length) return [tables[tables.length - 1]]
  return datasets.filter((d) => shouldShowChart(d)).slice(-1)
}

export interface ChartColors {
  text?: string
  border?: string
  background?: string
  palette?: string[]
}

/** pivot 分支要用的本地化文案：richText tooltip 只能拼纯文本，空组与空值的说法只有卡片知道。 */
export interface PivotLabels {
  noValue: string
  emptyGroup: string
}

/** tooltip 保留全精度千分位；轴/柱顶标签用 compactMetric 按中文量级缩写，两者不要混用。 */
export function formatMetric(value: number): string {
  return value.toLocaleString('zh-CN', { maximumFractionDigits: 2 })
}
export function compactMetric(value: number): string {
  const abs = Math.abs(value)
  const trim = (n: number) => String(Number(n.toFixed(2)))
  if (abs >= 1e8) return `${trim(value / 1e8)}亿`
  if (abs >= 1e4) return `${trim(value / 1e4)}万`
  return formatMetric(value)
}

// 与卡片 columnLabel 同一消歧规则：仅重名表头才追加列序号，否则图例会出现误导性的「(2)」。
function seriesName(data: QueryDataset, y: number): string {
  const name = data.columns[y]
  return data.columns.filter((n) => n === name).length > 1 ? `${name} (${y + 1})` : name
}

interface PivotTooltipParam {
  marker?: string
  seriesName?: string
  name?: string
  axisValue?: unknown
  value?: unknown
}

// richText 下 formatter 返回值直接作为 ZRender 文本（非 HTML），换行只能用 \n。
// itemMode 供饼图使用：扇区没有 axisValue，类别名在 name 上，且只有一条参数、不需要表头行。
function pivotTooltipFormatter(labels: PivotLabels, itemMode: boolean) {
  return (params: unknown) => {
    const list = (Array.isArray(params) ? params : [params]) as PivotTooltipParam[]
    const lines: string[] = []
    list.forEach((p, i) => {
      if (!itemMode && i === 0) lines.push(String(p.axisValue ?? ''))
      const text = typeof p.value === 'number' && Number.isFinite(p.value) ? formatMetric(p.value) : labels.noValue
      const who = itemMode ? (p.name ?? '') : (p.seriesName ?? '')
      lines.push(`${p.marker ?? ''} ${who}：${text}`)
    })
    return lines.join('\n')
  }
}

/** 所有图形统一走 pivot：类别与系列决定格子，聚合方式决定格内多行如何收敛成一个数。dataset 无法表达「多行收敛成一个值」，故不使用。 */
export function buildChartOption(
  data: QueryDataset,
  rows: string[][],
  s: ChartSelection,
  colors: ChartColors = {},
  labels: PivotLabels,
): EChartsOption {
  const pivot = computePivot(rows, s.x, s.seriesDim)
  const cellRows = (ci: number, gi: number) => pivot.cells.get(pivotKey(ci, gi))
  const category = {
    type: 'category' as const,
    axisLabel: { color: colors.text, hideOverlap: true },
    axisLine: { lineStyle: { color: colors.border } },
  }
  const value = {
    type: 'value' as const,
    axisLabel: { color: colors.text, formatter: (v: string | number) => compactMetric(Number(v)) },
    splitLine: { lineStyle: { color: colors.border } },
  }
  const shell = { animation: false, color: colors.palette, backgroundColor: colors.background, textStyle: { color: colors.text } }
  if (s.type === 'pie') {
    const y = s.ys[0]
    const pieData: { name: string; value: number }[] = []
    // 饼图不支持系列维度（控件已禁），每类别收敛成一个扇区；聚合不出数值的类别直接不画。
    pivot.cats.forEach((cat, ci) => {
      const cell = cellRows(ci, 0)
      if (!cell) return
      const val = aggregate(cell.map((r) => r[y]), s.agg)
      if (val === null) return
      pieData.push({ name: cat, value: val })
    })
    return {
      ...shell,
      tooltip: { trigger: 'item', renderMode: 'richText', confine: true, formatter: pivotTooltipFormatter(labels, true) },
      legend: { show: true, type: 'scroll', top: 0, textStyle: { color: colors.text } },
      series: [{ type: 'pie', name: seriesName(data, y), radius: '60%', center: ['50%', '55%'], data: pieData, label: { color: colors.text } }],
    }
  }
  // 未选系列维度且只有一个指标时只有一条系列，图例与系列同名，隐藏以免重复信息。
  const showLegend = s.seriesDim !== null || s.ys.length > 1
  const showLabel = pivot.cats.length <= 12
  const zoomed = pivot.cats.length > 30
  const series: SeriesOption[] = []
  s.ys.forEach((y) => {
    pivot.groups.forEach((group, gi) => {
      const cellValues = pivot.cats.map((_, ci) => {
        const cell = cellRows(ci, gi)
        return cell ? aggregate(cell.map((r) => r[y]), s.agg) : null
      })
      const groupName = group.trim() === '' ? labels.emptyGroup : group
      series.push({
        type: s.type as 'bar' | 'line',
        name: s.seriesDim === null ? seriesName(data, y) : s.ys.length > 1 ? `${seriesName(data, y)} · ${groupName}` : groupName,
        data: cellValues,
        barMaxWidth: s.type === 'bar' ? 56 : undefined,
        label: showLabel
          ? {
              show: true,
              position: s.horizontal ? 'right' : 'top',
              color: colors.text,
              formatter: (p: { value?: unknown }) => (typeof p.value === 'number' && Number.isFinite(p.value) ? compactMetric(p.value) : ''),
            }
          : undefined,
      })
    })
  })
  return {
    ...shell,
    tooltip: { trigger: 'axis', renderMode: 'richText', confine: true, formatter: pivotTooltipFormatter(labels, false) },
    legend: { show: showLegend, type: 'scroll', top: 0, textStyle: { color: colors.text } },
    grid: { left: 24, right: 48, top: showLegend ? 48 : 24, bottom: zoomed ? 76 : 40, containLabel: true },
    xAxis: s.horizontal ? value : { ...category, data: pivot.cats },
    yAxis: s.horizontal ? { ...category, data: pivot.cats } : value,
    dataZoom: zoomed
      ? [{ type: 'slider', ...(s.horizontal ? { yAxisIndex: 0 } : { xAxisIndex: 0 }), start: 0, end: Math.min(100, (30 / pivot.cats.length) * 100) }]
      : [],
    series,
  }
}
