/** 消息块类型 */
export type BlockType =
  | 'text'
  | 'intent'
  | 'reasoning'
  | 'sql'
  | 'table'
  | 'chart'
  | 'provenance'
  | 'clarification'

/** 通用消息块 */
export interface MessageBlock<P = unknown> {
  type: BlockType
  title?: string
  payload?: P
  traceRef?: string
}

/** text 块 payload */
export interface TextBlockPayload {
  text?: string
  /** 流式聚合用的分块ID（同 blockId 的增量累加到同一 text 块） */
  blockId?: string
  /** 该文本块是否仍在流式生成中（驱动打字机光标） */
  streaming?: boolean
}

/** intent 块 payload */
export interface IntentBlockPayload {
  classCode?: string
  className?: string
  properties?: string[]
  filters?: Array<{ property: string; op: string; value?: unknown }>
  joins?: Array<{ relation: string; condition: string }>
}

/** sql 块 payload */
export interface SqlBlockPayload {
  sql?: string
  explanation?: string
  datasourceName?: string
  sourceTable?: string
  steps?: SqlStepItem[]
}

/** SQL 步骤（多步查询） */
export interface SqlStepItem {
  purpose: string
  sql: string
  datasourceLabel: string
}

/** table 块 payload */
export interface TableBlockPayload {
  columns?: string[]
  rows?: unknown[][]
  records?: Record<string, unknown>[]
  rowCount?: number
  aggregated?: boolean
}

/** chart 块 payload */
export interface ChartBlockPayload {
  chartType?: 'bar' | 'line' | 'pie' | 'metric'
  dimensions?: Array<{ field: string; label: string }>
  metrics?: Array<{ field: string; label: string; agg?: string }>
  rows?: unknown[][]
  columns?: string[]
  aggregated?: boolean
  /** 新模式：引用 msg.datasets 中的数据集ID（DatasetChartCard 消费）；旧模式保留 columns/rows */
  datasetId?: string
}

/** 数据集快照（与后端 DatasetSnapshot 对齐） */
export interface QueryDataset {
  id: string
  title?: string
  /** 表头列名 */
  columns: string[]
  /** 仅数据行，不含表头，单元格一律字符串（避免 JS 大数精度与二次格式化） */
  rows: string[][]
  rowCount: number
  truncated?: boolean
  physicalSql?: string
  /** 数据来源：query=原始查询结果（默认），table=正文分析结论表（后端 Markdown 解析合成） */
  source?: 'query' | 'table'
}

/** 图表推荐（模型给出，前端必须严格校验后采用） */
export interface ChartHint {
  dataId: string
  type: 'bar' | 'line' | 'pie'
  xIndex: number
  yIndices: number[]
  /** 转置提示：类别在表头上（如本月/上月对比），成图前先转置表数据集视图 */
  transposed?: boolean
}

/** 图型 */
export type QueryChartType = 'bar' | 'line' | 'pie'

/** 同格多行收敛方式；count 统计行数，其余作用于指标列 */
export type QueryAggMethod = 'sum' | 'avg' | 'max' | 'min' | 'count'

/** 图表选择状态（卡片内部交互模型） */
export interface ChartSelection {
  type: QueryChartType
  x: number
  ys: number[]
  horizontal: boolean
  /** 系列维度列索引；null 表示不拆分。饼图不支持 pivot，必须为 null。 */
  seriesDim: number | null
  agg: QueryAggMethod
}

/** chart 块 payload 扩展字段见 ChartBlockPayload.datasetId；转置状态由卡片内部维护（hint 或用户切换） */

/** provenance 块 payload */
export interface ProvenanceBlockPayload {
  traceId?: string
  nlText?: string
  parsePath?: string
  status?: string
  datasourceName?: string
  sourceTable?: string
  className?: string
  translatedSql?: string
  rowCount?: number
  durationMs?: number
  errorMessage?: string
}

/** clarification 块 payload */
export interface ClarificationBlockPayload {
  prompt?: string
  options?: Array<{ label?: string; value?: string; kind?: string } | string>
  freeInput?: boolean
}

/** 助手消息（前端聊天视图模型） */
export interface AssistantMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  blocks: MessageBlock[]
  time: string
  traceId?: string
  parsePath?: string
  skillId?: string
  streaming: boolean
  error?: string
  /** 本轮工具产出的数据集快照（供图表卡片消费） */
  datasets?: QueryDataset[]
  /** 轮末图表推荐，按 dataId 关联 */
  chartHints?: Record<string, ChartHint>
}

/** 聊天会话 */
export interface ChatSession {
  /** 会话主键（后端 AssistantSessionVO.id，雪花 Long 序列化为字符串） */
  id: string
  title?: string
  ontologyId?: string
  messageCount?: number
  lastMessageAt?: string
  createTime: string
  updateTime?: string
}

/** SSE 事件类型 */
export type SSEEventName =
  | 'meta'
  | 'intent'
  | 'reasoning'
  | 'sql'
  | 'block'
  | 'clarification'
  | 'provenance'
  | 'draft'
  | 'done'
  | 'error'
  // 流式与数据集扩展事件（向后兼容，旧前端忽略未知事件）
  | 'text_delta'
  | 'text_end'
  | 'thinking_delta'
  | 'thinking_end'
  | 'tool_call'
  | 'dataset'
  | 'chart_hint'
  | (string & {})

/** 解析后的 SSE 事件 */
export interface SSEEvent<T = unknown> {
  event: SSEEventName
  data: T | null
  raw: string
}
