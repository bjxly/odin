import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { NoticeLogVO, NoticeTemplateVO, ImConfigVO, NoticeSendRequest } from '@/types/sys/notice'

// ==================== 通知记录 ====================

/** 分页查询通知记录 */
export function getNoticePage(params: Record<string, unknown>) {
  return get<PageResult<NoticeLogVO>>('/sys/notice/page', { params })
}

/** 手动发送通知 */
export function sendNotice(data: NoticeSendRequest) {
  return post<string>('/sys/notice/send', data)
}

/** 获取未读通知数量 */
export function getUnreadCount(userId: string) {
  return get<number>('/sys/notice/unread-count', { params: { userId } })
}

/** 标记单条通知为已读 */
export function markNoticeRead(id: string) {
  return put<void>(`/sys/notice/read/${id}`)
}

/** 标记全部通知为已读 */
export function markAllNoticeRead(userId: string) {
  return put<void>('/sys/notice/read-all', null, { params: { userId: userId } })
}

// ==================== 通知模板 ====================

/** 查询模板列表 */
export function getNoticeTemplateList() {
  return get<NoticeTemplateVO[]>('/sys/notice-template/list')
}

/** 根据ID查询模板 */
export function getNoticeTemplateById(id: string) {
  return get<NoticeTemplateVO>(`/sys/notice-template/${id}`)
}

/** 新增模板 */
export function createNoticeTemplate(data: NoticeTemplateVO) {
  return post<string>('/sys/notice-template', data)
}

/** 更新模板 */
export function updateNoticeTemplate(data: NoticeTemplateVO) {
  return put<void>('/sys/notice-template', data)
}

/** 删除模板（逻辑删除） */
export function deleteNoticeTemplate(id: string) {
  return del<void>(`/sys/notice-template/${id}`)
}

/** 模板变量预览 */
export function previewNoticeTemplate(data: { templateCode: string; variables: Record<string, unknown> }) {
  return post<{ title: string; content: string }>('/sys/notice-template/preview', data)
}

// ==================== IM 配置 ====================

/** 查询IM配置列表 */
export function getImConfigList(providerId?: string) {
  return get<ImConfigVO[]>('/sys/im-config/list', { params: { providerId } })
}

/** 根据ID查询IM配置 */
export function getImConfigById(id: string) {
  return get<ImConfigVO>(`/sys/im-config/${id}`)
}

/** 新增IM配置 */
export function createImConfig(data: ImConfigVO) {
  return post<string>('/sys/im-config', data)
}

/** 更新IM配置 */
export function updateImConfig(data: ImConfigVO) {
  return put<void>('/sys/im-config', data)
}

/** 删除IM配置 */
export function deleteImConfig(id: string) {
  return del<void>(`/sys/im-config/${id}`)
}
