/** OntologyQuery AST（结构化本体查询） */

export type FilterOp =
  | 'eq'
  | 'ne'
  | 'gt'
  | 'gte'
  | 'lt'
  | 'lte'
  | 'in'
  | 'notin'
  | 'like'
  | 'isnull'
  | 'isnotnull'

export const FILTER_OPS: { value: FilterOp; label: string; needValue: boolean }[] = [
  { value: 'eq', label: '=', needValue: true },
  { value: 'ne', label: '≠', needValue: true },
  { value: 'gt', label: '>', needValue: true },
  { value: 'gte', label: '≥', needValue: true },
  { value: 'lt', label: '<', needValue: true },
  { value: 'lte', label: '≤', needValue: true },
  { value: 'in', label: 'IN', needValue: true },
  { value: 'notin', label: 'NOT IN', needValue: true },
  { value: 'like', label: 'LIKE', needValue: true },
  { value: 'isnull', label: 'IS NULL', needValue: false },
  { value: 'isnotnull', label: 'IS NOT NULL', needValue: false },
]

// ---------------------------------------------------------------------------
// 聚合（Phase 1）：与后端冻结契约逐字对齐，字段名一律 snake_case
// ---------------------------------------------------------------------------

/**
 * 聚合函数枚举。
 * 后端四处同源：translator 实现 / handler.supportedAggFuncs /
 * ai.supportedAggregateFuncs / intent_tool.go 的 Enum，前端不得臆造新值。
 */
export type AggFunc = 'count' | 'sum' | 'avg' | 'min' | 'max'

/**
 * 仿 FILTER_OPS 的下拉选项集；needProperty=false 表示可省略属性（COUNT(*)）。
 *
 * ⚠️ label 必须与后端 `translator.aggFuncLabels` 逐字一致
 * （count=计数 / sum=合计 / avg=平均 / min=最小 / max=最大），
 * 因为后端缺省度量别名 = `aggFuncLabels[fn] + 属性中文 label`（前缀式），
 * 前端自动填充的 alias 必须用同一套词，否则同一个度量在
 * 「前端预填」与「后端生成」两条路径下会得到不同的列名。
 */
export const AGG_FUNCS: { value: AggFunc; label: string; needProperty: boolean }[] = [
  { value: 'count', label: '计数', needProperty: false },
  { value: 'sum', label: '合计', needProperty: true },
  { value: 'avg', label: '平均', needProperty: true },
  { value: 'min', label: '最小', needProperty: true },
  { value: 'max', label: '最大', needProperty: true },
]

/** 是否为受支持的聚合函数（后端白名单归一用） */
export function isAggFunc(v: unknown): v is AggFunc {
  const s = String(v ?? '').trim().toLowerCase()
  return AGG_FUNCS.some((a) => a.value === s)
}

/** 聚合函数 → 中文名（未知值回退大写原名） */
export function aggFuncLabel(func?: string | null): string {
  const f = String(func ?? '').trim().toLowerCase()
  if (!f) return ''
  return AGG_FUNCS.find((a) => a.value === f)?.label ?? f.toUpperCase()
}

/**
 * 缺省度量别名（镜像后端 `defaultAggAlias`）：前缀式「聚合函数中文名 + 属性中文 label」。
 * 例：SUM(quantity) → 「合计数量」；COUNT(*) → 「计数」。
 * 后端额外会做重名唯一化（追加 2/3…），前端无法预测，故只用于预填与展示兜底。
 */
export function defaultAggAlias(func?: string | null, propLabel?: string | null): string {
  const base = aggFuncLabel(func)
  if (!base) return ''
  const lb = String(propLabel ?? '').trim()
  return lb ? `${base}${lb}` : base
}

/**
 * 数值型属性 range（可作 SUM/AVG/MIN/MAX 的度量候选）。
 * 本体 range 是自由文本（xsd:decimal / INT / 金额 等），故用宽松关键词匹配。
 */
export function isNumericRange(range?: string | null): boolean {
  return /int|decimal|number|float|double|numeric|long|real|money|currency|quantity|amount|数量|金额|整数|小数/i.test(
    String(range ?? ''),
  )
}

/**
 * 维度型属性 range（文本 / 日期 / 枚举，可作分组键）。
 * QueryLab 的「分组维度」候选与 NL 解析共用同一判定，避免两处口径不一。
 */
export function isDimensionRange(range?: string | null): boolean {
  return /string|text|char|varchar|date|time|enum|bool|文本|字符|日期|时间|枚举/i.test(String(range ?? ''))
}

/** 一条聚合度量（前端形态，与后端 AST 的 aggregates 元素一一对应） */
export interface AggregateSpec {
  func: AggFunc
  /** 被聚合属性名；count 时可空表示 COUNT(*) */
  property: string
  /** 输出列别名；缺省时后端自动生成中文「聚合函数+label」 */
  alias?: string
}

/** 一条 HAVING（按 aggregate 别名引用度量，op 复用 filter 操作符集） */
export interface HavingSpec {
  alias: string
  op: FilterOp
  value: number
}

