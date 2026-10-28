<script setup lang="ts">
import { nextTick, onMounted, ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ChatDotRound,
  Plus,
  Position,
  Refresh,
  Delete,
  VideoPause,
  MagicStick,
  User,
} from '@element-plus/icons-vue'
import BlockRenderer from '@/views/odin/components/assistant/BlockRenderer.vue'
import ParsePathBadge from '@/views/odin/components/assistant/ParsePathBadge.vue'
import MessageMetaBar from '@/views/odin/components/assistant/MessageMetaBar.vue'
import OntologySwitcher from '@/views/odin/components/OntologySwitcher.vue'
import { useOdinAssistantStore } from '@/stores/odin/assistant'
import { useOdinQueryStore } from '@/stores/odin/query'
import { useOdinTraceStore } from '@/stores/odin/trace'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
// Types used by store & block components

const { t } = useI18n()
const assistantStore = useOdinAssistantStore()
const queryStore = useOdinQueryStore()
const traceStore = useOdinTraceStore()
const wsStore = useOdinWorkspaceStore()

const inputMessage = ref('')
const composerFocused = ref(false)
const messagesRef = ref<HTMLElement | null>(null)
const textareaRef = ref<any>(null)
const templatesExpanded = ref(false)

// ===== Computed =====
const sessions = computed(() => assistantStore.sessions)
const messages = computed(() => assistantStore.messages)
const streaming = computed(() => assistantStore.streaming)
const skills = computed(() => assistantStore.skills)
const activeSessionId = computed(() => assistantStore.activeSessionId)

const presetChips = computed<string[]>(() => {
  const qs = wsStore.activeOntology?.presetQuestions ?? []
  if (qs.length) return qs
  return [
    t('odinAssistant.chip1'),
    t('odinAssistant.chip2'),
    t('odinAssistant.chip3'),
    t('odinAssistant.chip4'),
  ]
})

