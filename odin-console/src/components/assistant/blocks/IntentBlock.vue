<script setup lang="ts">
import { computed } from 'vue'
import { ElTag } from 'element-plus'
import { FILTER_OPS } from '@/types/query'
import type { AggregateVO, HavingVO, QueryRequestVO } from '@/types/query'
import type { BlockVO, IntentPayload } from '@/types/assistant'
import { aggText, havingText } from '@/utils/format'

/**
 * intent 块（IntentPreview）：展示后端解析出的查询意图 AST。
 * 结构：主类 / 投影属性（聚合时为分组维度）/ 度量 / HAVING / 过滤条件 / 展开关系 / 排序分页。
 */
const props = defineProps<{ block: BlockVO }>()

const OP_LABELS: Record<string, string> = Object.fromEntries(FILTER_OPS.map((o) => [o.value, o.label]))

const intent = computed<QueryRequestVO | null>(() => {
  const p = props.block?.payload as IntentPayload | QueryRequestVO | null | undefined
  if (!p || typeof p !== 'object') return null
  // 兼容两种形态：{ intent: {...} } 与直接就是 intent
  return ((p as IntentPayload).intent ?? (p as QueryRequestVO)) as QueryRequestVO
})

const className = computed(() => intent.value?.class_name || '（未识别）')
const properties = computed<string[]>(() => intent.value?.properties ?? [])
const filters = computed(() => intent.value?.filters ?? [])
const relations = computed<string[]>(() => intent.value?.relations ?? [])
const limit = computed(() => Number(intent.value?.limit ?? 0))
const offset = computed(() => Number(intent.value?.offset ?? 0))
const orderBy = computed(() => {
  const ob = intent.value?.order_by
  if (!ob) return ''
  const dir = String(intent.value?.order_dir ?? '').toLowerCase() === 'desc' ? 'DESC' : 'ASC'
  return `${ob} ${dir}`
})

// --- 聚合（Phase 1）：group_by / aggregates / having ---
const aggregates = computed<AggregateVO[]>(() => intent.value?.aggregates ?? [])
const groupBy = computed<string[]>(() => intent.value?.group_by ?? [])
const havings = computed<HavingVO[]>(() => intent.value?.having ?? [])
/** 有度量即视为聚合查询（与后端判定同源） */
const isAgg = computed(() => aggregates.value.length > 0)
/**
 * 聚合时的维度行：group_by 为空表示纯标量聚合（全局一行）；
 * 此时不回退到 properties，因为 ONLY_FULL_GROUP_BY 下 SELECT 严格 = 分组键 ∪ 聚合列。
 */
const projectionLabel = computed(() => (isAgg.value ? '维度' : '投影'))
const projectionItems = computed<string[]>(() =>
  isAgg.value ? groupBy.value : properties.value,
)

function opLabel(op?: string): string {
  const o = String(op ?? '')
  return OP_LABELS[o] ?? o
}

function valueText(v: unknown): string {
  if (v == null) return 'NULL'
  if (Array.isArray(v)) return `[${v.map((x) => String(x)).join(', ')}]`
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}
</script>

<template>
  <div class="intent-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || '查询意图' }}</span>
      <ElTag size="small" type="primary" effect="plain">AST</ElTag>
    </div>

    <div v-if="!intent" class="empty">本次未产出结构化意图</div>

    <div v-else class="intent-grid">
      <div class="row">
        <span class="k">主类</span>
        <ElTag size="small" type="success" effect="dark">{{ className }}</ElTag>
        <ElTag v-if="intent.ontology_id" size="small" type="info" effect="plain">
          本体 #{{ intent.ontology_id }}
        </ElTag>
      </div>

      <div class="row">
        <span class="k">{{ projectionLabel }}</span>
        <template v-if="projectionItems.length">
          <ElTag v-for="p in projectionItems" :key="p" size="small" effect="plain">{{ p }}</ElTag>
        </template>
        <span v-else-if="isAgg" class="muted">无分组（全局聚合）</span>
        <span v-else class="muted">全部属性</span>
      </div>

      <div v-if="isAgg" class="row">
        <span class="k">度量</span>
        <ElTag
          v-for="(a, i) in aggregates"
          :key="i"
          size="small"
          type="danger"
          effect="plain"
          class="filter-tag"
        >
          {{ aggText(a) }}
        </ElTag>
      </div>

      <div v-if="havings.length" class="row">
        <span class="k">HAVING</span>
        <ElTag
          v-for="(h, i) in havings"
          :key="i"
          size="small"
          type="warning"
          effect="plain"
          class="filter-tag"
        >
          {{ havingText(h) }}
        </ElTag>
      </div>

      <div class="row">
        <span class="k">过滤</span>
        <template v-if="filters.length">
          <ElTag
            v-for="(f, i) in filters"
            :key="i"
            size="small"
            type="warning"
            effect="plain"
            class="filter-tag"
          >
            {{ f.property }} {{ opLabel(f.op) }} {{ valueText(f.value) }}
          </ElTag>
        </template>
        <span v-else class="muted">无</span>
      </div>

      <div class="row">
        <span class="k">关系</span>
        <template v-if="relations.length">
          <ElTag v-for="r in relations" :key="r" size="small" type="primary" effect="plain">{{ r }}</ElTag>
        </template>
        <span v-else class="muted">无展开</span>
      </div>

      <div class="row">
        <span class="k">排序分页</span>
        <span class="v">{{ orderBy || '默认排序' }}</span>
        <span class="v">
          {{ isAgg && orderBy && limit > 0 ? `TOP-${limit}` : `LIMIT ${limit > 0 ? limit : '未限制'}` }}
        </span>
        <span v-if="offset > 0" class="v">OFFSET {{ offset }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.intent-block {
  border: 1px solid #bee3f8;
  border-radius: 10px;
  background: linear-gradient(180deg, #f0f7ff, #ffffff);
  padding: 10px 12px;
}

.block-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.block-title {
  font-size: 13px;
  font-weight: 700;
  color: #1a365d;
}

.intent-grid {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: 12px;
}

.k {
  min-width: 56px;
  color: #718096;
  font-weight: 600;
}

.v {
  color: #2d3748;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.muted {
  color: #a0aec0;
}

.filter-tag {
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.empty {
  font-size: 12px;
  color: #a0aec0;
}
</style>
