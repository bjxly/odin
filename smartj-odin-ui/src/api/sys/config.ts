import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { ConfigVO, ConfigDTO, ConfigQuery } from '@/types/sys'

export function getConfigPage(params: ConfigQuery) {
  return get<PageResult<ConfigVO>>('/sys/config/page', { params })
}

export function getConfigById(id: string) {
  return get<ConfigVO>(`/sys/config/${id}`)
}

export function getConfigByKey(configKey: string) {
  return get<string>(`/sys/config/key/${configKey}`)
}

export function createConfig(data: ConfigDTO) {
  return post<string>('/sys/config', data)
}

export function updateConfig(data: ConfigDTO) {
  return put<void>('/sys/config', data)
}

export function deleteConfig(id: string) {
  return del<void>(`/sys/config/${id}`)
}

export function deleteConfigs(id: string) {
  return del<void>(`/sys/config/${id}`)
}
