import { get, post, put, del } from '@/api/request'
import type { PageResult } from '@/types'
import type {
  DatasourceVO,
  DatasourceDTO,
  DatasourceQuery,
  DatasourceTestResult,
  SchemaTableVO,
  SchemaColumnVO,
} from '@/types/odin'

const BASE = '/odin/datasource'

/** 分页查询数据源 */
export function getDatasourcePage(params: DatasourceQuery) {
  return get<PageResult<DatasourceVO>>(`${BASE}/page`, { params })
}

/** 根据ID查询数据源 */
export function getDatasourceById(id: string) {
  return get<DatasourceVO>(`${BASE}/${id}`)
}

/** 新增数据源，返回ID */
export function createDatasource(data: DatasourceDTO) {
  return post<string>(BASE, data)
}

/** 更新数据源 */
export function updateDatasource(data: DatasourceDTO) {
  return put<void>(BASE, data)
}

/** 删除数据源 */
export function deleteDatasource(id: string) {
  return del<void>(`${BASE}/${id}`)
}

/** 测试连接（远程内省，放宽超时） */
export function testDatasource(id: string) {
  return post<DatasourceTestResult>(`${BASE}/${id}/test`, undefined, { timeout: 60000 })
}

/** 同步元数据，返回表数量（大库耗时长，放宽超时） */
export function syncDatasource(id: string) {
  return post<number>(`${BASE}/${id}/sync`, undefined, { timeout: 300000 })
}

/** 查询数据源表清单 */
export function getDatasourceTables(id: string) {
  return get<SchemaTableVO[]>(`${BASE}/${id}/tables`)
}

/** 查询表列清单 */
export function getTableColumns(tableId: string) {
  return get<SchemaColumnVO[]>(`${BASE}/table/${tableId}/columns`)
}

/** 查询全部数据源列表（不分页，用于下拉选择） */
export function getDatasourceList() {
  return get<DatasourceVO[]>(`${BASE}/list`)
}
