<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, watch } from 'vue'
import {
  ElAlert,
  ElButton,
  ElCard,
  ElDialog,
  ElEmpty,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElOption,
  ElPopconfirm,
  ElSelect,
  ElSwitch,
  ElTable,
  ElTableColumn,
  ElTabPane,
  ElTabs,
  ElTag,
} from 'element-plus'
import { Share } from '@element-plus/icons-vue'
import { useOntologyStore } from '@/stores/ontology'
import { useQueryStore } from '@/stores/query'
import { useTraceStore } from '@/stores/trace'
import {
  AGG_FUNCS,
  FILTER_OPS,
  ROW_SOURCE,
  defaultAggAlias,
  isNumericRange,
  type FilterOp,
  type LLMMetaVO,
  type ParsePath,
} from '@/types/query'
import { formatNumber } from '@/utils/format'
import type { BlockVO } from '@/types/assistant'

/** P1(#43)：聚合结果图表——复用 EChartsBlock，动态 import 不进首屏 */
const AggChart = defineAsyncComponent(() => import('@/components/assistant/blocks/EChartsBlock.vue'))

const onto = useOntologyStore()
const qs = useQueryStore()
/** P3：全链路溯源抽屉由 trace store 驱动（TraceDrawer 挂在 App.vue） */
const traceStore = useTraceStore()

/** NL 示例句由 store 根据当前本体真实类名生成 */
const nlExamples = computed(() => qs.nlExamples)

const resultTab = ref('sql')
const saveName = ref('')
const saveOpen = ref(false)
const pendingJoin = ref('')
const nlText = ref('')
const nlNotes = ref<string[]>([])
const nlHits = ref<string[]>([])
const nlParsing = ref(false)
const activePresetId = ref('')
/** 服务端解析元信息：解析路径标签 / 是否命中 / LLM 审计 / 缓存键 / 链路 id */
const nlParsePath = ref<ParsePath | ''>('')
const nlMatched = ref(false)
const nlLLMMeta = ref<LLMMetaVO | null>(null)
const nlCanonicalKey = ref('')
const nlTraceId = ref('')

/** 当前可用的链路 id：优先 Dry-run，其次执行结果，最后 NL 解析 */
const activeTraceId = computed(
  () => qs.dryRun?.traceId || qs.executeResult?.traceId || nlTraceId.value || '',
)

/** 打开 P3 全链路溯源抽屉（NL→intent→AST→SQL→行） */
function openTrace(traceId?: string) {
  const id = traceId || activeTraceId.value
  if (!id) {
    ElMessage.warning('本次请求未返回 trace_id，无法查看全链路')
    return
  }
  void traceStore.open(id)
}

const baseProps = computed(() => qs.resolvedBase?.properties ?? [])

const selectOptions = computed(() =>
  baseProps.value.map((p) => ({
    id: p.id,
    label: `${p.label}（${p.id}）`,
    isKey: !!p.isKey,
    sensitive: !!p.sensitive,
  })),
)

const orderOptions = computed(() => selectOptions.value)

// ---------------------------------------------------------------------------
// 聚合 / 分组（Phase 1）
// ---------------------------------------------------------------------------

/**
 * 三个聚合字段的读写代理。
 * 用 getter/setter 而非直接 `v-model="qs.current.groupBy"`：旧模板 / 旧本地缓存
 * 可能没有这三个字段（undefined），ElSelect multiple 绑 undefined 会直接报错。
 */
const groupByModel = computed<string[]>({
  get: () => qs.current.groupBy ?? [],
  set: (v) => { qs.current.groupBy = v },
})
const aggregates = computed(() => qs.current.aggregates ?? [])
const havings = computed(() => qs.current.having ?? [])
/** 有度量即为聚合查询（与后端判定同源） */
const isAggQuery = computed(() => aggregates.value.length > 0)

const joinRelationOptions = computed(() =>
  qs.availableRelations
    .filter((r) => !qs.current.joins.some((j) => j.relationId === r.id))
    .map((r) => ({
      id: r.id,
      label: `${r.label}（${r.domain} → ${r.range} · ${r.cardinality}${r.crossSource ? ' · 跨源' : ''}）`,
    })),
)

function joinSelectOptions(relationId: string) {
  const rel = onto.relations.find((r) => r.id === relationId)
  if (!rel) return []
  const other =
    rel.domain === qs.current.classId
      ? onto.classes.find((c) => c.id === rel.range)
      : onto.classes.find((c) => c.id === rel.domain)
  return (other?.properties ?? []).map((p) => ({
    id: p.id,
    label: `${p.label}（${p.id}）`,
  }))
}

function joinFilterProps(relationId: string) {
  const rel = onto.relations.find((r) => r.id === relationId)
  if (!rel) return []
  const other =
    rel.domain === qs.current.classId
      ? onto.classes.find((c) => c.id === rel.range)
      : onto.classes.find((c) => c.id === rel.domain)
  return other?.properties ?? []
}

function needValue(op: FilterOp) {
  return FILTER_OPS.find((o) => o.value === op)?.needValue ?? true
}

function needAggProp(func: string): boolean {
  return AGG_FUNCS.find((f) => f.value === func)?.needProperty ?? true
}

/** 分组维度候选：主类 + 已关联关系目标类的「非度量」属性（数值型不适合做分组键） */
const groupOptions = computed<Array<{ id: string; label: string }>>(() => {
  const out: Array<{ id: string; label: string }> = []
  for (const p of baseProps.value) {
    if (isNumericRange(p.range)) continue
    out.push({ id: p.id, label: `${p.label}（${p.id}）` })
  }
  for (const j of qs.current.joins) {
    const rel = onto.relations.find((r) => r.id === j.relationId)
    if (!rel) continue
    for (const p of joinFilterProps(j.relationId)) {
      if (isNumericRange(p.range)) continue
      // 后端 group_by 支持 `relationName.property` 限定，relationName 即前端 relationId
      out.push({ id: `${j.relationId}.${p.id}`, label: `${rel.label} · ${p.label}` })
    }
  }
  return out
})

