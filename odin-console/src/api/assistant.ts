import request from './request'
import type {
  DoneEventPayload,
  ErrorEventPayload,
  MessageVO,
  SessionMessagesVO,
  SessionPageVO,
  SessionVO,
  SkillListVO,
  SkillVO,
  SSEEvent,
} from '@/types/assistant'

/**
 * P3 智能助手 API。
 *
 * 说明：`chat` 是 SSE（text/event-stream）流式端点，axios 不支持增量读取，
 * 因此改用原生 fetch + ReadableStream 手工解析 `event:` / `data:` 行。
 * 其余端点仍走统一封装的 axios（自动解包 data、自动错误提示）。
 */
const API_BASE = '/api/v1'

// ---------------------------------------------------------------------------
// 技能清单
// ---------------------------------------------------------------------------

/**
 * GET /assistant/skills
 * 后端返回 { skills: Manifest[], count, trace_id }；兼容裸数组形态。
 */
export async function getSkills(): Promise<SkillVO[]> {
  const r = await request.get('/assistant/skills')
  const data = r.data as SkillListVO | SkillVO[] | null | undefined
  if (Array.isArray(data)) return data
  return data?.skills ?? []
}

// ---------------------------------------------------------------------------
// 会话历史
// ---------------------------------------------------------------------------

/**
 * GET /assistant/sessions
 * 后端为分页结构 { records, total, current, size }；兼容裸数组形态。
 */
export async function getSessions(params?: { current?: number; size?: number; ontology_id?: number }): Promise<SessionVO[]> {
  const r = await request.get('/assistant/sessions', { params })
  const data = r.data as SessionPageVO | SessionVO[] | null | undefined
  if (Array.isArray(data)) return data
  return data?.records ?? []
}

/** GET /assistant/sessions/:id/messages → { session_id, messages[], count, trace_id } */
export async function getSessionMessages(sessionId: string): Promise<MessageVO[]> {
  const r = await request.get(`/assistant/sessions/${encodeURIComponent(sessionId)}/messages`)
  const data = r.data as SessionMessagesVO | MessageVO[] | null | undefined
  if (Array.isArray(data)) return data
  return data?.messages ?? []
}

// ---------------------------------------------------------------------------
// SSE 对话（POST /assistant/chat）
// ---------------------------------------------------------------------------

export interface ChatPayload {
  /** 空则由后端生成新会话（uuid） */
  session_id?: string
  text: string
  ontology_id?: number
}

export interface ChatSSEHandlers {
  /** 每条解析成功的事件（data 已尝试 JSON 解析） */
  onEvent?: (ev: SSEEvent) => void
  /** 收到 done 事件或流自然结束 */
  onDone?: (data: DoneEventPayload | null) => void
  /** 业务错误：error 事件、非 SSE 响应、HTTP 非 2xx */
  onError?: (message: string, traceId?: string) => void
  /** 主动取消或网络中断 */
  onAbort?: (reason: string) => void
  /** AbortSignal，用于「停止生成」 */
  signal?: AbortSignal
}

/**
 * 解析单个 SSE 事件块（已按空行切分）。
 * 规则遵循 WHATWG EventSource：忽略 `:` 注释行、`id:`/`retry:` 字段；
 * 多行 `data:` 以 \n 拼接后再整体 JSON 解析。
 */
export function parseSSEChunk(chunk: string): SSEEvent | null {
  const lines = chunk.split(/\r?\n/)
  let event = ''
  const dataLines: string[] = []
  for (const rawLine of lines) {
    const line = rawLine.endsWith('\r') ? rawLine.slice(0, -1) : rawLine
    if (!line) continue
    // 以 ':' 开头为注释/心跳，直接忽略
    if (line.startsWith(':')) continue
    const idx = line.indexOf(':')
    let field: string
    let value: string
    if (idx === -1) {
      field = line
      value = ''
    } else {
      field = line.slice(0, idx)
      value = line.slice(idx + 1)
      // 冒号后的单个空格属于分隔符，需剥离
      if (value.startsWith(' ')) value = value.slice(1)
    }
    if (field === 'event') {
      event = value.trim()
    } else if (field === 'data') {
      dataLines.push(value)
    }
    // id / retry 等字段当前无需处理
  }
  const raw = dataLines.join('\n')
  if (!event && !raw) return null
  let data: unknown = null
  if (raw) {
    try {
      data = JSON.parse(raw)
    } catch {
      // 非 JSON（纯文本 data）时原样保留字符串
      data = raw
    }
  }
  return { event: event || 'message', data, raw }
}

