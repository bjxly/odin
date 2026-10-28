import { get, post, put, del } from '../request'
import type { DeptVO, DeptDTO, DeptQuery } from '@/types/sys'

export function getDeptTree(params?: DeptQuery) {
  return get<DeptVO[]>('/sys/dept/tree', { params })
}

export function getDeptById(id: string) {
  return get<DeptVO>(`/sys/dept/${id}`)
}

export function createDept(data: DeptDTO) {
  return post<string>('/sys/dept', data)
}

export function updateDept(data: DeptDTO) {
  return put<void>('/sys/dept', data)
}

export function deleteDept(id: string) {
  return del<void>(`/sys/dept/${id}`)
}