/** 度量属性候选：数值型属性（SUM/AVG/MIN/MAX 只对数值有意义） */
const measurePropOptions = computed<Array<{ id: string; label: string }>>(() => {
  const nums = baseProps.value.filter((p) => isNumericRange(p.range))
  const list = nums.length ? nums : baseProps.value
  return list.map((p) => ({ id: p.id, label: `${p.label}（${p.id}）` }))
})

/** HAVING / 排序可选的度量别名（后端 HAVING 仅按别名寻址） */
const aggAliasOptions = computed<string[]>(() =>
  aggregates.value.map((a) => String(a.alias ?? '').trim()).filter(Boolean),
)

/** HAVING 只做数值比较，从 FILTER_OPS 裁剪（仍是后端 op 集合的子集，契约不变） */
const HAVING_OPS = FILTER_OPS.filter((o) =>
  ['eq', 'ne', 'gt', 'gte', 'lt', 'lte'].includes(o.value),
)

/** 已被用户手改过 alias 的度量行，不再自动覆盖 */
const aliasTouched = new WeakSet<object>()

function propLabelOf(id: string): string {
  const raw = String(id ?? '')
  const tail = raw.includes('.') ? raw.slice(raw.lastIndexOf('.') + 1) : raw
  return baseProps.value.find((x) => x.id === tail)?.label ?? tail
}

/** 自动补全度量别名：镜像后端 `defaultAggAlias`（前缀式「聚合函数中文名 + 属性中文 label」） */
function syncAggAlias(i: number) {
  const a = aggregates.value[i]
  if (!a || aliasTouched.has(a)) return
  a.alias = needAggProp(a.func) ? defaultAggAlias(a.func, propLabelOf(a.property)) : defaultAggAlias(a.func, '')
}

function onAliasInput(i: number) {
  const a = aggregates.value[i]
  if (a) aliasTouched.add(a)
}

/** 聚合模式下的排序候选：已选分组维度 + 度量别名（TOP-N 需能按度量排序） */
const aggOrderOptions = computed<Array<{ id: string; label: string }>>(() => {
  const out: Array<{ id: string; label: string }> = []
  for (const g of groupByModel.value) {
    const hit = groupOptions.value.find((o) => o.id === g)
    out.push({ id: g, label: hit ? hit.label : g })
  }
  for (const a of aggAliasOptions.value) out.push({ id: a, label: `${a}（度量）` })
  return out
})

/** 排序下拉候选：聚合查询切到「维度 + 度量别名」，否则仍是主类属性 */
const effectiveOrderOptions = computed(() =>
  isAggQuery.value ? aggOrderOptions.value : orderOptions.value,
)

/**
 * ElSelect 绑定对象时按引用比较，导致回显不同步（P2 #43）。
 * 改为字符串模型：`"propertyId::desc"` / `"propertyId::asc"`，ElSelect 按值比较即可正确回显。
 */
const orderByModel = computed<string>({
  get: () => {
    const ob = qs.current.orderBy
    if (!ob?.propertyId) return ''
    return `${ob.propertyId}::${ob.desc ? 'desc' : 'asc'}`
  },
  set: (v) => {
    if (!v) { qs.current.orderBy = null; return }
    const sep = v.lastIndexOf('::')
    if (sep < 0) { qs.current.orderBy = null; return }
    qs.current.orderBy = { propertyId: v.slice(0, sep), desc: v.slice(sep + 2) === 'desc' }
  },
})

/**
 * 聚合开关 / 分组维度变化后，旧排序字段可能既不在维度也不在度量别名内，
 * 后端会按普通属性下推到 ORDER BY 而生成非法 SQL（非分组列），故提前清空。
 */
watch(effectiveOrderOptions, (opts) => {
  const ob = qs.current.orderBy
  if (!ob?.propertyId) return
  if (!opts.some((o) => o.id === ob.propertyId)) qs.current.orderBy = null
})

/** 结果集中的度量列（优先 column_meta.role，退化按度量别名匹配） */
const measureCols = computed<Set<string>>(() => {
  const set = new Set<string>()
  for (const c of qs.executeResult?.columns ?? []) {
    const key = String(c.propertyId ?? '')
    if (c.role === 'measure') { set.add(key); continue }
    if (c.role === 'dimension') continue
    if (aggAliasOptions.value.includes(key) || aggAliasOptions.value.includes(String(c.label ?? ''))) {
      set.add(key)
    }
  }
  return set
})

const isAggResult = computed(() => !!qs.executeResult?.aggregated)

/** 结果计数文案：聚合结果是「N 个分组」，明细是「N 行」 */
const resultCountLabel = computed(() => {
  const n = qs.executeResult?.rowCount ?? 0
  return isAggResult.value ? `${n} 个分组` : `${n} 行`
})

/** 度量列走千分位格式化，维度列原样展示 */
function cellTextOf(key: string, v: unknown): string {
  if (measureCols.value.has(key)) return formatNumber(v)
  return v == null ? '' : String(v)
}

/**
 * P1(#43)：从执行结果构造 chart BlockVO，供 EChartsBlock 渲染。
 * 纯标量聚合（无维度 + 单度量）→ KPI 指标卡；分组聚合 → bar 图。
 * 列下标按 field 名反查，不用位置假设。
 */
const aggChartBlock = computed<BlockVO | null>(() => {
  const er = qs.executeResult
  if (!er?.aggregated || !er.columns.length || !er.rows.length) return null

  const dims = er.columns.filter((c) => c.role === 'dimension')
  const mets = er.columns.filter((c) => c.role === 'measure')
  if (!mets.length) return null

  // 列名数组（与 rows 的 key 对齐）
  const columns = er.columns.map((c) => c.propertyId)
  // QueryRow (Record) → unknown[][]（按 columns 顺序）
  const rows: unknown[][] = er.rows.map((row) => columns.map((col) => row[col] ?? null))

  const chartType = dims.length === 0 ? 'metric' : 'bar'

  return {
    type: 'chart',
    title: '聚合可视化',
    payload: {
      chart_type: chartType,
      dimensions: dims.map((c) => ({ field: c.propertyId, label: c.label })),
      metrics: mets.map((c) => ({ field: c.propertyId, label: c.label, agg: c.agg })),
      columns,
      rows,
      aggregated: true,
      top_n: er.truncated && qs.current.orderBy ? qs.current.limit : undefined,
    },
  }
})

