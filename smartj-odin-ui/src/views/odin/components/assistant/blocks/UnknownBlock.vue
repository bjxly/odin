<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MessageBlock } from '@/types/odin'

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()

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
      <span class="block-title">{{ block.title || t('odinAssistant.blockUnknown') }}</span>
      <el-tag size="small" type="warning" effect="plain">type={{ block.type || 'unknown' }}</el-tag>
      <el-tag v-if="block.traceRef" size="small" type="info" effect="plain">trace</el-tag>
    </div>
    <pre class="raw"><code>{{ pretty }}</code></pre>
    <div class="hint">registerBlock('{{ block.type }}', Component)</div>
  </div>
</template>

<style scoped>
.unknown-block {
  border: 1px dashed var(--el-border-color);
  border-radius: 10px;
  background: var(--el-fill-color-lighter);
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
  color: var(--el-text-color-secondary);
}
.raw {
  margin: 0;
  padding: 8px 10px;
  background: var(--el-fill-color);
  border-radius: 6px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
  max-height: 220px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
.hint {
  margin-top: 6px;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}
</style>
