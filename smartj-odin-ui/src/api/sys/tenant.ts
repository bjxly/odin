import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { TenantVO, TenantDTO, TenantQuery, RoleVO } from '@/types/sys'

export function getTenantPage(params: TenantQuery) {
  return get<PageResult<TenantVO>>('/tenant/page', { params })
}

export function getTenantById(id: string) {
  return get<TenantVO>(`/tenant/${id}`)
}

export function getTenantAdminRoles() {
  return get<RoleVO[]>('/tenant/admin-roles')
}

export function createTenant(data: TenantDTO) {
  return post<string>('/tenant', data)
}

export function updateTenant(data: TenantDTO) {
  return put<void>('/tenant', data)
}

export function deleteTenants(ids: string[]) {
  return del<void>('/tenant', { data: ids })
}
