<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { ElButton, ElIcon, ElInput, ElTag } from 'element-plus'
import {
  Back,
  ChatDotRound,
  Plus,
  Position,
  Refresh,
  VideoPause,
} from '@element-plus/icons-vue'
import BlockRenderer from '@/components/assistant/BlockRenderer.vue'
import ParsePathBadge from '@/components/assistant/ParsePathBadge.vue'
import MessageMetaBar from '@/components/assistant/MessageMetaBar.vue'
import { useAssistantStore } from '@/stores/assistant'
import { useOntologyStore } from '@/stores/ontology'
import { useTraceStore } from '@/stores/trace'
import { SKILL_LABELS } from '@/types/assistant'
import type { ChatAction, ChatMessage, ToolCallPayload } from '@/types/assistant'
import { shortId } from '@/utils/format'

/**
 * P3 智能助手：真实 SSE 流式对话。
 *
 * - POST /assistant/chat 以 fetch + ReadableStream 消费（见 api/assistant.ts）；
 * - 事件按类型装配为渲染块，经 BlockRenderer 注册表分发（见 components/assistant/blocks）；
 * - 预设问题 chips 来自 GET /assistant/skills 的 concept_qa.preset_questions；
 * - ProvenanceCard 点击 → trace store → TraceDrawer（挂载于 App.vue）；
 * - 左侧会话栏支持历史会话切换与消息回放。
 */
const as = useAssistantStore()
const onto = useOntologyStore()
const ts = useTraceStore()

const emit = defineEmits<{ goBack: [] }>()

const inputMessage = ref('')
const messagesRef = ref<HTMLElement | null>(null)

const WELCOME = `你好，我是 ODIN 智能助手。

我可以用业务语言帮你查数据，也可以解释本体与推理规则：
• 数据查询：例如「查询 VIP 客户的前 5 条」
• 概念问答：点击左侧预设问题即可
• 全链路溯源：每条数据答复都附带 trace，可一键查看 NL → 意图 → 推理 → SQL → 结果

答复中的「解析路径」标签会告诉你这次走了哪条链路：规则解析（零 LLM）、LLM 意图链、计划缓存或概念知识库。`

