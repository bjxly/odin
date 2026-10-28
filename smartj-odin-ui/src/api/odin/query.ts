import { get, post, put, del } from '@/api/request'
import type { PageResult } from '@/types'
import type {
  QueryResultVO,
  QueryDryRunVO,
  QueryTemplateVO,
  QueryTemplateDTO,
  QueryTemplateQuery,
  QueryHistoryVO,
  QueryHistoryQuery,
  QueryTraceVO,
} from '@/types/odin'

const BASE = '/odin/query'

// ===== 问数执行 =====

/** 执行本体逻辑 SQL 问数（联邦执行可能跨源，放宽超时） */
export function executeLogicalSql(data: { logicalSql: string; ontologyId?: string; params?: Record<string, unknown> }) {
  return post<QueryResultVO>(`${BASE}/execute`, data, { timeout: 120000 })
}

/** dry-run 校验逻辑 SQL（不真正执行，返回物理 SQL 与告警） */
export function dryRunQuery(data: { logicalSql: string; ontologyId?: string }) {
  return post<QueryDryRunVO>(`${BASE}/dry-run`, data, { timeout: 60000 })
}

// ===== 查询模板 =====

/** 分页查询模板 */
export function getQueryTemplatePage(params: QueryTemplateQuery) {
  return get<PageResult<QueryTemplateVO>>(`${BASE}/template/page`, { params })
}

/** 根据ID查询模板 */
export function getQueryTemplateById(id: string) {
  return get<QueryTemplateVO>(`${BASE}/template/${id}`)
}

/** 新增模板，返回ID */
export function createQueryTemplate(data: QueryTemplateDTO) {
  return post<string>(`${BASE}/template`, data)
}

/** 更新模板 */
export function updateQueryTemplate(data: QueryTemplateDTO) {
  return put<void>(`${BASE}/template`, data)
}

/** 删除模板 */
export function deleteQueryTemplate(id: string) {
  return del<void>(`${BASE}/template/${id}`)
}

/** 按模板执行查询 */
export function executeTemplate(id: string, params?: Record<string, unknown>) {
  return post<QueryResultVO>(`${BASE}/template/${id}/execute`, params, { timeout: 120000 })
}

// ===== 查询历史与溯源 =====

/** 分页查询历史 */
export function getQueryHistoryPage(params: QueryHistoryQuery) {
  return get<PageResult<QueryHistoryVO>>(`${BASE}/history/page`, { params })
}

/** 查询单条历史的执行溯源 */
export function getQueryTrace(historyId: string) {
  return get<QueryTraceVO>(`${BASE}/history/${historyId}/trace`)
}
