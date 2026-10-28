import { get, del } from '../request'
import type { PageResult } from '@/types'
import type { OperLogVO, OperLogQuery } from '@/types/sys'

export function getOperLogPage(params: OperLogQuery) {
  return get<PageResult<OperLogVO>>('/log/oper/page', { params })
}

export function cleanOperLog(days?: number) {
  return del<void>('/log/oper/clean', { params: { days } })
}
