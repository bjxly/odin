<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Connection } from '@element-plus/icons-vue'

/**
 * 消息元信息条：hop_count / llm_calls / cross_source / sources
 */
const { t } = useI18n()

export interface MsgMeta {
  hopCount?: number
  llmCalls?: number
  crossSource?: boolean
  sources?: string[]
  subIntentCount?: number
}

const props = defineProps<{ meta?: MsgMeta | null }>()

const m = computed<MsgMeta>(() => props.meta ?? {})

const hopCount = computed(() => Math.max(0, Number(m.value.hopCount) || 0))
const llmCalls = computed(() => Math.max(0, Number(m.value.llmCalls) || 0))
const subIntentCount = computed(() => Math.max(0, Number(m.value.subIntentCount) || 0))
const sources = computed<string[]>(() => (Array.isArray(m.value.sources) ? m.value.sources : []).filter(Boolean))
const crossSource = computed(() => !!m.value.crossSource || sources.value.length > 1)

const hasBadges = computed(
  () => hopCount.value > 0 || llmCalls.value > 0 || subIntentCount.value > 0 || crossSource.value,
)
const hasBanner = computed(() => crossSource.value && sources.value.length > 0)
</script>

<template>
  <div v-if="hasBadges || hasBanner" class="meta-bar">
    <div v-if="hasBadges" class="badges">
      <el-tooltip v-if="hopCount > 0" content="本次遍历的跳数" placement="top" :show-after="200">
        <el-tag size="small" type="primary" effect="plain">{{ t('odinAssistant.metaHops', { n: hopCount }) }}</el-tag>
      </el-tooltip>
      <el-tooltip v-if="llmCalls > 0" content="本次调用的大模型轮数" placement="top" :show-after="200">
        <el-tag size="small" type="warning" effect="plain">{{ t('odinAssistant.metaLlmCalls', { n: llmCalls }) }}</el-tag>
      </el-tooltip>
      <el-tooltip v-if="subIntentCount > 0" content="拆解出的子意图数量" placement="top" :show-after="200">
        <el-tag size="small" type="info" effect="plain">{{ t('odinAssistant.metaSubIntents', { n: subIntentCount }) }}</el-tag>
      </el-tooltip>
      <el-tag v-if="crossSource" size="small" type="success" effect="dark">{{ t('odinAssistant.metaCrossSource') }}</el-tag>
    </div>

    <div v-if="hasBanner" class="source-banner">
      <el-icon><Connection /></el-icon>
      <span>{{ t('odinAssistant.metaSources', { n: sources.length, list: sources.join('、') }) }}</span>
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
  padding: 6px 10px;
  border-radius: 8px;
  border: 1px solid var(--el-color-success-light-5);
  background: linear-gradient(135deg, var(--el-color-success-light-9), var(--el-fill-color-blank));
  color: var(--el-color-success-dark-2);
  font-size: 12px;
  line-height: 1.5;
  word-break: break-word;
}
</style>
