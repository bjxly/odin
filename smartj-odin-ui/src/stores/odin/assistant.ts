import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  getSkills,
  getSessions,
  createSession,
  removeSession,
  fetchSessionMessages,
} from '@/api/odin/assistant'
import { useOdinSSE } from '@/composables/useOdinSSE'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import { useOdinDraftStore } from '@/stores/odin/draft'
import { pickChartDatasets } from '@/utils/queryChart'
import type {
  AssistantSkillVO,
  ChatSession,
  AssistantMessageVO,
  AssistantMessage,
  MessageBlock,
  QueryDataset,
  ChartHint,
  SSEEvent,
} from '@/types/odin'

let seq = 0
function uid(prefix: string): string {
  seq += 1
  return `${prefix}-${Date.now().toString(36)}-${seq}`
}

function nowTime(): string {
  return new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

/** 流式增量缓冲区键：msgId + ':' + blockId */
function bufKey(msgId: string, blockId: string): string {
  return `${msgId}:${blockId}`
}

export const useOdinAssistantStore = defineStore('odinAssistant', () => {
  const sessions = ref<ChatSession[]>([])
  const activeSessionId = ref('')
  const messages = ref<AssistantMessage[]>([])
  const streaming = ref(false)
  const skills = ref<AssistantSkillVO[]>([])
  const loading = ref(false)

  // 当前工作区本体（限定问数语义上下文，随 chat 请求下发）
  const workspaceStore = useOdinWorkspaceStore()

  let abortController: AbortController | null = null

  // 流式增量 rAF 节流：delta 先入缓冲，每帧最多写一次响应式对象，避免逐 token 全量重解析
  const pendingText = new Map<string, string>()
  const pendingThinking = new Map<string, string>()
  let rafId: number | null = null

  /** 查找（或创建）指定 blockId 的块，返回其响应式对象 */
  function ensureBlock(
    msg: AssistantMessage,
    type: MessageBlock['type'],
    blockId: string,
  ): MessageBlock {
    let blk = msg.blocks.find(
      (b) => b.type === type && (b.payload as Record<string, unknown> | undefined)?.blockId === blockId,
    )
    if (!blk) {
      blk = {
        type,
        traceRef: msg.traceId,
        payload: { blockId, text: '', streaming: true },
      }
      msg.blocks.push(blk)
    }
    return blk
  }

  /** 把缓冲的增量落到响应式块正文（每帧一次） */
  function flushPending(msgId: string): void {
    const flush = (buf: Map<string, string>, type: MessageBlock['type']) => {
      for (const [key, delta] of buf) {
        if (!key.startsWith(`${msgId}:`)) continue
        const blockId = key.slice(msgId.length + 1)
        const blk = messages.value
          .find((m) => m.id === msgId)
          ?.blocks.find(
            (b) => b.type === type && (b.payload as Record<string, unknown> | undefined)?.blockId === blockId,
          )
        if (blk) {
          const p = blk.payload as Record<string, unknown>
          p.text = String(p.text ?? '') + delta
        }
        buf.delete(key)
      }
    }
    flush(pendingThinking, 'reasoning')
    flush(pendingText, 'text')
    // 同步正文到 content（供复制/兜底），取首个 text 块或拼接
    const m = messages.value.find((x) => x.id === msgId)
    if (m) {
      const texts = m.blocks
        .filter((b) => b.type === 'text')
        .map((b) => String((b.payload as Record<string, unknown>)?.text ?? ''))
        .join('\n\n')
      if (texts) m.content = texts
    }
  }

  /** 调度一次 rAF 刷新 */
  function scheduleFlush(msgId: string): void {
    if (rafId != null) return
    rafId = requestAnimationFrame(() => {
      rafId = null
      flushPending(msgId)
    })
  }

  /** 立即冲刷并取消挂起的 rAF（流结束/错误/中止时调用） */
  function forceFlush(msgId: string): void {
    if (rafId != null) {
      cancelAnimationFrame(rafId)
      rafId = null
    }
    flushPending(msgId)
  }

  const activeSession = computed(() =>
    sessions.value.find((s) => s.id === activeSessionId.value) ?? null,
  )

  /** 加载技能列表 */
  async function loadSkills(): Promise<void> {
    try {
      skills.value = await getSkills()
    } catch (e) {
      console.error('[odinAssistant] loadSkills failed', e)
    }
  }

  /** 加载会话列表（默认仅当前本体，保持列表与问数上下文一致） */
  async function loadSessions(): Promise<void> {
    loading.value = true
    try {
      sessions.value = await getSessions(workspaceStore.activeOntologyId || undefined)
    } catch (e) {
      console.error('[odinAssistant] loadSessions failed', e)
    } finally {
      loading.value = false
    }
  }

  /** 创建新会话 */
  async function newSession(ontologyId?: string): Promise<string> {
    try {
      const sessionId = await createSession({ ontologyId })
      activeSessionId.value = sessionId
      messages.value = []
      await loadSessions()
      return sessionId
    } catch (e) {
      console.error('[odinAssistant] createSession failed', e)
      return ''
    }
  }

  /** 删除会话 */
  async function deleteSession(id: string): Promise<void> {
    try {
      await removeSession(id)
      sessions.value = sessions.value.filter((s) => s.id !== id)
      if (activeSessionId.value === id) {
        activeSessionId.value = ''
        messages.value = []
      }
    } catch (e) {
      console.error('[odinAssistant] deleteSession failed', e)
    }
  }

  /** 切换到指定会话 */
  async function setActiveSession(sessionId: string): Promise<void> {
    if (!sessionId || streaming.value) return
    activeSessionId.value = sessionId
    messages.value = []
    try {
      const list = await fetchSessionMessages(sessionId)
      messages.value = (list ?? []).map(mapHistoryMessage)
    } catch (e) {
      console.error('[odinAssistant] setActiveSession failed', e)
    }
  }

  /** 历史消息 → 前端消息模型（解析 metadata 还原 text/reasoning/dataset/chart，解决刷新后图表丢失） */
  function mapHistoryMessage(vo: AssistantMessageVO, idx: number): AssistantMessage {
    const msg: AssistantMessage = {
      id: `h-${vo.id ?? idx}`,
      role: vo.role === 'user' ? 'user' : 'assistant',
      content: vo.content ?? '',
      blocks: [],
      time: vo.createTime ? new Date(vo.createTime).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) : '',
      traceId: '',
      streaming: false,
    }
    if (msg.role === 'assistant') {
      if (vo.metadata) {
        try {
          const meta = JSON.parse(vo.metadata) as {
            blocks?: Array<{ type: string; title?: string; payload?: unknown }>
            datasets?: QueryDataset[]
            chartHints?: Record<string, ChartHint>
          }
          if (Array.isArray(meta.blocks)) {
            msg.blocks = meta.blocks.map((b) => ({
              type: b.type as MessageBlock['type'],
              title: b.title ?? '',
              payload: b.payload,
              traceRef: '',
            }))
          }
          if (Array.isArray(meta.datasets)) msg.datasets = meta.datasets
          if (meta.chartHints && typeof meta.chartHints === 'object') msg.chartHints = meta.chartHints
        } catch {
          /* metadata 解析失败：降级为纯文本正文 */
        }
      }
      // 按统一规则（推荐优先、分析表优先、启发式闸门）重建可出图的 chart 块（与实流一致）
      ensureChartBlocks(msg)
      // 无块但有正文（旧数据/解析失败）：整体作为 text 块渲染
      if (!msg.blocks.length && msg.content) {
        msg.blocks = [{ type: 'text', payload: { text: msg.content }, traceRef: '' }]
      }
    }
    return msg
  }

  /** 发送消息（SSE 流式） */
  async function sendMessage(text: string): Promise<void> {
    const t = (text ?? '').trim()
    if (!t || streaming.value) return

    // 添加用户消息
    messages.value.push({
      id: uid('u'),
      role: 'user',
      content: t,
      blocks: [],
      time: nowTime(),
      streaming: false,
    })

    // 添加空的助手消息占位
    const assistantMsg: AssistantMessage = {
      id: uid('a'),
      role: 'assistant',
      content: '',
      blocks: [],
      time: nowTime(),
      streaming: true,
    }
    messages.value.push(assistantMsg)

    streaming.value = true
    abortController = new AbortController()

    const sse = useOdinSSE()

    await sse.connect(
      '/api/odin/assistant/chat/stream',
      {
        sessionId: activeSessionId.value || '',
        message: t,
        // 当前选中本体ID（后端据此注入本体词表并启用问数工具）；以字符串下发避免 JS 大数精度丢失
        ontologyId: workspaceStore.activeOntologyId || undefined,
      },
      (event: SSEEvent) => {
        handleSSEEvent(assistantMsg, event)
      },
      (error: Error) => {
        forceFlush(assistantMsg.id)
        assistantMsg.error = error.message || '助手处理失败'
        endStreaming(assistantMsg)
      },
      () => {
        forceFlush(assistantMsg.id)
        endStreaming(assistantMsg)
        streaming.value = false
        abortController = null
        // 刷新会话列表
        void loadSessions()
      },
      abortController.signal,
    )

    forceFlush(assistantMsg.id)
    endStreaming(assistantMsg)
    streaming.value = false
    abortController = null
  }

  /** 处理 SSE 事件 → 装配消息块 */
  function handleSSEEvent(msg: AssistantMessage, ev: SSEEvent): void {
    const data = ev.data as Record<string, unknown> | null
    switch (ev.event) {
      case 'meta': {
        if (data?.traceId) msg.traceId = String(data.traceId)
        if (data?.sessionId) activeSessionId.value = String(data.sessionId)
        if (data?.parsePath) msg.parsePath = String(data.parsePath)
        if (data?.skillId) msg.skillId = String(data.skillId)
        break
      }
      case 'intent':
      case 'reasoning':
      case 'sql':
      case 'clarification':
      case 'provenance':
      case 'block': {
        const block: MessageBlock = {
          type: (ev.event === 'block' ? (data as any)?.type : ev.event) as MessageBlock['type'],
          title: (data as any)?.title ?? '',
          payload: data,
          traceRef: msg.traceId,
        }
        // 兜底全文：后端非流式/关闭态下发整块 text；若已有同 blockId 流式块则整体替换，避免重复
        if (block.type === 'text' && (data as any)?.blockId) {
          const existing = msg.blocks.find(
            (b) => b.type === 'text' && (b.payload as any)?.blockId === (data as any).blockId,
          )
          if (existing) {
            existing.payload = data
            break
          }
        }
        msg.blocks.push(block)
        break
      }
      case 'text_delta': {
        const blockId = String((data as any)?.blockId ?? 'default')
        const delta = String((data as any)?.delta ?? '')
        if (!delta) break
        ensureBlock(msg, 'text', blockId)
        const key = bufKey(msg.id, blockId)
        pendingText.set(key, (pendingText.get(key) ?? '') + delta)
        scheduleFlush(msg.id)
        break
      }
      case 'text_end': {
        forceFlush(msg.id)
        const blockId = String((data as any)?.blockId ?? 'default')
        const blk = msg.blocks.find(
          (b) => b.type === 'text' && (b.payload as any)?.blockId === blockId,
        )
        if (blk) (blk.payload as Record<string, unknown>).streaming = false
        break
      }
      case 'thinking_delta': {
        const blockId = String((data as any)?.blockId ?? 'default')
        const delta = String((data as any)?.delta ?? '')
        if (!delta) break
        ensureBlock(msg, 'reasoning', blockId)
        const key = bufKey(msg.id, blockId)
        pendingThinking.set(key, (pendingThinking.get(key) ?? '') + delta)
        scheduleFlush(msg.id)
        break
      }
      case 'thinking_end': {
        forceFlush(msg.id)
        const blockId = String((data as any)?.blockId ?? 'default')
        const blk = msg.blocks.find(
          (b) => b.type === 'reasoning' && (b.payload as any)?.blockId === blockId,
        )
        if (blk) {
          const p = blk.payload as Record<string, unknown>
          p.streaming = false
          p.collapsed = true
        }
        break
      }
      case 'dataset': {
        // 数据集快照（含分析表）：缓存到 msg.datasets；chart 块统一在流末按 pickChartDatasets 规则创建
        const ds = data as any
        if (ds && ds.id) {
          if (!msg.datasets) msg.datasets = []
          msg.datasets.push(ds)
        }
        break
      }
      case 'chart_hint': {
        const hint = data as any
        if (hint?.dataId) {
          if (!msg.chartHints) msg.chartHints = {}
          msg.chartHints[String(hint.dataId)] = hint
        }
        break
      }
      case 'tool_call': {
        // 工具执行占位：本轮暂不建块（P5 图表/工具细化时再补），保持事件不报错
        break
      }
      case 'draft': {
        // 本轮 AI 产生了草稿：累加未读角标并刷新草稿列表（在函数内取 store，避免初始化期循环依赖）
        useOdinDraftStore().notifyNewDraft()
        break
      }
      case 'done': {
        forceFlush(msg.id)
        msg.streaming = false
        if (data?.sessionId) activeSessionId.value = String(data.sessionId)
        if (data?.traceId && !msg.traceId) msg.traceId = String(data.traceId)
        break
      }
      case 'error': {
        msg.error = (data as any)?.message || '助手处理失败'
        break
      }
      default:
        break
    }
  }

  /** 中止流式生成 */
  function abortStream(): void {
    if (abortController) {
      abortController.abort()
      abortController = null
    }
    streaming.value = false
    const last = messages.value[messages.value.length - 1]
    if (last?.streaming) {
      forceFlush(last.id)
      endStreaming(last)
    }
  }

  /**
   * 按统一规则创建 chart 块（实流 done / 历史还原共用）：
   * 模型推荐的分析表/数据集优先；无推荐时分析表优于原始查询，且最多一图。
   */
  function ensureChartBlocks(msg: AssistantMessage): void {
    for (const ds of pickChartDatasets(msg.datasets, msg.chartHints)) {
      if (!msg.blocks.some((b) => b.type === 'chart' && (b.payload as Record<string, unknown>)?.datasetId === ds.id)) {
        msg.blocks.push({ type: 'chart', title: ds.title || '', payload: { datasetId: ds.id }, traceRef: msg.traceId })
      }
    }
  }

  /** 流结束：关闭消息及所有块的流式标志（移除打字机光标） */
  function endStreaming(msg: AssistantMessage): void {
    msg.streaming = false
    ensureChartBlocks(msg)
    for (const b of msg.blocks) {
      const p = b.payload as Record<string, unknown> | undefined
      if (p && p.streaming === true) p.streaming = false
    }
  }

  function reset() {
    sessions.value = []
    activeSessionId.value = ''
    messages.value = []
    streaming.value = false
    skills.value = []
  }

  return {
    sessions,
    activeSessionId,
    activeSession,
    messages,
    streaming,
    skills,
    loading,
    loadSkills,
    loadSessions,
    createSession: newSession,
    deleteSession,
    setActiveSession,
    sendMessage,
    abortStream,
    reset,
  }
})
