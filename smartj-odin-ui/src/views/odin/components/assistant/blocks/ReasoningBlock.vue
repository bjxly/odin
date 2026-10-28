<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowRight } from '@element-plus/icons-vue'
import type { MessageBlock } from '@/types/odin'

interface ReasonStep {
  title?: string
  content?: string
  duration?: number | string
  [key: string]: unknown
}

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()

const payload = computed<Record<string, unknown> | null>(() => {
  const p = props.block?.payload as Record<string, unknown> | null | undefined
  return p && typeof p === 'object' ? p : null
})

// 流式思考块：streaming 时默认展开让用户看到进度，collapsed 到达后自动折叠（仍可手动展开回看）
const streaming = computed(() => payload.value?.streaming === true)
const collapsedFlag = computed(() => payload.value?.collapsed === true)
const expanded = ref(streaming.value)
watch(collapsedFlag, (v) => {
  if (v) expanded.value = false
})

// 流式聚合的纯文本思考增量
const liveText = computed(() => String(payload.value?.text ?? ''))

const steps = computed<ReasonStep[]>(() => {
  const p = payload.value
  if (!p) return []
  // 兼容多种形态
  if (Array.isArray((p as any).steps)) return (p as any).steps
  if (Array.isArray((p as any).hops)) return (p as any).hops
  if (Array.isArray(p)) return p as ReasonStep[]
  // 单步
  if ((p as any).content || (p as any).title) return [p as unknown as ReasonStep]
  return []
})

const summary = computed(() => {
  const p = payload.value
  return String((p as any)?.summary ?? (p as any)?.content ?? '')
})

// 是否有可展示的思考内容（流式文本或步骤或摘要）
const hasContent = computed(() => !!liveText.value || steps.value.length > 0 || !!summary.value)
</script>

<template>
  <div class="reasoning-block">
    <div class="block-header" @click="expanded = !expanded">
      <span class="block-title">{{ block.title || t('odinAssistant.blockReasoning') }}</span>
      <el-tag v-if="steps.length" size="small" type="info" effect="plain">{{ steps.length }} steps</el-tag>
      <span v-if="streaming" class="live-dot"></span>
      <el-icon class="arrow" :class="{ 'is-expanded': expanded }"><ArrowRight /></el-icon>
    </div>

    <div v-show="expanded" class="block-body">
      <div v-if="!hasContent" class="empty">—</div>
      <!-- 流式思考增量：纯文本实时展示 -->
      <div v-if="liveText" class="step-content live-text">{{ liveText }}</div>
      <div v-if="summary" class="summary">{{ summary }}</div>
      <div v-for="(step, i) in steps" :key="i" class="step">
        <div class="step-head">
          <span class="step-idx">{{ i + 1 }}</span>
          <span class="step-title">{{ step.title || `Step ${i + 1}` }}</span>
          <span v-if="step.duration" class="step-dur">{{ step.duration }}ms</span>
        </div>
        <div v-if="step.content" class="step-content">{{ step.content }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reasoning-block {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
  overflow: hidden;
}
.block-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  background: var(--el-fill-color-lighter);
  user-select: none;
}
.block-header:hover {
  background: var(--el-fill-color-light);
}
.block-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.arrow {
  margin-left: auto;
  transition: transform 0.2s;
}
.arrow.is-expanded {
  transform: rotate(90deg);
}
.block-body {
  padding: 10px 14px;
  border-top: 1px solid var(--el-border-color-lighter);
}
.summary {
  font-size: 13px;
  color: var(--el-text-color-regular);
  margin-bottom: 8px;
  line-height: 1.6;
}
.step {
  padding: 8px 0;
  border-bottom: 1px dashed var(--el-border-color-lighter);
}
.step:last-child {
  border-bottom: none;
}
.step-head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.step-idx {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 11px;
  flex-shrink: 0;
}
.step-title {
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.step-dur {
  margin-left: auto;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-family: 'JetBrains Mono', Consolas, monospace;
}
.step-content {
  margin-top: 4px;
  padding-left: 28px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  word-break: break-word;
}
.empty {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}
.live-text {
  padding-left: 0;
}
.live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-color-primary);
  animation: live-pulse 1s ease-in-out infinite;
}
@keyframes live-pulse {
  0%,
  100% {
    opacity: 0.3;
  }
  50% {
    opacity: 1;
  }
}
</style>
