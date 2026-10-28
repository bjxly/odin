/** 洞察监控 VO */
export interface InsightMonitorVO {
  id: string
  name: string
  description?: string
  ontologyId?: string
  ontologyName?: string
  ruleExpression?: string
  scheduleCron?: string
  thresholds?: string
  notificationChannels?: string
  enabled?: number
  status: number
  remark?: string
  alertCount?: number
  lastAlertTime?: string
  createTime: string
}

/** 洞察监控 DTO */
export interface InsightMonitorDTO {
  id?: string
  name: string
  description?: string
  ontologyId?: string
  ruleExpression?: string
  scheduleCron?: string
  thresholds?: string
  notificationChannels?: string
  enabled?: number
  status?: number
  remark?: string
}

/** 洞察监控分页条件 */
export interface InsightMonitorQuery {
  current: number
  size: number
  name?: string
  ontologyId?: string
  enabled?: number
  status?: number
}

/** 洞察告警 VO */
export interface InsightAlertVO {
  id: string
  monitorId: string
  monitorName?: string
  triggerTime: string
  severity: string
  title?: string
  detail?: string
  metricValues?: string
  handled?: number
  handledBy?: string
  handledTime?: string
  handleNote?: string
  createTime: string
}

/** 告警分页条件 */
export interface AlertQuery {
  current: number
  size: number
  monitorId?: string
  severity?: string
  handled?: number
}

/** 变更历史 VO */
export interface ChangeHistoryVO {
  id: string
  entityType: string
  entityId: string
  action: string
  beforeSnapshot?: string
  afterSnapshot?: string
  operator?: string
  reason?: string
  createTime: string
}

/** 变更历史分页条件 */
export interface ChangeHistoryQuery {
  current: number
  size: number
  entityType?: string
  entityId?: string
  action?: string
}

/** 洞察看板 VO（AI 助手图表另存） */
export interface InsightDashboardVO {
  id: string
  name: string
  ontologyId?: string
  title?: string
  chartType?: string
  /** 数据集列头（JSON 字符串） */
  columnsJson?: string
  /** 数据集行（JSON 字符串） */
  rowsJson?: string
  /** 图表配置快照（JSON 字符串） */
  chartOption?: string
  source?: string
  status: number
  remark?: string
  createTime: string
}

/** 图表快照 DTO（另存为看板入参） */
export interface ChartSnapshotDTO {
  name: string
  ontologyId?: string
  title?: string
  chartType?: string
  columns: string[]
  rows: string[][]
  chartOption?: unknown
}

/** 洞察看板分页条件 */
export interface InsightDashboardQuery {
  current: number
  size: number
  name?: string
  ontologyId?: string
  status?: number
}