/** 后端 AST 中的聚合度量（QueryRequest.aggregates 元素） */
export interface AggregateVO {
  func: string
  /** count 时可空（COUNT(*)） */
  property?: string
  alias?: string
}

/** 后端 AST 中的 HAVING 条件（QueryRequest.having 元素） */
export interface HavingVO {
  alias: string
  op: string
  value?: number
}

/** 结果列角色：维度 / 度量 */
export type ColumnRole = 'dimension' | 'measure'

/** query / dry-run 响应的 column_meta 元素（与 columns 顺序对齐） */
export interface ColumnMetaVO {
  name: string
  role: ColumnRole
  /** 度量列携带的聚合函数名 */
  agg?: string
}

export interface QueryFilter {
  propertyId: string
  op: FilterOp
  value?: string
}

export interface QueryJoin {
  /** 关系 id，如 Order.customer */
  relationId: string
  select?: string[]
  filters?: QueryFilter[]
}

export interface OntologyQuery {
  classId: string
  select: string[]
  filters: QueryFilter[]
  joins: QueryJoin[]
  orderBy?: { propertyId: string; desc?: boolean } | null
  limit: number
  /** 是否启用推理展开（虚拟类 / 规则） */
  useReasoning: boolean
  /** 限定数据源，空 = 全部已映射源 */
  sourceIds?: string[]
  /** 分组维度：属性名，跨类时用 `relationName.property` */
  groupBy?: string[]
  /** 聚合度量；非空即视为聚合查询 */
  aggregates?: AggregateSpec[]
  /** HAVING：按度量别名过滤聚合结果 */
  having?: HavingSpec[]
}

export interface ReasonStep {
  kind: 'subclass' | 'rule' | 'filter' | 'note'
  message: string
}

// ---------------------------------------------------------------------------
// 后端真实推理轨迹（对齐 internal/ont/reason/trace.go 的 ReasonTrace）
// ---------------------------------------------------------------------------

/** 一条规则在推理过程中的评估与执行结果 */
export interface RuleFireVO {
  rule_id: number
  name: string
  rule_type: string
  /** 条件是否命中 */
  condition_result: boolean
  /** 是否以「派生当前类」方式注入过滤器 */
  derived: boolean
  /** 动作摘要，如 expand_class:VipCustomer */
  action: string
  detail: string
}

/** 虚拟类展开的一个步骤 */
export interface ExpansionStepVO {
  from: string
  to: string
  /** 展开原因：subclass / self-mapping / parent-fallback */
  reason: string
}

/** 由规则派生注入的过滤条件 */
export interface DerivedFilterVO {
  property: string
  op: string
  value?: unknown
  /** 来源规则名 */
  source: string
}

/** 被推导出的关系 */
export interface InferredRelationVO {
  name: string
  label: string
  from_class: string
  to_class: string
}

/** 一次推理过程的全部可解释信息（query/dry-run 与 query 响应均携带） */
export interface ReasonTraceVO {
  rules_fired?: RuleFireVO[] | null
  expansion?: ExpansionStepVO[] | null
  derived_filters?: DerivedFilterVO[] | null
  inferred_relations?: InferredRelationVO[] | null
}

/** 翻译过程中生成的一个 JOIN（对齐 query.JoinInfo） */
export interface JoinInfoVO {
  relation: string
  table_name: string
  condition: string
}

// ---------------------------------------------------------------------------
// 推理单步解释（POST /reason/explain）
// ---------------------------------------------------------------------------

/** 展开后的落地目标类（虚拟类 → 真实表/数据源） */
export interface ExpandedTargetVO {
  class_name: string
  source_table: string
  datasource_id?: number
  datasource_name?: string
  datasource_type?: string
}

/** POST /reason/explain 响应：只推理不执行 SQL */
export interface ReasonExplainVO {
  ontology_id?: number
  class_name?: string
  /** ont_class.class_type：后端默认 normal，虚拟类为 virtual */
  class_type?: string
  /** 注意：后端将轨迹嵌在 reason_trace 字段内，非平铺 */
  reason_trace?: ReasonTraceVO | null
  expanded_targets?: ExpandedTargetVO[] | null
  trace_id?: string
}

// ---------------------------------------------------------------------------
// 服务端意图解析（POST /query/nl-parse）
// ---------------------------------------------------------------------------

/**
 * 解析路径。
 * rule=规则零LLM、llm=Eino 意图链、cache=计划缓存、concept=概念知识库、
 * clarification=澄清、meta=词表/溯源等元信息、direct=直接查询（非 NL 入口，
 * 如 QueryLab 构建器 Dry-run / 执行）、local=前端离线兜底。
 */
export type ParsePath =
  | 'rule'
  | 'llm'
  | 'cache'
  | 'concept'
  | 'clarification'
  | 'meta'
  | 'direct'
  | 'local'
  | (string & {})

/** 后端查询意图 AST（query.QueryRequest） */
export interface QueryRequestVO {
  ontology_id?: number
  class_name?: string
  /** 要查询的属性名列表，空 = 全部 */
  properties?: string[] | null
  filters?: { property: string; op: string; value?: unknown }[] | null
  /** 要展开的关系名列表 */
  relations?: string[] | null
  limit?: number
  offset?: number
  /** 可填属性名，也可填 aggregate 别名（TOP-N = 别名 + desc + limit） */
  order_by?: string
  /** asc / desc */
  order_dir?: string
  /** 分组维度：属性名或 `relationName.property` */
  group_by?: string[] | null
  aggregates?: AggregateVO[] | null
  having?: HavingVO[] | null
}

