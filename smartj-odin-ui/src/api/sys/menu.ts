import { get, post, put, del } from '../request'
import type { MenuVO, MenuDTO, MenuTreeVO } from '@/types/sys'

export function getMyMenus() {
  return get<MenuTreeVO[]>('/sys/menu/me')
}

export function getMyPermissions() {
  return get<string[]>('/sys/menu/permission/me')
}

export function getMenuTree() {
  return get<MenuVO[]>('/sys/menu/tree')
}

export function getMenuById(id: string) {
  return get<MenuVO>(`/sys/menu/${id}`)
}

export function createMenu(data: MenuDTO) {
  return post<string>('/sys/menu', data)
}

export function updateMenu(data: MenuDTO) {
  return put<void>('/sys/menu', data)
}

export function deleteMenu(id: string) {
  return del<void>(`/sys/menu/${id}`)
}
