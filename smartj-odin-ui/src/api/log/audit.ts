import { get } from '../request'
import type { PageResult } from '@/types'
import type { AuditLogVO, AuditLogQuery } from '@/types/sys'

export function getAuditLogPage(params: AuditLogQuery) {
  return get<PageResult<AuditLogVO>>('/log/audit/page', { params })
}
