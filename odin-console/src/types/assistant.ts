/**
 * P3 智能助手类型定义（对齐后端 internal/handler/assistant.go、internal/skills/*）
 *
 * 设计文档：设计文档/推理和AI助手设计.md §8（P3）
 */
import type { ParsePath, QueryRequestVO, ReasonTraceVO, JoinInfoVO } from './query'

// ---------------------------------------------------------------------------
// 技能清单（GET /assistant/skills）
// ---------------------------------------------------------------------------

/** 技能类型（skills.TypeDataQuery / TypeConceptQA / TypeReason / TypeProvenance / TypeMeta） */
export type SkillType = 'data_query' | 'concept_qa' | 'reason' | 'provenance' | 'meta' | (string & {})

/** 技能示例（few-shot，也用于前端展示可问样例） */
export interface SkillExampleVO {
  /** 自然语言问题 */
  nl: string
  /** 期望的意图 AST（可选） */
  intent?: unknown
  /** 期望的答案摘要（可选） */
  answer?: string
}

export interface SkillVO {
  id: string
  name: string
  type: SkillType
  description: string
  /** 是否确定性技能（零 LLM） */
  deterministic: boolean
  /** 预设问题（concept_qa 含 6 条概念库问题） */
  preset_questions?: string[] | null
  examples?: SkillExampleVO[] | null
}

/** GET /assistant/skills 解包后的响应体 */
export interface SkillListVO {
  skills?: SkillVO[] | null
  count?: number
  trace_id?: string
}

// ---------------------------------------------------------------------------
// 渲染块（Block）—— 助手消息的结构化产物
// ---------------------------------------------------------------------------

/**
 * 块类型（skills.BlockXxx 常量）。
 * text/reasoning/intent/sql/table/chart/provenance/clarification，
 * 预留 markdown/pivot/file 等，新增类型只需注册渲染器，不改核心。
 *
 * 注：新 SQL 生成架构（#66）已退役 thinking/plan/derivation/hop 四类块，
 * 相应渲染器与类型一并移除；`(string & {})` 兜底仍可渲染历史遗留块。
 */
export type BlockType =
  | 'text'
  | 'reasoning'
  | 'intent'
  | 'sql'
  | 'table'
  | 'chart'
  | 'provenance'
  | 'clarification'
  | (string & {})

/** 通用渲染块。payload 形态由 type 决定，见下方各 *Payload 接口 */
export interface BlockVO<P = unknown> {
  type: BlockType
  title?: string
  payload?: P
  /** 关联链路 id，可一键溯源 */
  trace_ref?: string
}

/** text 块：{ text } */
export interface TextPayload {
  text?: string
}

/**
 * chart / table 块的维度或度量字段描述（后端对象契约）。
 * field = 列名（与 rows/columns 中的名字一致），label = 展示名（本体中文 label）。
 */
export interface ChartFieldVO {
  field: string
  label: string
  /** 仅度量携带：count / sum / avg / min / max */
  agg?: string
}

/** table 块：columns/rows 为对齐数组，records 为对象数组（含 _src_* 行级溯源列） */
export interface TablePayload {
  columns?: string[] | null
  rows?: unknown[][] | null
  records?: Record<string, unknown>[] | null
  row_count?: number
  /** 是否为聚合结果（表头「N 行」改「N 个分组」） */
  aggregated?: boolean
  /** 维度列（聚合结果可选下发） */
  dimensions?: ChartFieldVO[] | string[] | null
  /** 度量列（聚合结果可选下发） */
  metrics?: ChartFieldVO[] | string[] | null
  /** 后端图表建议；metric 表示纯标量聚合，应渲染指标卡 */
  chart_type?: string
}

/** chart 块：后端产出 chart_type/dimensions/metrics/rows */
export interface ChartPayload {
  /** bar / line / pie / metric（metric = 纯标量聚合，渲染 KPI 指标卡） */
  chart_type?: string
  /**
   * 维度列。新契约为 `{field,label}[]`；旧契约为 `string[]`，
   * 统一由 `normalizeChartPayload` 归一。
   */
  dimensions?: ChartFieldVO[] | string[] | null
  /** 度量列，形态同 dimensions（对象项额外携带 agg） */
  metrics?: ChartFieldVO[] | string[] | null
  rows?: unknown[][] | null
  /** 与 rows 同序的列名（后端可选下发，用于按 field 名反查列下标） */
  columns?: string[] | null
  /** 是否为聚合结果 */
  aggregated?: boolean
  /** TOP-N 提示（非「截断」语义） */
  top_n?: number
}

