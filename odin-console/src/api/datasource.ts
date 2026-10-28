import request from './request'
import type { DataSourceSchemaVO, SchemaRefreshResultVO } from '@/types/api'

/** 数据源列表（分页） */
export function getDatasources(params?: { current?: number; size?: number; type?: string; status?: string }) {
  return request.get('/datasources', { params }).then(r => r.data)
}

export function getDatasource(id: number) {
  return request.get(`/datasources/${id}`).then(r => r.data)
}

export function createDatasource(data: any) {
  return request.post('/datasources', data).then(r => r.data)
}

export function updateDatasource(id: number, data: any) {
  return request.put(`/datasources/${id}`, data).then(r => r.data)
}

export function deleteDatasource(id: number) {
  return request.delete(`/datasources/${id}`).then(r => r.data)
}

export function testDatasource(id: number) {
  return request.post(`/datasources/${id}/test`).then(r => r.data)
}

export function getDatasourceSchema(id: number) {
  return request.get(`/datasources/${id}/schema`).then(r => r.data as DataSourceSchemaVO)
}

/**
 * 差量刷新 Schema：后端重新内省、覆盖写缓存并做 diff。
 * 响应为**嵌套结构** `{ schema, diff }`：
 *   - `schema` 为刷新后的完整 schema（含 tables / table_count / column_count）；
 *   - `diff` 为 `{ added_tables, removed_tables, added_columns, removed_columns, changed_columns }`，
 *     其中列级差异是 `{table, column}` / `{table, column, old_type, new_type}` 对象数组。
 * 调用方请用 `utils/schemaDiff#normalizeSchemaDiff` 归一化后再消费。
 */
export function refreshSchema(id: number) {
  return request.post(`/datasources/${id}/schema/refresh`).then(r => r.data as SchemaRefreshResultVO)
}

/** 连接器（驱动）信息 */
export interface DriverInfo {
  name: string
  display_name: string
  category: string
  driver_type: 'native' | 'jdbc' | string
  driver_lib: string
  status: 'ready' | 'need_driver' | string
  description: string
  config_required: boolean
}

/** 获取系统支持的连接器 / 驱动列表 */
export function getDrivers() {
  return request.get('/drivers').then(r => r.data as DriverInfo[])
}

/** 上传（或更新）指定连接器的 JDBC 代理驱动文件 */
export function uploadDriver(name: string, formData: FormData) {
  return request
    .post(`/drivers/${name}/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then(r => r.data)
}
