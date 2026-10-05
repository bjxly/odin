<script setup lang="ts">
import { computed } from 'vue'
import ReasonTraceView from '@/components/reason/ReasonTraceView.vue'
import type { ReasonTraceVO } from '@/types/query'
import type { BlockVO, ReasoningPayload } from '@/types/assistant'

/** reasoning 块：把技能产出的 reason_trace 交给 ReasonTraceView 渲染 */
const props = defineProps<{ block: BlockVO }>()

const trace = computed<ReasonTraceVO | null>(() => {
  const p = props.block?.payload as ReasoningPayload | ReasonTraceVO | null | undefined
  if (!p || typeof p !== 'object') return null
  // 兼容 { reason_trace: {...} } 与直接就是 ReasonTrace 两种形态
  return ((p as ReasoningPayload).reason_trace ?? (p as ReasonTraceVO)) as ReasonTraceVO
})
</script>

<template>
  <div class="reasoning-block">
    <div class="block-title">{{ block.title || '推理轨迹' }}</div>
    <ReasonTraceView :trace="trace" compact />
  </div>
</template>

<style scoped>
.reasoning-block {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
  padding: 10px 12px;
}

.block-title {
  font-size: 13px;
  font-weight: 700;
  color: #1a365d;
  margin-bottom: 8px;
}
</style>