// ===== Scroll =====
function scrollToBottom() {
  void nextTick(() => {
    const el = messagesRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

// ===== Auto-resize textarea =====
function autoResize() {
  void nextTick(() => {
    const el = textareaRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement | null
    if (!el) return
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 160) + 'px'
  })
}

// ===== Session management =====
async function handleNewSession() {
  await assistantStore.createSession(wsStore.activeOntologyId || undefined)
  scrollToBottom()
}

async function handleSelectSession(sessionId: string) {
  if (sessionId === activeSessionId.value) return
  await assistantStore.setActiveSession(sessionId)
  scrollToBottom()
}

async function handleDeleteSession(sessionId: string) {
  try {
    await ElMessageBox.confirm(
      t('odinAssistant.confirmDeleteSession'),
      t('common.tip'),
      { type: 'warning' },
    )
    await assistantStore.deleteSession(sessionId)
    ElMessage.success(t('common.deleteSuccess'))
  } catch { /* cancelled */ }
}

// ===== Send =====
async function send(text?: string) {
  const content = (text ?? inputMessage.value).trim()
  if (!content || streaming.value) return
  if (!text) inputMessage.value = ''
  autoResize()
  scrollToBottom()
  await assistantStore.sendMessage(content)
  scrollToBottom()
}

// ===== Block actions =====
function onBlockAction(a: { type: string; value?: string; traceId?: string }) {
  switch (a.type) {
    case 'clarify-option':
      if (a.value) void send(a.value)
      break
    case 'open-trace': {
      const id = a.traceId || a.value
      if (id) void traceStore.openTrace(id)
      break
    }
  }
}

function openTrace(traceId: string) {
  if (traceId) void traceStore.openTrace(traceId)
}

// ===== Templates =====
async function loadTemplates() {
  await queryStore.loadTemplates()
}

function insertTemplate(sql: string) {
  inputMessage.value = sql
  autoResize()
}

// ===== Session time format =====
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

// ===== Watch for auto-scroll =====
watch(() => messages.value.length, () => scrollToBottom())
watch(
  () => {
    const last = messages.value[messages.value.length - 1]
    return last ? `${last.id}:${last.blocks.length}:${last.streaming ? 1 : 0}:${last.content.length}` : ''
  },
  () => scrollToBottom(),
)

// ===== Lifecycle =====
onMounted(async () => {
  if (!wsStore.ontologies.length) void wsStore.loadOntologies()
  await Promise.all([
    assistantStore.loadSkills(),
    assistantStore.loadSessions(),
    loadTemplates(),
  ])
  scrollToBottom()
})

// 切换本体时刷新会话列表，避免历史会话与当前本体不匹配
watch(
  () => wsStore.activeOntologyId,
  async () => {
    await assistantStore.loadSessions()
    if (assistantStore.activeSessionId && !assistantStore.sessions.some((s) => s.id === assistantStore.activeSessionId)) {
      // 原活跃会话不属于当前本体：清空上下文，不自动进入某个历史会话
      assistantStore.activeSessionId = ''
      assistantStore.messages = []
    }
  },
)
</script>

<template>
  <div class="assistant-page">
    <!-- 左侧面板 -->
    <aside class="side-panel">
      <div class="side-head">
        <el-button type="primary" size="small" :icon="Plus" @click="handleNewSession">
          {{ t('odinAssistant.newSession') }}
        </el-button>
        <el-button
          size="small"
          :icon="Refresh"
          :loading="assistantStore.loading"
          @click="assistantStore.loadSessions()"
        />
      </div>

      <div class="side-body">
        <!-- 会话列表 -->
        <section class="side-sec">
          <h4>{{ t('odinAssistant.sessions') }}</h4>
          <div v-if="assistantStore.loading && !sessions.length" class="side-empty">...</div>
          <div v-else-if="!sessions.length" class="side-empty">{{ t('odinAssistant.noSessions') }}</div>
          <ul v-else class="session-list">
            <li
              v-for="s in sessions"
              :key="s.id"
              :class="{ active: s.id === activeSessionId }"
              @click="handleSelectSession(s.id)"
            >
              <div class="s-content">
                <div class="s-title">{{ s.title || t('odinAssistant.untitled') }}</div>
                <div class="s-meta">
                  {{ t('odinAssistant.msgCount', { n: s.messageCount ?? 0 }) }}
                  <template v-if="s.lastMessageAt || s.createTime">
                    · {{ fmtSessionTime(s.lastMessageAt || s.createTime) }}
                  </template>
                </div>
              </div>
              <el-button
                v-permission="'odin:assistant:session'"
                class="s-delete"
                link
                type="danger"
                size="small"
                :icon="Delete"
                @click.stop="handleDeleteSession(s.id)"
              />
            </li>
          </ul>
        </section>

        <!-- 预设问题 -->
        <section class="side-sec">
          <h4>{{ t('odinAssistant.presets') }}</h4>
          <div class="preset-list">
            <button
              v-for="q in presetChips"
              :key="q"
              type="button"
              class="preset-btn"
              :disabled="streaming"
              @click="send(q)"
            >
              {{ q }}
            </button>
          </div>
        </section>

        <!-- 可用技能 -->
        <section v-if="skills.length" class="side-sec">
          <h4>
            <span>{{ t('odinAssistant.skills') }}</span>
            <el-tag size="small" effect="plain" type="info">{{ skills.length }}</el-tag>
          </h4>
          <ul class="skill-list">
            <li v-for="s in skills" :key="s.code">
              <div class="sk-name">{{ s.name }}</div>
              <div class="sk-desc">{{ s.description }}</div>
            </li>
          </ul>
        </section>

        <!-- 查询模板 -->
        <section class="side-sec">
          <h4 class="clickable-header" @click="templatesExpanded = !templatesExpanded">
            <span>{{ t('odinAssistant.queryTemplates') }}</span>
            <el-icon :class="{ rotated: templatesExpanded }" class="expand-icon">
              <svg viewBox="0 0 1024 1024" width="12" height="12"><path d="M384 256l256 256-256 256z" fill="currentColor"/></svg>
            </el-icon>
          </h4>
          <div v-show="templatesExpanded" class="template-list">
            <div v-if="!queryStore.templates.length" class="side-empty">{{ t('odinAssistant.noTemplates') }}</div>
            <button
              v-for="tpl in queryStore.templates"
              :key="tpl.id"
              type="button"
              class="template-btn"
              @click="insertTemplate(tpl.logicalSql || tpl.name)"
            >
              {{ tpl.name }}
            </button>
          </div>
        </section>
      </div>
    </aside>

    <!-- 右侧对话区 -->
    <section class="chat-area">
      <!-- Header -->
      <header class="chat-header">
        <div class="head-left">
          <div class="avatar"><el-icon><ChatDotRound /></el-icon></div>
          <div class="head-text">
            <h2>{{ t('odinAssistant.title') }}</h2>
            <p>{{ t('odinAssistant.subtitle') }}</p>
          </div>
        </div>
        <div class="head-right">
          <OntologySwitcher />
          <el-button v-if="streaming" size="small" type="danger" plain :icon="VideoPause" @click="assistantStore.abortStream()">
            {{ t('odinAssistant.abort') }}
          </el-button>
        </div>
      </header>

      <!-- 消息流 -->
      <div ref="messagesRef" class="messages-container">
        <!-- 空状态 -->
        <div v-if="!messages.length" class="empty-state">
          <div class="empty-icon"><el-icon :size="48"><MagicStick /></el-icon></div>
          <p class="empty-text">{{ t('odinAssistant.emptyHint') }}</p>
          <div class="empty-chips">
            <button
              v-for="chip in presetChips"
              :key="chip"
              type="button"
              class="chip"
              :disabled="streaming"
              @click="send(chip)"
            >
              {{ chip }}
            </button>
          </div>
        </div>

        <!-- 消息列表 -->
        <div v-for="msg in messages" :key="msg.id" :class="['msg-row', msg.role]">
          <div class="msg-avatar">
            <el-icon v-if="msg.role === 'assistant'"><MagicStick /></el-icon>
            <el-icon v-else><User /></el-icon>
          </div>

          <div class="msg-main">
            <!-- 助手消息顶部：解析路径 + trace -->
            <div v-if="msg.role === 'assistant' && (msg.parsePath || msg.traceId)" class="msg-meta-top">
              <ParsePathBadge :path="msg.parsePath" />
              <button v-if="msg.traceId" type="button" class="trace-link" @click="openTrace(msg.traceId!)">
                {{ t('odinAssistant.traceId') }} {{ msg.traceId.slice(0, 8) }}
              </button>
              <span class="meta-time">{{ msg.time }}</span>
            </div>

            <!-- 消息气泡 -->
            <div class="msg-bubble" :class="msg.role">
              <!-- content 仅作兜底：无渲染块时（用户消息、旧数据/解析失败）才当纯文本显示，避免与 text 块重复 -->
              <div v-if="msg.content && !msg.blocks.length" class="msg-text">{{ msg.content }}</div>

              <!-- 渲染块 -->
              <div v-if="msg.blocks.length" class="msg-blocks">
                <BlockRenderer
                  v-for="(b, i) in msg.blocks"
                  :key="`${msg.id}-b${i}`"
                  :block="b"
                  :datasets="msg.datasets"
                  :chart-hints="msg.chartHints"
                  :streaming="msg.streaming"
                  @action="onBlockAction"
                />
              </div>

              <!-- 错误 -->
              <div v-if="msg.error" class="msg-error">{{ msg.error }}</div>

              <!-- 打字动画 -->
              <div v-if="msg.streaming && !msg.blocks.length && !msg.content && !msg.error" class="typing">
                <span /><span /><span />
              </div>
              <div v-else-if="msg.streaming" class="streaming-hint">{{ t('odinAssistant.streaming') }}</div>
            </div>

            <!-- MessageMetaBar -->
            <MessageMetaBar
              v-if="msg.role === 'assistant' && (msg as any).meta"
              :meta="(msg as any).meta"
            />

            <!-- 用户消息时间 -->
            <div v-if="msg.role === 'user'" class="msg-time-right">{{ msg.time }}</div>
          </div>
        </div>
      </div>

      <!-- 快捷 chips -->
      <div v-if="messages.length && presetChips.length" class="chips-row">
        <span class="chips-label">{{ t('odinAssistant.presetHint') }}</span>
        <button
          v-for="q in presetChips"
          :key="q"
          type="button"
          class="chip"
          :disabled="streaming"
          @click="send(q)"
        >
          {{ q }}
        </button>
      </div>

      <!-- 输入区 -->
      <div class="input-area">
        <div class="composer" :class="{ 'is-focused': composerFocused }">
          <el-input
            ref="textareaRef"
            v-model="inputMessage"
            type="textarea"
            :rows="1"
            resize="none"
            :placeholder="t('odinAssistant.inputPlaceholder')"
            :disabled="streaming"
            @input="autoResize"
            @focus="composerFocused = true"
            @blur="composerFocused = false"
            @keydown.enter.exact.prevent="send()"
          />
          <div class="composer__toolbar">
            <span class="composer__hint">{{ t('odinAssistant.enterHint') }}</span>
            <div class="composer__actions">
              <el-button
                v-if="streaming"
                class="composer__stop"
                type="danger"
                plain
                circle
                :icon="VideoPause"
                :title="t('odinAssistant.abort')"
                @click="assistantStore.abortStream()"
              />
              <button
                v-else
                type="button"
                class="composer__send"
                :disabled="!inputMessage.trim()"
                :title="t('odinAssistant.send')"
                @click="send()"
              >
                <el-icon><Position /></el-icon>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.assistant-page {
  height: 100%;
  display: flex;
  overflow: hidden;
  background: var(--el-bg-color-page);
}

/* ===== 左侧面板 ===== */
.side-panel {
  width: 288px;
  flex-shrink: 0;
  background: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color-lighter);
  display: flex;
  flex-direction: column;
}
.side-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.side-head .el-button--primary {
  flex: 1;
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
  color: var(--el-text-color-primary);
  letter-spacing: 0.3px;
}
.clickable-header {
  cursor: pointer;
  user-select: none;
}
.expand-icon {
  margin-left: auto;
  transition: transform 0.2s;
}
.expand-icon.rotated {
  transform: rotate(90deg);
}
.side-empty {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  padding: 6px 2px;
}

/* 会话列表 */
.session-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 240px;
  overflow-y: auto;
}
.session-list li {
  display: flex;
  align-items: center;
  padding: 8px 10px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.18s ease;
  background: var(--el-fill-color-blank);
}
.session-list li:hover {
  border-color: var(--el-color-primary-light-5);
  background: var(--el-color-primary-light-9);
}
.session-list li.active {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  box-shadow: inset 3px 0 0 var(--el-color-primary);
}
.s-content {
  flex: 1;
  min-width: 0;
}
.s-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.s-meta {
  margin-top: 2px;
  font-size: 11px;
  color: var(--el-text-color-secondary);
}
.s-delete {
  opacity: 0;
  transition: opacity 0.2s;
  flex-shrink: 0;
}
.session-list li:hover .s-delete {
  opacity: 1;
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
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.18s ease;
  word-break: break-word;
}
.preset-btn:hover:not(:disabled) {
  background: var(--el-color-primary-light-8);
  border-color: var(--el-color-primary-light-3);
  transform: translateX(2px);
}
.preset-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