function scrollToBottom() {
  void nextTick(() => {
    const el = messagesRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

function skillLabel(id: string): string {
  return SKILL_LABELS[id] ?? id
}

/** 消息是否带有可展示的 meta（遍历规模 / 跨源），用于控制消息头渲染 */
function hasMsgMeta(msg: ChatMessage): boolean {
  const m = msg.meta
  if (!m) return false
  return (
    Number(m.hop_count) > 0 ||
    Number(m.llm_calls) > 0 ||
    Number(m.sub_intent_count) > 0 ||
    !!m.cross_source ||
    (Array.isArray(m.sources) && m.sources.length > 0)
  )
}

/**
 * P3-1(#56)：tool_call 标记去重。plan_execute 会在响应末尾重复下发
 * 「工具：plan · start/end」×N；按 (tool, intent_id, phase/status) 归并，
 * 同一键仅渲染一次，消除重复堆叠。
 */
function dedupeToolCalls(calls: ToolCallPayload[]): ToolCallPayload[] {
  if (!Array.isArray(calls) || calls.length <= 1) return calls ?? []
  const seen = new Set<string>()
  const out: ToolCallPayload[] = []
  for (const tc of calls) {
    const tool = String(tc.name || tc.skill_id || '未知')
    const args = (tc.args ?? {}) as Record<string, unknown>
    const intent = args.intent_id != null ? String(args.intent_id) : ''
    const phase = args.phase != null ? String(args.phase) : String(tc.status ?? '')
    const key = `${tool}|${intent}|${phase}`
    if (seen.has(key)) continue
    seen.add(key)
    out.push(tc)
  }
  return out
}

function fmtSessionTime(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  const hm = d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  if (sameDay) return hm
  return `${d.getMonth() + 1}/${d.getDate()} ${hm}`
}

/** 发送一条消息（预设问题 chips / 输入框 / 澄清选项共用） */
async function send(text?: string) {
  const t = (text ?? inputMessage.value).trim()
  if (!t) return
  if (as.sending) return
  if (!text) inputMessage.value = ''
  scrollToBottom()
  await as.sendMessage(t)
  scrollToBottom()
}

/** 渲染块交互统一入口 */
function onAction(a: ChatAction) {
  switch (a.type) {
    case 'clarify-option':
    case 'resend':
      void send(a.value)
      break
    case 'open-trace': {
      const id = a.traceId || a.value
      if (id) void ts.open(id)
      break
    }
    default:
      break
  }
}

function openTrace(traceId: string) {
  if (traceId) void ts.open(traceId)
}

function onNewSession() {
  as.newSession()
  as.pushLocalAssistantMessage(WELCOME)
  scrollToBottom()
}

async function onOpenSession(sessionId: string) {
  if (sessionId === as.activeSessionId) return
  await as.openSession(sessionId)
  scrollToBottom()
}

async function reloadSessions() {
  await as.loadSessions()
}

// 消息条数 / 末条消息块数变化时自动滚动，保证流式增量始终可见
watch(() => as.messages.length, () => scrollToBottom())
watch(
  () => {
    const last = as.messages[as.messages.length - 1]
    return last ? `${last.id}:${last.blocks.length}:${last.streaming ? 1 : 0}:${last.content.length}` : ''
  },
  () => scrollToBottom(),
)

onMounted(() => {
  void as.loadSkills()
  void as.loadSessions()
  if (!as.messages.length) as.pushLocalAssistantMessage(WELCOME)
  scrollToBottom()
})
</script>

<template>
  <div class="assistant-page">
    <!-- 左侧：会话 + 预设问题 + 技能 -->
    <aside class="side-panel">
      <div class="side-head">
        <ElButton :icon="Back" text size="small" @click="emit('goBack')">返回</ElButton>
        <ElButton type="primary" size="small" :icon="Plus" @click="onNewSession">新会话</ElButton>
      </div>

      <div class="side-body">
        <section class="side-sec">
          <h4>
            <span>会话历史</span>
            <ElButton
              class="sec-btn"
              text
              size="small"
              :icon="Refresh"
              :loading="as.sessionsLoading"
              @click="reloadSessions"
            />
          </h4>
          <div v-if="as.sessionsLoading && !as.sessions.length" class="side-empty">加载中…</div>
          <div v-else-if="!as.sessions.length" class="side-empty">
            暂无历史会话，发送第一条消息即自动创建
          </div>
          <ul v-else class="session-list">
            <li
              v-for="s in as.sessions"
              :key="s.session_id"
              :class="{ active: s.session_id === as.activeSessionId }"
              @click="onOpenSession(s.session_id)"
            >
              <div class="s-title">{{ s.title || '未命名会话' }}</div>
              <div class="s-meta">
                {{ s.message_count ?? 0 }} 条
                <template v-if="s.last_message_at || s.created_at">
                  · {{ fmtSessionTime(s.last_message_at || s.created_at) }}
                </template>
              </div>
            </li>
          </ul>
        </section>

        <section class="side-sec">
          <h4><span>预设问题</span></h4>
          <div v-if="as.skillsLoading && !as.skills.length" class="side-empty">加载技能清单中…</div>
          <div v-else-if="!as.presetQuestions.length" class="side-empty">后端未提供预设问题</div>
          <div v-else class="preset-list">
            <button
              v-for="q in as.presetQuestions"
              :key="q"
              type="button"
              class="preset-btn"
              :disabled="as.sending"
              @click="send(q)"
            >
              {{ q }}
            </button>
          </div>
        </section>

        <section v-if="as.skills.length" class="side-sec">
          <h4>
            <span>可用技能</span>
            <ElTag size="small" effect="plain" type="info">{{ as.skills.length }}</ElTag>
          </h4>
          <ul class="skill-list">
            <li v-for="s in as.skills" :key="s.id">
              <div class="sk-line">
                <span class="sk-name">{{ s.name }}</span>
                <ElTag size="small" :type="s.deterministic ? 'success' : 'primary'" effect="plain">
                  {{ s.deterministic ? '确定性' : 'LLM' }}
                </ElTag>
              </div>
              <div class="sk-desc">{{ s.description }}</div>
            </li>
          </ul>
        </section>
      </div>
    </aside>

    <!-- 右侧：对话区 -->
    <section class="chat-area">
      <header class="chat-header">
        <div class="head-left">
          <div class="avatar"><ElIcon><ChatDotRound /></ElIcon></div>
          <div class="head-text">
            <h2>ODIN 智能助手</h2>
            <p>本体驱动 · 确定性优先 · 每条答复可溯源</p>
          </div>
        </div>
        <div class="head-right">
          <ElTag v-if="as.activeSessionId" size="small" effect="plain" type="info">
            会话 {{ shortId(as.activeSessionId, 6, 4) }}
          </ElTag>
          <ElTag size="small" effect="plain" type="success">
            本体 {{ onto.currentOntologyId != null ? `#${onto.currentOntologyId}` : '未选择' }}
          </ElTag>
          <ElButton v-if="as.sending" size="small" type="danger" plain :icon="VideoPause" @click="as.stop()">
            停止
          </ElButton>
        </div>
      </header>

      <div ref="messagesRef" class="messages-container">
        <div v-for="msg in as.messages" :key="msg.id" :class="['msg-row', msg.role]">
          <div class="msg-avatar">{{ msg.role === 'user' ? '我' : 'AI' }}</div>

          <div class="msg-main">
            <!-- 助手消息元信息：技能 / 解析路径 / trace -->
            <div v-if="msg.role === 'assistant' && (msg.skillId || msg.parsePath || msg.traceId)" class="msg-meta">
              <ElTag v-if="msg.skillId" size="small" effect="plain" type="primary">
                {{ skillLabel(msg.skillId) }}
              </ElTag>
              <ParsePathBadge :path="msg.parsePath" />
              <button v-if="msg.traceId" type="button" class="trace-link" @click="openTrace(msg.traceId)">
                trace {{ shortId(msg.traceId, 6, 4) }}
              </button>
              <span class="meta-time">{{ msg.time }}</span>
            </div>

            <!-- P1-1/P3-5(#56)：消息级 meta 摘要徽章 + 跨源横幅 -->
            <MessageMetaBar
              v-if="msg.role === 'assistant' && hasMsgMeta(msg)"
              :meta="msg.meta"
            />

            <div class="msg-bubble" :class="msg.role">
              <div v-if="msg.content" class="msg-text">{{ msg.content }}</div>

              <div v-if="msg.blocks.length" class="msg-blocks">
                <BlockRenderer
                  v-for="(b, i) in msg.blocks"
                  :key="`${msg.id}-b${i}`"
                  :block="b"
                  @action="onAction"
                />
              </div>

              <div v-if="msg.toolCalls.length" class="tool-calls">
                <ElTag
                  v-for="(tc, i) in dedupeToolCalls(msg.toolCalls)"
                  :key="i"
                  size="small"
                  type="warning"
                  effect="plain"
                >
                  工具：{{ tc.name || tc.skill_id || '未知' }}{{ tc.status ? ` · ${tc.status}` : '' }}
                </ElTag>
              </div>

              <div v-if="msg.error" class="msg-error">{{ msg.error }}</div>

              <div
                v-if="msg.streaming && !msg.blocks.length && !msg.content && !msg.error"
                class="typing"
              >
                <span /><span /><span />
              </div>
              <div v-else-if="msg.streaming" class="streaming-hint">正在生成…</div>
            </div>

            <div v-if="msg.role === 'user'" class="msg-time-right">{{ msg.time }}</div>
          </div>
        </div>
      </div>

      <!-- 预设问题 chips：点击即发送 -->
      <div v-if="as.presetQuestions.length" class="chips-row">
        <span class="chips-label">试试</span>
        <button
          v-for="q in as.presetQuestions"
          :key="q"
          type="button"
          class="chip"
          :disabled="as.sending"
          @click="send(q)"
        >
          {{ q }}
        </button>
      </div>

      <div class="input-area">
        <ElInput
          v-model="inputMessage"
          type="textarea"
          :rows="2"
          resize="none"
          placeholder="用业务语言提问，例如：查询 VIP 客户的前 5 条"
          @keydown.enter.exact.prevent="send()"
        />
        <div class="input-actions">
          <span class="input-hint">Enter 发送 · Shift + Enter 换行</span>
          <ElButton v-if="as.sending" :icon="VideoPause" @click="as.stop()">停止生成</ElButton>
          <ElButton
            type="primary"
            :icon="Position"
            :loading="as.sending"
            :disabled="!inputMessage.trim()"
            @click="send()"
          >
            发送
          </ElButton>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.assistant-page {
  height: 100%;
  display: flex;
  background: linear-gradient(180deg, #f0f7ff 0%, #eef2f7 100%);
  overflow: hidden;
}

/* ---------------- 左侧面板 ---------------- */
.side-panel {
  width: 288px;
  flex-shrink: 0;
  background: #ffffff;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
}

.side-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(135deg, #2b6cb0, #3182ce);
}

.side-head :deep(.el-button) {
  color: #ffffff;
}

.side-head :deep(.el-button--primary) {
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.4);
}

.side-head :deep(.el-button--primary:hover) {
  background: rgba(255, 255, 255, 0.3);
}

.side-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.side-sec h4 {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 8px;
  font-size: 12px;
  font-weight: 700;
  color: #1a365d;
  letter-spacing: 0.4px;
  text-transform: none;
}

.side-sec h4 .sec-btn {
  margin-left: auto;
  padding: 2px;
  height: auto;
}

.side-empty {
  font-size: 12px;
  color: #a0aec0;
  line-height: 1.6;
  padding: 6px 2px;
}

/* 会话列表 */
.session-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 240px;
  overflow-y: auto;
}

.session-list li {
  padding: 8px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.18s ease;
  background: #fbfdff;
}

.session-list li:hover {
  border-color: #4c8bf5;
  background: #f0f7ff;
}

.session-list li.active {
  border-color: #2b6cb0;
  background: linear-gradient(135deg, #ebf4ff, #e0edff);
  box-shadow: inset 3px 0 0 #2b6cb0;
}

.s-title {
  font-size: 12px;
  font-weight: 600;
  color: #2d3748;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.s-meta {
  margin-top: 2px;
  font-size: 11px;
  color: #a0aec0;
}

/* 预设问题 */
.preset-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.preset-btn {
  width: 100%;
  text-align: left;
  padding: 8px 10px;
  font-size: 12px;
  line-height: 1.5;
  color: #2c5282;
  background: #f0f7ff;
  border: 1px solid #bee3f8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.18s ease;
  word-break: break-word;
}

.preset-btn:hover:not(:disabled) {
  background: #dcebff;
  border-color: #4c8bf5;
  transform: translateX(2px);
}

.preset-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

/* 技能清单 */
.skill-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sk-line {
  display: flex;
  align-items: center;
  gap: 6px;
}

.sk-name {
  font-size: 12px;
  font-weight: 600;
  color: #2d3748;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.sk-desc {
  margin-top: 2px;
  font-size: 11px;
  line-height: 1.55;
  color: #718096;
}

/* ---------------- 对话区 ---------------- */
.chat-area {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.chat-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.head-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 38px;
  height: 38px;
  border-radius: 11px;
  background: linear-gradient(135deg, #4c8bf5, #2b6cb0);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 19px;
  box-shadow: 0 4px 12px rgba(43, 108, 176, 0.28);
}

.head-text h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #1a365d;
}

.head-text p {
  margin: 2px 0 0;
  font-size: 12px;
  color: #718096;
}

.head-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 消息容器 */
.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.msg-row {
  display: flex;
  gap: 10px;
  max-width: 88%;
}

.msg-row.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #4c8bf5, #2b6cb0);
}

.msg-row.user .msg-avatar {
  background: linear-gradient(135deg, #64748b, #475569);
}

.msg-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.msg-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.meta-time {
  font-size: 11px;
  color: #a0aec0;
}

.trace-link {
  border: none;
  background: transparent;
  padding: 0;
  cursor: pointer;
  font-size: 11px;
  color: #2b6cb0;
  font-family: 'JetBrains Mono', Consolas, monospace;
  text-decoration: underline dotted;
}

.trace-link:hover {
  color: #4c8bf5;
}

.msg-bubble {
  padding: 12px 14px;
  border-radius: 14px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.05);
  min-width: 0;
}

.msg-row.assistant .msg-bubble {
  border-top-left-radius: 4px;
}

.msg-row.user .msg-bubble {
  border-top-right-radius: 4px;
  background: linear-gradient(135deg, #4c8bf5, #2b6cb0);
  border-color: transparent;
  color: #ffffff;
}

.msg-text {
  font-size: 14px;
  line-height: 1.75;
  white-space: pre-wrap;
  word-break: break-word;
}

.msg-blocks {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.msg-text + .msg-blocks {
  margin-top: 10px;
}

.msg-row.assistant .msg-bubble {
  /* 助手气泡承载结构化块，放宽宽度 */
  width: 100%;
}

.msg-row.assistant {
  width: 100%;
  max-width: 100%;
}

.tool-calls {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.msg-error {
  margin-top: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  background: #fff5f5;
  border: 1px solid #feb2b2;
  color: #c53030;
  font-size: 12px;
  line-height: 1.6;
  word-break: break-word;
}

.msg-time-right {
  align-self: flex-end;
  font-size: 11px;
  color: #a0aec0;
}

/* 打字指示器 */
.typing {
  display: flex;
  gap: 5px;
  padding: 4px 0;
}

.typing span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #4c8bf5;
  animation: typing 1.4s infinite;
}

.typing span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing span:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.5; }
  30% { transform: translateY(-6px); opacity: 1; }
}

.streaming-hint {
  margin-top: 8px;
  font-size: 11px;
  color: #4c8bf5;
  animation: pulse 1.4s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.45; }
  50% { opacity: 1; }
}

/* chips 快捷区 */
.chips-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 20px;
  background: rgba(255, 255, 255, 0.7);
  border-top: 1px solid #e2e8f0;
  overflow-x: auto;
  flex-wrap: wrap;
}

.chips-label {
  font-size: 11px;
  color: #718096;
  flex-shrink: 0;
}

.chip {
  flex-shrink: 0;
  padding: 5px 12px;
  font-size: 12px;
  color: #2c5282;
  background: #ffffff;
  border: 1px solid #bee3f8;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.18s ease;
  white-space: nowrap;
}

.chip:hover:not(:disabled) {
  background: #2b6cb0;
  border-color: #2b6cb0;
  color: #ffffff;
}

.chip:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

/* 输入区 */
.input-area {
  padding: 12px 20px 16px;
  background: #ffffff;
  border-top: 1px solid #e2e8f0;
}

.input-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
}

.input-hint {
  font-size: 11px;
  color: #a0aec0;
  margin-right: auto;
}
</style>
