import request from './request'
import type { TraceVO } from '@/types/query'

/**
 * GET /api/v1/query/trace/:traceId
 * 一次查询的全链路溯源：自然语言 → 意图 → 推理 → SQL → 执行结果 → 节点耗时。
 */
export function getTrace(traceId: string) {
  return request
    .get(`/query/trace/${encodeURIComponent(traceId)}`)
    .then((r) => r.data as TraceVO)
}