/* 技能 */
.skill-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.sk-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  font-family: 'JetBrains Mono', Consolas, monospace;
}
.sk-desc {
  margin-top: 2px;
  font-size: 11px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

/* 模板 */
.template-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.template-btn {
  width: 100%;
  text-align: left;
  padding: 6px 10px;
  font-size: 12px;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.template-btn:hover {
  background: var(--el-color-primary-light-9);
  border-color: var(--el-color-primary-light-7);
  color: var(--el-color-primary);
}

/* ===== 对话区 ===== */
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
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-lighter);
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
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 19px;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--el-color-primary) 28%, transparent);
}
.head-text h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.head-text p {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
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

/* 空状态 */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}
.empty-icon {
  color: var(--el-color-primary-light-3);
}
.empty-text {
  font-size: 15px;
  color: var(--el-text-color-secondary);
  margin: 0;
}
.empty-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
  max-width: 560px;
}

/* 消息行 */
.msg-row {
  display: flex;
  gap: 10px;
  max-width: 88%;
}
.msg-row.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}
.msg-row.assistant {
  width: 100%;
  max-width: 100%;
}
.msg-avatar {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #fff;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
}
.msg-row.user .msg-avatar {
  background: linear-gradient(135deg, var(--el-text-color-secondary), var(--el-text-color-primary));
}
.msg-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.msg-meta-top {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}
.meta-time {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}
.trace-link {
  border: none;
  background: transparent;
  padding: 0;
  cursor: pointer;
  font-size: 11px;
  color: var(--el-color-primary);
  font-family: 'JetBrains Mono', Consolas, monospace;
  text-decoration: underline dotted;
}
.trace-link:hover {
  color: var(--el-color-primary-light-3);
}

