<script setup lang="ts">
import { computed } from 'vue'
import { ElIcon, ElTag } from 'element-plus'
import {
  ChatDotRound,
  Aim,
  MagicStick,
  Document,
  DataAnalysis,
  Timer,
} from '@element-plus/icons-vue'
import ReasonTraceView from '@/components/reason/ReasonTraceView.vue'
import ParsePathBadge from '@/components/assistant/ParsePathBadge.vue'
import { aggText, formatCount, formatMs, havingText, opLabel, valueText } from '@/utils/format'
import type { AggregateVO, HavingVO, TraceVO, TraceNodeVO } from '@/types/query'

/**
 * 全链路溯源时间线：NL → 意图 → 推理 → SQL → 执行 → 节点耗时。
 * 由 TraceDrawer（抽屉）承载，ProvenanceCard 点击 trace_id 打开。
 */
const props = defineProps<{ trace: TraceVO | null }>()

const t = computed<TraceVO | null>(() => props.trace)

const intent = computed(() => t.value?.intent ?? t.value?.ontology_query ?? null)
const filters = computed(() => intent.value?.filters ?? [])
const properties = computed<string[]>(() => intent.value?.properties ?? [])
const relations = computed<string[]>(() => intent.value?.relations ?? [])
// --- 聚合（Phase 1）：与 IntentBlock 保持同一套展示语义 ---
const aggregates = computed<AggregateVO[]>(() => intent.value?.aggregates ?? [])
const groupBy = computed<string[]>(() => intent.value?.group_by ?? [])
const havings = computed<HavingVO[]>(() => intent.value?.having ?? [])
/** 有度量即为聚合查询；此时「投影」语义变为「维度」（SELECT 严格 = 分组键 ∪ 聚合列） */
const isAgg = computed(() => aggregates.value.length > 0)
const projectionLabel = computed(() => (isAgg.value ? '维度' : '投影'))
const projectionItems = computed<string[]>(() => (isAgg.value ? groupBy.value : properties.value))
const joins = computed(() => t.value?.joins ?? [])
const params = computed<unknown[]>(() => (Array.isArray(t.value?.params) ? (t.value!.params as unknown[]) : []))
const nodes = computed<TraceNodeVO[]>(() => t.value?.nodes ?? [])
const reasonTrace = computed(() => t.value?.reason_trace ?? null)

const hasReason = computed(
  () =>
    !!reasonTrace.value
    && ((reasonTrace.value.rules_fired?.length ?? 0) > 0
      || (reasonTrace.value.expansion?.length ?? 0) > 0
      || (reasonTrace.value.derived_filters?.length ?? 0) > 0
      || (reasonTrace.value.inferred_relations?.length ?? 0) > 0),
)

const statusText = computed(() => {
  const s = String(t.value?.status ?? '')
  const map: Record<string, string> = {
    success: '执行成功',
    failed: '执行失败',
    dry_run: '试算（未执行）',
    parse_only: '仅解析',
    translated: '已翻译',
  }
  return map[s] ?? (s || '未知')
})

const statusType = computed<'success' | 'danger' | 'warning' | 'info'>(() => {
  const s = String(t.value?.status ?? '')
  if (s === 'success') return 'success'
  if (s === 'failed') return 'danger'
  if (s === 'dry_run' || s === 'parse_only') return 'warning'
  return 'info'
})

/** 节点耗时条宽度：相对最大耗时归一化 */
const maxNodeMs = computed(() => {
  let max = 0
  for (const n of nodes.value) {
    const d = Number(n.duration_ms ?? 0)
    if (Number.isFinite(d) && d > max) max = d
  }
  return max
})

function nodeWidth(n: TraceNodeVO): string {
  const d = Number(n.duration_ms ?? 0)
  if (!maxNodeMs.value || !Number.isFinite(d)) return '2%'
  return `${Math.max(2, Math.min(100, (d / maxNodeMs.value) * 100)).toFixed(1)}%`
}

const totalNodeMs = computed(() =>
  nodes.value.reduce((sum, n) => sum + (Number(n.duration_ms) || 0), 0),
)

