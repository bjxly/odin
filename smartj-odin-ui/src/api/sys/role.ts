import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { RoleVO, RoleDTO, RoleQuery } from '@/types/sys'

export function getRolePage(params: RoleQuery) {
  return get<PageResult<RoleVO>>('/sys/role/page', { params })
}

export function getRoleById(id: string) {
  return get<RoleVO>(`/sys/role/${id}`)
}

export function createRole(data: RoleDTO) {
  return post<string>('/sys/role', data)
}

export function updateRole(data: RoleDTO) {
  return put<void>('/sys/role', data)
}

export function deleteRoles(id: string) {
  return del<void>(`/sys/role/${id}`)
}

export function getRoleList() {
  return get<RoleVO[]>('/sys/role/list')
}
