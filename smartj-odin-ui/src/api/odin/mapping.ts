import { get, post, put, del } from '@/api/request'
import type { PageResult } from '@/types'
import type { MappingConfigVO, MappingConfigDTO, MappingConfigQuery, MappingSuggestVO } from '@/types/odin'

const BASE = '/odin/mapping'

/** 分页查询映射配置 */
export function getMappingPage(params: MappingConfigQuery) {
  return get<PageResult<MappingConfigVO>>(`${BASE}/page`, { params })
}

/** 根据ID查询映射配置 */
export function getMappingById(id: string) {
  return get<MappingConfigVO>(`${BASE}/${id}`)
}

/** 新增映射配置，返回ID */
export function createMapping(data: MappingConfigDTO) {
  return post<string>(BASE, data)
}

/** 更新映射配置 */
export function updateMapping(data: MappingConfigDTO) {
  return put<void>(BASE, data)
}

/** 删除映射配置 */
export function deleteMapping(id: string) {
  return del<void>(`${BASE}/${id}`)
}

/** AI 建议列映射（按类属性与表列智能匹配，放宽超时） */
export function suggestColumnMappings(classId: string, tableId: string) {
  return get<MappingSuggestVO[]>(`${BASE}/suggest`, { params: { classId, tableId }, timeout: 120000 })
}

/** 校验映射配置（列映射完整性/可执行性），返回后端校验结果描述 */
export function validateMapping(id: string) {
  return post<string>(`${BASE}/${id}/validate`, undefined, { timeout: 60000 })
}
