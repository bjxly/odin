import { get, post, put, del } from '@/api/request'
import type { PageResult } from '@/types'
import type { ReasonRuleVO, ReasonRuleDTO, ReasonRuleQuery, ReasonTraceResult, ReasonPath, ReasonExplainParam } from '@/types/odin'

const BASE = '/odin/reason'

/** 推理试运行（规则应用 + 虚拟类展开 + 关系推导） */
export function explainRule(data: ReasonExplainParam) {
  return post<ReasonTraceResult>(`${BASE}/explain`, data, { timeout: 120000 })
}

/** 分析两个类之间的推理路径 */
export function analyzePath(params: { ontologyId: string; sourceClass: string; targetClass: string }) {
  return get<ReasonPath>(`${BASE}/path`, { params, timeout: 60000 })
}

/** 分页查询推理规则 */
export function getRulePage(params: ReasonRuleQuery) {
  return get<PageResult<ReasonRuleVO>>(`${BASE}/rule/page`, { params })
}

/** 新增推理规则，返回ID */
export function createRule(data: ReasonRuleDTO) {
  return post<string>(`${BASE}/rule`, data)
}

/** 更新推理规则 */
export function updateRule(data: ReasonRuleDTO) {
  return put<void>(`${BASE}/rule`, data)
}

/** 批量删除推理规则 */
export function deleteRules(ids: string[]) {
  return del<void>(`${BASE}/rule`, { data: { ids } })
}

/** 启用/禁用推理规则 */
export function toggleRule(id: string, enabled: number) {
  return put<void>(`${BASE}/rule/${id}/toggle`, { enabled })
}