/** 从未知响应体中提取可读错误文案 */
function extractMessage(text: string, fallback: string): string {
  if (!text) return fallback
  try {
    const body = JSON.parse(text) as { message?: string; error?: string; data?: { message?: string } }
    return body?.message || body?.error || body?.data?.message || fallback
  } catch {
    return text.length > 300 ? `${text.slice(0, 300)}…` : text
  }
}

/**
 * 以流式方式发起一次助手问答。
 *
 * 稳健性保障：
 * - 响应 Content-Type 非 text/event-stream 时（后端参数校验失败走 JSON）读取 JSON 提取 message；
 * - 增量解码使用 `{ stream: true }`，避免多字节字符被切断；
 * - 事件以空行（

 或 \r
\r
）切分，跨 chunk 残片保留在 buffer；
 * - 流结束后 flush decoder 并处理尾部无空行的残片；
 * - reader 异常/主动取消均回收锁并回调，不抛出未捕获异常。
 */
export async function chatSSE(payload: ChatPayload, handlers: ChatSSEHandlers = {}): Promise<void> {
  const { onEvent, onDone, onError, onAbort, signal } = handlers

  let resp: Response
  try {
    resp = await fetch(`${API_BASE}/assistant/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'text/event-stream',
      },
      body: JSON.stringify({
        session_id: payload.session_id ?? '',
        text: payload.text ?? '',
        ontology_id: payload.ontology_id ?? 0,
      }),
      signal,
    })
  } catch (e) {
    const err = e as Error
    if (err?.name === 'AbortError') {
      onAbort?.('已取消')
      return
    }
    onError?.(err?.message || '无法连接助手服务，请确认后端已启动')
    return
  }

  const contentType = resp.headers.get('content-type') || ''

  // 非 SSE：可能是参数校验失败（fail() 返回 JSON）或网关错误页
  if (!resp.ok || !contentType.includes('text/event-stream')) {
    const text = await resp.text().catch(() => '')
    const fallback = resp.ok ? '助手服务未返回事件流' : `请求失败（HTTP ${resp.status}）`
    onError?.(extractMessage(text, fallback))
    return
  }

  if (!resp.body) {
    onError?.('当前浏览器不支持流式响应（ReadableStream）')
    return
  }

  const reader = resp.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let doneData: DoneEventPayload | null = null
  let failed = false
  let aborted = false

  /** 分发单条事件；返回是否为终止事件 */
  const dispatch = (ev: SSEEvent): void => {
    onEvent?.(ev)
    if (ev.event === 'done') {
      doneData = (ev.data as DoneEventPayload | null) ?? null
    } else if (ev.event === 'error') {
      const p = (ev.data as ErrorEventPayload | null) ?? null
      failed = true
      const msg = typeof ev.data === 'string' ? ev.data : p?.message || '助手处理失败'
      onError?.(msg, p?.trace_id)
    }
  }

  try {
    for (;;) {
      const { value, done } = await reader.read()
      if (done) break
      if (value) buffer += decoder.decode(value, { stream: true })
      // 切出所有完整事件（以空行为界）
      for (;;) {
        const m = /\r?\n\r?\n/.exec(buffer)
        if (!m) break
        const chunk = buffer.slice(0, m.index)
        buffer = buffer.slice(m.index + m[0].length)
        const ev = chunk.trim() ? parseSSEChunk(chunk) : null
        if (ev) dispatch(ev)
      }
    }
    // flush 解码器，处理最后一字节
    buffer += decoder.decode()
    if (buffer.trim()) {
      const ev = parseSSEChunk(buffer)
      if (ev) dispatch(ev)
      buffer = ''
    }
  } catch (e) {
    const err = e as Error
    if (err?.name === 'AbortError') {
      aborted = true
    } else if (!failed) {
      failed = true
      onError?.(err?.message || '事件流中断')
    }
  } finally {
    try {
      reader.releaseLock()
    } catch {
      // 流已关闭时 releaseLock 可能抛错，忽略
    }
  }

  if (aborted) onAbort?.('已停止生成')
  else onDone?.(doneData)
}