function addJoinFromSelect() {
  if (!pendingJoin.value) return
  qs.addJoin(pendingJoin.value)
  pendingJoin.value = ''
}

const dryRunning = ref(false)

async function onDryRun() {
  if (!qs.current.classId) {
    ElMessage.warning('请先选择本体类')
    return
  }
  dryRunning.value = true
  try {
    await qs.dryRunQuery({
      ontology_query: qs.current,
      ontology_id: onto.currentOntologyId ?? undefined,
    })
    const r = qs.dryRun
    resultTab.value = r?.ok ? 'sql' : 'ast'
    if (r?.ok) ElMessage.success('Dry-run 完成，已生成物理 SQL')
    else ElMessage.error(r?.errors?.[0] ?? 'Dry-run 失败')
  } finally {
    dryRunning.value = false
  }
}

async function onExecute() {
  const r = await qs.runExecute()
  resultTab.value = 'result'
  if (r) ElMessage.success(`执行完成：${r.rowCount} 行 · ${r.elapsedMs}ms`)
  else ElMessage.error('执行失败，请先看 AST / SQL 诊断')
}

function onSaveQuery() {
  const item = qs.saveQuery(saveName.value)
  saveOpen.value = false
  saveName.value = ''
  ElMessage.success(`已保存查询「${item.name}」`)
}

function applyPreset(id: string) {
  const ok = qs.loadSavedQuery(id)
  if (!ok) {
    ElMessage.error('未找到该模板')
    return
  }
  activePresetId.value = id
  nlNotes.value = []
  nlHits.value = []
  nlParsePath.value = ''
  nlMatched.value = false
  nlLLMMeta.value = null
  nlCanonicalKey.value = ''
  nlTraceId.value = ''
  ElMessage.success('已载入查询模板，请查看下方构建器')
}

/** 解析路径标签：rule 零LLM / llm 意图链 / cache 同问同SQL / local 离线兜底 */
const PARSE_PATH_META: Record<string, { label: string; type: 'success' | 'primary' | 'warning' | 'info' }> = {
  rule: { label: '规则解析 · 零 LLM', type: 'success' },
  llm: { label: 'LLM 意图链', type: 'primary' },
  cache: { label: '计划缓存 · 同问同 SQL', type: 'warning' },
  local: { label: '离线兜底', type: 'info' },
}

const parsePathBadge = computed(() => {
  const p = String(nlParsePath.value || '')
  if (!p) return null
  return PARSE_PATH_META[p] ?? { label: p, type: 'info' as const }
})

const llmMetaText = computed(() => {
  const m = nlLLMMeta.value
  if (!m?.model) return ''
  const parts = [m.model]
  if (m.total_tokens != null) parts.push(`${m.total_tokens} tokens`)
  if (m.latency_ms != null) parts.push(`${m.latency_ms}ms`)
  if (m.repairs) parts.push(`修复 ${m.repairs} 次`)
  if (m.seed != null) parts.push(`seed=${m.seed}`)
  return parts.join(' · ')
})

async function runNLParse() {
  const text = nlText.value.trim()
  if (!text) {
    ElMessage.warning('请输入自然语言查询')
    return
  }
  nlParsing.value = true
  try {
    const parsed = await qs.parseNaturalLanguage(text)
    // 未命中本体类时不覆盖构建器，避免清空用户已选择的类
    if (parsed.query.classId) qs.applyQuery(parsed.query)
    nlNotes.value = parsed.notes
    nlHits.value = parsed.hits
    nlParsePath.value = parsed.parsePath
    nlMatched.value = parsed.matched
    nlLLMMeta.value = parsed.llmMeta ?? null
    nlCanonicalKey.value = parsed.canonicalKey ?? ''
    nlTraceId.value = parsed.traceId ?? ''
    activePresetId.value = ''
    if (parsed.matched) {
      ElMessage.success('已解析为本体查询，可继续 Dry-run / 执行')
      resultTab.value = 'ast'
    } else {
      ElMessage.warning('未命中本体类，请补充关键词或手动选择类')
    }
  } finally {
    nlParsing.value = false
  }
}

function fillExample(text: string) {
  nlText.value = text
  void runNLParse()
}

function clearNL() {
  nlText.value = ''
  nlNotes.value = []
  nlHits.value = []
  nlParsePath.value = ''
  nlMatched.value = false
  nlLLMMeta.value = null
  nlCanonicalKey.value = ''
  nlTraceId.value = ''
}

const astJson = computed(() =>
  JSON.stringify(qs.dryRun?.ast ?? qs.current, null, 2),
)

const reasonSteps = computed(() => qs.dryRun?.reasons ?? [])

// --- 后端真实推理轨迹（reason_trace）分区渲染 ---
const traceRules = computed(() => qs.dryRun?.reasonTrace?.rules_fired ?? [])
const traceExpansion = computed(() => qs.dryRun?.reasonTrace?.expansion ?? [])
const traceDerived = computed(() => qs.dryRun?.reasonTrace?.derived_filters ?? [])
const traceInferred = computed(() => qs.dryRun?.reasonTrace?.inferred_relations ?? [])
const traceJoins = computed(() => qs.dryRun?.joins ?? [])
const traceHasAny = computed(
  () =>
    !!qs.dryRun?.explanation ||
    traceRules.value.length > 0 ||
    traceExpansion.value.length > 0 ||
    traceDerived.value.length > 0 ||
    traceInferred.value.length > 0 ||
    traceJoins.value.length > 0,
)

/** 行级来源：虚拟类 UNION 时展示「数据源.表」 */
function rowSource(row: Record<string, any>): string {
  const d = String(row?.[ROW_SOURCE.datasource] ?? '').trim()
  const t = String(row?.[ROW_SOURCE.table] ?? '').trim()
  if (d && t) return `${d}.${t}`
  return d || t || '—'
}