const llmText = computed(() => {
  const m = t.value?.llm
  if (!m) return ''
  const parts: string[] = []
  if (m.model) parts.push(m.model)
  if (m.total_tokens) parts.push(`${m.total_tokens} tokens`)
  if (m.latency_ms != null) parts.push(formatMs(m.latency_ms))
  if (m.repairs) parts.push(`修复 ${m.repairs} 次`)
  if (m.prompt_version) parts.push(`prompt ${m.prompt_version}`)
  return parts.join(' · ')
})

function fmtCreatedAt(v?: string): string {
  if (!v) return ''
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  return d.toLocaleString('zh-CN', { hour12: false })
}
</script>

<template>
  <div v-if="!t" class="trace-empty">暂无链路数据</div>

  <div v-else class="trace-timeline">
    <!-- 概要头 -->
    <div class="trace-head">
      <div class="head-line">
        <ParsePathBadge :path="t.parse_path" show-empty />
        <ElTag size="small" :type="statusType" effect="dark">{{ statusText }}</ElTag>
        <ElTag size="small" type="info" effect="plain">
          {{ formatCount(t.result_count) }} 行 · {{ formatMs(t.execution_time_ms) }}
        </ElTag>
        <span v-if="t.created_at" class="created">{{ fmtCreatedAt(t.created_at) }}</span>
      </div>
      <div class="trace-id">
        <span class="lbl">trace_id</span>
        <code>{{ t.trace_id }}</code>
      </div>
    </div>

    <!-- 1. 自然语言 -->
    <section class="stage">
      <div class="stage-head">
        <span class="dot"><ElIcon><ChatDotRound /></ElIcon></span>
        <h4>自然语言</h4>
      </div>
      <div class="stage-body">
        <p class="nl">{{ t.nl_text || '（无原始问句，可能由结构化查询直接发起）' }}</p>
      </div>
    </section>

    <!-- 2. 意图解析 -->
    <section class="stage">
      <div class="stage-head">
        <span class="dot"><ElIcon><Aim /></ElIcon></span>
        <h4>意图解析（AST）</h4>
        <ElTag v-if="intent" size="small" type="success" effect="plain">{{ intent.class_name || '未识别' }}</ElTag>
        <ElTag v-if="isAgg" size="small" type="warning" effect="plain">
          聚合 · {{ aggregates.length }} 度量 / {{ groupBy.length }} 维度
        </ElTag>
      </div>
      <div class="stage-body">
        <div v-if="!intent" class="muted">本次链路未记录意图 AST</div>
        <template v-else>
          <div class="kv"><span class="k">{{ projectionLabel }}</span>
            <span v-if="projectionItems.length" class="v">{{ projectionItems.join('、') }}</span>
            <span v-else-if="isAgg" class="muted">无分组（全局聚合）</span>
            <span v-else class="muted">全部属性</span>
          </div>
          <div v-if="isAgg" class="kv"><span class="k">度量</span>
            <span class="v mono">{{ aggregates.map((a) => aggText(a)).join(' , ') }}</span>
          </div>
          <div v-if="havings.length" class="kv"><span class="k">HAVING</span>
            <span class="v mono">{{ havings.map((h) => havingText(h)).join('  AND  ') }}</span>
          </div>
          <div class="kv"><span class="k">过滤</span>
            <span v-if="filters.length" class="v mono">
              {{ filters.map((f) => `${f.property} ${opLabel(f.op)} ${valueText(f.value)}`).join('  AND  ') }}
            </span>
            <span v-else class="muted">无</span>
          </div>
          <div class="kv"><span class="k">关系</span>
            <span v-if="relations.length" class="v">{{ relations.join('、') }}</span>
            <span v-else class="muted">无展开</span>
          </div>
          <div class="kv"><span class="k">排序分页</span>
            <span class="v mono">
              {{ intent.order_by ? `${intent.order_by} ${String(intent.order_dir || 'asc').toUpperCase()}` : '默认' }}
              <template v-if="isAgg && intent.order_by && intent.limit"> · TOP-{{ intent.limit }}</template>
              <template v-else> · LIMIT {{ intent.limit || '未限制' }}</template>
              <template v-if="intent.offset"> · OFFSET {{ intent.offset }}</template>
            </span>
          </div>
          <div v-if="llmText" class="kv"><span class="k">LLM 审计</span><span class="v mono">{{ llmText }}</span></div>
        </template>
      </div>
    </section>

    <!-- 3. 本体推理 -->
    <section class="stage">
      <div class="stage-head">
        <span class="dot"><ElIcon><MagicStick /></ElIcon></span>
        <h4>本体推理</h4>
        <ElTag size="small" :type="hasReason ? 'success' : 'info'" effect="plain">
          {{ hasReason ? '有规则命中/展开' : '无推理动作' }}
        </ElTag>
      </div>
      <div class="stage-body">
        <ReasonTraceView :trace="reasonTrace" />
        <div v-if="t.explanation" class="explain">{{ t.explanation }}</div>
      </div>
    </section>

    <!-- 4. SQL 翻译 -->
    <section class="stage">
      <div class="stage-head">
        <span class="dot"><ElIcon><Document /></ElIcon></span>
        <h4>SQL 翻译</h4>
        <ElTag v-if="t.source_table" size="small" type="primary" effect="plain">{{ t.source_table }}</ElTag>
        <ElTag v-if="t.datasource_id" size="small" effect="plain">数据源 #{{ t.datasource_id }}</ElTag>
      </div>
      <div class="stage-body">
        <pre v-if="t.translated_sql" class="sql"><code>{{ t.translated_sql }}</code></pre>
        <div v-else class="muted">本次链路未产出 SQL</div>
        <div v-if="joins.length" class="joins">
          <div v-for="(j, i) in joins" :key="i" class="join-item">
            <ElTag size="small" type="primary" effect="plain">{{ j.relation }}</ElTag>
            <code>JOIN {{ j.table_name }} ON {{ j.condition }}</code>
          </div>
        </div>
        <div v-if="params.length" class="kv">
          <span class="k">绑定参数</span><code class="v">{{ JSON.stringify(params) }}</code>
        </div>
      </div>
    </section>

    <!-- 5. 执行结果 -->
    <section class="stage">
      <div class="stage-head">
        <span class="dot"><ElIcon><DataAnalysis /></ElIcon></span>
        <h4>执行</h4>
        <ElTag size="small" :type="statusType" effect="plain">{{ statusText }}</ElTag>
      </div>
      <div class="stage-body">
        <div class="metrics">
          <div class="metric">
            <span class="mk">返回行数</span>
            <span class="mv">{{ formatCount(t.result_count) }}</span>
          </div>
          <div class="metric">
            <span class="mk">执行耗时</span>
            <span class="mv">{{ formatMs(t.execution_time_ms) }}</span>
          </div>
        </div>
        <div v-if="t.error_message" class="err">{{ t.error_message }}</div>
      </div>
    </section>

    <!-- 6. 节点耗时 -->
    <section v-if="nodes.length" class="stage">
      <div class="stage-head">
        <span class="dot"><ElIcon><Timer /></ElIcon></span>
        <h4>节点耗时</h4>
        <ElTag size="small" type="info" effect="plain">{{ nodes.length }} 节点 · 合计 {{ formatMs(totalNodeMs) }}</ElTag>
      </div>
      <div class="stage-body">
        <div v-for="(n, i) in nodes" :key="i" class="node-item" :class="{ failed: !!n.error }">
          <div class="node-line1">
            <span class="node-name">{{ n.name }}</span>
            <ElTag size="small" :type="n.source === 'eino' ? 'warning' : 'success'" effect="plain">
              {{ n.source === 'eino' ? 'Eino' : '管线' }}
            </ElTag>
            <ElTag v-if="n.component" size="small" effect="plain">{{ n.component }}</ElTag>
            <ElTag v-if="n.type" size="small" type="info" effect="plain">{{ n.type }}</ElTag>
            <span class="node-ms">{{ formatMs(n.duration_ms) }}</span>
          </div>
          <div class="node-bar-track">
            <div class="node-bar" :style="{ width: nodeWidth(n) }" />
          </div>
          <div v-if="n.input_summary" class="node-sum">
            <span class="lbl">in</span><span>{{ n.input_summary }}</span>
          </div>
          <div v-if="n.output_summary" class="node-sum">
            <span class="lbl">out</span><span>{{ n.output_summary }}</span>
          </div>
          <div v-if="n.total_tokens" class="node-sum">
            <span class="lbl">tokens</span>
            <span>
              {{ n.total_tokens }}（prompt {{ n.prompt_tokens ?? 0 }} / completion {{ n.completion_tokens ?? 0 }}）
            </span>
          </div>
          <div v-if="n.error" class="node-err">{{ n.error }}</div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.trace-timeline {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.trace-empty {
  padding: 40px 0;
  text-align: center;
  color: #a0aec0;
  font-size: 13px;
}

/* 概要头 */
.trace-head {
  border: 1px solid #bee3f8;
  border-radius: 10px;
  background: linear-gradient(135deg, #f0f7ff, #ffffff);
  padding: 10px 12px;
}

.head-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.created {
  margin-left: auto;
  font-size: 11px;
  color: #a0aec0;
}

.trace-id {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  font-size: 11px;
}

.trace-id .lbl {
  color: #718096;
  font-weight: 600;
}

.trace-id code {
  font-family: 'JetBrains Mono', Consolas, monospace;
  color: #2b6cb0;
  word-break: break-all;
}

/* 阶段 */
.stage {
  position: relative;
  padding-left: 30px;
}

.stage::before {
  content: '';
  position: absolute;
  left: 11px;
  top: 24px;
  bottom: -14px;
  width: 2px;
  background: linear-gradient(180deg, #4c8bf5, #bee3f8);
}

.stage:last-child::before {
  display: none;
}

.stage-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 6px;
}

.stage-head .dot {
  position: absolute;
  left: 0;
  top: 0;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4c8bf5, #2b6cb0);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  box-shadow: 0 2px 6px rgba(43, 108, 176, 0.3);
}

.stage-head h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #1a365d;
}

.stage-body {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.nl {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: #2d3748;
}

.kv {
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 12px;
  flex-wrap: wrap;
}

.kv .k {
  min-width: 62px;
  color: #718096;
  font-weight: 600;
  flex-shrink: 0;
}

.kv .v {
  color: #2d3748;
  word-break: break-word;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.muted {
  font-size: 12px;
  color: #a0aec0;
}

.explain {
  margin-top: 4px;
  padding-top: 8px;
  border-top: 1px dashed #e2e8f0;
  font-size: 12px;
  line-height: 1.7;
  color: #4a5568;
}

.sql {
  margin: 0;
  padding: 10px 12px;
  background: #1a202c;
  color: #e2e8f0;
  border-radius: 8px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 12px;
  line-height: 1.65;
  overflow-x: auto;
  white-space: pre;
}

.joins {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.join-item {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.join-item code,
.kv code {
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
  color: #2c5282;
  background: #edf2f7;
  border-radius: 4px;
  padding: 1px 5px;
  word-break: break-all;
}

.metrics {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.metric {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mk {
  font-size: 11px;
  color: #718096;
}

.mv {
  font-size: 18px;
  font-weight: 700;
  color: #2b6cb0;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.err {
  font-size: 12px;
  color: #c53030;
  background: #fff5f5;
  border-radius: 6px;
  padding: 6px 10px;
}

/* 节点耗时 */
.node-item {
  border-left: 3px solid #4c8bf5;
  background: #f7fbff;
  border-radius: 0 8px 8px 0;
  padding: 8px 10px;
}

.node-item.failed {
  border-left-color: #e53e3e;
  background: #fff5f5;
}

.node-line1 {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: 12px;
}

.node-name {
  font-weight: 700;
  color: #1a365d;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.node-ms {
  margin-left: auto;
  font-weight: 700;
  color: #2b6cb0;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.node-bar-track {
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  margin: 6px 0;
  overflow: hidden;
}

.node-bar {
  height: 100%;
  background: linear-gradient(90deg, #4c8bf5, #2b6cb0);
  border-radius: 3px;
}

.node-sum {
  display: flex;
  gap: 6px;
  font-size: 11px;
  color: #4a5568;
  line-height: 1.6;
}

.node-sum .lbl {
  color: #a0aec0;
  font-weight: 700;
  min-width: 42px;
  flex-shrink: 0;
}

.node-err {
  margin-top: 4px;
  font-size: 11px;
  color: #c53030;
  word-break: break-all;
}
</style>
