import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import * as assistantApi from '@/api/assistant'
import { useOntologyStore } from './ontology'
import type {
  BlockVO,
  ChatMessage,
  ClarificationPayload,
  DoneEventPayload,
  IntentPayload,
  MetaEventPayload,
  MessageMeta,
  MessageVO,
  ProvenancePayload,
  ReasoningPayload,
  SessionVO,
  SkillVO,
  SqlPayload,
  SSEEvent,
} from '@/types/assistant'

/**
 * P3 智能助手 store。
 *
 * 职责：技能清单（含预设问题）、会话列表、当前会话消息流，
 * 以及通过 SSE 增量装配助手消息（meta/intent/reasoning/sql/block/
 * clarification/provenance/done/error）。
 *
 * #66 新 SQL 生成架构：已退役 thinking / tool_call 事件处理，
 * SQL 以 type="sql" 的 block 事件整体下发（payload.steps）。
 */
let seq = 0
function uid(prefix: string): string {
  seq += 1
  return `${prefix}-${Date.now().toString(36)}-${seq}`
}

function nowTime(): string {
  return new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function fmtTime(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function emptyAssistantMessage(): ChatMessage {
  return {
    id: uid('a'),
    role: 'assistant',
    content: '',
    blocks: [],
    time: nowTime(),
    traceId: '',
    parsePath: '',
    skillId: '',
    matched: false,
    streaming: true,
    thinking: [],
    toolCalls: [],
    error: '',
  }
}

export const useAssistantStore = defineStore('assistant', () => {
  // --- 技能清单 ---
  const skills = ref<SkillVO[]>([])
  const skillsLoading = ref(false)
  const skillsLoaded = ref(false)

  /**
   * 预设问题 chips 数据源：优先取 concept_qa 技能的 preset_questions，
   * 其次汇总其余技能的预设问题与示例问句。
   */
  const presetQuestions = computed<string[]>(() => {
    const concept = skills.value.find((s) => s.id === 'concept_qa' || s.type === 'concept_qa')
    const own = (concept?.preset_questions ?? []).filter(Boolean)
    if (own.length) return own
    const acc: string[] = []
    for (const s of skills.value) {
      for (const q of s.preset_questions ?? []) if (q && !acc.includes(q)) acc.push(q)
      for (const ex of s.examples ?? []) if (ex?.nl && !acc.includes(ex.nl)) acc.push(ex.nl)
    }
    return acc
  })

  async function loadSkills(force = false): Promise<SkillVO[]> {
    if (skillsLoaded.value && !force) return skills.value
    skillsLoading.value = true
    try {
      skills.value = await assistantApi.getSkills()
      skillsLoaded.value = true
    } catch (e: any) {
      console.error('[assistant] loadSkills failed', e)
    } finally {
      skillsLoading.value = false
    }
    return skills.value
  }

  // --- 会话 ---
  const sessions = ref<SessionVO[]>([])
  const sessionsLoading = ref(false)
  const activeSessionId = ref('')
  const messages = ref<ChatMessage[]>([])
  const sending = ref(false)
  let abortCtrl: AbortController | null = null

  const activeSession = computed<SessionVO | null>(
    () => sessions.value.find((s) => s.session_id === activeSessionId.value) ?? null,
  )
  const hasMessages = computed(() => messages.value.length > 0)

  async function loadSessions(): Promise<SessionVO[]> {
    sessionsLoading.value = true
    try {
      sessions.value = await assistantApi.getSessions({ current: 1, size: 50 })
    } catch (e: any) {
      console.error('[assistant] loadSessions failed', e)
    } finally {
      sessionsLoading.value = false
    }
    return sessions.value
  }

  /**
   * P4(#43)：统一 live 流式渲染与历史回放的块标签。
   * 后端持久化时可能用不同标题（如「解析意图」「生成的SQL」「推理链路」「溯源」），
   * 前端 live 路径统一用下方规范名，历史回放时按 type 归一即可。
   */
  const CANONICAL_BLOCK_TITLES: Record<string, string> = {
    intent: '查询意图',
    sql: '查询语句',
    reasoning: '推理轨迹',
    provenance: '数据溯源',
    clarification: '需要澄清',
  }

  function normalizeBlockTitle(b: BlockVO): BlockVO {
    const canonical = CANONICAL_BLOCK_TITLES[b.type]
    if (!canonical) return b
    // 已经一致则不创建新对象（避免无谓的响应式触发）
    if (b.title === canonical) return b
    return { ...b, title: canonical }
  }

  /** 历史消息视图 → 前端聊天消息（有 blocks 时不再重复渲染 content） */
  function mapHistoryMessage(vo: MessageVO, idx: number): ChatMessage {
    const blocks = (Array.isArray(vo.blocks) ? vo.blocks : []).filter(
      (b): b is BlockVO => !!b && typeof b === 'object' && !!b.type,
    ).map(normalizeBlockTitle)
    return {
      id: `h-${vo.id ?? idx}`,
      role: vo.role === 'user' ? 'user' : 'assistant',
      content: blocks.length ? '' : vo.content ?? '',
      blocks,
      time: fmtTime(vo.created_at),
      traceId: vo.trace_id ?? '',
      parsePath: vo.parse_path ?? '',
      skillId: vo.skill_id ?? '',
      matched: false,
      streaming: false,
      thinking: [],
      toolCalls: [],
      error: '',
      meta: metaFromRecord(vo.meta),
      history: true,
    }
  }

  /** 切换到指定会话并加载历史消息 */
  async function openSession(sessionId: string): Promise<void> {
    const sid = (sessionId ?? '').trim()
    if (!sid || sending.value) return
    stop()
    activeSessionId.value = sid
    messages.value = []
    try {
      const list = await assistantApi.getSessionMessages(sid)
      messages.value = (list ?? []).map(mapHistoryMessage)
    } catch (e: any) {
      console.error('[assistant] openSession failed', e)
    }
  }

  /** 新建会话：清空上下文，session_id 由后端在首次 chat 时生成 */
  function newSession(): void {
    stop()
    activeSessionId.value = ''
    messages.value = []
  }

  /** 中止当前流式生成 */
  function stop(): void {
    if (abortCtrl) {
      try {
        abortCtrl.abort()
      } catch {
        // 已结束的控制器的 abort 是空操作
      }
      abortCtrl = null
    }
    sending.value = false
    const last = messages.value[messages.value.length - 1]
    if (last?.streaming) last.streaming = false
  }

  /** 当前本体 id（未选择则为 0，后端按默认本体处理） */
  function ontologyIdOf(): number {
    const onto = useOntologyStore()
    return onto.currentOntologyId ?? 0
  }

  // --- SSE 事件 → 消息块装配 ---

  /**
   * P1-1/P3-5(#56)：把消息级 meta（遍历规模 + 跨源）合并到 ChatMessage.meta。
   * live 由 meta / done 事件写入，历史由 mapHistoryMessage 从 MessageVO.meta 还原，
   * 供 Assistant 页渲染「N 跳 · N 轮 LLM · N 子意图」摘要徽章与跨源横幅。
   */
  function applyMeta(msg: ChatMessage, p: MetaEventPayload): void {
    const has =
      p.sub_intent_count != null ||
      p.hop_count != null ||
      p.llm_calls != null ||
      p.cross_source != null ||
      p.sources != null ||
      p.parse_path != null
    if (!has) return
    const prev = msg.meta ?? {}
    msg.meta = {
      parse_path: p.parse_path ?? prev.parse_path,
      sub_intent_count: p.sub_intent_count ?? prev.sub_intent_count,
      hop_count: p.hop_count ?? prev.hop_count,
      llm_calls: p.llm_calls ?? prev.llm_calls,
      cross_source: p.cross_source ?? prev.cross_source,
      sources: p.sources ?? prev.sources,
    }
  }

  /** 历史消息 meta（Record<string,unknown>）→ MessageMeta，做字段与类型归一 */
  function metaFromRecord(m?: Record<string, unknown> | null): MessageMeta | null {
    if (!m || typeof m !== 'object') return null
    const meta: MessageMeta = {}
    if (m.parse_path != null) meta.parse_path = String(m.parse_path) as MessageMeta['parse_path']
    if (m.sub_intent_count != null) meta.sub_intent_count = Number(m.sub_intent_count)
    if (m.hop_count != null) meta.hop_count = Number(m.hop_count)
    if (m.llm_calls != null) meta.llm_calls = Number(m.llm_calls)
    if (m.cross_source != null) meta.cross_source = !!m.cross_source
    if (Array.isArray(m.sources)) meta.sources = m.sources.map((s) => String(s))
    return Object.keys(meta).length ? meta : null
  }

  function pushBlock(msg: ChatMessage, block: BlockVO): void {
    msg.blocks.push(block)
  }

  function applyEvent(msg: ChatMessage, ev: SSEEvent): void {
    const data = ev.data
    switch (ev.event) {
      case 'meta': {
        const p = (data ?? {}) as MetaEventPayload
        if (p.trace_id) msg.traceId = p.trace_id
        if (p.session_id) activeSessionId.value = p.session_id
        if (p.parse_path) msg.parsePath = p.parse_path
        if (p.skill_id) msg.skillId = p.skill_id
        msg.matched = !!p.matched
        applyMeta(msg, p)
        break
      }
      case 'intent':
        pushBlock(msg, {
          type: 'intent',
          title: '查询意图',
          payload: (data ?? {}) as IntentPayload,
          trace_ref: msg.traceId,
        })
        break
      case 'reasoning':
        pushBlock(msg, {
          type: 'reasoning',
          title: '推理轨迹',
          payload: (data ?? {}) as ReasoningPayload,
          trace_ref: msg.traceId,
        })
        break
      case 'sql':
        pushBlock(msg, {
          type: 'sql',
          title: '翻译 SQL',
          payload: (data ?? {}) as SqlPayload,
          trace_ref: msg.traceId,
        })
        break
      case 'clarification':
        pushBlock(msg, {
          type: 'clarification',
          title: '需要澄清',
          payload: (data ?? {}) as ClarificationPayload,
          trace_ref: msg.traceId,
        })
        break
      case 'provenance': {
        const p = (data ?? {}) as ProvenancePayload
        pushBlock(msg, {
          type: 'provenance',
          title: '数据溯源',
          payload: p,
          trace_ref: p.trace_id || msg.traceId,
        })
        break
      }
      case 'block': {
        // 后端整块下发（含 type/title/payload/trace_ref）
        const b = data as BlockVO | null
        if (b && typeof b === 'object' && b.type) {
          pushBlock(msg, { ...b, trace_ref: b.trace_ref || msg.traceId })
        }
        break
      }
      case 'done':
      case 'error':
        // 由 chatSSE 的 onDone / onError 统一收口
        break
      default: {
        // 前向兼容：未来新增事件若本身即为 Block 形态，直接渲染
        const b = data as BlockVO | null
        if (b && typeof b === 'object' && b.type) pushBlock(msg, b)
        break
      }
    }
  }

  /**
   * 发送一条消息并以 SSE 流式接收回复。
   * 增量渲染：每个事件到达即写入 blocks，界面实时刷新。
   */
  async function sendMessage(text: string): Promise<void> {
    const t = (text ?? '').trim()
    if (!t) return
    if (sending.value) {
      ElMessage.warning('上一条回复仍在生成中，请先停止或等待完成')
      return
    }

    messages.value.push({
      id: uid('u'),
      role: 'user',
      content: t,
      blocks: [],
      time: nowTime(),
      traceId: '',
      parsePath: '',
      skillId: '',
      matched: false,
      streaming: false,
      thinking: [],
      toolCalls: [],
      error: '',
    })
    messages.value.push(emptyAssistantMessage())
    const msg = messages.value[messages.value.length - 1]

    sending.value = true
    abortCtrl = new AbortController()

    await assistantApi.chatSSE(
      { session_id: activeSessionId.value || '', text: t, ontology_id: ontologyIdOf() },
      {
        signal: abortCtrl.signal,
        onEvent: (ev) => applyEvent(msg, ev),
        onDone: (d: DoneEventPayload | null) => {
          msg.streaming = false
          if (d?.session_id) activeSessionId.value = d.session_id
          if (d?.trace_id && !msg.traceId) msg.traceId = d.trace_id
          if (d?.skill_id && !msg.skillId) msg.skillId = d.skill_id
          if (d?.parse_path && !msg.parsePath) msg.parsePath = d.parse_path
          if (d) applyMeta(msg, d as MetaEventPayload)
          if (!msg.blocks.length && !msg.error) {
            msg.content = msg.content || '（本次未返回可渲染内容）'
          }
          void loadSessions()
        },
        onError: (message: string, traceId?: string) => {
          msg.error = msg.error || message || '助手处理失败'
          if (traceId && !msg.traceId) msg.traceId = traceId
          ElMessage.error(msg.error)
        },
        onAbort: () => {
          msg.streaming = false
          if (!msg.blocks.length && !msg.error) msg.content = '（已停止生成）'
        },
      },
    )

    msg.streaming = false
    sending.value = false
    abortCtrl = null
  }

  /** 插入一条本地助手消息（欢迎语等，不走后端） */
  function pushLocalAssistantMessage(content: string): void {
    messages.value.push({
      ...emptyAssistantMessage(),
      id: uid('w'),
      content,
      streaming: false,
    })
  }

  function clearMessages(): void {
    messages.value = []
  }

  return {
    // 技能
    skills, skillsLoading, skillsLoaded, presetQuestions, loadSkills,
    // 会话
    sessions, sessionsLoading, activeSessionId, activeSession, messages, hasMessages, sending,
    loadSessions, openSession, newSession, stop,
    // 对话
    sendMessage, pushLocalAssistantMessage, clearMessages,
  }
})
