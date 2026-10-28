<script setup lang="ts">
import { computed } from 'vue'
import { ElIcon, ElTag, ElTooltip } from 'element-plus'
import { Connection } from '@element-plus/icons-vue'
import type { MessageMeta } from '@/types/assistant'

/**
 * 消息级 meta 摘要条（#56 P1-1 / P3-5）。
 *
 * P1-1：从 meta 读取 hop_count / llm_calls / sub_intent_count，渲染「N 跳 · N 轮 LLM · N 子意图」
 *        摘要徽章，让用户感知遍历规模（替代被移除的 thinking spam）。
 *        ReAct 路径的「N 轮」仍由 ThinkingBlock 承载，本组件不重复。
 * P3-5：从 meta.cross_source / sources 渲染确定性「综合 N 个数据源：…」横幅，
 *        不依赖结论自由文本措辞。
 */
const props = defineProps<{ meta?: MessageMeta | null }>()

const m = computed<MessageMeta>(() => props.meta ?? {})

const hopCount = computed(() => Math.max(0, Number(m.value.hop_count) || 0))
const llmCalls = computed(() => Math.max(0, Number(m.value.llm_calls) || 0))
const subIntentCount = computed(() => Math.max(0, Number(m.value.sub_intent_count) || 0))

const sources = computed<string[]>(() =>
  (Array.isArray(m.value.sources) ? m.value.sources : []).map((s) => String(s)).filter(Boolean),
)
const crossSource = computed(() => !!m.value.cross_source || sources.value.length > 1)

const hasBadges = computed(
  () => hopCount.value > 0 || llmCalls.value > 0 || subIntentCount.value > 0 || crossSource.value,
)
const hasBanner = computed(() => crossSource.value && sources.value.length > 0)
</script>

<template>
  <div v-if="hasBadges || hasBanner" class="meta-bar">
    <div v-if="hasBadges" class="badges">
      <ElTooltip v-if="hopCount > 0" content="本次遍历的跳数" placement="top" :show-after="200">
        <ElTag size="small" type="primary" effect="plain">{{ hopCount }} 跳</ElTag>
      </ElTooltip>
      <ElTooltip v-if="llmCalls > 0" content="本次调用的大模型轮数" placement="top" :show-after="200">
        <ElTag size="small" type="warning" effect="plain">{{ llmCalls }} 轮 LLM</ElTag>
      </ElTooltip>
      <ElTooltip v-if="subIntentCount > 0" content="拆解出的子意图数量" placement="top" :show-after="200">
        <ElTag size="small" type="info" effect="plain">{{ subIntentCount }} 子意图</ElTag>
      </ElTooltip>
      <ElTag v-if="crossSource" size="small" type="success" effect="dark">跨源综合</ElTag>
    </div>

    <div v-if="hasBanner" class="source-banner">
      <ElIcon><Connection /></ElIcon>
      <span>综合 {{ sources.length }} 个数据源：{{ sources.join('、') }}</span>
    </div>
  </div>
</template>

<style scoped>
.meta-bar {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.badges {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.source-banner {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-basis: 100%;
  padding: 6px 10px;
  border-radius: 8px;
  border: 1px solid #9ae6b4;
  background: linear-gradient(135deg, #f0fff4, #ffffff);
  color: #276749;
  font-size: 12px;
  line-height: 1.5;
  word-break: break-word;
}
</style>
