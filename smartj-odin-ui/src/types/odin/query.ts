/** 问数查询结果 VO */
export interface QueryResultVO {
  success: boolean
  columns?: string[]
  rows?: Record<string, unknown>[]
  rowCount?: number
  latencyMs?: number
  logicalSql?: string
  physicalSql?: string
  traceId?: string
  errorMessage?: string
}

/** 查询 dry-run 校验结果 VO */
export interface QueryDryRunVO {
  valid: boolean
  logicalSql?: string
  physicalSql?: string
  estimatedRows?: number
  warnings?: string[]
  errorMessage?: string
}

/** 查询模板 VO */
export interface QueryTemplateVO {
  id: string
  name: string
  description?: string
  ontologyId?: string
  ontologyName?: string
  logicalSql?: string
  parameters?: string
  tags?: string
  usageCount?: number
  status: number
  createTime: string
  updateTime?: string
}

/** 查询模板 DTO */
export interface QueryTemplateDTO {
  id?: string
  name: string
  description?: string
  ontologyId?: string
  logicalSql?: string
  parameters?: string
  tags?: string
  status?: number
}

/** 查询模板分页条件 */
export interface QueryTemplateQuery {
  current: number
  size: number
  name?: string
  ontologyId?: string
  status?: number
}

/** 查询历史 VO */
export interface QueryHistoryVO {
  id: string
  sessionId?: string
  userId?: string
  naturalLanguage?: string
  logicalSql?: string
  physicalSql?: string
  resultSummary?: string
  latencyMs?: number
  status: string
  errorMessage?: string
  createTime: string
}

/** 查询历史分页条件 */
export interface QueryHistoryQuery {
  current: number
  size: number
  sessionId?: string
  userId?: string
  status?: string
}

/** 查询溯源 VO */
export interface QueryTraceVO {
  id: string
  historyId: string
  steps: TraceStep[]
}

/** 溯源步骤 */
export interface TraceStep {
  stepIndex: number
  stepType: string
  datasourceId?: string
  datasourceName?: string
  sqlText?: string
  rowCount?: number
  latencyMs?: number
  details?: string
}
