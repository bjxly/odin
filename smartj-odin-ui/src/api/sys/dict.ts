import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { DictTypeVO, DictTypeDTO, DictTypeQuery, DictDataVO, DictDataDTO, DictDataQuery } from '@/types/sys'

// ==================== 字典类型 ====================
export function getDictTypePage(params: DictTypeQuery) {
  return get<PageResult<DictTypeVO>>('/sys/dict/type/page', { params })
}

export function getDictTypeById(id: string) {
  return get<DictTypeVO>(`/sys/dict/type/${id}`)
}

export function createDictType(data: DictTypeDTO) {
  return post<string>('/sys/dict/type', data)
}

export function updateDictType(data: DictTypeDTO) {
  return put<void>('/sys/dict/type', data)
}

export function deleteDictTypes(ids: string[]) {
  return del<void>('/sys/dict/type', { data: ids })
}

export function getDictTypeList() {
  return get<DictTypeVO[]>('/sys/dict/type/list')
}

// ==================== 字典数据 ====================
export function getDictDataPage(params: DictDataQuery) {
  return get<PageResult<DictDataVO>>('/sys/dict/data/page', { params })
}

export function getDictDataById(id: string) {
  return get<DictDataVO>(`/sys/dict/data/${id}`)
}

export function createDictData(data: DictDataDTO) {
  return post<string>('/sys/dict/data', data)
}

export function updateDictData(data: DictDataDTO) {
  return put<void>('/sys/dict/data', data)
}

export function deleteDictData(id: string) {
  return del<void>(`/sys/dict/data/${id}`)
}

export function getDictDataByType(dictType: string) {
  return get<DictDataVO[]>(`/sys/dict/data/items/${dictType}`)
}
