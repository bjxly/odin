import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { SysSmsConfigVO, SysSmsConfigDTO, SysSmsConfigQuery } from '@/types/sys'

export function getSmsConfigPage(params: SysSmsConfigQuery) {
  return get<PageResult<SysSmsConfigVO>>('/sys/sms/page', { params })
}

export function getSmsConfigById(id: string) {
  return get<SysSmsConfigVO>(`/sys/sms/${id}`)
}

export function createSmsConfig(data: SysSmsConfigDTO) {
  return post<string>('/sys/sms', data)
}

export function updateSmsConfig(data: SysSmsConfigDTO) {
  return put<void>('/sys/sms', data)
}

export function deleteSmsConfig(id: string) {
  return del<void>(`/sys/sms/${id}`)
}

/** 启用短信服务商（同时禁用其他） */
export function enableSmsConfig(id: string) {
  return put<void>(`/sys/sms/${id}/enable`)
}

/** 测试短信发送 */
export function testSmsSend(data: { providerId: string; phone: string; templateCode: string }) {
  return post<string>('/sys/sms/test', data)
}
