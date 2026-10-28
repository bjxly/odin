import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { UserVO, UserDTO, UserQuery } from '@/types/sys'

export function getUserPage(params: UserQuery) {
  return get<PageResult<UserVO>>('/sys/user/page', { params })
}

export function getUserById(id: string) {
  return get<UserVO>(`/sys/user/${id}`)
}

export function createUser(data: UserDTO) {
  return post<string>('/sys/user', data)
}

export function updateUser(data: UserDTO) {
  return put<void>('/sys/user', data)
}

export function deleteUsers(ids: string[]) {
  return del<void>('/sys/user', { data: ids })
}

export function resetPassword(id: string, newPassword: string) {
  return put<void>(`/sys/user/${id}/resetPassword?newPassword=${encodeURIComponent(newPassword)}`)
}

// ==================== 个人中心 ====================

/** 获取当前用户资料（含社交绑定） */
export function getUserProfile() {
  return get<UserVO>('/sys/user/profile')
}

/** 更新个人资料 */
export function updateUserProfile(data: Partial<UserDTO>) {
  return put<void>('/sys/user/profile', data)
}

/** 修改密码 */
export function changePassword(data: { oldPassword: string; newPassword: string }) {
  return put<void>('/sys/user/profile/password', data)
}
