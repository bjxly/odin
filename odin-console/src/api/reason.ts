import request from './request'
import type { ReasonExplainVO } from '@/types/query'

/** 推理试运行的过滤条件（对齐后端 query.Filter） */
export interface ReasonFilterParam {
  property: string
  op: string
  value?: unknown
}

export interface ReasonExplainParam {
  ontology_id: number
  class_name: string
  filters?: ReasonFilterParam[]
}

/**
 * POST /api/v1/reason/explain
 * 只做推理（规则应用 + 虚拟类展开 + 关系推导），返回 reason_trace，不翻译也不执行 SQL。
 * 用于「规则试运行 / 单步解释」。
 */
export function reasonExplain(data: ReasonExplainParam) {
  return request
    .post('/reason/explain', {
      ontology_id: data.ontology_id,
      class_name: data.class_name,
      filters: data.filters ?? [],
    })
    .then((r) => r.data as ReasonExplainVO)
}
