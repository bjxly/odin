import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { CloudRouteVO, CloudRouteDTO, CloudRouteQuery } from '@/types/sys'

export function getCloudRoutePage(params: CloudRouteQuery) {
  return get<PageResult<CloudRouteVO>>('/sys/cloud-route/page', { params })
}

export function getCloudRouteById(id: string) {
  return get<CloudRouteVO>(`/sys/cloud-route/${id}`)
}

export function createCloudRoute(data: CloudRouteDTO) {
  return post<string>('/sys/cloud-route', data)
}

export function updateCloudRoute(data: CloudRouteDTO) {
  return put<void>('/sys/cloud-route', data)
}

export function deleteCloudRoute(id: string) {
  return del<void>(`/sys/cloud-route/${id}`)
}
