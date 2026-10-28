<template>
  <div class="sm-json-editor">
    <div class="sm-json-editor__toolbar">
      <el-tooltip
        content="格式化"
        placement="top"
      >
        <el-button
          size="small"
          text
          @click="format"
        >
          <el-icon><Document /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip
        content="压缩"
        placement="top"
      >
        <el-button
          size="small"
          text
          @click="compress"
        >
          <el-icon><Minus /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip
        content="复制"
        placement="top"
      >
        <el-button
          size="small"
          text
          @click="copy"
        >
          <el-icon><CopyDocument /></el-icon>
        </el-button>
      </el-tooltip>
      <span
        v-if="jsonError"
        class="sm-json-editor__status sm-json-editor__status--error"
      >
        <el-icon><WarningFilled /></el-icon>
        JSON 格式错误
      </span>
      <span
        v-else-if="modelValue"
        class="sm-json-editor__status sm-json-editor__status--ok"
      >
        <el-icon><CircleCheckFilled /></el-icon>
        格式正确
      </span>
    </div>
    <textarea
      ref="textareaRef"
      class="sm-json-editor__textarea"
      :class="{ 'is-error': jsonError }"
      :value="modelValue"
      :readonly="readonly"
      :style="{ height }"
      spellcheck="false"
      @input="onInput"
    />
    <div
      v-if="jsonError"
      class="sm-json-editor__error"
    >
      {{ jsonError }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Document, Minus, CopyDocument, WarningFilled, CircleCheckFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    readonly?: boolean
    height?: string
  }>(),
  {
    modelValue: '',
    readonly: false,
    height: '300px',
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

const textareaRef = ref<HTMLTextAreaElement>()

const jsonError = computed(() => {
  if (!props.modelValue || !props.modelValue.trim()) return ''
  try {
    JSON.parse(props.modelValue)
    return ''
  } catch (err: any) {
    return err.message || 'Invalid JSON'
  }
})

function onInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLTextAreaElement).value)
}

function format() {
  if (!props.modelValue || !props.modelValue.trim()) return
  try {
    const obj = JSON.parse(props.modelValue)
    emit('update:modelValue', JSON.stringify(obj, null, 2))
  } catch {
    ElMessage.warning('JSON 格式错误，无法格式化')
  }
}

function compress() {
  if (!props.modelValue || !props.modelValue.trim()) return
  try {
    const obj = JSON.parse(props.modelValue)
    emit('update:modelValue', JSON.stringify(obj))
  } catch {
    ElMessage.warning('JSON 格式错误，无法压缩')
  }
}

async function copy() {
  if (!props.modelValue) return
  try {
    await navigator.clipboard.writeText(props.modelValue)
    ElMessage.success('已复制到剪贴板')
  } catch {
    // fallback
    const textarea = document.createElement('textarea')
    textarea.value = props.modelValue
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    ElMessage.success('已复制到剪贴板')
  }
}
</script>

<style scoped>
.sm-json-editor {
  display: flex;
  flex-direction: column;
  gap: var(--sm-spacing-xs);
  width: 100%;
}

.sm-json-editor__toolbar {
  display: flex;
  align-items: center;
  gap: var(--sm-spacing-xs);
  padding: var(--sm-spacing-xs) var(--sm-spacing-sm);
  background: var(--sm-bg-header);
  border: 1px solid var(--sm-border-light);
  border-bottom: none;
  border-radius: var(--sm-radius-md) var(--sm-radius-md) 0 0;
}

.sm-json-editor__status {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: var(--sm-font-size-xs);
}

.sm-json-editor__status--error {
  color: var(--sm-color-danger);
}

.sm-json-editor__status--ok {
  color: var(--sm-color-success);
}

.sm-json-editor__textarea {
  width: 100%;
  padding: var(--sm-spacing-md);
  font-family: 'Cascadia Code', 'Fira Code', 'JetBrains Mono', Consolas, 'Courier New', monospace;
  font-size: var(--sm-font-size-sm);
  line-height: 1.6;
  color: #d4d4d4;
  background: #1e1e1e;
  border: 1px solid var(--sm-border-light);
  border-radius: 0 0 var(--sm-radius-md) var(--sm-radius-md);
  outline: none;
  resize: vertical;
  tab-size: 2;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.sm-json-editor__textarea:focus {
  border-color: var(--sm-color-primary);
}

.sm-json-editor__textarea.is-error {
  border-color: var(--sm-color-danger);
}

.sm-json-editor__textarea:read-only {
  opacity: 0.7;
  cursor: not-allowed;
}

.sm-json-editor__error {
  padding: var(--sm-spacing-xs) var(--sm-spacing-sm);
  font-size: var(--sm-font-size-xs);
  color: var(--sm-color-danger);
  background: #fef0f0;
  border-radius: var(--sm-radius-sm);
  word-break: break-all;
}
</style>
