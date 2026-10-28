<script setup lang="ts">
import { computed } from 'vue'
import { MagicStick } from '@element-plus/icons-vue'
import type { BlockVO, TextPayload } from '@/types/assistant'

/**
 * text 块（#66 增强为结论展示）：自然语言结论卡片。
 *
 * - 顶部「AI 分析」图标前缀，区分于普通气泡文本；
 * - 支持基础 Markdown 渲染（**粗体** → <strong>，- / * 列表 → <ul><li>）；
 * - 先转义 HTML 再套用 Markdown，v-html 无 XSS 风险。
 */
const props = defineProps<{ block: BlockVO }>()

const text = computed<string>(() => {
  const p = props.block?.payload as (TextPayload & { content?: string }) | string | null | undefined
  if (typeof p === 'string') return p.trim()
  return String(p?.content ?? p?.text ?? '').trim()
})

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

/** 行内 Markdown：仅处理 **粗体**（输入已转义，安全） */
function inlineMd(s: string): string {
  return s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
}

/** 极简 Markdown → HTML：粗体 + 无序列表 + 换行，转义在前防注入 */
function renderMarkdown(src: string): string {
  const lines = escapeHtml(src).split('\n')
  const out: string[] = []
  let listOpen = false
  for (const raw of lines) {
    const m = /^\s*[-*]\s+(.*)$/.exec(raw)
    if (m) {
      if (!listOpen) {
        out.push('<ul class="md-list">')
        listOpen = true
      }
      out.push(`<li>${inlineMd(m[1])}</li>`)
      continue
    }
    if (listOpen) {
      out.push('</ul>')
      listOpen = false
    }
    out.push(inlineMd(raw))
    out.push('<br/>')
  }
  if (listOpen) out.push('</ul>')
  let html = out.join('')
  if (html.endsWith('<br/>')) html = html.slice(0, -'<br/>'.length)
  return html
}
</script>

<template>
  <div v-if="text" class="text-block">
    <div class="text-block__head">
      <el-icon class="text-block__icon"><MagicStick /></el-icon>
      <span class="text-block__label">{{ block.title || 'AI 分析' }}</span>
    </div>
    <!-- eslint-disable-next-line vue/no-v-html -->
    <div class="text-block__body" v-html="renderMarkdown(text)"></div>
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

.text-block__body :deep(strong) {
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.text-block__body :deep(.md-list) {
  margin: 4px 0;
  padding-left: 20px;
}

.text-block__body :deep(.md-list li) {
  margin: 2px 0;
}
</style>