onMounted(async () => {
  await Promise.all([
    qs.fetchHistory({ ontology_id: onto.currentOntologyId ?? undefined }),
    qs.fetchTemplates({ ontology_id: onto.currentOntologyId ?? undefined }),
  ])
})

</script>

<template>
  <div class="query-page">
    <div class="page-head">
      <div>
        <h2>查询工作台</h2>
        <p class="sub">
          本体查询 → 推理展开 → 映射翻译 → 多源 SQL
          <span v-if="qs.highlightClasses.length" class="hl">
            涉及类：{{ qs.highlightClasses.join('、') }}
            <ElButton size="small" text type="primary" @click="qs.clearHighlight()">
              清除高亮
            </ElButton>
          </span>
        </p>
      </div>
    </div>

    <div class="layout">
      <!-- ========== 左：构建器 ========== -->
      <aside class="builder">
        <ElCard shadow="never" class="block">
          <template #header>自然语言查询</template>
          <ElInput
            v-model="nlText"
            size="small"
            placeholder="如：查询华东区 VIP 客户今年的订单"
            clearable
            @keyup.enter="runNLParse"
          />
          <div class="nl-actions">
            <ElButton size="small" type="primary" @click="runNLParse" :loading="nlParsing">
              解析
            </ElButton>
            <ElButton size="small" text @click="clearNL">
              清空
            </ElButton>
          </div>
          <div class="nl-badges" v-if="parsePathBadge">
            <ElTag size="small" :type="parsePathBadge.type" effect="dark">
              {{ parsePathBadge.label }}
            </ElTag>
            <ElTag size="small" :type="nlMatched ? 'success' : 'danger'" effect="plain">
              {{ nlMatched ? '已命中本体类' : '未命中本体类' }}
            </ElTag>
          </div>
          <div class="nl-hits" v-if="nlHits.length">
            <ElTag v-for="h in nlHits" :key="h" size="small" type="info">{{ h }}</ElTag>
          </div>
          <div class="nl-notes" v-if="nlNotes.length">
            <div v-for="(n, i) in nlNotes" :key="i" class="nl-note">{{ n }}</div>
          </div>
          <div class="nl-meta" v-if="llmMetaText || nlCanonicalKey || nlTraceId">
            <div v-if="llmMetaText" class="nl-meta-row">
              <span class="lbl">LLM 审计</span><code>{{ llmMetaText }}</code>
            </div>
            <div v-if="nlCanonicalKey" class="nl-meta-row">
              <span class="lbl">缓存键</span><code>{{ nlCanonicalKey }}</code>
            </div>
            <div v-if="nlTraceId" class="nl-meta-row">
              <span class="lbl">链路 id</span><code>{{ nlTraceId }}</code>
              <ElButton link type="primary" size="small" @click="openTrace(nlTraceId)">
                查看全链路
              </ElButton>
            </div>
          </div>
          <div class="nl-examples">
            <span class="lbl">试试：</span>
            <ElButton
              v-for="ex in nlExamples"
              :key="ex"
              size="small"
              text
              type="primary"
              @click="fillExample(ex)"
            >
              {{ ex }}
            </ElButton>
          </div>
        </ElCard>

        <ElCard shadow="never" class="block" :key="'b' + qs.loadedAt">
          <template #header>
            <div class="card-head">
              <span>查询构建</span>
              <ElSwitch
                v-model="qs.current.useReasoning"
                active-text="推理展开"
                size="small"
              />
            </div>
          </template>

          <div class="field">
            <span class="lbl">本体类</span>
            <ElSelect
              :model-value="qs.current.classId"
              filterable
              size="small"
              style="width: 100%"
              @change="(v: string) => qs.setClass(v)"
            >
              <ElOption
                v-for="c in qs.queryableClasses"
                :key="c.id"
                :value="c.id"
                :label="`${c.label}（${c.id}）${c.abstract ? ' · 虚拟' : ''}${c.mapped ? '' : ' · 未映射'}`"
              />
            </ElSelect>
          </div>

          <div class="field">
            <div class="lbl-row">
              <span class="lbl">{{ isAggQuery ? '选择属性（聚合下忽略）' : '选择属性' }}</span>
              <span v-if="isAggQuery" class="muted">输出列由「分组维度 + 度量」决定</span>
            </div>
            <ElSelect
              v-model="qs.current.select"
              multiple
              filterable
              size="small"
              style="width: 100%"
              :disabled="isAggQuery"
              :placeholder="isAggQuery ? '聚合查询不使用投影' : '选择要投影的本体属性'"
            >
              <ElOption
                v-for="p in selectOptions"
                :key="p.id"
                :value="p.id"
                :label="p.label"
              >
                <span>{{ p.label }}</span>
                <ElTag v-if="p.isKey" size="small" type="warning" style="margin-left: 6px">键</ElTag>
                <ElTag v-if="p.sensitive" size="small" type="danger" style="margin-left: 4px">敏</ElTag>
              </ElOption>
            </ElSelect>
          </div>

          <div class="field">
            <div class="lbl-row">
              <span class="lbl">过滤条件</span>
              <ElButton size="small" text type="primary" @click="qs.addFilter()">+ 添加</ElButton>
            </div>
            <div v-if="!qs.current.filters.length" class="muted">无过滤</div>
            <div
              v-for="(f, i) in qs.current.filters"
              :key="i"
              class="filter-row"
            >
              <ElSelect v-model="f.propertyId" size="small" style="flex: 1.2" filterable>
                <ElOption
                  v-for="p in baseProps"
                  :key="p.id"
                  :value="p.id"
                  :label="p.label"
                />
              </ElSelect>
              <ElSelect v-model="f.op" size="small" style="width: 88px">
                <ElOption
                  v-for="o in FILTER_OPS"
                  :key="o.value"
                  :value="o.value"
                  :label="o.label"
                />
              </ElSelect>
              <ElInput
                v-if="needValue(f.op)"
                v-model="f.value"
                size="small"
                style="flex: 1"
                :placeholder="f.op === 'in' ? '逗号分隔' : '值'"
              />
              <ElButton size="small" text type="danger" @click="qs.removeFilter(i)">
                删
              </ElButton>
            </div>
          </div>

          <div class="field">
            <div class="lbl-row">
              <span class="lbl">关系 Join</span>
            </div>
            <div class="join-add">
              <ElSelect
                v-model="pendingJoin"
                size="small"
                style="flex: 1"
                placeholder="选择关系"
                clearable
              >
                <ElOption
                  v-for="r in joinRelationOptions"
                  :key="r.id"
                  :value="r.id"
                  :label="r.label"
                />
              </ElSelect>
              <ElButton size="small" type="primary" plain @click="addJoinFromSelect">
                添加
              </ElButton>
            </div>
            <div v-for="(j, i) in qs.current.joins" :key="j.relationId" class="join-block">
              <div class="join-title">
                <span>{{ onto.relations.find((r) => r.id === j.relationId)?.label ?? j.relationId }}</span>
                <ElButton size="small" text type="danger" @click="qs.removeJoin(i)">移除</ElButton>
              </div>
              <ElSelect
                v-model="j.select"
                multiple
                size="small"
                style="width: 100%"
                placeholder="关联类属性"
              >
                <ElOption
                  v-for="p in joinSelectOptions(j.relationId)"
                  :key="p.id"
                  :value="p.id"
                  :label="p.label"
                />
              </ElSelect>
              <div
                v-for="(f, fi) in j.filters"
                :key="fi"
                class="filter-row compact"
              >
                <ElSelect v-model="f.propertyId" size="small" style="flex: 1" filterable>
                  <ElOption
                    v-for="p in joinFilterProps(j.relationId)"
                    :key="p.id"
                    :value="p.id"
                    :label="p.label"
                  />
                </ElSelect>
                <ElSelect v-model="f.op" size="small" style="width: 72px">
                  <ElOption
                    v-for="o in FILTER_OPS"
                    :key="o.value"
                    :value="o.value"
                    :label="o.label"
                  />
                </ElSelect>
                <ElInput
                  v-if="needValue(f.op)"
                  v-model="f.value"
                  size="small"
                  style="width: 90px"
                />
                <ElButton
                  size="small"
                  text
                  type="danger"
                  @click="j.filters?.splice(fi, 1)"
                >
                  删
                </ElButton>
              </div>
              <ElButton
                size="small"
                text
                type="primary"
                @click="j.filters = [...(j.filters ?? []), { propertyId: joinFilterProps(j.relationId)[0]?.id ?? '', op: 'eq', value: '' }]"
              >
                + 关联过滤
              </ElButton>
            </div>
          </div>

          <!-- 聚合 / 分组（Phase 1）：有度量即为聚合查询，group_by / having 仅在聚合下生效 -->
          <div class="field">
            <div class="lbl-row">
              <span class="lbl">聚合 / 分组</span>
              <span class="agg-head-right">
                <ElTag v-if="isAggQuery" size="small" type="warning" effect="plain">
                  {{ aggregates.length }} 度量 / {{ groupByModel.length }} 维度
                </ElTag>
                <span v-else class="muted">未配置度量，当前为明细查询</span>
                <ElButton
                  v-if="isAggQuery || groupByModel.length || havings.length"
                  size="small"
                  text
                  type="danger"
                  @click="qs.clearAggregation()"
                >
                  清空聚合
                </ElButton>
              </span>
            </div>

            <span class="lbl">度量</span>
            <div v-for="(a, i) in aggregates" :key="i" class="filter-row">
              <ElSelect
                v-model="a.func"
                size="small"
                style="width: 84px"
                @change="syncAggAlias(i)"
              >
                <ElOption
                  v-for="f in AGG_FUNCS"
                  :key="f.value"
                  :value="f.value"
                  :label="f.label"
                />
              </ElSelect>
              <ElSelect
                v-model="a.property"
                size="small"
                style="flex: 1"
                filterable
                clearable
                :disabled="!needAggProp(a.func)"
                :placeholder="needAggProp(a.func) ? '被聚合属性' : 'COUNT(*) 无需属性'"
                @change="syncAggAlias(i)"
              >
                <ElOption
                  v-for="p in measurePropOptions"
                  :key="p.id"
                  :value="p.id"
                  :label="p.label"
                />
              </ElSelect>
              <ElInput
                v-model="a.alias"
                size="small"
                style="width: 124px"
                placeholder="列别名"
                @input="onAliasInput(i)"
              />
              <ElButton size="small" text type="danger" @click="qs.removeAggregate(i)">
                删
              </ElButton>
            </div>
            <ElButton size="small" text type="primary" @click="qs.addAggregate()">
              + 度量
            </ElButton>

            <span class="lbl agg-sub">分组维度</span>
            <ElSelect
              v-model="groupByModel"
              multiple
              clearable
              collapse-tags
              collapse-tags-tooltip
              size="small"
              style="width: 100%"
              :disabled="!isAggQuery"
              placeholder="不分组（全局聚合）"
            >
              <ElOption
                v-for="g in groupOptions"
                :key="g.id"
                :value="g.id"
                :label="g.label"
              />
            </ElSelect>

            <span class="lbl agg-sub">HAVING</span>
            <div v-for="(h, i) in havings" :key="i" class="filter-row">
              <ElSelect
                v-model="h.alias"
                size="small"
                style="flex: 1"
                placeholder="度量别名"
              >
                <ElOption v-for="al in aggAliasOptions" :key="al" :value="al" :label="al" />
              </ElSelect>
              <ElSelect v-model="h.op" size="small" style="width: 72px">
                <ElOption
                  v-for="o in HAVING_OPS"
                  :key="o.value"
                  :value="o.value"
                  :label="o.label"
                />
              </ElSelect>
              <ElInputNumber
                v-if="needValue(h.op)"
                v-model="h.value"
                size="small"
                style="width: 108px"
                :controls="false"
              />
              <ElButton size="small" text type="danger" @click="qs.removeHaving(i)">
                删
              </ElButton>
            </div>
            <ElButton
              size="small"
              text
              type="primary"
              :disabled="!aggAliasOptions.length"
              @click="qs.addHaving()"
            >
              + HAVING
            </ElButton>
            <div v-if="isAggQuery && !aggAliasOptions.length" class="muted agg-sub">
              度量未填别名时无法配置 HAVING
            </div>
          </div>

          <div class="field row-2">
            <div>
              <span class="lbl">{{ isAggQuery ? '排序（维度 / 度量别名）' : '排序属性' }}</span>
              <ElSelect
                v-model="orderByModel"
                clearable
                size="small"
                style="width: 100%"
                placeholder="不排序"
              >
                <ElOption
                  v-for="p in effectiveOrderOptions"
                  :key="p.id + '-desc'"
                  :value="`${p.id}::desc`"
                  :label="`${p.label} ↓`"
                />
                <ElOption
                  v-for="p in effectiveOrderOptions"
                  :key="p.id + '-asc'"
                  :value="`${p.id}::asc`"
                  :label="`${p.label} ↑`"
                />
              </ElSelect>
            </div>
            <div>
              <span class="lbl">Limit</span>
              <ElInputNumber
                v-model="qs.current.limit"
                :min="1"
                :max="200"
                size="small"
                style="width: 100%"
              />
            </div>
          </div>

          <div class="actions">
            <ElButton size="small" type="primary" @click="onExecute" :loading="qs.running">
              执行
            </ElButton>
            <ElButton size="small" :loading="dryRunning" @click="onDryRun">Dry-run</ElButton>
            <ElButton size="small" @click="saveOpen = true">保存</ElButton>
            <ElButton size="small" text @click="qs.resetQuery()">重置</ElButton>
          </div>
        </ElCard>

        <ElCard shadow="never" class="block">
          <template #header>模板 / 最近</template>
          <div class="saved-list">
            <div v-for="s in qs.savedQueries.slice(0, 6)" :key="s.id" class="saved-item">
              <span class="name" @click="applyPreset(s.id)">{{ s.name }}</span>
              <ElPopconfirm title="删除该模板？" @confirm="qs.removeSavedQuery(s.id)">
                <template #reference>
                  <ElButton size="small" text type="danger">删</ElButton>
                </template>
              </ElPopconfirm>
            </div>
          </div>
          <div v-if="qs.queryHistory.length" class="history">
            <div class="hist-title">执行历史</div>
            <div v-for="h in qs.queryHistory.slice(0, 5)" :key="h.id" class="hist-item">
              <ElTag size="small" :type="h.ok ? 'success' : 'danger'">
                {{ h.mode }}
              </ElTag>
              <span class="hist-sum">{{ h.summary }}</span>
              <span v-if="h.rowCount != null" class="muted">{{ h.rowCount }} 行</span>
            </div>
          </div>
        </ElCard>
      </aside>

      <!-- ========== 右：结果 ========== -->
      <main class="result">
        <ElTabs v-model="resultTab">
          <ElTabPane label="SQL" name="sql">
            <ElEmpty
              v-if="!qs.dryRun"
              description="点左侧「Dry-run」或「执行」生成物理 SQL"
            />
            <template v-else>
              <ElAlert
                v-for="(e, i) in qs.dryRun.errors"
                :key="'e' + i"
                type="error"
                :title="e"
                :closable="false"
                style="margin-bottom: 8px"
              />
              <ElAlert
                v-for="(w, i) in qs.dryRun.warnings"
                :key="'w' + i"
                type="warning"
                :title="w"
                :closable="false"
                style="margin-bottom: 8px"
              />
              <div v-for="p in qs.dryRun.plans" :key="p.sourceId" class="sql-block">
                <div class="sql-head">
                  <ElTag type="success">{{ p.sourceLabel }}</ElTag>
                  <ElTag size="small" effect="plain">{{ p.sourceType }}</ElTag>
                  <code>{{ p.schema ? p.schema + '.' : '' }}{{ p.table }}</code>
                  <span class="muted">预估 ~{{ p.estimatedRows }} 行</span>
                </div>
                <pre class="sql">{{ p.sql }}</pre>
              </div>
              <div v-if="qs.dryRun.crossSource" class="cross-note">
                跨源计划：各 SQL 下推执行后本地合并（演示环境不真正连库）
              </div>
            </template>
          </ElTabPane>

          <ElTabPane label="AST" name="ast">
            <pre class="ast">{{ astJson }}</pre>
          </ElTabPane>

          <ElTabPane label="结果" name="result">
            <ElEmpty
              v-if="!qs.executeResult"
              description="点「执行」查看结果集"
            />
            <template v-else>
              <div class="exec-meta">
                <ElTag type="success">{{ resultCountLabel }}</ElTag>
                <ElTag v-if="isAggResult" type="warning" effect="plain">
                  聚合 · {{ aggregates.length || measureCols.size }} 度量 / {{ (qs.executeResult?.columns ?? []).length - measureCols.size }} 维度
                </ElTag>
                <ElTag type="info" effect="plain">{{ qs.executeResult.elapsedMs }} ms</ElTag>
                <ElTag
                  v-for="s in qs.executeResult.sources"
                  :key="s"
                  effect="plain"
                >
                  {{ onto.sourceInfo(s)?.label ?? s }}
                </ElTag>
                <ElTag v-if="qs.executeResult.truncated" type="warning" size="small">
                  {{ isAggResult && qs.current.orderBy ? `TOP-${qs.current.limit}` : '已截断至 limit' }}
                </ElTag>
                <ElTag v-if="qs.executeResult.hasRowSource" type="primary" size="small" effect="plain">
                  行级溯源
                </ElTag>
                <span class="muted">
                  {{ isAggResult ? '列名为本体中文 label，度量列右对齐' : '列名已回填本体属性语义' }}
                </span>
                <ElButton
                  v-if="qs.executeResult.traceId"
                  size="small"
                  text
                  type="primary"
                  :icon="Share"
                  @click="openTrace(qs.executeResult.traceId)"
                >
                  全链路溯源
                </ElButton>
              </div>
              <ElTable :data="qs.executeResult.rows" size="small" border max-height="420">
                <ElTableColumn
                  v-for="c in qs.executeResult.columns"
                  :key="c.propertyId"
                  :prop="c.propertyId"
                  :label="c.label"
                  :align="measureCols.has(c.propertyId) ? 'right' : 'left'"
                  min-width="120"
                >
                  <template #header>
                    <div>
                      {{ c.label }}
                      <span v-if="c.agg" class="col-agg">{{ c.agg.toUpperCase() }}</span>
                      <div class="col-id">{{ c.propertyId }}</div>
                    </div>
                  </template>
                  <template #default="{ row }">
                    {{ cellTextOf(c.propertyId, row[c.propertyId]) }}
                  </template>
                </ElTableColumn>
                <!-- 虚拟类 UNION 时后端逐行注入 _src_datasource / _src_table，此处统一展示为「数据源.来源表」 -->
                <ElTableColumn
                  v-if="qs.executeResult.hasRowSource"
                  label="来源"
                  min-width="180"
                  fixed="right"
                >
                  <template #header>
                    <div>
                      来源
                      <div class="col-id">数据源.来源表</div>
                    </div>
                  </template>
                  <template #default="{ row }">
                    <ElTag size="small" effect="plain" type="primary">{{ rowSource(row) }}</ElTag>
                  </template>
                </ElTableColumn>
              </ElTable>
              <!-- P1(#43)：聚合结果图表 / KPI 指标卡（复用 EChartsBlock，动态 import 不进首屏） -->
              <AggChart v-if="aggChartBlock" :block="aggChartBlock" class="agg-chart" />
            </template>
          </ElTabPane>

          <ElTabPane label="推理 / 计划" name="reason">
            <ElEmpty v-if="!qs.dryRun" description="先 Dry-run 查看推理与计划说明" />
            <template v-else>
              <div class="trace-head">
                <ElTag v-if="qs.dryRun.traceId" size="small" effect="plain" type="info">
                  trace_id：{{ qs.dryRun.traceId }}
                </ElTag>
                <ElTag
                  size="small"
                  :type="traceHasAny ? 'success' : 'warning'"
                  effect="plain"
                >
                  {{ traceHasAny ? '后端真实 reason_trace' : '无 reason_trace（前端降级视图）' }}
                </ElTag>
                <!-- P3：复用 TraceDrawer 查看 NL→intent→AST→SQL→行 全链路与节点耗时 -->
                <ElButton
                  size="small"
                  type="primary"
                  plain
                  :icon="Share"
                  :disabled="!activeTraceId"
                  @click="openTrace(qs.dryRun.traceId)"
                >
                  查看全链路
                </ElButton>
              </div>

              <h3 class="sec">翻译说明</h3>
              <p v-if="qs.dryRun.explanation" class="explain">{{ qs.dryRun.explanation }}</p>
              <p v-else class="muted">后端未返回 explanation</p>

              <h3 class="sec">虚拟展开链</h3>
              <div v-if="traceExpansion.length" class="reason-list">
                <div v-for="(e, i) in traceExpansion" :key="i" class="reason-item">
                  <ElTag size="small" type="info">expand</ElTag>
                  <code>{{ e.from }}</code>
                  <span class="arrow">→</span>
                  <code>{{ e.to }}</code>
                  <span v-if="e.reason" class="muted">{{ e.reason }}</span>
                </div>
              </div>
              <p v-else class="muted">无虚拟类展开（查询类为具体类）</p>

              <h3 class="sec">规则命中</h3>
              <div v-if="traceRules.length" class="reason-list">
                <div v-for="(r, i) in traceRules" :key="i" class="reason-item">
                  <ElTag
                    size="small"
                    :type="r.condition_result ? 'warning' : 'info'"
                    effect="plain"
                  >
                    {{ r.rule_type || 'rule' }}
                  </ElTag>
                  <strong>{{ r.name || `#${r.rule_id}` }}</strong>
                  <span :class="{ muted: !r.condition_result }">
                    {{ r.condition_result ? '条件命中' : '条件未命中' }}
                  </span>
                  <span v-if="r.action" class="muted">{{ r.action }}</span>
                  <span v-if="r.detail" class="muted">{{ r.detail }}</span>
                  <ElTag v-if="r.derived" size="small" type="success" effect="plain">派生</ElTag>
                </div>
              </div>
              <p v-else class="muted">无规则命中</p>

              <h3 class="sec">派生过滤</h3>
              <div v-if="traceDerived.length" class="reason-list">
                <div v-for="(f, i) in traceDerived" :key="i" class="reason-item">
                  <ElTag size="small" type="primary" effect="plain">filter</ElTag>
                  <code>{{ f.property }}</code>
                  <span>{{ f.op }}</span>
                  <span>{{ f.value == null ? '空' : f.value }}</span>
                  <span v-if="f.source" class="muted">规则 {{ f.source }}</span>
                </div>
              </div>
              <p v-else class="muted">无派生过滤</p>

              <h3 class="sec">推导关系</h3>
              <div v-if="traceInferred.length" class="reason-list">
                <div v-for="(r, i) in traceInferred" :key="i" class="reason-item">
                  <ElTag size="small" type="info" effect="plain">relation</ElTag>
                  <strong>{{ r.label || r.name }}</strong>
                  <code>{{ r.from_class }}</code>
                  <span class="arrow">→</span>
                  <code>{{ r.to_class }}</code>
                </div>
              </div>
              <p v-else class="muted">无推导关系</p>

              <h3 class="sec">推理摘要</h3>
              <div v-if="reasonSteps.length" class="reason-list">
                <div v-for="(r, i) in reasonSteps" :key="i" class="reason-item">
                  <ElTag
                    size="small"
                    :type="r.kind === 'rule' ? 'warning' : r.kind === 'filter' ? 'primary' : 'info'"
                  >
                    {{ r.kind }}
                  </ElTag>
                  <span>{{ r.message }}</span>
                </div>
              </div>
              <p v-else class="muted">本次无推理步骤</p>

              <h3 class="sec">改写后过滤</h3>
              <div v-if="qs.dryRun.rewrittenFilters.length" class="reason-list">
                <div
                  v-for="(f, i) in qs.dryRun.rewrittenFilters"
                  :key="i"
                  class="reason-item"
                >
                  <code>{{ f.propertyId }}</code>
                  <span>{{ f.op }}</span>
                  <span>{{ f.value }}</span>
                </div>
              </div>
              <p v-else class="muted">无</p>

              <h3 class="sec">Join 计划</h3>
              <div v-if="traceJoins.length" class="reason-list">
                <div v-for="(j, i) in traceJoins" :key="i" class="reason-item">
                  <ElTag size="small" type="primary" effect="plain">join</ElTag>
                  <strong>{{ j.relation }}</strong>
                  <code>{{ j.table_name }}</code>
                  <span class="muted">ON {{ j.condition }}</span>
                </div>
              </div>
              <div v-else-if="qs.dryRun.joinNotes.length" class="reason-list">
                <div v-for="(n, i) in qs.dryRun.joinNotes" :key="i" class="reason-item">
                  <span>{{ n }}</span>
                </div>
              </div>
              <p v-else class="muted">无关系 Join</p>

              <h3 class="sec">解析类</h3>
              <p>
                <code>{{ qs.dryRun.resolvedClassId }}</code>
                <span class="muted"> · 基类映射 {{ qs.dryRun.plans.length }} 个源</span>
              </p>
            </template>
          </ElTabPane>
        </ElTabs>
      </main>
    </div>

    <ElDialog v-model="saveOpen" title="保存查询模板" width="400px">
      <ElInput v-model="saveName" placeholder="模板名称，如 VIP 客户订单" @keyup.enter="onSaveQuery" />
      <template #footer>
        <ElButton @click="saveOpen = false">取消</ElButton>
        <ElButton type="primary" @click="onSaveQuery">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.query-page {
  height: 100%;
  overflow: auto;
  padding: 16px 20px 32px;
  background: #f8fafc;
}

.page-head {
  margin-bottom: 10px;
}

.page-head h2 {
  margin: 0 0 4px;
  font-size: 18px;
}

.sub {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.hl {
  color: #2563eb;
}

.layout {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 12px;
  align-items: start;
  min-height: 0;
}

.builder {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.block {
  margin: 0;
}

.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.field {
  margin-bottom: 12px;
}

.field.row-2 {
  display: grid;
  grid-template-columns: 1fr 110px;
  gap: 8px;
}

.lbl {
  display: block;
  font-size: 12px;
  color: #64748b;
  margin-bottom: 4px;
}

.lbl-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.lbl-row .lbl {
  margin-bottom: 0;
}

.agg-head-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.lbl.agg-sub {
  margin-top: 10px;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.filter-row.compact {
  margin-top: 6px;
}

.join-add {
  display: flex;
  gap: 6px;
  margin-bottom: 6px;
}

.join-block {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 8px;
  margin-top: 8px;
  background: #f8fafc;
}

.join-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 6px;
}

.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding-top: 4px;
}

.saved-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.saved-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.saved-item .name {
  cursor: pointer;
  color: #2563eb;
}

.saved-item .name:hover {
  text-decoration: underline;
}

.history {
  margin-top: 12px;
  border-top: 1px solid #e2e8f0;
  padding-top: 8px;
}

.hist-title {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 6px;
}

.hist-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  margin-bottom: 4px;
}

.hist-sum {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 8px 14px 16px;
  min-height: 480px;
}

.sql-block {
  margin-bottom: 12px;
}

.sql-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}

