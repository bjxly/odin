import { get, post, put, del } from '@/api/request'
import type { PageResult } from '@/types'
import type { ConnectorVO, ConnectorDTO, ConnectorQuery, DriverInfo } from '@/types/odin'

const BASE = '/odin/connector'

/** 分页查询连接器 */
export function getConnectorPage(params: ConnectorQuery) {
  return get<PageResult<ConnectorVO>>(`${BASE}/page`, { params })
}

/** 查询全部连接器列表（不分页） */
export function getConnectorList() {
  return get<ConnectorVO[]>(`${BASE}/list`)
}

/** 新增连接器，返回ID */
export function createConnector(data: ConnectorDTO) {
  return post<string>(BASE, data)
}

/** 更新连接器 */
export function updateConnector(data: ConnectorDTO) {
  return put<void>(BASE, data)
}

/** 批量删除连接器 */
export function deleteConnectors(ids: string[]) {
  return del<void>(BASE, { data: { ids } })
}

/** 查询可用驱动列表 */
export function getDrivers() {
  return get<DriverInfo[]>(`${BASE}/drivers`)
}

/** 上传自定义驱动 */
export function uploadDriver(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return post<void>(`${BASE}/driver/upload`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
  })
}

/** 测试连接器 */
export function testConnector(id: string) {
  return post<void>(`${BASE}/${id}/test`, undefined, { timeout: 60000 })
}
