/** 助手对话请求 */
export interface ChatRequest {
  message: string
  sessionId?: string
  ontologyId?: string
}

/** 助手对话响应 VO */
export interface ChatResponseVO {
  sessionId: string
  messageId: string
  content: string
  contentType: string
  done: boolean
  /** 本轮产生的首个 AI 草稿ID（无草稿时为空） */
  draftId?: string
  /** 本轮草稿明细项总数 */
  draftItemCount?: number
}

/** 助手会话 VO */
export interface AssistantSessionVO {
  id: string
  userId?: string
  title?: string
  ontologyId?: string
  ontologyName?: string
  status: string
  messageCount?: number
  createTime: string
  updateTime?: string
}

/** 助手消息 VO */
export interface AssistantMessageVO {
  id: string
  sessionId: string
  role: string
  content: string
  contentType?: string
  metadata?: string
  tokensUsed?: number
  latencyMs?: number
  createTime: string
}

/** 助手技能 VO */
export interface AssistantSkillVO {
  code: string
  name: string
  description: string
  icon?: string
}

/** 会话分页条件 */
export interface SessionQuery {
  current: number
  size: number
  userId?: string
  status?: string
}

/** AI 草稿 VO */
export interface AiDraftVO {
  id: string
  draftType: string
  ontologyId?: string
  ontologyName?: string
  status: string
  createdByAi?: number
  summary?: string
  itemCount?: number
  /** 明细项（仅详情接口返回） */
  items?: AiDraftItemVO[]
  createTime: string
}

/** AI 草稿明细项 VO */
export interface AiDraftItemVO {
  id: string
  draftId: string
  itemType?: string
  action: string
  targetId?: string
  beforeSnapshot?: string
  afterSnapshot?: string
  reviewStatus: string
  reviewerComment?: string
  createTime: string
}

/** 草稿分页条件 */
export interface DraftQuery {
  current: number
  size: number
  draftType?: string
  ontologyId?: string
  status?: string
}