/** 归一化后的图表字段 */
export interface ChartField {
  field: string
  label: string
  agg?: string
}

/** 归一化后的图表规格（组件直接消费，无需再兼容旧形态） */
export interface NormalizedChart {
  chartType: 'bar' | 'line' | 'pie' | 'metric'
  dimensions: ChartField[]
  metrics: ChartField[]
  /** 与 rows 同序的列名，用于按 field 反查列下标 */
  columns: string[]
  rows: unknown[][]
  aggregated: boolean
  /** 0 = 后端未下发 TOP-N */
  topN: number
}

const CHART_TYPES = ['bar', 'line', 'pie', 'metric'] as const

/**
 * 字段列表归一：兼容 `{field,label,agg}[]` 与旧 `string[]` 两种形态。
 * 对象项缺 label 时回退 field；字符串项 field=label=字符串本身。
 */
export function normalizeChartFields(v: unknown): ChartField[] {
  if (!Array.isArray(v)) return []
  const out: ChartField[] = []
  for (const item of v) {
    if (item == null) continue
    if (typeof item === 'string') {
      const s = item.trim()
      if (s) out.push({ field: s, label: s })
      continue
    }
    if (typeof item === 'object') {
      const o = item as Partial<ChartFieldVO>
      const field = String(o.field ?? o.label ?? '').trim()
      if (!field) continue
      const agg = String(o.agg ?? '').trim().toLowerCase()
      out.push({ field, label: String(o.label ?? field).trim() || field, ...(agg ? { agg } : {}) })
    }
  }
  return out
}

/**
 * chart payload 归一化（兼容旧 `string[]` 形态）。
 * columns 优先取 payload.columns，否则由「维度在前、度量在后」推导，
 * 使图表能按 field 名反查列下标，不依赖 i+1 位置假设。
 */
export function normalizeChartPayload(payload: unknown): NormalizedChart {
  const p = (payload && typeof payload === 'object' ? payload : {}) as ChartPayload
  const rawType = String(p.chart_type ?? '').trim().toLowerCase()
  const chartType = (CHART_TYPES as readonly string[]).includes(rawType)
    ? (rawType as NormalizedChart['chartType'])
    : 'bar'
  const dimensions = normalizeChartFields(p.dimensions)
  const metrics = normalizeChartFields(p.metrics)
  const explicit = Array.isArray(p.columns)
    ? p.columns.map((c) => String(c ?? '').trim()).filter(Boolean)
    : []
  const columns = explicit.length
    ? explicit
    : [...dimensions.map((d) => d.field), ...metrics.map((m) => m.field)]
  const rows = Array.isArray(p.rows) ? (p.rows as unknown[][]) : []
  const topN = Number(p.top_n)
  return {
    chartType,
    dimensions,
    metrics,
    columns,
    rows,
    aggregated: !!p.aggregated,
    topN: Number.isFinite(topN) && topN > 0 ? topN : 0,
  }
}

/** intent 块：{ intent } 为后端查询 AST */
export interface IntentPayload {
  intent?: QueryRequestVO | null
}

/** reasoning 块：{ reason_trace } */
export interface ReasoningPayload {
  reason_trace?: ReasonTraceVO | null
}

/** sql 块（旧契约）：单条翻译结果 + 说明 + JOIN 计划 + 参数 */
export interface SqlPayload {
  sql?: string
  explanation?: string
  joins?: JoinInfoVO[] | null
  params?: unknown[] | null
  datasource_id?: number
  datasource_name?: string
  datasource_type?: string
  source_table?: string
  class_name?: string
}

/**
 * SQL 步骤（#66 新 SQL 生成架构）。
 * 一次问答可能拆为多条子查询（多跳/跨源/聚合），每条含用途说明、SQL 文本与数据源标签。
 */
export interface SQLStepItem {
  /** 中文用途说明（如「查询 VIP 客户订单」） */
  purpose: string
  /** 该步生成的 SQL 文本 */
  sql: string
  /** 数据源展示名（el-tag 标签） */
  datasource_label: string
}

