<template>
  <div
    class="sm-code-editor"
    :class="{ 'is-readonly': readonly }"
  >
    <div class="sm-code-editor__header">
      <span class="sm-code-editor__lang">{{ languageLabel }}</span>
      <span
        v-if="readonly"
        class="sm-code-editor__label"
      >
        只读
      </span>
    </div>
    <div
      class="sm-code-editor__body"
      :style="bodyStyle"
    >
      <div
        ref="lineNumRef"
        class="sm-code-editor__lines"
      >
        <div
          v-for="n in lineCount"
          :key="n"
          class="sm-code-editor__line-num"
        >
          {{ n }}
        </div>
      </div>
      <div class="sm-code-editor__input-area">
        <textarea
          ref="textareaRef"
          :value="modelValue"
          class="sm-code-editor__textarea"
          :readonly="readonly"
          :placeholder="placeholder"
          spellcheck="false"
          @input="onInput"
          @keydown="onKeydown"
          @scroll="onScroll"
        />
        <pre
          v-if="highlightedHtml"
          class="sm-code-editor__highlight"
          aria-hidden="true"
          v-html="highlightedHtml"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'
import 'highlight.js/styles/github-dark.css'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    language?: string
    height?: number | string
    placeholder?: string
    readonly?: boolean
  }>(),
  {
    modelValue: '',
    language: 'json',
    height: 400,
    placeholder: '',
    readonly: false,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

const textareaRef = ref<HTMLTextAreaElement>()
const lineNumRef = ref<HTMLElement>()
const lineCount = computed(() => {
  return (props.modelValue || '').split('\n').length
})

const languageLabel = computed(() => {
  const labels: Record<string, string> = {
    json: 'JSON',
    javascript: 'JavaScript',
    typescript: 'TypeScript',
    sql: 'SQL',
    python: 'Python',
    shell: 'Shell / Bash',
    yaml: 'YAML',
    xml: 'XML',
    html: 'HTML',
    css: 'CSS',
  }
  return labels[props.language] || props.language.toUpperCase()
})

const bodyStyle = computed(() => ({
  height: typeof props.height === 'number' ? `${props.height}px` : props.height,
}))

const highlightedHtml = computed(() => {
  if (!props.modelValue) return ''
  try {
    const lang = props.language === 'shell' ? 'bash' : props.language
    if (hljs.getLanguage(lang)) {
      const result = hljs.highlight(props.modelValue, { language: lang })
      return result.value
    }
    return escapeHtml(props.modelValue)
  } catch {
    return escapeHtml(props.modelValue)
  }
})

function escapeHtml(text: string): string {
  return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function onInput(e: Event) {
  const value = (e.target as HTMLTextAreaElement).value
  emit('update:modelValue', value)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Tab') {
    e.preventDefault()
    const textarea = textareaRef.value
    if (!textarea) return
    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const value = props.modelValue
    const newValue = value.substring(0, start) + '  ' + value.substring(end)
    emit('update:modelValue', newValue)
    nextTick(() => {
      textarea.selectionStart = textarea.selectionEnd = start + 2
    })
  }
}

function onScroll() {
  if (textareaRef.value && lineNumRef.value) {
    lineNumRef.value.scrollTop = textareaRef.value.scrollTop
  }
}

watch(
  () => props.modelValue,
  () => {
    // Keep highlight in sync
  }
)
</script>

<style>
/* highlight.js 主题适配：浅色模式默认 github.css，深色模式用 github-dark.css 覆盖 */
[data-theme='dark'] .sm-code-editor .hljs {
  background: transparent;
  color: #c9d1d9;
}
[data-theme='dark'] .sm-code-editor .hljs-doctag,
[data-theme='dark'] .sm-code-editor .hljs-keyword,
[data-theme='dark'] .sm-code-editor .hljs-meta .hljs-keyword,
[data-theme='dark'] .sm-code-editor .hljs-template-tag,
[data-theme='dark'] .sm-code-editor .hljs-template-variable,
[data-theme='dark'] .sm-code-editor .hljs-type,
[data-theme='dark'] .sm-code-editor .hljs-variable.language_ {
  color: #ff7b72;
}
[data-theme='dark'] .sm-code-editor .hljs-title,
[data-theme='dark'] .sm-code-editor .hljs-title.class_,
[data-theme='dark'] .sm-code-editor .hljs-title.class_.inherited__,
[data-theme='dark'] .sm-code-editor .hljs-title.function_ {
  color: #d2a8ff;
}
[data-theme='dark'] .sm-code-editor .hljs-attr,
[data-theme='dark'] .sm-code-editor .hljs-attribute,
[data-theme='dark'] .sm-code-editor .hljs-literal,
[data-theme='dark'] .sm-code-editor .hljs-meta,
[data-theme='dark'] .sm-code-editor .hljs-number,
[data-theme='dark'] .sm-code-editor .hljs-operator,
[data-theme='dark'] .sm-code-editor .hljs-selector-attr,
[data-theme='dark'] .sm-code-editor .hljs-selector-class,
[data-theme='dark'] .sm-code-editor .hljs-selector-id,
[data-theme='dark'] .sm-code-editor .hljs-variable {
  color: #79c0ff;
}
[data-theme='dark'] .sm-code-editor .hljs-meta .hljs-string,
[data-theme='dark'] .sm-code-editor .hljs-regexp,
[data-theme='dark'] .sm-code-editor .hljs-string {
  color: #a5d6ff;
}
[data-theme='dark'] .sm-code-editor .hljs-built_in,
[data-theme='dark'] .sm-code-editor .hljs-symbol {
  color: #ffa657;
}
[data-theme='dark'] .sm-code-editor .hljs-code,
[data-theme='dark'] .sm-code-editor .hljs-comment,
[data-theme='dark'] .sm-code-editor .hljs-formula {
  color: #8b949e;
}
[data-theme='dark'] .sm-code-editor .hljs-name,
[data-theme='dark'] .sm-code-editor .hljs-quote,
[data-theme='dark'] .sm-code-editor .hljs-selector-pseudo,
[data-theme='dark'] .sm-code-editor .hljs-selector-tag {
  color: #7ee787;
}
[data-theme='dark'] .sm-code-editor .hljs-subst {
  color: #c9d1d9;
}
[data-theme='dark'] .sm-code-editor .hljs-section {
  color: #1f6feb;
  font-weight: bold;
}
[data-theme='dark'] .sm-code-editor .hljs-bullet {
  color: #f2cc60;
}
[data-theme='dark'] .sm-code-editor .hljs-emphasis {
  color: #c9d1d9;
  font-style: italic;
}
[data-theme='dark'] .sm-code-editor .hljs-strong {
  color: #c9d1d9;
  font-weight: bold;
}
[data-theme='dark'] .sm-code-editor .hljs-addition {
  color: #aff5b4;
  background-color: #033a16;
}
[data-theme='dark'] .sm-code-editor .hljs-deletion {
  color: #ffdcd7;
  background-color: #67060c;
}
</style>

<style scoped>
.sm-code-editor {
  border: 1px solid var(--sm-code-border);
  border-radius: var(--sm-radius-md);
  overflow: hidden;
  background: var(--sm-code-bg);
  width: 100%;
}
.sm-code-editor__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--sm-code-header-bg);
  border-bottom: 1px solid var(--sm-code-border);
}
.sm-code-editor__lang {
  font-size: 12px;
  color: var(--sm-code-text-secondary);
  text-transform: uppercase;
}
.sm-code-editor__label {
  font-size: 11px;
  color: var(--sm-code-text-muted);
  background: var(--sm-code-badge-bg);
  padding: 1px 6px;
  border-radius: 3px;
}
.sm-code-editor__body {
  display: flex;
  overflow: hidden;
}
.sm-code-editor__lines {
  flex-shrink: 0;
  width: 44px;
  overflow: hidden;
  padding: 12px 0;
  text-align: right;
  background: var(--sm-code-bg);
  border-right: 1px solid var(--sm-code-border);
  user-select: none;
}
.sm-code-editor__line-num {
  padding-right: 10px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--sm-code-text-secondary);
  font-family: 'Consolas', 'Courier New', monospace;
}
.sm-code-editor__input-area {
  flex: 1;
  position: relative;
  overflow: hidden;
}
.sm-code-editor__textarea {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  padding: 12px 16px;
  border: none;
  outline: none;
  resize: none;
  font-size: 13px;
  line-height: 1.6;
  font-family: 'Consolas', 'Courier New', monospace;
  color: transparent;
  caret-color: var(--sm-code-caret);
  background: transparent;
  z-index: 2;
  overflow: auto;
  white-space: pre;
  tab-size: 2;
}
.sm-code-editor__textarea::placeholder {
  color: var(--sm-code-text-muted);
}
.sm-code-editor__highlight {
  margin: 0;
  padding: 12px 16px;
  font-size: 13px;
  line-height: 1.6;
  font-family: 'Consolas', 'Courier New', monospace;
  white-space: pre;
  overflow: auto;
  pointer-events: none;
  min-height: 100%;
  color: var(--sm-code-text);
}
.sm-code-editor__highlight :deep(.hljs) {
  background: transparent;
  padding: 0;
}
.sm-code-editor.is-readonly .sm-code-editor__textarea {
  color: var(--sm-code-readonly-text);
  caret-color: transparent;
}
</style>
