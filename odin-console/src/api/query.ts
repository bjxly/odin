import request from './request'
import type { NLParseResultVO } from '@/types/query'

/**
 * 服务端意图解析：自然语言 → 受本体词表约束的查询 AST。
 * 后端先走确定性规则（parse_path=rule），未命中且 ai.enabled 时回退 Eino 意图链（llm）。
 */
export function nlParse(ontologyId: number, text: string) {
  return request
    .post('/query/nl-parse', { ontology_id: ontologyId, text })
    .then(r => r.data as NLParseResultVO)
}

export function executeQuery(data: any) {
  return request.post('/query', data).then(r => r.data)
}

export function dryRunQuery(data: any) {
  return request.post('/query/dry-run', data).then(r => r.data)
}

export function explainQuery(data: any) {
  return request.post('/query/explain', data).then(r => r.data)
}

export function getQueryHistory(params?: { current?: number; size?: number; ontology_id?: number }) {
  return request.get('/query/history', { params }).then(r => r.data)
}

export function getQueryTemplates(params?: { ontology_id?: number }) {
  return request.get('/query/templates', { params }).then(r => r.data)
}

export function saveQueryTemplate(data: any) {
  return request.post('/query/templates', data).then(r => r.data)
}