/* 气泡 */
.msg-bubble {
  padding: 12px 14px;
  border-radius: 14px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  min-width: 0;
}
.msg-row.assistant .msg-bubble {
  border-top-left-radius: 4px;
  width: 100%;
}
.msg-row.user .msg-bubble {
  border-top-right-radius: 4px;
  background: var(--el-color-primary);
  border-color: transparent;
  color: #fff;
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
.msg-error {
  margin-top: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  background: var(--el-color-danger-light-9);
  border: 1px solid var(--el-color-danger-light-5);
  color: var(--el-color-danger);
  font-size: 12px;
  line-height: 1.6;
  word-break: break-word;
}
.msg-time-right {
  align-self: flex-end;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

/* 打字动画 */
.typing {
  display: flex;
  gap: 5px;
  padding: 4px 0;
}
.typing span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--el-color-primary);
  animation: typing-bounce 1.4s infinite;
}
.typing span:nth-child(2) { animation-delay: 0.2s; }
.typing span:nth-child(3) { animation-delay: 0.4s; }
@keyframes typing-bounce {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.5; }
  30% { transform: translateY(-6px); opacity: 1; }
}
.streaming-hint {
  margin-top: 8px;
  font-size: 11px;
  color: var(--el-color-primary);
  animation: pulse-hint 1.4s infinite;
}
@keyframes pulse-hint {
  0%, 100% { opacity: 0.45; }
  50% { opacity: 1; }
}