/** sql 块（#66 新契约）：{ steps } 多步查询语句集合 */
export interface SQLBlockPayload {
  steps: SQLStepItem[]
}

/**
 * 新 SQL 生成架构的消息级 meta（#66）。
 * 由 SSE meta 事件下发，描述本次问答的解析路径、LLM 调用次数、步骤数与涉及数据源。
 */
export interface SQLGenMeta {
  /** 解析路径：rule / llm / cache 等 */
  parse_path: string
  /** 本次问答调用 LLM 的次数 */
  llm_calls: number
  /** 生成的 SQL 步骤数 */
  step_count: number
  /** 涉及的数据源 id 列表 */
  datasource_ids: number[]
}

/** provenance 块：溯源摘要卡片数据 */
export interface ProvenancePayload {
  trace_id?: string
  nl_text?: string
  parse_path?: ParsePath
  status?: string
  datasource_id?: number
  datasource_name?: string
  datasource_type?: string
  source_table?: string
  class_name?: string
  joins?: JoinInfoVO[] | null
  translated_sql?: string
  explanation?: string
  result_count?: number
  execution_time_ms?: number
  /** 统一块契约（#56）：行数与耗时，优先于 result_count/execution_time_ms */
  row_count?: number
  duration_ms?: number
  error_message?: string
  created_at?: string
  /** SQL 指纹（#66）：翻译产出的 SQL 规范化后的哈希，用于「同问同 SQL」校验与去重 */
  sql_hash?: string
}

/**
 * 消息级 meta（#56 P1-1 / P3-5）：一次问答的遍历规模与跨源信息。
 * 由 SSE meta 事件下发，历史消息则来自 MessageVO.meta。
 */
export interface MessageMeta {
  parse_path?: ParsePath
  sub_intent_count?: number
  hop_count?: number
  llm_calls?: number
  cross_source?: boolean
  sources?: string[] | null
}

/** 澄清选项：后端为 { label, value, kind }，也兼容纯字符串 */
export interface ClarificationOptionVO {
  label?: string
  value?: string
  kind?: string
}

/** clarification 块：{ prompt, options[], free_input } */
export interface ClarificationPayload {
  prompt?: string
  options?: (ClarificationOptionVO | string)[] | null
  free_input?: boolean
}

// ---------------------------------------------------------------------------
// SSE 事件（POST /assistant/chat，text/event-stream）
// ---------------------------------------------------------------------------

/**
 * SSE 事件名：
 * meta / thinking / tool_call / reasoning / intent / sql / block /
 * clarification / provenance / done / error
 */
export type SSEEventName =
  | 'meta'
  | 'thinking'
  | 'tool_call'
  | 'reasoning'
  | 'intent'
  | 'sql'
  | 'block'
  | 'clarification'
  | 'provenance'
  | 'done'
  | 'error'
  | (string & {})

/** 解析后的单条 SSE 事件（data 已 JSON 解析；解析失败时保留 raw） */
export interface SSEEvent<T = unknown> {
  event: SSEEventName
  data: T | null
  /** 原始 data 文本（多行已用 \n 拼接） */
  raw: string
}

/** meta 事件：一次问答的链路元信息 */
export interface MetaEventPayload {
  trace_id?: string
  session_id?: string
  parse_path?: ParsePath
  skill_id?: string
  matched?: boolean
  /** #56 P1-1：遍历规模摘要 */
  sub_intent_count?: number
  hop_count?: number
  llm_calls?: number
  /** #56 P3-5：跨源综合信息 */
  cross_source?: boolean
  sources?: string[] | null
}

/** tool_call 事件：技能/工具调用（前向兼容，当前后端未发出） */
export interface ToolCallPayload {
  name?: string
  skill_id?: string
  args?: Record<string, unknown> | null
  status?: string
  elapsed_ms?: number
}

/** done 事件 */
export interface DoneEventPayload {
  trace_id?: string
  session_id?: string
  skill_id?: string
  parse_path?: ParsePath
  matched?: boolean
  ok?: boolean
}

/** error 事件 */
export interface ErrorEventPayload {
  message?: string
  code?: number
  trace_id?: string
}

// ---------------------------------------------------------------------------
// 会话与消息（GET /assistant/sessions、GET /assistant/sessions/:id/messages）
// ---------------------------------------------------------------------------

export interface SessionVO {
  id?: number
  session_id: string
  title?: string
  ontology_id?: number
  message_count?: number
  last_message_at?: string
  created_at?: string
  updated_at?: string
}

