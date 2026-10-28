import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as queryApi from '@/api/query'
import type { QueryResultVO, QueryDryRunVO, QueryHistoryVO, QueryTemplateVO, PageResult } from '@/types/api'
import type {
  AggFunc,
  AggregateSpec,
  ColumnMetaVO,
  DryRunResult,
  ExecuteResult,
  FilterOp,
  HavingSpec,
  JoinInfoVO,
  LLMMetaVO,
  NLParseResultVO,
  OntologyQuery,
  ParsePath,
  QueryFilter,
  QueryHistoryItem,
  QueryRequestVO,
  QueryRow,
  ReasonStep,
  ReasonTraceVO,
  SavedQuery,
  SourcePlan,
} from '@/types/query'
import { AGG_FUNCS, FILTER_OPS, ROW_SOURCE, aggFuncLabel, defaultAggAlias, isAggFunc, isDimensionRange, isNumericRange } from '@/types/query'
import type { OntoClass, OntoProperty, OntoRelation } from '@/types/ontology'
import { useOntologyStore } from './ontology'
import { useMappingStore } from './mapping'
import { useDataSourceStore } from './datasource'

function emptyQuery(): OntologyQuery {
  return {
    classId: '',
    select: [],
    filters: [],
    joins: [],
    orderBy: null,
    limit: 20,
    useReasoning: true,
    // 聚合三字段恒为数组（而非 undefined），避免 applyQuery / 模板回填时出现 undefined 混入
    groupBy: [],
    aggregates: [],
    having: [],
  }
}

/** 自然语言解析结果 */
export interface NLParseResult {
  query: OntologyQuery
  notes: string[]
  hits: string[]
  /** 解析路径：rule 规则零LLM / llm 意图链 / cache 计划缓存 / local 前端离线兜底 */
  parsePath: ParsePath
  /** 服务端是否命中本体类 */
  matched: boolean
  /** LLM 路径的可复现审计元数据 */
  llmMeta?: LLMMetaVO | null
  /** 规范化缓存键（同问同 SQL） */
  canonicalKey?: string
  /** 服务端链路 id，可一键溯源 */
  traceId?: string
}

/** 属性别名候选 */
interface NLPropAlias {
  alias: string
  classId: string
  prop: OntoProperty
}

/** NL 比较词 → FilterOp（按优先级排列，eq 最后） */
const NL_OPS: { re: string; op: FilterOp }[] = [
  { re: '(?:>=|≥|大于等于|不低于|不小于)', op: 'gte' },
  { re: '(?:<=|≤|小于等于|不超过|不大于)', op: 'lte' },
  { re: '(?:!=|<>|≠|不等于)', op: 'ne' },
  { re: '(?:>|大于|超过|高于)', op: 'gt' },
  { re: '(?:<|小于|低于)', op: 'lt' },
  { re: '(?:包含|含有|模糊匹配|like)', op: 'like' },
  { re: '(?:等于|为|是|=|:|：)', op: 'eq' },
]

const NL_NULL_RE = /^\s*(?:为空|是空|不存在|没值|is\s+null)/i
const NL_OP_RE = new RegExp(`^\\s*(?:${NL_OPS.map((o) => o.re).join('|')})`, 'i')

/**
 * NL 聚合关键词 → AggFunc（与服务端 internal/ai/nlparse 的关键词集对齐）。
 * 顺序即优先级：先具体词（平均/总和/最大）后泛词（多少/统计）。
 */
const NL_AGG_RULES: { re: RegExp; func: AggFunc }[] = [
  { re: /(?:平均|均值|人均|avg|average)/i, func: 'avg' },
  { re: /(?:总和|求和|总量|合计|汇总|累加|sum|total)/i, func: 'sum' },
  { re: /(?:最大|最多|最高|峰值|max|maximum)/i, func: 'max' },
  { re: /(?:最小|最少|最低|min|minimum)/i, func: 'min' },
  { re: /(?:统计|总数|多少|几笔|几条|几项|几个|计数|共计|共有|count)/i, func: 'count' },
]

/** TOP-N 复合模式：「TOP 3」/「前 3」 */
const NL_TOP_N_RE = /(?:top|前)\s*(\d+)/i

/**
 * 度量别名统一走 `defaultAggAlias`（镜像后端 `translator.defaultAggAlias`）：
 * 前缀式「聚合函数中文名 + 属性中文 label」，如 SUM(数量) → 「合计数量」、COUNT(*) → 「计数」。
 * 不再维护前端自有的后缀词表，避免同一度量在前端预填与后端生成两条路径下列名不一致。
 */

/** 在文本中定位被聚合的数值属性（按别名包含匹配，长名优先） */
function findMeasureProp(text: string, aliases: NLPropAlias[]): OntoProperty | null {
  const lower = text.toLowerCase()
  for (const a of aliases) {
    if (!isNumericRange(a.prop.range)) continue
    if (lower.includes(a.alias.toLowerCase())) return a.prop
  }
  return null
}

/** 在文本中定位可作分组键的维度属性（去重、按命中位置升序，最多 max 个） */
function findDimensionProps(text: string, aliases: NLPropAlias[], max = 2): OntoProperty[] {
  const lower = text.toLowerCase()
  const hit: Array<{ at: number; prop: OntoProperty }> = []
  const seen = new Set<string>()
  for (const a of aliases) {
    if (!isDimensionRange(a.prop.range)) continue
    if (seen.has(a.prop.id)) continue
    const at = lower.indexOf(a.alias.toLowerCase())
    if (at < 0) continue
    seen.add(a.prop.id)
    hit.push({ at, prop: a.prop })
  }
  hit.sort((x, y) => x.at - y.at)
  return hit.slice(0, max).map((h) => h.prop)
}

