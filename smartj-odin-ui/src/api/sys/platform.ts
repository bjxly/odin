import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type {
  UserVO,
  PlatformUserVO,
  UserQuery,
  UserTenantAuthorizationDTO,
  RoleVO,
  RoleDTO,
} from '@/types/sys'

// ==================== 全部用户（平台账号） ====================

/** 全局账号分页列表（仅平台超管在系统租户可用） */
export function getPlatformUserPage(params: UserQuery) {
  return get<PageResult<UserVO>>('/platform/users/page', { params })
}

/** 账号详情及全部租户授权（授权弹窗回显） */
export function getPlatformUserById(id: string) {
  return get<PlatformUserVO>(`/platform/users/${id}`)
}

/** 原子保存账号的全部租户归属、角色与默认租户 */
export function saveUserAuthorization(id: string, data: UserTenantAuthorizationDTO) {
  return put<void>(`/platform/users/${id}/authorization`, data)
}

// ==================== 租户角色权限（跨租户维护） ====================

/** 查询目标租户的全部角色及已勾选菜单权限 */
export function getTenantRoles(tenantId: string) {
  return get<RoleVO[]>(`/platform/tenants/${tenantId}/roles`)
}

/** 在目标租户内新增角色 */
export function createTenantRole(tenantId: string, data: RoleDTO) {
  return post<string>(`/platform/tenants/${tenantId}/roles`, data)
}

/** 更新目标租户内的既有角色 */
export function updateTenantRole(tenantId: string, data: RoleDTO) {
  return put<string>(`/platform/tenants/${tenantId}/roles`, data)
}

/** 删除目标租户内无用户绑定的角色 */
export function deleteTenantRole(tenantId: string, id: string) {
  return del<void>(`/platform/tenants/${tenantId}/roles/${id}`)
}

/** 查询目标租户内某角色绑定的用户列表 */
export function getTenantRoleBoundUsers(tenantId: string, id: string) {
  return get<UserVO[]>(`/platform/tenants/${tenantId}/roles/${id}/bound-users`)
}
