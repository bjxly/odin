import { get, post, del } from '@/api/request'
import type { PageResult } from '@/types'
import type {
  ChatRequest,
  ChatResponseVO,
  AssistantSessionVO,
  AssistantMessageVO,
  AssistantSkillVO,
  SessionQuery,
  AiDraftVO,
  DraftQuery,
  ChatSession,
} from '@/types/odin'

const BASE = '/odin/assistant'
const DRAFT_BASE = '/odin/ai/drafts'

// ===== 对话 =====

/** 发送对话消息（LLM 推理耗时较长，放宽超时） */
export function chat(data: ChatRequest) {
  return post<ChatResponseVO>(`${BASE}/chat`, data, { timeout: 300000 })
}

/** 查询助手技能列表 */
export function getAssistantSkills() {
  return get<AssistantSkillVO[]>(`${BASE}/skills`)
}

// ===== 会话 =====

/** 分页查询会话 */
export function getSessionPage(params: SessionQuery) {
  return get<PageResult<AssistantSessionVO>>(`${BASE}/session/page`, { params })
}

/** 查询会话消息列表 */
export function getSessionMessages(sessionId: string) {
  return get<AssistantMessageVO[]>(`${BASE}/session/${sessionId}/messages`)
}

/** 删除会话 */
export function deleteSession(sessionId: string) {
  return del<void>(`${BASE}/session/${sessionId}`)
}

// ===== AI 草稿审阅 =====

/** 分页查询 AI 草稿 */
export function getDraftPage(params: DraftQuery) {
  return get<PageResult<AiDraftVO>>(`${DRAFT_BASE}/page`, { params })
}

/** 查询草稿详情（含明细项） */
export function getDraftDetail(draftId: string) {
  return get<AiDraftVO>(`${DRAFT_BASE}/${draftId}`)
}

/** 审阅草稿明细项（approved/rejected），后端 body 字段为 comment */
export function reviewDraftItem(
  draftId: string,
  itemId: string,
  data: { reviewStatus: string; comment?: string },
) {
  return post<void>(`${DRAFT_BASE}/${draftId}/items/${itemId}/review`, data)
}

/** 发布草稿（将已通过的明细项应用到本体） */
export function publishDraft(draftId: string) {
  return post<void>(`${DRAFT_BASE}/${draftId}/publish`, undefined, { timeout: 60000 })
}

/** 驳回草稿（整体丢弃，可附原因） */
export function rejectDraft(draftId: string, reason?: string) {
  return post<void>(`${DRAFT_BASE}/${draftId}/reject`, undefined, { params: { reason } })
}

// ===== 会话管理（轻量接口） =====

/** 查询助手技能列表（简写） */
export function getSkills() {
  return get<AssistantSkillVO[]>(`${BASE}/skills`)
}

/** 查询会话列表（不分页，可按本体过滤） */
export function getSessions(ontologyId?: string) {
  return get<ChatSession[]>(`${BASE}/sessions`, { params: { ontologyId } })
}

/** 创建会话，返回 sessionId */
export function createSession(data: { ontologyId?: string; title?: string }) {
  return post<string>(`${BASE}/session`, data)
}

/** 删除会话 */
export function removeSession(id: string) {
  return del<void>(`${BASE}/session/${id}`)
}

/** 获取会话消息列表 */
export function fetchSessionMessages(sessionId: string) {
  return get<AssistantMessageVO[]>(`${BASE}/session/${sessionId}/messages`)
}
