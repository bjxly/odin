<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Connection, Grid, Timer, DataLine, Share } from '@element-plus/icons-vue'
import { useOdinTraceStore } from '@/stores/odin/trace'
import type { MessageBlock, ProvenanceBlockPayload } from '@/types/odin'

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()
const traceStore = useOdinTraceStore()

const p = computed<ProvenanceBlockPayload>(() => (props.block?.payload ?? {}) as ProvenanceBlockPayload)
const traceId = computed(() => String(p.value.traceId || props.block?.traceRef || ''))

const datasourceText = computed(() => String(p.value.datasourceName || '').trim() || '—')
const tableText = computed(() => String(p.value.sourceTable || '').trim() || '—')
const classText = computed(() => String(p.value.className || '').trim() || '—')

const rowCountText = computed(() => {
  if (String(p.value.status ?? '') === 'dry_run') return t('odinAssistant.provDryRun')
  return `${p.value.rowCount ?? 0} 行`
})

const durationText = computed(() => {
  const ms = p.value.durationMs
  if (ms == null) return '—'
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`
})

const statusType = computed<'success' | 'danger' | 'warning' | 'info'>(() => {
  const s = String(p.value.status ?? '')
  if (s === 'success') return 'success'
  if (s === 'failed') return 'danger'
  if (s === 'dry_run' || s === 'parse_only') return 'warning'
  return 'info'
})

function openTrace() {
  if (traceId.value) void traceStore.openTrace(traceId.value)
}
</script>

<template>
  <div class="provenance-card" :class="{ clickable: !!traceId }" @click="openTrace">
    <div class="card-head">
      <span class="card-title">{{ block.title || t('odinAssistant.blockProvenance') }}</span>
      <el-tag v-if="p.status" size="small" :type="statusType" effect="plain">{{ p.status }}</el-tag>
      <el-button
        v-if="traceId"
        class="open-btn"
        size="small"
        text
        type="primary"
        :icon="Share"
        @click.stop="openTrace"
      >
        {{ t('odinAssistant.traceLink') }}
      </el-button>
    </div>

    <div class="card-grid">
      <div class="item">
        <span class="k"><el-icon><Connection /></el-icon>{{ t('odinAssistant.provDatasource') }}</span>
        <span class="v">{{ datasourceText }}</span>
      </div>
      <div class="item">
        <span class="k"><el-icon><Grid /></el-icon>{{ t('odinAssistant.provTable') }}</span>
        <span class="v mono">{{ tableText }}</span>
      </div>
      <div class="item">
        <span class="k"><el-icon><DataLine /></el-icon>{{ t('odinAssistant.provClass') }}</span>
        <span class="v mono">{{ classText }}</span>
      </div>
      <div class="item">
        <span class="k"><el-icon><Timer /></el-icon>{{ t('odinAssistant.provRowCount') }}</span>
        <span class="v mono">{{ rowCountText }} · {{ durationText }}</span>
      </div>
      <div v-if="traceId" class="item wide">
        <span class="k">trace_id</span>
        <span class="v mono trace">{{ traceId.length > 20 ? traceId.slice(0, 12) + '…' + traceId.slice(-8) : traceId }}</span>
      </div>
    </div>

    <div v-if="p.errorMessage" class="err">{{ p.errorMessage }}</div>
  </div>
</template>

<style scoped>
.provenance-card {
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 10px;
  background: linear-gradient(135deg, var(--el-color-primary-light-9) 0%, var(--el-fill-color-blank) 60%);
  padding: 10px 12px;
}
.provenance-card.clickable {
  cursor: pointer;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}
.provenance-card.clickable:hover {
  border-color: var(--el-color-primary-light-3);
  box-shadow: 0 4px 14px rgba(64, 158, 255, 0.15);
}
.card-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}
.card-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.open-btn {
  margin-left: auto;
}
.card-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px 14px;
}
.item {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 12px;
  min-width: 0;
}
.item.wide {
  grid-column: 1 / -1;
}
.item .k {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  min-width: 78px;
}
.item .v {
  color: var(--el-text-color-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.mono {
  font-family: 'JetBrains Mono', Consolas, monospace;
}
.trace {
  color: var(--el-color-primary);
}
.err {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed var(--el-color-danger-light-5);
  font-size: 12px;
  color: var(--el-color-danger);
}
</style>
