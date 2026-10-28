/** 溯源阶段 */
export type TracePhase =
  | 'nl_parse'
  | 'intent_resolve'
  | 'ontology_map'
  | 'reason'
  | 'sql_translate'
  | 'execute'
  | 'format'

/** 溯源步骤 */
export interface TraceStepDetail {
  stepIndex: number
  phase: TracePhase
  label: string
  input?: string
  output?: string
  durationMs?: number
  status: 'success' | 'error' | 'skipped'
  errorMessage?: string
  metadata?: Record<string, unknown>
}

/** 溯源数据 */
export interface TraceData {
  traceId: string
  queryId?: string
  naturalLanguage?: string
  logicalSql?: string
  physicalSql?: string
  totalDurationMs?: number
  status: string
  steps: TraceStepDetail[]
  createdAt?: string
}

/** 变更日志 VO */
export interface ChangeLogVO {
  id: string
  ontologyId?: string
  entityType: string
  entityId: string
  action: string
  beforeSnapshot?: string
  afterSnapshot?: string
  operator?: string
  operatorName?: string
  reason?: string
  createTime: string
}

/** 变更日志查询条件 */
export interface ChangeLogQuery {
  current: number
  size: number
  ontologyId?: string
  entityType?: string
  action?: string
}
