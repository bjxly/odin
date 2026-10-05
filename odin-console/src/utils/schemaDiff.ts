import type {
  SchemaDiffColumnChangeVO,
  SchemaDiffColumnVO,
  SchemaDiffCountsVO,
  SchemaDiffVO,
  SchemaRawDiffVO,
} from '@/types/api'

/**
 * Schema 差量刷新的 diff 归一化工具。
 *
 * 后端 `POST /api/v1/datasources/:id/schema/refresh`（handler.RefreshDatasourceSchema）
 * 返回 `data = { schema, diff }`，其中 diff 对应 handler.schemaDiff：
 *
 * ```jsonc
 * {
 *   "added_tables":    ["zz_probe"],                                        // string[]
 *   "removed_tables":  ["zz_probe"],                                        // string[]
 *   "added_columns":   [{ "table": "t", "column": "c" }],                   // 对象数组
 *   "removed_columns": [{ "table": "t", "column": "c" }],                   // 对象数组
 *   "changed_columns": [{ "table": "t", "column": "c",
 *                         "old_type": "varchar", "new_type": "text" }]       // 对象数组
 * }
 * ```
 *
 * 本模块把它统一映射为前端友好的 {@link SchemaDiffVO}
 * （保留对象数组 + 追加 `table.column` 展示串 + 计数），
 * 页面 / 组件无需再关心后端字段名，也不要把列差异当字符串数组使用。
 */

/** 空 diff（表示本次刷新无结构变化） */
export function emptySchemaDiff(): SchemaDiffVO {
  return {
    addedTables: [],
    removedTables: [],
    addedColumns: [],
    removedColumns: [],
    changedColumns: [],
    added: [],
    removed: [],
    changed: [],
    counts: {
      addedTables: 0,
      removedTables: 0,
      addedColumns: 0,
      removedColumns: 0,
      changedColumns: 0,
      total: 0,
    },
  }
}

/** 任意值 → 字符串数组（丢弃非字符串与空白项） */
function toStrArray(v: unknown): string[] {
  if (!Array.isArray(v)) return []
  return v
    .filter((x): x is string => typeof x === 'string' && x.trim() !== '')
    .map((x) => x.trim())
}

/** 任意值 → `{table, column}` 数组；容忍后端退化为纯字符串（`t.c` 或 `c`） */
function toColumnArray(v: unknown): SchemaDiffColumnVO[] {
  if (!Array.isArray(v)) return []
  const out: SchemaDiffColumnVO[] = []
  for (const item of v) {
    if (typeof item === 'string') {
      const s = item.trim()
      if (!s) continue
      const dot = s.indexOf('.')
      out.push(dot > 0 ? { table: s.slice(0, dot), column: s.slice(dot + 1) } : { table: '', column: s })
      continue
    }
    if (item && typeof item === 'object') {
      const o = item as Record<string, unknown>
      const column = String(o.column ?? o.name ?? '').trim()
      if (!column) continue
      out.push({ table: String(o.table ?? o.table_name ?? '').trim(), column })
    }
  }
  return out
}

/** 任意值 → 类型变更数组（额外提取 old_type / new_type） */
function toChangeArray(v: unknown): SchemaDiffColumnChangeVO[] {
  if (!Array.isArray(v)) return []
  const out: SchemaDiffColumnChangeVO[] = []
  for (const item of v) {
    if (!item || typeof item !== 'object') continue
    const o = item as Record<string, unknown>
    const column = String(o.column ?? o.name ?? '').trim()
    if (!column) continue
    out.push({
      table: String(o.table ?? o.table_name ?? '').trim(),
      column,
      old_type: String(o.old_type ?? o.oldType ?? '').trim(),
      new_type: String(o.new_type ?? o.newType ?? '').trim(),
    })
  }
  return out
}

/** 列差异的可读标识：`table.column`（缺表名时只给列名） */
export function columnLabel(c: SchemaDiffColumnVO): string {
  return c.table ? `${c.table}.${c.column}` : c.column
}

/**
 * 归一化：后端 `schemaDiff` → 前端 {@link SchemaDiffVO}。
 * 入参可为 null / undefined / 已归一化对象，均安全返回（永不抛异常）。
 */
export function normalizeSchemaDiff(
  raw?: SchemaRawDiffVO | Partial<SchemaDiffVO> | null,
): SchemaDiffVO {
  if (!raw || typeof raw !== 'object') return emptySchemaDiff()
  const r = raw as Record<string, unknown>

  const addedTables = toStrArray(r.added_tables ?? r.addedTables)
  const removedTables = toStrArray(r.removed_tables ?? r.removedTables)
  const addedColumns = toColumnArray(r.added_columns ?? r.addedColumns)
  const removedColumns = toColumnArray(r.removed_columns ?? r.removedColumns)
  const changedColumns = toChangeArray(r.changed_columns ?? r.changedColumns)

  const added = addedColumns.map(columnLabel)
  const removed = removedColumns.map(columnLabel)
  const changed = changedColumns.map((c) => {
    const label = columnLabel(c)
    if (!c.old_type && !c.new_type) return label
    return `${label}: ${c.old_type || '?'} → ${c.new_type || '?'}`
  })

  const counts: SchemaDiffCountsVO = {
    addedTables: addedTables.length,
    removedTables: removedTables.length,
    addedColumns: added.length,
    removedColumns: removed.length,
    changedColumns: changed.length,
    total: 0,
  }
  counts.total =
    counts.addedTables +
    counts.removedTables +
    counts.addedColumns +
    counts.removedColumns +
    counts.changedColumns

  return {
    addedTables,
    removedTables,
    addedColumns,
    removedColumns,
    changedColumns,
    added,
    removed,
    changed,
    counts,
  }
}

/** 汇总文案：`X 新增 / Y 删除 / Z 变更`（表与列合并计数） */
export function diffHeadline(diff?: SchemaDiffVO | null): string {
  if (!diff) return '0 新增 / 0 删除 / 0 变更'
  const c = diff.counts
  return `${c.addedTables + c.addedColumns} 新增 / ${c.removedTables + c.removedColumns} 删除 / ${c.changedColumns} 变更`
}

/** 一个 diff 明细分组 */
export interface SchemaDiffGroup {
  label: string
  items: string[]
}

/** 分组明细（仅返回非空分组），供通知 / 抽屉按类展示 */
export function diffGroups(diff?: SchemaDiffVO | null): SchemaDiffGroup[] {
  if (!diff) return []
  const all: SchemaDiffGroup[] = [
    { label: '新增表', items: diff.addedTables },
    { label: '删除表', items: diff.removedTables },
    { label: '新增列', items: diff.added },
    { label: '删除列', items: diff.removed },
    { label: '变更列', items: diff.changed },
  ]
  return all.filter((g) => g.items.length > 0)
}
