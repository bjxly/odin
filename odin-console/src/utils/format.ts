import { FILTER_OPS } from '@/types/query'
import type { AggregateVO, HavingVO } from '@/types/query'

/** 展示层通用格式化工具（推理解释 / SQL 块 / 溯源卡片共用） */

const OP_LABELS: Record<string, string> = Object.fromEntries(FILTER_OPS.map((o) => [o.value, o.label]))

/** 过滤操作符 → 可读符号，未知操作符原样返回 */
export function opLabel(op?: string | null): string {
  const o = String(op ?? '')
  if (!o) return ''
  return OP_LABELS[o] ?? o
}

/** 过滤值 → 可读文本（数组/对象/NULL 分别处理） */
export function valueText(v: unknown): string {
  if (v == null || v === '') return 'NULL'
  if (Array.isArray(v)) return `[${v.map((x) => valueText(x)).join(', ')}]`
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

/** 毫秒数 → 友好文本；负数（后端 dry_run 占位 -1）视为无数据 */
export function formatMs(ms?: number | null): string {
  const n = Number(ms)
  if (!Number.isFinite(n) || n < 0) return '—'
  if (n < 1000) return `${Math.round(n)} ms`
  return `${(n / 1000).toFixed(2)} s`
}

/** 行数 → 友好文本；负数视为无数据 */
export function formatCount(n?: number | null): string {
  const v = Number(n)
  if (!Number.isFinite(v) || v < 0) return '—'
  return String(v)
}

/**
 * 计数归一（#56 统一块契约）：缺失 / 负数 / 非数值一律视为 0，
 * 用于 row_count / count 等「必含 number」字段的展示，杜绝「—」占位符。
 */
export function normalizeCount(n?: number | null): number {
  const v = Number(n)
  return Number.isFinite(v) && v >= 0 ? v : 0
}

/**
 * 耗时归一（#56 统一块契约）：缺失 / 负数 / 非数值一律视为 0，
 * 再走 formatMs，保证 0 显示「0 ms」而非「—」。
 */
export function formatDuration(ms?: number | null): string {
  const v = Number(ms)
  return formatMs(Number.isFinite(v) && v >= 0 ? v : 0)
}

/**
 * 度量列数值格式化：千分位 + 非整数保留 2 位小数。
 * 无法数值化的值（文本 / 日期 / null）原样返回，避免度量列出现 `NaN`。
 */
export function formatNumber(v: unknown): string {
  if (v == null || v === '') return ''
  if (typeof v === 'boolean') return v ? 'true' : 'false'
  const n = typeof v === 'number' ? v : Number(String(v).trim().replace(/,/g, ''))
  if (!Number.isFinite(n)) return typeof v === 'object' ? JSON.stringify(v) : String(v)
  if (Number.isInteger(n)) return n.toLocaleString('zh-CN')
  const a = Math.abs(n)
  // 过大/过小的数走科学计数，避免表格被一长串数字撑爆
  if (a >= 1e12 || (a > 0 && a < 1e-4)) return n.toExponential(2)
  return n.toLocaleString('zh-CN', { maximumFractionDigits: 2 })
}

/**
 * 聚合度量 → SQL 形态表达式：`SUM(quantity)`；count 无属性时为 `COUNT(*)`。
 * 函数名统一大写，属性名用后端下发的原样（本体中文 label 或属性 id）。
 */
export function aggExpr(func?: string | null, property?: string | null): string {
  const f = String(func ?? '').trim().toUpperCase()
  if (!f) return ''
  const p = String(property ?? '').trim()
  return `${f}(${f === 'COUNT' && !p ? '*' : p})`
}

/**
 * 聚合度量 → 意图展示文本：`SUM(quantity) AS 数量合计`。
 * alias 缺省时用「聚合函数中文名 + 属性名」兜底（与后端自动命名策略同源）。
 */
export function aggText(spec: AggregateVO | null | undefined): string {
  if (!spec) return ''
  const expr = aggExpr(spec.func, spec.property)
  if (!expr) return ''
  const alias = String(spec.alias ?? '').trim()
  if (!alias) return expr
  return `${expr} AS ${alias}`
}

/** HAVING 条件 → 可读文本：`合计数量 > 100` */
export function havingText(h: HavingVO | null | undefined): string {
  if (!h) return ''
  const alias = String(h.alias ?? '').trim()
  if (!alias) return ''
  return `${alias} ${opLabel(h.op)} ${formatNumber(h.value)}`
}

/** 截断长 id（trace_id 等）中间部分，便于展示 */
export function shortId(id?: string | null, head = 8, tail = 6): string {
  const s = String(id ?? '')
  if (!s) return ''
  if (s.length <= head + tail + 1) return s
  return `${s.slice(0, head)}…${s.slice(-tail)}`
}

/** 复制文本到剪贴板，返回是否成功（不抛异常） */
export async function copyText(text: string): Promise<boolean> {
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
      return true
    }
  } catch {
    // 落到下方兜底
  }
  try {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    return ok
  } catch {
    return false
  }
}