.sql {
  margin: 0;
  background: #0f172a;
  color: #e2e8f0;
  padding: 12px 14px;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre;
}

.ast {
  margin: 0;
  background: #0f172a;
  color: #a5f3fc;
  padding: 12px 14px;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  max-height: 520px;
}

.cross-note {
  font-size: 12px;
  color: #b45309;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 6px;
  padding: 8px 10px;
}

.exec-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.col-id {
  font-size: 10px;
  color: #9ca3af;
  font-weight: 400;
}

/* 度量列表头的聚合函数小标（SUM / COUNT 等） */
.col-agg {
  margin-left: 4px;
  font-size: 10px;
  font-weight: 600;
  color: #d97706;
}

/* P1(#43)：聚合结果图表与表格的间距 */
.agg-chart {
  margin-top: 12px;
}

.sec {
  margin: 14px 0 8px;
  font-size: 13px;
  color: #374151;
  border-left: 3px solid #4c8bf5;
  padding-left: 8px;
}

.reason-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.reason-item {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 13px;
}

.muted {
  color: #9ca3af;
  font-size: 12px;
}

.nl-actions {
  display: flex;
  gap: 6px;
  margin-top: 8px;
}

.nl-hits {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 8px;
}

.nl-notes {
  margin-top: 8px;
  font-size: 12px;
  color: #475569;
  line-height: 1.6;
}

.nl-note::before {
  content: '→ ';
  color: #94a3b8;
}

/* --- P2：解析路径标签与服务端元信息 --- */
.nl-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 8px;
}

.nl-meta {
  margin-top: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  background: #f5f8ff;
  border: 1px solid #e2ebff;
}

.nl-meta-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 12px;
  color: #475569;
  line-height: 1.7;
}

.nl-meta-row .lbl {
  display: inline;
  flex: 0 0 auto;
  margin-bottom: 0;
  color: #94a3b8;
}

.nl-meta-row code {
  font-size: 11px;
  color: #1d4ed8;
  word-break: break-all;
}

/* --- P2：真实 reason_trace 展示 --- */
.trace-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.explain {
  margin: 0;
  font-size: 12px;
  color: #475569;
  line-height: 1.7;
  white-space: pre-wrap;
}

.arrow {
  color: #4c8bf5;
  font-weight: 600;
}

.nl-examples {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.nl-examples .lbl {
  margin-bottom: 0;
}
</style>
