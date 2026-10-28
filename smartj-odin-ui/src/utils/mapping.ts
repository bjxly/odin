import type { ColumnMappingRow } from '@/types/odin/mapping'

/**
 * 列映射（column_mappings JSON）统一编解码。
 *
 * 历史上同一字段存在四套键约定：
 * - 种子/SQL：property_name + column_name + validation_status
 * - 旧工作台/MappingEditor：propertyCode + columnName
 * - ClassDrawer：按属性 name（property）
 * 读取端统一容错归一到规范 ColumnMappingRow；写入端只产规范 camelCase 键。
 */

/** propertyCode 的历史别名（按优先级） */
const PROP_KEYS = ['propertyCode', 'property_code', 'propertyName', 'property_name', 'property'] as const
/** columnName 的历史别名（按优先级） */
const COL_KEYS = ['columnName', 'column_name', 'column'] as const
/** validationStatus 的历史别名 */
const STATUS_KEYS = ['validationStatus', 'validation_status'] as const

/** 从记录中按候选键顺序取第一个存在的值 */
function pick(rec: Record<string, unknown>, keys: readonly string[]): unknown {
  for (const k of keys) {
    if (k in rec && rec[k] != null) return rec[k]
  }
  return undefined
}

/** 转字符串，null/undefined → '' */
function toStr(v: unknown): string {
  return v == null ? '' : String(v)
}

/** valueMap 归一：对象保持、字符串尝试 JSON.parse、其余置空 */
function normalizeValueMap(v: unknown): Record<string, string> | string | null {
  if (v == null || v === '') return null
  if (typeof v === 'object') return v as Record<string, string>
  if (typeof v === 'string') {
    try {
      const parsed = JSON.parse(v)
      return parsed && typeof parsed === 'object' ? (parsed as Record<string, string>) : v
    } catch {
      return v
    }
  }
  return null
}

/**
 * 容错解析 column_mappings JSON 为规范行数组。
 *
 * @param json 原始 JSON 字符串（可空）
 * @returns 规范行数组；非法/空输入返回 []
 */
export function parseColumnMappings(json?: string | null): ColumnMappingRow[] {
  if (!json) return []
  try {
    const parsed = JSON.parse(json)
    if (!Array.isArray(parsed)) return []
    return parsed
      .map((raw): ColumnMappingRow | null => {
        if (!raw || typeof raw !== 'object') return null
        const rec = raw as Record<string, unknown>
        const propertyCode = toStr(pick(rec, PROP_KEYS))
        const columnName = toStr(pick(rec, COL_KEYS))
        // 两个核心键都为空则视为无效行
        if (!propertyCode && !columnName) return null
        const conf = rec.confidence
        const status = pick(rec, STATUS_KEYS)
        const row: ColumnMappingRow = { propertyCode, columnName }
        const transform = toStr(rec.transform)
        if (transform) row.transform = transform
        const vm = normalizeValueMap(rec.valueMap ?? rec.value_map)
        if (vm != null) row.valueMap = vm
        if (typeof conf === 'number') row.confidence = conf
        if (status != null) row.validationStatus = toStr(status)
        // propertyName 快照：仅当历史键显式提供且与 code 不同时保留
        const nameSnapshot = toStr(rec.propertyName ?? rec.property_name)
        if (nameSnapshot && nameSnapshot !== propertyCode) row.propertyName = nameSnapshot
        return row
      })
      .filter((r): r is ColumnMappingRow => r != null)
  } catch {
    return []
  }
}

/**
 * 将规范行数组序列化为 column_mappings JSON（只产规范 camelCase 键）。
 *
 * @param rows 规范行数组
 * @returns JSON 字符串；空数组返回 '[]'
 */
export function serializeColumnMappings(rows: ColumnMappingRow[]): string {
  const out = (rows ?? []).map((r) => {
    const o: Record<string, unknown> = {
      propertyCode: r.propertyCode ?? '',
      columnName: r.columnName ?? '',
    }
    if (r.propertyName) o.propertyName = r.propertyName
    if (r.transform) o.transform = r.transform
    const vm = normalizeValueMap(r.valueMap)
    if (vm != null) o.valueMap = vm
    if (r.confidence != null) o.confidence = r.confidence
    if (r.validationStatus) o.validationStatus = r.validationStatus
    return o
  })
  return JSON.stringify(out)
}

/** 统计一段 column_mappings JSON 的有效列映射条数 */
export function countColumnMappings(json?: string | null): number {
  return parseColumnMappings(json).length
}