/* Chips */
.chips-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 20px;
  background: color-mix(in srgb, var(--el-bg-color) 70%, transparent);
  border-top: 1px solid var(--el-border-color-lighter);
  overflow-x: auto;
  flex-wrap: wrap;
}
.chips-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}
.chip {
  flex-shrink: 0;
  padding: 5px 12px;
  font-size: 12px;
  color: var(--el-color-primary);
  background: var(--el-bg-color);
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.18s ease;
  white-space: nowrap;
}
.chip:hover:not(:disabled) {
  background: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: #fff;
}
.chip:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

/* 输入区 */
.input-area {
  padding: 12px 20px 18px;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color-lighter);
}
.composer {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 10px 8px 12px;
  background: var(--el-fill-color-blank);
  border: 1px solid var(--el-border-color);
  border-radius: 18px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}
.composer.is-focused {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--el-color-primary) 12%, transparent);
}
.composer :deep(.el-textarea__inner) {
  padding: 4px 2px;
  font-size: 14px;
  line-height: 1.6;
  background: transparent;
  border: none;
  box-shadow: none;
  resize: none;
}
.composer :deep(.el-textarea__inner:focus) {
  box-shadow: none;
}
.composer__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 2px;
}
.composer__hint {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}
.composer__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}
.composer__send {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  padding: 0;
  font-size: 16px;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  border: none;
  border-radius: 50%;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--el-color-primary) 32%, transparent);
  transition: transform 0.15s ease, filter 0.15s ease, opacity 0.15s ease;
}
.composer__send:hover:not(:disabled) {
  filter: brightness(1.06);
  transform: translateY(-1px);
}
.composer__send:active:not(:disabled) {
  transform: translateY(0);
}
.composer__send:disabled {
  cursor: not-allowed;
  opacity: 0.45;
  box-shadow: none;
}
.composer__stop {
  width: 34px;
  height: 34px;
}
</style>
