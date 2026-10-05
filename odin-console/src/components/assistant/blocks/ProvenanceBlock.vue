<script setup lang="ts">
import { computed } from 'vue'
import { ElButton, ElIcon, ElTag } from 'element-plus'
import { Share, Connection, Grid, Timer, DataLine } from '@element-plus/icons-vue'
import ParsePathBadge from '@/components/assistant/ParsePathBadge.vue'
import { formatDuration, normalizeCount, shortId } from '@/utils/format'
import type { BlockVO, ChatAction, ProvenancePayload } from '@/types/assistant'

/**
 * provenance 块（ProvenanceCard）：一行数据的「出身证明」。
 * 展示 trace_id、parse_path、数据源/表、行数与耗时；点击可打开全链路溯源抽屉。
 */
const props = defineProps<{ block: BlockVO }>()
const emit = defineEmits<{ action: [ChatAction] }>()

const p = computed<ProvenancePayload>(() => (props.block?.payload ?? {}) as ProvenancePayload)
const traceId = computed(() => String(p.value.trace_id || props.block?.trace_ref || ''))
const joins = computed(() => p.value.joins ?? [])

const datasourceText = computed(() => {
  const name = String(p.value.datasource_name || '').trim()
  const type = String(p.value.datasource_type || '').trim()
  const id = p.value.datasource_id
  if (name && type) return `${name}（${type}）`
  if (name) return name
  if (id) return `数据源 #${id}`
  return isAgentLevel.value ? '多源取证' : '—'
})

/** P3(#43)：agent 级溯源无单一归属（数据源/表/类均缺省） */
const isAgentLevel = computed(() => {
  const hasDs = !!(p.value.datasource_name || p.value.datasource_id)
  const hasTable = !!String(p.value.source_table || '').trim()
  const hasClass = !!String(p.value.class_name || '').trim()
  return !hasDs && !hasTable && !hasClass
})

/**
 * P2-2(#56)：统一块契约下 provenance payload 必含 row_count/duration_ms，
 * 旧字段 result_count/execution_time_ms 作为回退。dry_run / result_count<0 时
 * 显示「试运行」而非「— 行」；其余情况即使为 0 / 缺失也显示「0 行 · 0 ms」，
 * 彻底消除「— 行 · —」占位符（live 与 reload 一致）。
 */
const isDryRun = computed(
  () => String(p.value.status ?? '') === 'dry_run' || Number(p.value.result_count) < 0,
)

const rowCountText = computed(() => {
  if (isDryRun.value) return '试运行（不执行）'
  return `${normalizeCount(p.value.row_count ?? p.value.result_count)} 行`
})

const durationText = computed(() =>
  formatDuration(p.value.duration_ms ?? p.value.execution_time_ms),
)

const statusText = computed(() => {
  const s = String(p.value.status ?? '')
  if (!s) return ''
  const map: Record<string, string> = {
    success: '执行成功',
    failed: '执行失败',
    dry_run: '试算（未执行）',
    parse_only: '仅解析',
    translated: '已翻译',
  }
  return map[s] ?? s
})

const statusType = computed<'success' | 'danger' | 'warning' | 'info'>(() => {
  const s = String(p.value.status ?? '')
  if (s === 'success') return 'success'
  if (s === 'failed') return 'danger'
  if (s === 'dry_run' || s === 'parse_only') return 'warning'
  return 'info'
})

function openTrace() {
  if (!traceId.value) return
  emit('action', { type: 'open-trace', value: traceId.value, traceId: traceId.value })
}
</script>

<template>
  <div class="provenance-card" :class="{ clickable: !!traceId }" @click="openTrace">
    <div class="card-head">
      <span class="card-title">{{ block.title || '数据溯源' }}</span>
      <ParsePathBadge :path="p.parse_path" />
      <ElTag v-if="statusText" size="small" :type="statusType" effect="plain">{{ statusText }}</ElTag>
      <ElButton
        v-if="traceId"
        class="open-btn"
        size="small"
        text
        type="primary"
        :icon="Share"
        @click.stop="openTrace"
      >
        查看全链路
      </ElButton>
    </div>

    <div class="card-grid">
      <div class="item">
        <span class="k"><ElIcon><Connection /></ElIcon>数据源</span>
        <span class="v">{{ datasourceText }}</span>
      </div>
      <div class="item">
        <span class="k"><ElIcon><Grid /></ElIcon>物理表</span>
        <span class="v mono">{{ p.source_table || (isAgentLevel ? '多表联合' : '—') }}</span>
      </div>
      <div class="item">
        <span class="k"><ElIcon><DataLine /></ElIcon>本体类</span>
        <span class="v mono">{{ p.class_name || (isAgentLevel ? '多类协作' : '—') }}</span>
      </div>
      <div class="item">
        <span class="k"><ElIcon><Timer /></ElIcon>行数 / 耗时</span>
        <span class="v mono">
          {{ rowCountText }}<template v-if="!isDryRun"> · {{ durationText }}</template>
        </span>
      </div>
      <div v-if="joins.length" class="item wide">
        <span class="k">JOIN</span>
        <span class="v mono">{{ joins.map((j) => j.relation).join('、') }}（{{ joins.length }}）</span>
      </div>
      <div v-if="traceId" class="item wide">
        <span class="k">trace_id</span>
        <span class="v mono trace" :title="traceId">{{ shortId(traceId, 12, 8) }}</span>
      </div>
    </div>

    <div v-if="p.error_message" class="err">{{ p.error_message }}</div>
    <div v-else-if="p.explanation" class="exp">{{ p.explanation }}</div>
  </div>
</template>

<style scoped>
.provenance-card {
  border: 1px solid #bee3f8;
  border-radius: 10px;
  background: linear-gradient(135deg, #f0f7ff 0%, #ffffff 60%);
  padding: 10px 12px;
}

.provenance-card.clickable {
  cursor: pointer;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.provenance-card.clickable:hover {
  border-color: #4c8bf5;
  box-shadow: 0 4px 14px rgba(76, 139, 245, 0.18);
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
  color: #1a365d;
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
  color: #718096;
  flex-shrink: 0;
  min-width: 78px;
}

.item .v {
  color: #2d3748;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.trace {
  color: #2b6cb0;
}

.exp {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #cbd5e0;
  font-size: 12px;
  line-height: 1.6;
  color: #4a5568;
}

.err {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #feb2b2;
  font-size: 12px;
  color: #c53030;
}
</style>
