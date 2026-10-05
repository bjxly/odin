<script setup lang="ts">
import { computed } from 'vue'
import { ElTag } from 'element-plus'
import type { BlockVO } from '@/types/assistant'

/**
 * 兜底渲染器：未注册类型的块以原始 JSON 展示，
 * 保证后端新增块类型时前端「不白屏」，同时提示需要注册对应渲染器。
 */
const props = defineProps<{ block: BlockVO }>()

const pretty = computed(() => {
  try {
    return JSON.stringify(props.block?.payload ?? null, null, 2)
  } catch {
    return String(props.block?.payload ?? '')
  }
})
</script>

<template>
  <div class="unknown-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || '未识别的渲染块' }}</span>
      <ElTag size="small" type="warning" effect="plain">type={{ block.type || 'unknown' }}</ElTag>
      <ElTag v-if="block.trace_ref" size="small" type="info" effect="plain">trace</ElTag>
    </div>
    <pre class="raw"><code>{{ pretty }}</code></pre>
    <div class="hint">该块类型尚未注册渲染器，可通过 registerBlockRenderer('{{ block.type }}', Component) 扩展。</div>
  </div>
</template>

<style scoped>
.unknown-block {
  border: 1px dashed #cbd5e0;
  border-radius: 10px;
  background: #f8fafc;
  padding: 10px 12px;
}

.block-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}

.block-title {
  font-size: 13px;
  font-weight: 700;
  color: #4a5568;
}

.raw {
  margin: 0;
  padding: 8px 10px;
  background: #edf2f7;
  border-radius: 6px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
  line-height: 1.6;
  color: #2d3748;
  max-height: 220px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.hint {
  margin-top: 6px;
  font-size: 11px;
  color: #a0aec0;
}
</style>