/** LLM 路径的可复现审计元数据（对齐 ai.LLMMeta） */
export interface LLMMetaVO {
  model?: string
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens?: number
  seed?: number
  temperature?: number
  prompt_version?: string
  vocab_version?: string
  latency_ms?: number
  /** IntentValidator 有界修复次数 */
  repairs?: number
}

/** POST /query/nl-parse 响应 */
export interface NLParseResultVO {
  query?: QueryRequestVO | null
  notes?: string[] | null
  hits?: string[] | null
  parse_path?: ParsePath
  matched?: boolean
  llm_meta?: LLMMetaVO | null
  canonical_key?: string
  trace_id?: string
}

/**
 * 链路中的一个执行节点（对齐 internal/recorder.Node）。
 * 由 GET /query/trace/:traceId 的 nodes 字段返回，用于渲染耗时时间线。
 */
export interface TraceNodeVO {
  name: string
  /** 节点类型，如 chain / model / tool */
  type?: string
  /** 所属组件，如 intent / translate / execute */
  component?: string
  /** 来源：eino=大模型编排框架、pipeline=确定性管线 */
  source?: 'eino' | 'pipeline' | (string & {})
  input_summary?: string
  output_summary?: string
  /** 相对链路起点的开始偏移（毫秒） */
  start_offset_ms?: number
  duration_ms?: number
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens?: number
  error?: string
}

/** GET /query/trace/:traceId 响应（全链路溯源） */
export interface TraceVO {
  trace_id: string
  nl_text?: string
  parse_path?: ParsePath
  intent?: QueryRequestVO | null
  ontology_query?: QueryRequestVO | null
  llm?: LLMMetaVO | null
  reason_trace?: ReasonTraceVO | null
  translated_sql?: string
  params?: unknown[] | null
  datasource_id?: number
  source_table?: string
  joins?: JoinInfoVO[] | null
  explanation?: string
  result_count?: number
  execution_time_ms?: number
  status?: string
  error_message?: string
  /** 节点级耗时与摘要（nodes_json） */
  nodes?: TraceNodeVO[] | null
  created_at?: string
}

/** 行级溯源隐藏列名（translator.srcLineageCols 注入） */
export const ROW_SOURCE = {
  datasource: '_src_datasource',
  table: '_src_table',
} as const

export interface SourcePlan {
  sourceId: string
  sourceLabel: string
  sourceType: string
  table: string
  schema?: string
  sql: string
  estimatedRows?: number
}

export interface DryRunResult {
  ok: boolean
  errors: string[]
  warnings: string[]
  /** 推理展开步骤 */
  reasons: ReasonStep[]
  /** 最终作用类（推理后） */
  resolvedClassId: string
  /** 追加的本体层过滤（来自规则） */
  rewrittenFilters: QueryFilter[]
  /** 各源物理计划 */
  plans: SourcePlan[]
  /** AST 预览 */
  ast: OntologyQuery & {
    resolvedClassId: string
    rewrittenFilters: QueryFilter[]
  }
  /** 跨源说明 */
  crossSource: boolean
  joinNotes: string[]
  /** 后端真实推理轨迹（reason_trace），无则为 null */
  reasonTrace: ReasonTraceVO | null
  /** 后端翻译说明（explanation） */
  explanation: string
  /** 后端真实 JOIN 计划 */
  joins: JoinInfoVO[]
  /** 本次翻译/执行的链路 id，可一键溯源 */
  traceId: string
}

export interface QueryRow {
  [propertyId: string]: string | number | null
}

export interface QueryColumn {
  propertyId: string
  /**
   * 表头文本。
   * 聚合/明细查询的列名由后端直接返回本体中文 label，前端不再二次映射；
   * 故正常情况下 label === propertyId。
   */
  label: string
  range: string
  /** 列角色（来自响应 column_meta，非聚合查询为空） */
  role?: ColumnRole
  /** 度量列的聚合函数（来自响应 column_meta） */
  agg?: string
}

export interface ExecuteResult {
  columns: QueryColumn[]
  rows: QueryRow[]
  /** 每行来源（演示 provenance） */
  sources: string[]
  elapsedMs: number
  truncated: boolean
  rowCount: number
  /** 参与查询的类，供图上高亮 */
  involvedClasses: string[]
  /** 结果行是否携带行级溯源列（_src_datasource/_src_table） */
  hasRowSource: boolean
  /** 本次执行的链路 id */
  traceId: string
  /** 是否为聚合结果（响应 aggregated） */
  aggregated: boolean
}

export interface SavedQuery {
  id: string
  name: string
  query: OntologyQuery
  at: number
}

export interface QueryHistoryItem {
  id: string
  at: number
  mode: 'dry-run' | 'execute'
  summary: string
  ok: boolean
  rowCount?: number
}