/** GET /assistant/sessions 为分页结构（middleware.SuccessPage） */
export interface SessionPageVO {
  records?: SessionVO[] | null
  total?: number
  current?: number
  size?: number
}

/** 历史消息视图（handler.assistantMessageView） */
export interface MessageVO {
  id?: number
  role: 'user' | 'assistant' | (string & {})
  content?: string
  blocks?: BlockVO[] | null
  skill_id?: string
  parse_path?: ParsePath
  trace_id?: string
  meta?: Record<string, unknown> | null
  created_at?: string
}

/** GET /assistant/sessions/:id/messages 解包后的响应体 */
export interface SessionMessagesVO {
  session_id?: string
  messages?: MessageVO[] | null
  count?: number
  trace_id?: string
}

// ---------------------------------------------------------------------------
// 前端聊天消息模型（由 SSE 事件增量装配）
// ---------------------------------------------------------------------------

/** 一次渲染块交互：澄清重发 / 打开溯源抽屉 */
export interface ChatAction {
  type: 'clarify-option' | 'open-trace' | 'resend'
  value: string
  /** open-trace 时携带 */
  traceId?: string
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  /** 纯文本内容（欢迎语 / text 块聚合 / 用户输入） */
  content: string
  /** 结构化渲染块，按到达顺序追加 */
  blocks: BlockVO[]
  time: string
  traceId: string
  parsePath: ParsePath | ''
  skillId: string
  matched: boolean
  /** 流式接收中 */
  streaming: boolean
  /** thinking 事件累积 */
  thinking: string[]
  /** tool_call 事件累积 */
  toolCalls: ToolCallPayload[]
  /** error 事件文案 */
  error: string
  /** 消息级 meta（遍历规模 / 跨源），live 由 meta 事件写入，历史由 MessageVO.meta 还原 */
  meta?: MessageMeta | null
  /** 来自历史会话（非本次流式产生） */
  history?: boolean
}

// ---------------------------------------------------------------------------
// 解析路径徽章元数据（含助手特有的 concept / clarification / meta）
// ---------------------------------------------------------------------------

export interface ParsePathBadge {
  label: string
  type: 'success' | 'primary' | 'warning' | 'info' | 'danger'
  /** 悬浮说明 */
  tip: string
}

/**
 * parse_path → 徽章。
 * rule=规则解析·零LLM / llm=LLM意图链 / cache=计划缓存 /
 * concept=概念知识库 / clarification=澄清 / meta=元信息技能 / local=前端离线兜底
 */
export const PARSE_PATH_BADGES: Record<string, ParsePathBadge> = {
  rule: { label: '规则解析 · 零 LLM', type: 'success', tip: '确定性规则命中，未调用大模型' },
  llm: { label: 'LLM 意图链', type: 'primary', tip: '由 Eino 意图链解析生成 AST' },
  cache: { label: '计划缓存 · 同问同 SQL', type: 'warning', tip: '命中意图缓存，直接复用已翻译计划' },
  concept: { label: '概念知识库', type: 'primary', tip: '命中概念知识库（concept_qa），返回权威口径' },
  clarification: { label: '澄清', type: 'info', tip: '未能确定意图，已返回澄清选项' },
  meta: { label: '元信息技能', type: 'info', tip: '词表/溯源等元信息类技能' },
  direct: {
    label: '直接查询',
    type: 'info',
    tip: '非自然语言入口（如 QueryLab 构建器 Dry-run / 执行），直接翻译本体 AST',
  },
  local: { label: '离线兜底', type: 'info', tip: '前端本地正则解析（服务端不可用）' },
}

export function parsePathBadge(path?: ParsePath | string | null): ParsePathBadge | null {
  const p = String(path ?? '').trim()
  if (!p) return null
  return PARSE_PATH_BADGES[p] ?? { label: p, type: 'info', tip: '未知解析路径' }
}

/** 技能 id → 中文短名（用于消息头部标记） */
export const SKILL_LABELS: Record<string, string> = {
  ontology_query: '本体查询',
  ontology_dry_run: '查询试算',
  ontology_explain: '查询解释',
  reason_explain: '推理说明',
  get_provenance: '数据溯源',
  list_vocabulary: '本体词表',
  concept_qa: '概念问答',
}
