<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MagicStick } from '@element-plus/icons-vue'
import hljs from 'highlight.js/lib/common'
import { renderMarkdown } from '@/utils/markdown'
import type { MessageBlock } from '@/types/odin'

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()

const bodyRef = ref<HTMLElement | null>(null)

const text = computed<string>(() => {
  const p = props.block?.payload as Record<string, unknown> | string | null | undefined
  if (typeof p === 'string') return p.trim()
  return String((p as any)?.content ?? (p as any)?.text ?? '').trim()
})

// 流式生成中（驱动打字机光标）：streaming 由 store 写入 payload，结束时置 false
const streaming = computed<boolean>(() => {
  const p = props.block?.payload as Record<string, unknown> | undefined
  return !!p && p.streaming === true
})

const html = computed<string>(() => renderMarkdown(text.value))

// 代码高亮后处理（marked v12 无内置 highlight）：渲染后对未标注过的 code 逐个高亮
watch(
  html,
  () => {
    nextTick(() => {
      const root = bodyRef.value
      if (!root) return
      for (const code of root.querySelectorAll('pre code')) {
        const el = code as HTMLElement
        if (el.dataset.highlighted !== 'yes') {
          try {
            hljs.highlightElement(el)
          } catch {
            /* 语言不可识别等，忽略 */
          }
        }
      }
    })
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  bodyRef.value = null
})
</script>

<template>
  <div v-if="text" class="text-block">
    <div class="text-block__head">
      <el-icon class="text-block__icon"><MagicStick /></el-icon>
      <span class="text-block__label">{{ block.title || t('odinAssistant.blockText') }}</span>
    </div>
    <!-- eslint-disable-next-line vue/no-v-html -->
    <div ref="bodyRef" class="text-block__body chat-markdown" v-html="html"></div>
    <span v-if="streaming" class="typing-cursor" aria-hidden="true"></span>
  </div>
</template>

<style scoped>
.text-block {
  border: 1px solid var(--el-border-color-lighter);
  border-left: 3px solid var(--el-color-primary);
  border-radius: 8px;
  padding: 10px 14px;
  background: var(--el-fill-color-blank);
}
.text-block__head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}
.text-block__icon {
  color: var(--el-color-primary);
  font-size: 15px;
}
.text-block__label {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-color-primary);
}
.text-block__body {
  font-size: 14px;
  line-height: 1.75;
  color: var(--el-text-color-primary);
  word-break: break-word;
}

/* Markdown 正文样式 */
.chat-markdown :deep(p) {
  margin: 0 0 8px;
}
.chat-markdown :deep(p:last-child) {
  margin-bottom: 0;
}
.chat-markdown :deep(h1),
.chat-markdown :deep(h2),
.chat-markdown :deep(h3),
.chat-markdown :deep(h4) {
  margin: 12px 0 8px;
  font-weight: 600;
}
.chat-markdown :deep(ul),
.chat-markdown :deep(ol) {
  padding-left: 20px;
  margin: 4px 0;
}
.chat-markdown :deep(li) {
  margin: 2px 0;
}
.chat-markdown :deep(strong) {
  font-weight: 700;
}
.chat-markdown :deep(code) {
  background: var(--el-fill-color-light);
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
.chat-markdown :deep(pre) {
  background: #282c34;
  color: #abb2bf;
  padding: 12px 14px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}
.chat-markdown :deep(pre code) {
  background: transparent;
  padding: 0;
  font-size: 13px;
  line-height: 1.6;
}
.chat-markdown :deep(blockquote) {
  border-left: 3px solid var(--el-border-color);
  padding-left: 12px;
  margin: 8px 0;
  color: var(--el-text-color-secondary);
}
.chat-markdown :deep(table) {
  border-collapse: collapse;
  margin: 8px 0;
  display: block;
  max-width: 100%;
  overflow-x: auto;
}
.chat-markdown :deep(th),
.chat-markdown :deep(td) {
  border: 1px solid var(--el-border-color-lighter);
  padding: 6px 10px;
  font-size: 13px;
}
.chat-markdown :deep(th) {
  background: var(--el-fill-color-light);
  font-weight: 600;
}
.chat-markdown :deep(a) {
  color: var(--el-color-primary);
}
.chat-markdown :deep(hr) {
  border: none;
  border-top: 1px solid var(--el-border-color-lighter);
  margin: 12px 0;
}

/* 打字机光标 */
.typing-cursor {
  display: inline-block;
  width: 0;
  margin-left: 1px;
  vertical-align: text-bottom;
}
.typing-cursor::after {
  content: '▍';
  color: var(--el-color-primary);
  animation: typing-blink 1s step-end infinite;
}
@keyframes typing-blink {
  50% {
    opacity: 0;
  }
}
</style>
