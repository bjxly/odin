<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Position, VideoPause } from '@element-plus/icons-vue'
import BlockRenderer from './BlockRenderer.vue'
import { useOdinAssistantStore } from '@/stores/odin/assistant'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'

/**
 * ODIN AI 副驾对话面板。
 *
 * 复用全局助手 store（SSE 流式对话 + 本体上下文），消息块经 BlockRenderer 分发渲染。
 * 与助手页共享同一 store，因此在抽屉中发起的对话与助手页会话一致。
 */
const { t } = useI18n()
const assistantStore = useOdinAssistantStore()
const workspaceStore = useOdinWorkspaceStore()

const input = ref('')
const inputFocused = ref(false)
const listRef = ref<HTMLElement | null>(null)
const textareaRef = ref<any>(null)

/** 自适应高度：随内容生长，最高 160px */
function autoResize(): void {
  void nextTick(() => {
    const el = textareaRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement | null
    if (!el) return
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 160) + 'px'
  })
}

/** 滚动到底部（流式追加时保持跟随） */
async function scrollToBottom(): Promise<void> {
  await nextTick()
  const el = listRef.value
  if (el) {
    el.scrollTop = el.scrollHeight
  }
}

watch(
  () => assistantStore.messages.length,
  () => void scrollToBottom(),
)

/** Enter 发送，Shift+Enter 换行 */
function onKeydown(e: KeyboardEvent): void {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    void send()
  }
}

/** 发送消息 */
async function send(): Promise<void> {
  const text = input.value.trim()
  if (!text || assistantStore.streaming) {
    return
  }
  input.value = ''
  autoResize()
  await assistantStore.sendMessage(text)
  await scrollToBottom()
}

/** 新建会话（携带当前本体上下文） */
async function onNewSession(): Promise<void> {
  await assistantStore.createSession(workspaceStore.activeOntologyId || undefined)
}

/** 会话与当前本体不匹配时清空上下文（不自动进入历史会话，与助手页行为一致） */
function resetIfOntologyMismatch(): void {
  const sid = assistantStore.activeSessionId
  if (!sid) return
  const sess = assistantStore.sessions.find((s) => s.id === sid)
  if (!sess || (sess.ontologyId ?? '') !== (workspaceStore.activeOntologyId || '')) {
    assistantStore.activeSessionId = ''
    assistantStore.messages = []
  }
}

onMounted(async () => {
  // 抽屉关闭时组件会卸载，重新打开时旧会话可能已属于其它本体
  resetIfOntologyMismatch()
  if (!assistantStore.activeSessionId) {
    await assistantStore.createSession(workspaceStore.activeOntologyId || undefined)
  }
  await scrollToBottom()
})

// 抽屉开着时切换本体：同步重置会话上下文
watch(
  () => workspaceStore.activeOntologyId,
  () => resetIfOntologyMismatch(),
)
</script>

<template>
  <div class="copilot-chat">
    <div class="copilot-chat__toolbar">
      <span class="copilot-chat__ontology">
        {{ workspaceStore.activeOntology?.name || t('odinAssistant.copilotNoOntology') }}
      </span>
      <el-button link type="primary" :icon="Plus" :disabled="assistantStore.streaming" @click="onNewSession">
        {{ t('odinAssistant.newSession') }}
      </el-button>
    </div>

    <div ref="listRef" class="copilot-chat__list">
      <el-empty
        v-if="!assistantStore.messages.length"
        :description="t('odinAssistant.copilotEmpty')"
        :image-size="72"
      />
      <div
        v-for="m in assistantStore.messages"
        :key="m.id"
        :class="['copilot-msg', `copilot-msg--${m.role}`]"
      >
        <div class="copilot-msg__bubble">
          <template v-if="m.role === 'user'">{{ m.content }}</template>
          <template v-else>
            <BlockRenderer
              v-for="(b, i) in m.blocks"
              :key="i"
              :block="b"
              :datasets="m.datasets"
              :chart-hints="m.chartHints"
              :streaming="m.streaming"
            />
            <span v-if="m.streaming && !m.blocks.length" class="copilot-msg__typing">…</span>
            <el-alert v-if="m.error" type="error" :closable="false" :title="m.error" />
          </template>
        </div>
      </div>
    </div>

    <div class="copilot-chat__input">
      <div class="cp-composer" :class="{ 'is-focused': inputFocused }">
        <el-input
          ref="textareaRef"
          v-model="input"
          type="textarea"
          :rows="1"
          resize="none"
          :placeholder="t('odinAssistant.inputPlaceholder')"
          :disabled="assistantStore.streaming"
          @input="autoResize"
          @focus="inputFocused = true"
          @blur="inputFocused = false"
          @keydown="onKeydown"
        />
        <div class="cp-composer__toolbar">
          <span class="cp-composer__hint">{{ t('odinAssistant.enterHint') }}</span>
          <div class="cp-composer__actions">
            <el-button
              v-if="assistantStore.streaming"
              class="cp-composer__stop"
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
              class="cp-composer__send"
              :disabled="!input.trim()"
              :title="t('odinAssistant.send')"
              @click="send"
            >
              <el-icon><Position /></el-icon>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.copilot-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.copilot-chat__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--sm-border-lighter);
}

.copilot-chat__ontology {
  font-size: 12px;
  color: var(--sm-text-secondary, #909399);
}

.copilot-chat__list {
  flex: 1;
  overflow-y: auto;
  padding: 12px 0;
}

.copilot-msg {
  display: flex;
  margin-bottom: 12px;
}

.copilot-msg--user {
  justify-content: flex-end;
}

.copilot-msg__bubble {
  max-width: 86%;
  padding: 8px 12px;
  font-size: 13px;
  line-height: 1.6;
  background: var(--sm-bg-page);
  border-radius: var(--sm-radius-md);
}

.copilot-msg--user .copilot-msg__bubble {
  color: #fff;
  background: var(--sm-color-primary);
}

.copilot-msg__typing {
  color: var(--sm-text-secondary, #909399);
}

.copilot-chat__input {
  padding-top: 10px;
  border-top: 1px solid var(--sm-border-lighter);
}
.cp-composer {
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
.cp-composer.is-focused {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--el-color-primary) 12%, transparent);
}
.cp-composer :deep(.el-textarea__inner) {
  padding: 4px 2px;
  font-size: 13px;
  line-height: 1.6;
  background: transparent;
  border: none;
  box-shadow: none;
  resize: none;
}
.cp-composer :deep(.el-textarea__inner:focus) {
  box-shadow: none;
}
.cp-composer__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 2px;
}
.cp-composer__hint {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}
.cp-composer__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}
.cp-composer__send {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  font-size: 15px;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  border: none;
  border-radius: 50%;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--el-color-primary) 32%, transparent);
  transition: transform 0.15s ease, filter 0.15s ease, opacity 0.15s ease;
}
.cp-composer__send:hover:not(:disabled) {
  filter: brightness(1.06);
  transform: translateY(-1px);
}
.cp-composer__send:active:not(:disabled) {
  transform: translateY(0);
}
.cp-composer__send:disabled {
  cursor: not-allowed;
  opacity: 0.45;
  box-shadow: none;
}
.cp-composer__stop {
  width: 32px;
  height: 32px;
}
</style>
