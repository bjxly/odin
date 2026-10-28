import { get, post, put, del } from '@/api/request'
import type { PageResult } from '@/types'
import type {
  InsightMonitorVO,
  InsightMonitorDTO,
  InsightMonitorQuery,
  InsightAlertVO,
  AlertQuery,
  ChangeHistoryVO,
  ChangeHistoryQuery,
  InsightDashboardVO,
  InsightDashboardQuery,
  ChartSnapshotDTO,
} from '@/types/odin'

const BASE = '/odin/insight'
const CHANGE_BASE = '/odin/changes'

// ===== 洞察监控 =====

/** 分页查询监控规则 */
export function getMonitorPage(params: InsightMonitorQuery) {
  return get<PageResult<InsightMonitorVO>>(`${BASE}/monitor/page`, { params })
}

/** 根据ID查询监控规则 */
export function getMonitorById(id: string) {
  return get<InsightMonitorVO>(`${BASE}/monitor/${id}`)
}

/** 新增监控规则，返回ID */
export function createMonitor(data: InsightMonitorDTO) {
  return post<string>(`${BASE}/monitor`, data)
}

/** 更新监控规则 */
export function updateMonitor(data: InsightMonitorDTO) {
  return put<void>(`${BASE}/monitor`, data)
}

/** 删除监控规则 */
export function deleteMonitor(id: string) {
  return del<void>(`${BASE}/monitor/${id}`)
}

/** 立即执行一次监控检查 */
export function runMonitor(id: string) {
  return post<void>(`${BASE}/monitor/${id}/run`, undefined, { timeout: 120000 })
}

// ===== 告警 =====

/** 分页查询告警 */
export function getAlertPage(params: AlertQuery) {
  return get<PageResult<InsightAlertVO>>(`${BASE}/alert/page`, { params })
}

/** 处理告警 */
export function handleAlert(id: string, data: { handleNote?: string }) {
  return put<void>(`${BASE}/alert/${id}/handle`, data)
}

// ===== 变更历史 =====

/** 分页查询变更历史 */
export function getChangeHistoryPage(params: ChangeHistoryQuery) {
  return get<PageResult<ChangeHistoryVO>>(`${CHANGE_BASE}/page`, { params })
}

// ===== 洞察看板（AI 助手图表另存） =====

/** 图表快照另存为洞察看板，返回ID */
export function createFromChart(data: ChartSnapshotDTO) {
  return post<string>(`${BASE}/dashboard`, data)
}

/** 分页查询洞察看板 */
export function getDashboardPage(params: InsightDashboardQuery) {
  return get<PageResult<InsightDashboardVO>>(`${BASE}/dashboard/page`, { params })
}

/** 删除洞察看板 */
export function deleteDashboard(id: string) {
  return del<void>(`${BASE}/dashboard/${id}`)
}