/** TOP-N 无显式维度时的兜底分组键：名称/编码类属性优先，否则首个维度型属性 */
function pickDefaultDimension(pool: OntoProperty[]): OntoProperty | null {
  const dims = pool.filter((p) => isDimensionRange(p.range))
  if (!dims.length) return null
  return (
    dims.find((p) => /name|名称|title|标题|code|编码|no$/i.test(`${p.label}${p.id}`)) ??
    dims.find((p) => p.isKey) ??
    dims[0] ??
    null
  )
}

function opLabel(op: FilterOp): string {
  return FILTER_OPS.find((o) => o.value === op)?.label ?? op
}

export const useQueryStore = defineStore('query', () => {
  const onto = useOntologyStore()
  const maps = useMappingStore()
  const ds = useDataSourceStore()

  // --- API-backed state ---
  const results = ref<QueryResultVO | null>(null)
  const generatedSQL = ref('')
  const history = ref<QueryHistoryVO[]>([])
  const templates = ref<QueryTemplateVO[]>([])
  const loading = ref(false)
  const error = ref('')

  // --- Legacy compat state ---
  const current = ref<OntologyQuery>(emptyQuery())
  const dryRunResult = ref<DryRunResult | null>(null)
  const executeResult = ref<ExecuteResult | null>(null)
  const running = ref(false)
  const savedQueries = ref<SavedQuery[]>([])
  const highlightClasses = ref<string[]>([])
  const loadedAt = ref(0)

  // --- 基于本体 store 的真实派生数据 ---
  const queryableClasses = computed(() =>
    onto.classes.map(c => ({
      id: c.id,
      label: c.label,
      abstract: !!c.abstract,
      mapped: (c.mappedSources?.length ?? 0) > 0 || maps.classMaps.some(m => m.classId === c.id),
    }))
  )

  const currentClass = computed(() => onto.classes.find(c => c.id === current.value.classId) ?? null)

  /** NL 示例句：由当前本体的真实类名生成（不再硬编码领域词） */
  const nlExamples = computed<string[]>(() => {
    const concrete = onto.classes.filter(c => !c.abstract)
    const virtual = onto.classes.filter(c => c.abstract)
    const out: string[] = []
    for (const c of concrete.slice(0, 3)) out.push(`查询${c.label}前 20 条`)
    for (const c of virtual.slice(0, 1)) out.push(`查询${c.label}（推理展开）`)
    if (concrete[0] && concrete[1]) {
      const rel = onto.relations.find(
        r => (r.domain === concrete[0].id && r.range === concrete[1].id)
          || (r.range === concrete[0].id && r.domain === concrete[1].id),
      )
      if (rel) out.push(`${concrete[0].label}的${concrete[1].label}`)
    }
    return out.length ? out : ['查询本体类前 20 条']
  })

  /** 沿 subclassOf 向上回溯到具体基类 */
  const resolvedBase = computed(() => {
    let c = currentClass.value
    const seen = new Set<string>()
    while (c && c.subclassOf && !seen.has(c.id)) {
      seen.add(c.id)
      const parent = onto.classes.find(x => x.id === c!.subclassOf)
      if (!parent) break
      c = parent
    }
    return c
  })

  const availableRelations = computed<OntoRelation[]>(() => {
    const base = resolvedBase.value?.id
    if (!base) return onto.relations
    return onto.relations.filter(r => r.domain === base || r.range === base)
  })

  function summarizeHistory(h: QueryHistoryVO): string {
    // 后端字段：query_text（查询文本）、ontology_name
    const text = (h.query_text || '').trim()
    if (text) return text.slice(0, 60)
    return h.error_message || (h.ontology_name ?? '')
  }

  /** QueryHistoryVO → 页面使用的 QueryHistoryItem */
  const historyItems = computed<QueryHistoryItem[]>(() =>
    history.value.map(h => ({
      id: String(h.id),
      at: h.created_at ? new Date(h.created_at).getTime() : 0,
      mode: 'execute' as const,
      summary: summarizeHistory(h),
      ok: h.status === 'success',
      rowCount: h.result_count,
    }))
  )

  function collectClasses(q: OntologyQuery, resolvedClassId: string): string[] {
    const set = new Set<string>()
    if (q.classId) set.add(q.classId)
    if (resolvedClassId) set.add(resolvedClassId)
    for (const j of q.joins ?? []) {
      const rel = onto.relations.find(r => r.id === j.relationId)
      if (rel) { set.add(rel.domain); set.add(rel.range) }
    }
    return [...set]
  }

  /** 前端 FilterOp → 后端 op 枚举 */
  const OP_TO_BACKEND: Record<string, string> = {
    eq: 'eq', ne: 'neq', gt: 'gt', gte: 'gte', lt: 'lt', lte: 'lte',
    in: 'in', notin: 'not_in', like: 'like',
    isnull: 'is_null', isnotnull: 'is_not_null',
  }

  /** 后端 op 枚举 → 前端 FilterOp（含 neq/ge/le/nin/not_null 等别名归一） */
  const BACKEND_TO_OP: Record<string, FilterOp> = {
    eq: 'eq', ne: 'ne', neq: 'ne',
    gt: 'gt', gte: 'gte', ge: 'gte',
    lt: 'lt', lte: 'lte', le: 'lte',
    like: 'like', in: 'in', not_in: 'notin', nin: 'notin',
    is_null: 'isnull', is_not_null: 'isnotnull', not_null: 'isnotnull',
  }

  /** 后端过滤值 → 前端字符串值（数组按逗号拼接，空值转 undefined） */
  function toFilterValue(v: unknown): string | undefined {
    if (v == null) return undefined
    if (Array.isArray(v)) return v.map((x) => String(x)).join(',')
    const s = String(v)
    return s === '' ? undefined : s
  }

  /** 后端 op → 可读标签（未知则原样返回） */
  function backendOpLabel(op: string): string {
    const f = BACKEND_TO_OP[op]
    return f ? opLabel(f) : op
  }

  /** 过滤值文案化 */
  function filterValueText(v: unknown): string {
    if (v == null || v === '') return '空'
    return Array.isArray(v) ? v.join(',') : String(v)
  }

  /** OntologyQuery → 后端 QueryRequest（ontology_id / class_name / properties / filters / relations / limit / order_by / group_by / aggregates / having） */
  function buildQueryRequest(q: OntologyQuery): any {
    const req: any = {
      ontology_id: onto.currentOntologyId ?? undefined,
      class_name: q.classId,
      properties: q.select ?? [],
      limit: q.limit ?? 20,
    }
    const filters = (q.filters ?? [])
      .filter(f => f.propertyId)
      .map(f => ({ property: f.propertyId, op: OP_TO_BACKEND[f.op] ?? f.op, value: f.value }))
    if (filters.length) req.filters = filters
    const rels = (q.joins ?? []).map(j => j.relationId).filter(Boolean)
    if (rels.length) req.relations = rels
    // --- 聚合三字段（snake_case，与后端冻结契约逐字对齐）---
    const groups = (q.groupBy ?? []).map(g => String(g ?? '').trim()).filter(Boolean)
    if (groups.length) req.group_by = groups
    const aggregates = (q.aggregates ?? [])
      .filter(a => a && isAggFunc(a.func))
      .map(a => {
        const item: Record<string, unknown> = { func: a.func }
        // count 可省略 property（COUNT(*)）；其余函数必须携带
        const prop = String(a.property ?? '').trim()
        if (prop) item.property = prop
        const alias = String(a.alias ?? '').trim()
        if (alias) item.alias = alias
        return item
      })
      .filter(item => item.func === 'count' || !!item.property)
    if (aggregates.length) req.aggregates = aggregates
    const having = (q.having ?? [])
      .filter(h => h && String(h.alias ?? '').trim())
      .map(h => {
        const n = Number(h.value)
        return {
          alias: String(h.alias).trim(),
          op: OP_TO_BACKEND[h.op] ?? h.op,
          value: Number.isFinite(n) ? n : 0,
        }
      })
    if (having.length) req.having = having
    // order_by 可为属性名，也可为 aggregate 别名（TOP-N）
    if (q.orderBy?.propertyId) {
      req.order_by = q.orderBy.propertyId
      req.order_dir = q.orderBy.desc ? 'desc' : 'asc'
    }
    return req
  }

  /** 后端 AST（QueryRequestVO）→ 前端 OntologyQuery；class/property/relation 名称与前端 id 同源，无需查表 */
  function mapQueryRequest(req: QueryRequestVO | null | undefined, fallback: OntologyQuery): OntologyQuery {
    if (!req) return { ...fallback }
    const q = emptyQuery()
    q.classId = req.class_name ?? ''
    q.select = [...(req.properties ?? [])]
    q.filters = (req.filters ?? [])
      .filter((f) => f && f.property)
      .map((f) => ({
        propertyId: f.property,
        op: BACKEND_TO_OP[f.op] ?? 'eq',
        value: toFilterValue(f.value),
      }))
    q.joins = (req.relations ?? [])
      .filter((name) => !!name)
      .map((name) => ({ relationId: name, select: [], filters: [] }))
    // --- 聚合三字段反向映射（白名单归一，非法函数直接丢弃）---
    q.groupBy = (req.group_by ?? []).map((g) => String(g ?? '').trim()).filter(Boolean)
    q.aggregates = (req.aggregates ?? [])
      .filter((a) => a && isAggFunc(a.func))
      .map((a) => ({
        func: String(a.func).trim().toLowerCase() as AggFunc,
        property: String(a.property ?? ''),
        alias: String(a.alias ?? ''),
      }))
    q.having = (req.having ?? [])
      .filter((h) => h && String(h.alias ?? '').trim())
      .map((h) => {
        const n = Number(h.value)
        return {
          alias: String(h.alias).trim(),
          op: BACKEND_TO_OP[String(h.op ?? '')] ?? 'gt',
          value: Number.isFinite(n) ? n : 0,
        }
      })
    q.orderBy = req.order_by
      ? { propertyId: req.order_by, desc: (req.order_dir ?? '').toLowerCase() === 'desc' }
      : null
    // 服务端未给 limit（规则路径仅在命中「前N/topN」时赋值）时沿用当前面板值，避免全表扫描
    const limit = Number(req.limit)
    q.limit = Number.isFinite(limit) && limit > 0 ? limit : (fallback.limit || 20)
    // 后端翻译管线始终应用推理，无需前端开关
    q.useReasoning = true
    return q
  }

  /** 后端 reason_trace / explanation / joins → 前端推理步骤（真实轨迹，不再前端拼装） */
  function buildReasonSteps(
    trace: ReasonTraceVO | null,
    explanation: string,
    resolvedClassId: string,
    q: OntologyQuery,
  ): ReasonStep[] {
    const steps: ReasonStep[] = []
    for (const e of trace?.expansion ?? []) {
      steps.push({
        kind: 'subclass',
        message: `虚拟类展开：${e.from} → ${e.to}${e.reason ? `（${e.reason}）` : ''}`,
      })
    }
    for (const r of trace?.rules_fired ?? []) {
      const name = r.name || `#${r.rule_id}`
      const cond = r.condition_result ? '条件命中' : '条件未命中'
      const extra = [r.action, r.detail].filter(Boolean).join(' · ')
      steps.push({
        kind: 'rule',
        message: `规则 ${name}（${r.rule_type || 'rule'}）${cond}${extra ? ' · ' + extra : ''}`,
      })
    }
    for (const f of trace?.derived_filters ?? []) {
      steps.push({
        kind: 'filter',
        message: `派生过滤：${f.property} ${backendOpLabel(f.op)} ${filterValueText(f.value)}${f.source ? `（规则 ${f.source}）` : ''}`,
      })
    }
    for (const r of trace?.inferred_relations ?? []) {
      steps.push({
        kind: 'note',
        message: `推导关系：${r.label || r.name}（${r.from_class} → ${r.to_class}）`,
      })
    }
    if (explanation) steps.push({ kind: 'note', message: explanation })
    // 后端未返回轨迹时保留最小可读信息（离线/降级场景）
    if (!steps.length && q.useReasoning && resolvedClassId !== q.classId) {
      steps.push({ kind: 'subclass', message: `虚拟类 ${q.classId} 展开为 ${resolvedClassId}` })
    }
    return steps
  }

  /** reason_trace.derived_filters → 前端改写后过滤条件 */
  function mapDerivedFilters(trace: ReasonTraceVO | null): QueryFilter[] {
    return (trace?.derived_filters ?? []).map((f) => ({
      propertyId: f.property,
      op: BACKEND_TO_OP[f.op] ?? 'eq',
      value: toFilterValue(f.value),
    }))
  }

  /** 后端 joins → 可读 Join 说明；后端无 joins 时回退前端本体关系描述 */
  function buildJoinNotes(joins: JoinInfoVO[], q: OntologyQuery): string[] {
    if (joins.length) {
      return joins.map((j) => `${j.relation}：JOIN ${j.table_name} ON ${j.condition}`)
    }
    return (q.joins ?? []).map((j) => {
      const rel = onto.relations.find((r) => r.id === j.relationId)
      return rel ? `${rel.label}：${rel.domain} → ${rel.range}（${rel.cardinality}）` : j.relationId
    })
  }

  /** QueryDryRunVO → DryRunResult（后端为单条翻译结果，组装为单元素 plans） */
  function mapDryRunVO(res: QueryDryRunVO | QueryResultVO | null, q: OntologyQuery, failMessage?: string): DryRunResult {
    const errors = failMessage ? [failMessage] : []
    const resolvedClassId = res?.class_name || q.classId
    const plans: SourcePlan[] = []
    if (res && (res.datasource_id != null || res.sql)) {
      const sid = res.datasource_id != null ? String(res.datasource_id) : ''
      const info = sid ? ds.sourceInfo(sid) : undefined
      const cm = maps.classMaps.find(m => m.classId === resolvedClassId && m.sourceId === sid)
      plans.push({
        sourceId: sid,
        sourceLabel: res.datasource_name || info?.label || sid,
        sourceType: res.datasource_type || info?.type || '',
        table: res.source_table || cm?.table || '',
        schema: cm?.schema,
        sql: res.sql || '',
        estimatedRows: 0,
      })
    }
    const trace: ReasonTraceVO | null = res?.reason_trace ?? null
    const explanation = res?.explanation ?? ''
    const joins: JoinInfoVO[] = Array.isArray(res?.joins) ? (res!.joins as JoinInfoVO[]) : []
    const reasons = buildReasonSteps(trace, explanation, resolvedClassId, q)
    const rewrittenFilters = mapDerivedFilters(trace)
    return {
      ok: errors.length === 0,
      errors,
      warnings: [],
      reasons,
      resolvedClassId,
      rewrittenFilters,
      plans,
      ast: { ...q, resolvedClassId, rewrittenFilters },
      crossSource: plans.length > 1,
      joinNotes: buildJoinNotes(joins, q),
      reasonTrace: trace,
      explanation,
      joins,
      traceId: res?.trace_id ?? '',
    }
  }

  // --- API Actions ---
  async function executeQuery(params: any) {
    loading.value = true
    error.value = ''
    try {
      results.value = await queryApi.executeQuery(params) as QueryResultVO
      return results.value
    } catch (e: any) {
      error.value = e.message || '查询执行失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function dryRun(params: any) {
    loading.value = true
    error.value = ''
    const q: OntologyQuery = params?.ontology_query ?? current.value
    try {
      const res = await queryApi.dryRunQuery(buildQueryRequest(q)) as QueryDryRunVO
      generatedSQL.value = res?.sql || ''
      dryRunResult.value = mapDryRunVO(res ?? null, q)
      highlightClasses.value = collectClasses(q, dryRunResult.value.resolvedClassId)
      return res
    } catch (e: any) {
      error.value = e?.message || '查询解析失败'
      dryRunResult.value = mapDryRunVO(null, q, error.value)
      return null
    } finally {
      loading.value = false
    }
  }

  async function explain(params: any) {
    loading.value = true
    try {
      return await queryApi.explainQuery(params)
    } catch (e: any) {
      error.value = e.message || '查询解释失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function fetchHistory(params?: { current?: number; size?: number; ontology_id?: number }) {
    try {
      const res = await queryApi.getQueryHistory(params) as PageResult<QueryHistoryVO> | QueryHistoryVO[]
      if (Array.isArray(res)) {
        history.value = res
      } else {
        history.value = res.records ?? []
      }
    } catch (e) {
      console.error('[query] fetchHistory failed', e)
    }
  }

  async function fetchTemplates(params?: { ontology_id?: number }) {
    try {
      const res = await queryApi.getQueryTemplates(params) as PageResult<QueryTemplateVO> | QueryTemplateVO[]
      // 后端返回分页对象 { records, total, ... }，兼容裸数组
      templates.value = Array.isArray(res) ? res : (res.records ?? [])
      syncSavedFromTemplates()
    } catch (e) {
      console.error('[query] fetchTemplates failed', e)
    }
  }

  /** 将服务端模板合并进 savedQueries（仅替换 t- 前缀项） */
  function syncSavedFromTemplates() {
    const fromApi: SavedQuery[] = templates.value.map(t => {
      let q = emptyQuery()
      try {
        const parsed = JSON.parse(t.query_json || '{}')
        q = {
          ...emptyQuery(),
          ...parsed,
          filters: parsed?.filters ?? [],
          joins: parsed?.joins ?? [],
          select: parsed?.select ?? [],
          // 旧模板无聚合字段，显式兜底为数组（避免 null 覆盖 emptyQuery 默认值）
          groupBy: parsed?.groupBy ?? [],
          aggregates: parsed?.aggregates ?? [],
          having: parsed?.having ?? [],
        }
      } catch { /* ignore */ }
      return {
        id: `t-${t.id}`,
        name: t.name,
        query: q,
        at: t.created_at ? new Date(t.created_at).getTime() : 0,
      }
    })
    const local = savedQueries.value.filter(s => !s.id.startsWith('t-'))
    savedQueries.value = [...fromApi, ...local]
  }

  async function saveTemplate(data: any) {
    const res = await queryApi.saveQueryTemplate(data) as QueryTemplateVO
    templates.value.push(res)
    syncSavedFromTemplates()
    return res
  }

  // --- NL 解析：默认走服务端 POST /query/nl-parse（规则零LLM → LLM 意图链），本地正则仅作离线兜底 ---

  /** 收集全部属性别名（label / id / id 尾段 / 同义词），长名优先 */
  function propAliases(): NLPropAlias[] {
    const out: NLPropAlias[] = []
    for (const c of onto.classes) {
      for (const p of c.properties) {
        const names = new Set<string>([
          p.label,
          p.id,
          p.id.split('.').pop() ?? '',
          ...(p.synonyms ?? []),
        ])
        for (const n of names) {
          const t = (n ?? '').trim()
          if (t.length >= 2) out.push({ alias: t, classId: c.id, prop: p })
        }
      }
    }
    out.sort((a, b) => b.alias.length - a.alias.length)
    return out
  }

  /**
   * 前端离线兜底解析（同步，纯正则 + 本体元数据）。
   * 仅在未选择本体、服务端不可达或后端报错时使用；常规路径请调 async parseNaturalLanguage。
   */
  function parseNaturalLanguageLocal(input: string): NLParseResult {
    const raw = (input ?? '').trim()
    const notes: string[] = []
    const hits: string[] = []
    const q = emptyQuery()
    if (!raw) {
      notes.push('请输入查询意图')
      return { query: q, notes, hits, parsePath: 'local', matched: false }
    }
    const lower = raw.toLowerCase()

    // 1) 主类识别：按 label / id / 同义词包含匹配
    const classHits: OntoClass[] = []
    for (const c of onto.classes) {
      const keys = [c.label, c.id, ...(c.synonyms ?? [])].filter(Boolean)
      if (keys.some((k) => lower.includes(String(k).toLowerCase()))) classHits.push(c)
    }
    classHits.sort((a, b) => (b.label?.length ?? 0) - (a.label?.length ?? 0))

    const primary: OntoClass | undefined = classHits[0]
    if (primary) {
      q.classId = primary.id
      hits.push(primary.label)
      notes.push(`主类：${primary.label}（${primary.id}）`)
      if (primary.abstract) {
        q.useReasoning = true
        notes.push(`${primary.label} 为推理/虚拟类，将展开父类并套用规则`)
      }
    } else {
      q.useReasoning = false
      notes.push('未识别到本体类，请在构建器中手动选择')
    }

    // 2) 解析基类（虚拟类落到父类属性空间）
    let base: OntoClass | undefined = primary
    const seen = new Set<string>()
    while (base?.subclassOf && !seen.has(base.id)) {
      seen.add(base.id)
      const parent = onto.classes.find((c) => c.id === base!.subclassOf)
      if (!parent) break
      base = parent
    }
    if (base && primary && base.id !== primary.id) {
      notes.push(`属性空间取自基类：${base.label}（${base.id}）`)
    }

    // 3) 默认投影：键属性优先，最多 4 列
    const pool = base?.properties ?? primary?.properties ?? []
    const keyIds = pool.filter((p) => p.isKey).map((p) => p.id)
    const restIds = pool.filter((p) => !p.isKey).map((p) => p.id)
    q.select = [...keyIds, ...restIds].slice(0, 4)

    // 4) 过滤条件：<属性别名><比较词><值>
    const scopeId = base?.id ?? primary?.id
    const aliases = propAliases().filter((a) => !scopeId || a.classId === scopeId)
    const consumed: Array<[number, number]> = []
    for (const a of aliases) {
      const idx = lower.indexOf(a.alias.toLowerCase())
      if (idx < 0) continue
      const end = idx + a.alias.length
      if (consumed.some(([s, e]) => idx < e && end > s)) continue
      const restText = raw.slice(end)

      const nullMatch = NL_NULL_RE.exec(restText)
      if (nullMatch) {
        q.filters.push({ propertyId: a.prop.id, op: 'isnull' })
        hits.push(`${a.alias} 为空`)
        notes.push(`过滤：${a.prop.label} IS NULL`)
        consumed.push([idx, end + nullMatch[0].length])
        continue
      }

      const opMatch = NL_OP_RE.exec(restText)
      if (!opMatch) continue
      const opDef = NL_OPS.find((o) => new RegExp(`^${o.re}`, 'i').test(opMatch[0].trim()))
      if (!opDef) continue
      const after = restText.slice(opMatch[0].length)
      const valMatch = /^\s*["'“]?([^"'“”，,。;；、\s]+)["'”]?/.exec(after)
      if (!valMatch) continue
      q.filters.push({ propertyId: a.prop.id, op: opDef.op, value: valMatch[1] })
      hits.push(`${a.alias} ${opLabel(opDef.op)} ${valMatch[1]}`)
      notes.push(`过滤：${a.prop.label} ${opLabel(opDef.op)} ${valMatch[1]}`)
      consumed.push([idx, end + opMatch[0].length + valMatch[0].length])
    }

    // 5) 聚合意图：<聚合关键词> + 就近数值属性；「TOP N」= 分组 + 按度量降序 + limit
    //    与服务端 nlparse 关键词集对齐，避免降级路径静默丢失统计意图
    const aggRule = NL_AGG_RULES.find((r) => r.re.test(raw))
    const topNMatch = NL_TOP_N_RE.exec(raw)
    // 度量属性：先在当前类作用域找，再退到全本体（跨类聚合由后端挂载关系）
    const measureProp = findMeasureProp(raw, aliases) ?? findMeasureProp(raw, propAliases())
    const aggFunc: AggFunc | null = aggRule
      ? aggRule.func
      : topNMatch && measureProp
        ? 'sum'
        : null
    let topNApplied = false
    if (aggFunc && (aggFunc === 'count' || measureProp)) {
      const needProp = AGG_FUNCS.find((a) => a.value === aggFunc)?.needProperty ?? true
      const prop = needProp ? measureProp : null
      const alias = defaultAggAlias(aggFunc, prop?.label)
      const aggregates: AggregateSpec[] = [{ func: aggFunc, property: prop?.id ?? '', alias }]
      hits.push(`聚合 ${aggFuncLabel(aggFunc)}（${prop?.label ?? '全部记录'}）`)
      notes.push(`聚合：${aggFunc.toUpperCase()}(${prop ? prop.label : '*'}) AS ${alias}`)
      // 分组维度：文本命中的维度型属性；TOP-N 未命中时回退到名称/编码类属性
      let dims = findDimensionProps(raw, aliases)
      if (!dims.length && topNMatch) {
        const d = pickDefaultDimension(pool)
        if (d) dims = [d]
      }
      if (dims.length) notes.push(`分组：${dims.map((d) => d.label).join('、')}`)
      // TOP-N：按度量别名降序取前 N（非「截断」语义）
      if (topNMatch) {
        q.limit = Number(topNMatch[1])
        q.orderBy = { propertyId: alias, desc: true }
        topNApplied = true
        notes.push(`TOP-${q.limit}：按 ${alias} 降序`)
      }
      q.aggregates = aggregates
      q.groupBy = dims.map((d) => d.id)
      notes.push(`聚合查询：${aggregates.length} 度量 / ${dims.length} 维度`)
    }

    // 6) 排序：按 <属性> 降序/升序（TOP-N 已按度量别名排序时不覆盖）
    const obMatch = /(?:按|根据|依)\s*([^\s，,。;；、]{2,20})\s*(降序|从高到低|从大到小|倒序|升序|从低到高|从小到大|排序)?/.exec(raw)
    if (obMatch && !q.orderBy) {
      const name = obMatch[1].trim()
      const p = pool.find(
        (x) => x.label === name || x.id === name || x.id.endsWith(`.${name}`),
      )
      if (p) {
        const desc = !!obMatch[2] && /降序|从高到低|从大到小|倒序/.test(obMatch[2])
        q.orderBy = { propertyId: p.id, desc }
        hits.push(`按 ${p.label} ${desc ? '降序' : '升序'}`)
        notes.push(`排序：${p.label} ${desc ? '降序' : '升序'}`)
      }
    }

    // 7) limit
    const limitMatch =
      /(?:前|top)\s*(\d+)/i.exec(raw) ??
      /(?:限制|limit)\s*(\d+)/i.exec(raw) ??
      /(\d+)\s*[条项个行]/.exec(raw)
    if (limitMatch && !topNApplied) {
      q.limit = Number(limitMatch[1])
      hits.push(`limit=${q.limit}`)
      notes.push(`Limit = ${q.limit}`)
    }

    // 8) Join：其余命中的类若与主类/基类存在关系，自动挂载
    const anchor = base?.id ?? q.classId
    for (const oc of classHits.slice(1)) {
      if (q.joins.length >= 2) break
      const rel = onto.relations.find(
        (r) =>
          (r.domain === anchor && r.range === oc.id) ||
          (r.range === anchor && r.domain === oc.id) ||
          (r.domain === q.classId && r.range === oc.id) ||
          (r.range === q.classId && r.domain === oc.id),
      )
      if (!rel || q.joins.some((j) => j.relationId === rel.id)) continue
      q.joins.push({
        relationId: rel.id,
        select: oc.properties.slice(0, 3).map((p) => p.id),
        filters: [],
      })
      hits.push(oc.label)
      notes.push(`Join：${rel.label}（${rel.domain} → ${rel.range}）`)
    }

    // 9) 规则：命中当前类的启用规则则开启推理
    if (q.classId) {
      const enabledRules = onto.rulesOf(q.classId).filter((r) => r.enabled !== false)
      if (enabledRules.length) {
        q.useReasoning = true
        notes.push(`命中 ${enabledRules.length} 条规则：${enabledRules.map((r) => r.label).join('、')}`)
      }
    }

    if (!hits.length) notes.push('未识别到明确条件，可在构建器中手动调整')
    return { query: q, notes, hits, parsePath: 'local', matched: !!q.classId }
  }

  /**
   * 服务端意图解析（P2）：调用 POST /query/nl-parse，返回受本体词表约束的查询 AST。
   * 后端优先确定性规则（parse_path=rule），未命中且 ai.enabled 时回退 Eino 意图链（llm），
   * 命中意图缓存则为 cache（同问同 SQL）。
   * 未选择本体或服务端异常时自动降级为 parseNaturalLanguageLocal，并在 notes 中说明。
   */
  async function parseNaturalLanguage(input: string): Promise<NLParseResult> {
    const text = (input ?? '').trim()
    if (!text) {
      return { query: emptyQuery(), notes: ['请输入查询意图'], hits: [], parsePath: 'local', matched: false }
    }
    const ontologyId = onto.currentOntologyId
    if (ontologyId == null) {
      const local = parseNaturalLanguageLocal(text)
      return { ...local, notes: [...local.notes, '未选择本体，已使用前端离线解析'] }
    }
    try {
      const res = (await queryApi.nlParse(ontologyId, text)) as NLParseResultVO
      const notes = [...(res?.notes ?? [])]
      const query = mapQueryRequest(res?.query, current.value)
      const aggCount = (query.aggregates ?? []).length
      const groupCount = (query.groupBy ?? []).length
      if (!res?.matched) notes.push('服务端未命中本体类，请手动选择或补充关键词')
      if (res?.parse_path === 'cache') notes.push('命中意图缓存：同问同 SQL')
      if (aggCount) {
        // 聚合场景：不再提「投影全部属性」（SELECT 严格 = 分组键 ∪ 聚合列）
        notes.push(`聚合查询：${aggCount} 度量 / ${groupCount} 维度`)
        const havingCount = (query.having ?? []).length
        if (havingCount) notes.push(`HAVING ${havingCount} 条（按度量别名过滤）`)
        if (query.orderBy && query.limit > 0) {
          notes.push(`TOP-${query.limit}：按 ${query.orderBy.propertyId} ${query.orderBy.desc ? '降序' : '升序'}`)
        }
      } else if (res?.matched && !query.select.length) {
        notes.push('投影：全部属性（服务端意图未指定列，可在构建器中收窄）')
      }
      if (res?.llm_meta?.model) {
        notes.push(`LLM：${res.llm_meta.model}${res.llm_meta.repairs ? ` · 修复 ${res.llm_meta.repairs} 次` : ''}`)
      }
      if (!res?.hits?.length) notes.push('未识别到明确条件，可在构建器中手动调整')
      return {
        query,
        notes,
        hits: res?.hits ?? [],
        parsePath: (res?.parse_path ?? 'rule') as ParsePath,
        matched: !!res?.matched,
        llmMeta: res?.llm_meta ?? null,
        canonicalKey: res?.canonical_key ?? '',
        traceId: res?.trace_id ?? '',
      }
    } catch (e: any) {
      // request.ts 已统一弹错，此处仅降级并补充说明，避免重复提示
      const local = parseNaturalLanguageLocal(text)
      return {
        ...local,
        notes: [...local.notes, `服务端解析不可用（${e?.message || '网络错误'}），已回退前端本地解析`],
      }
    }
  }

  // --- Legacy compatibility methods ---
  function propertyLabel(propertyId: string): string {
    return propertyId
  }

  function resetQuery(partial?: Partial<OntologyQuery>) {
    current.value = { ...emptyQuery(), ...partial }
    dryRunResult.value = null
    executeResult.value = null
    highlightClasses.value = []
  }

  function setClass(classId: string) {
    current.value.classId = classId
    current.value.select = []
    current.value.filters = []
    current.value.joins = []
    current.value.orderBy = null
    // 切类时属性空间变化，聚合三字段一并重置（否则残留旧类的属性名/别名）
    current.value.groupBy = []
    current.value.aggregates = []
    current.value.having = []
    dryRunResult.value = null
    executeResult.value = null
  }

  function addFilter() {
    current.value.filters = [
      ...current.value.filters,
      { propertyId: '', op: 'eq' as FilterOp, value: '' },
    ]
  }

  function removeFilter(i: number) {
    current.value.filters = current.value.filters.filter((_, idx) => idx !== i)
  }

  function addJoin(relationId: string) {
    current.value.joins = [
      ...current.value.joins,
      { relationId, select: [], filters: [] },
    ]
  }

  function removeJoin(i: number) {
    current.value.joins = current.value.joins.filter((_, idx) => idx !== i)
  }

  /** 新增一行度量（缺省 COUNT(*)，无需属性） */
  function addAggregate(spec?: Partial<AggregateSpec>) {
    current.value.aggregates = [
      ...(current.value.aggregates ?? []),
      {
        func: isAggFunc(spec?.func) ? spec!.func : ('count' as AggFunc),
        property: String(spec?.property ?? ''),
        alias: String(spec?.alias ?? ''),
      },
    ]
  }

  function removeAggregate(i: number) {
    const prev = current.value.aggregates ?? []
    const removedAlias = String(prev[i]?.alias ?? '').trim()
    current.value.aggregates = prev.filter((_, idx) => idx !== i)
    // 度量被删后，其别名不可再被 HAVING / 排序引用（否则会生成非法 SQL）
    if (removedAlias) {
      current.value.having = (current.value.having ?? []).filter((h) => h.alias !== removedAlias)
      if (current.value.orderBy?.propertyId === removedAlias) current.value.orderBy = null
    }
  }

  /** 新增一行 HAVING（默认引用首个度量别名） */
  function addHaving(spec?: Partial<HavingSpec>) {
    const firstAlias = String((current.value.aggregates ?? [])[0]?.alias ?? '').trim()
    const n = Number(spec?.value)
    current.value.having = [
      ...(current.value.having ?? []),
      {
        alias: String(spec?.alias ?? firstAlias),
        op: spec?.op ?? ('gt' as FilterOp),
        value: Number.isFinite(n) ? n : 0,
      },
    ]
  }

  function removeHaving(i: number) {
    current.value.having = (current.value.having ?? []).filter((_, idx) => idx !== i)
  }

  /** 一键清空聚合配置（回到明细查询） */
  function clearAggregation() {
    current.value.groupBy = []
    current.value.aggregates = []
    current.value.having = []
  }

  function runDryRun(): DryRunResult | null {
    // In API mode, dry-run is server-side
    return dryRunResult.value
  }

  async function runExecute(): Promise<ExecuteResult | null> {
    running.value = true
    try {
      const res = await executeQuery(buildQueryRequest(current.value))
      if (!res) return null
      // columns 为 string[]（列名）；兼容旧对象形式
      const colNames: string[] = (res.columns ?? []).map(c =>
        typeof c === 'string' ? c : ((c as any)?.name ?? ''),
      ).filter(Boolean)
      // rows 优先用 records（对象数组），否则由 rows（二维）转对象
      let rows: QueryRow[] = []
      if (Array.isArray(res.records) && res.records.length) {
        rows = res.records as QueryRow[]
      } else if (Array.isArray(res.rows)) {
        rows = (res.rows as any[][]).map(r => {
          const obj: QueryRow = {}
          colNames.forEach((cn, i) => { obj[cn] = r?.[i] ?? null })
          return obj
        })
      }
      const sourceId = res.datasource_id != null ? String(res.datasource_id) : ''
      // 行级溯源：虚拟类 UNION 时后端为每行注入 _src_datasource / _src_table 常量列
      const hasRowSource = rows.some(
        (r) => r != null && (r[ROW_SOURCE.datasource] != null || r[ROW_SOURCE.table] != null),
      )
      // 溯源列不作为业务数据列渲染，由「来源」列统一展示
      const dataCols = hasRowSource
        ? colNames.filter((c) => c !== ROW_SOURCE.datasource && c !== ROW_SOURCE.table)
        : colNames
      const rowSources = hasRowSource
        ? [
            ...new Set(
              rows
                .map((r) => String(r?.[ROW_SOURCE.datasource] ?? '').trim())
                .filter(Boolean),
            ),
          ]
        : []
      // 聚合结果：列角色/聚合函数来自响应 column_meta（按列名反查，不依赖下标位置）
      const columnMeta: ColumnMetaVO[] = Array.isArray(res.column_meta) ? res.column_meta : []
      const metaOf = (name: string): ColumnMetaVO | undefined =>
        columnMeta.find((m) => m && m.name === name)
      const aggregated = !!res.aggregated
      const exec: ExecuteResult = {
        // 列名已由后端返回本体中文 label，直接用作表头；role/agg 仅供对齐与渲染区分
        columns: dataCols.map((cn) => {
          const m = metaOf(cn)
          return { propertyId: cn, label: cn, range: '', role: m?.role, agg: m?.agg }
        }),
        rows,
        sources: rowSources.length
          ? rowSources
          : (sourceId ? [res.datasource_name || sourceId] : []),
        elapsedMs: res.duration_ms || 0,
        truncated: res.truncated || false,
        rowCount: res.row_count ?? rows.length,
        involvedClasses: collectClasses(current.value, current.value.classId),
        hasRowSource,
        traceId: res.trace_id || '',
        aggregated,
      }
      executeResult.value = exec
      highlightClasses.value = exec.involvedClasses
      if (res.sql) generatedSQL.value = res.sql
      if (!dryRunResult.value) {
        // 复用执行结果中的 datasource/sql/reason_trace 信息组装 dry-run 视图
        dryRunResult.value = mapDryRunVO(res, current.value)
      } else if (res.reason_trace || res.trace_id) {
        // 已有 dry-run 视图：用执行响应的真实轨迹刷新推理面板
        const trace = res.reason_trace ?? dryRunResult.value.reasonTrace
        const explanation = res.explanation || dryRunResult.value.explanation
        const joins: JoinInfoVO[] = Array.isArray(res.joins)
          ? (res.joins as JoinInfoVO[])
          : dryRunResult.value.joins
        dryRunResult.value.reasonTrace = trace
        dryRunResult.value.explanation = explanation
        dryRunResult.value.joins = joins
        dryRunResult.value.reasons = buildReasonSteps(
          trace,
          explanation,
          dryRunResult.value.resolvedClassId,
          current.value,
        )
        dryRunResult.value.rewrittenFilters = mapDerivedFilters(trace)
        dryRunResult.value.joinNotes = buildJoinNotes(joins, current.value)
        if (res.trace_id) dryRunResult.value.traceId = res.trace_id
      }
      return exec
    } finally {
      running.value = false
    }
  }

  function saveQuery(name: string) {
    const item: SavedQuery = {
      id: `q-${Date.now()}`,
      name: name.trim() || `${current.value.classId} query`,
      query: structuredClone(current.value),
      at: Date.now(),
    }
    savedQueries.value = [item, ...savedQueries.value]
    // 尽力持久化为服务端模板（失败不影响本地使用）
    if (onto.currentOntologyId != null) {
      void queryApi
        .saveQueryTemplate({
          ontology_id: onto.currentOntologyId,
          name: item.name,
          query_json: JSON.stringify(item.query),
        })
        .catch(() => { /* ignore */ })
    }
    return item
  }

  function applyQuery(q: OntologyQuery) {
    current.value.classId = q.classId
    current.value.select = [...q.select]
    current.value.filters = (q.filters ?? []).map(f => ({ ...f }))
    current.value.joins = (q.joins ?? []).map(j => ({
      ...j,
      select: [...(j.select ?? [])],
      filters: (j.filters ?? []).map(f => ({ ...f })),
    }))
    // ❗ 聚合三字段必须显式逐字段深拷贝：本函数是逐字段赋值而非整体替换，
    // 漏写任一字段都会让 NL 解析出的聚合意图静默丢失（不报错、无提示）。
    current.value.groupBy = [...(q.groupBy ?? [])]
    current.value.aggregates = (q.aggregates ?? []).map(a => ({ ...a }))
    current.value.having = (q.having ?? []).map(h => ({ ...h }))
    current.value.orderBy = q.orderBy ? { ...q.orderBy } : null
    current.value.limit = q.limit ?? 20
    current.value.useReasoning = !!q.useReasoning
    current.value.sourceIds = q.sourceIds ? [...q.sourceIds] : undefined
    dryRunResult.value = null
    executeResult.value = null
    highlightClasses.value = []
    loadedAt.value = Date.now()
  }

  function loadSavedQuery(id: string): boolean {
    const s = savedQueries.value.find(x => x.id === id)
    if (!s) return false
    applyQuery(structuredClone(s.query))
    return true
  }

  function removeSavedQuery(id: string) {
    savedQueries.value = savedQueries.value.filter(x => x.id !== id)
  }

  function clearHighlight() {
    highlightClasses.value = []
  }

  return {
    // API state
    results, generatedSQL, history, templates, loading, error,
    // Legacy compat state
    current, dryRun: dryRunResult, executeResult, running,
    savedQueries, queryHistory: historyItems, historyItems, highlightClasses, loadedAt,
    queryableClasses, currentClass, resolvedBase, availableRelations, nlExamples,
    // API actions
    executeQuery, dryRunQuery: dryRun, explain,
    fetchHistory, fetchTemplates, saveTemplate,
    parseNaturalLanguage, parseNaturalLanguageLocal,
    // Legacy actions
    propertyLabel, resetQuery, setClass, applyQuery,
    addFilter, removeFilter, addJoin, removeJoin,
    addAggregate, removeAggregate, addHaving, removeHaving, clearAggregation,
    runDryRun, runExecute, saveQuery, loadSavedQuery, removeSavedQuery,
    clearHighlight,
  }
})
