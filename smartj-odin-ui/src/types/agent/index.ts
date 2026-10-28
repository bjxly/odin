// ==================== ID 类型约定 ====================
// 后端雪花 Long ID 经 Jackson 全局 ToStringSerializer 序列化为 JSON 字符串，
// 前端所有 ID 字段一律声明为 string，禁止 number（防 2^53 精度丢失）。

// ==================== 智能体（ag_agent） ====================

/** 智能体列表视图对象 — 对应 AgentVO */
export interface AgentVO {
  id: string
  agentCode: string
  agentName: string
  agentType: AgentType
  description?: string
  avatar?: string
  visibility: AgentVisibility
  ownerId: string
  ownerName?: string
  modelConfigId?: string
  workflowId?: string
  /** 0=草稿 1=已发布 2=已下线 */
  status: number
  remark?: string
  createTime: string
  updateTime?: string
}

/** 智能体类型 */
export type AgentType = 'chat' | 'react' | 'multi-agent' | 'workflow'

/** 智能体可见性 */
export type AgentVisibility = 'private' | 'team' | 'tenant'

/** 智能体详情视图对象 — 对应 AgentDetailVO（含各类绑定明细） */
export interface AgentDetailVO extends AgentVO {
  systemPrompt?: string
  openingMessage?: string
  maxIterations?: number
  temperature?: number
  topP?: number
  maxTokens?: number
  permissionMode?: string
  extraConfig?: Record<string, unknown>
  models?: AgentModelBindVO[]
  tools?: AgentToolBindVO[]
  skills?: AgentSkillBindVO[]
  subagents?: AgentSubagentBindVO[]
  knowledgeBases?: AgentKnowledgeBindVO[]
}

/** 智能体新增/编辑请求 — 对应 AgentDTO（绑定关系整体替换） */
export interface AgentDTO {
  id?: string
  agentCode: string
  agentName: string
  agentType?: AgentType
  description?: string
  systemPrompt?: string
  openingMessage?: string
  modelConfigId?: string
  maxIterations?: number
  temperature?: number
  topP?: number
  maxTokens?: number
  permissionMode?: string
  extraConfig?: Record<string, unknown>
  avatar?: string
  visibility?: AgentVisibility
  workflowId?: string
  remark?: string
  models?: AgentModelBindVO[]
  tools?: AgentToolBindVO[]
  skills?: AgentSkillBindVO[]
  subagents?: AgentSubagentBindVO[]
  knowledgeBases?: AgentKnowledgeBindVO[]
}

/** 智能体查询条件 — 对应 AgentQuery（BasePageQuery） */
export interface AgentQuery {
  current: number
  size: number
  agentCode?: string
  agentName?: string
  agentType?: AgentType
  visibility?: AgentVisibility
  ownerId?: string
  /** 0=草稿 1=已发布 2=已下线 */
  status?: number
}

/** 智能体发布请求 — 对应 PublishAgentRequest */
export interface PublishAgentRequest {
  changeLog?: string
}

/** 智能体场景模板 — 对应 AgentTemplateVO（GET /agent/agents/templates） */
export interface AgentTemplateVO {
  /** 模板编码（from-template 实例化入参） */
  templateCode: string
  templateName?: string
  /** 场景类别（写作/知识问答/智能问数据/工作流程/知识治理） */
  scene?: string
  agentType?: AgentType
  description?: string
  /** 展示图标（可选） */
  icon?: string
}

/** 场景模板实例化请求 — POST /agent/agents/from-template */
export interface FromTemplateRequest {
  templateCode: string
}

/** 智能体版本快照 — 对应 AgentVersionVO */
export interface AgentVersionVO {
  id: string
  agentId: string
  versionNo: number
  systemPrompt?: string
  extraConfig?: Record<string, unknown>
  modelConfigId?: string
  snapshot?: Record<string, unknown>
  publishedBy?: string
  publishTime?: string
  changeLog?: string
  /** 1=启用 0=禁用 */
  status: number
  createTime?: string
}

/** 智能体-模型绑定 — 对应 AgentModelBindVO */
export interface AgentModelBindVO {
  id?: string
  modelConfigId?: string
  modelDisplayName?: string
  /** primary=主模型 fallback=兜底模型 */
  role?: string
}

/** 智能体-工具绑定 — 对应 AgentToolBindVO */
export interface AgentToolBindVO {
  id?: string
  toolId?: string
  toolCode?: string
  toolName?: string
  /** builtin=内置 mcp=MCP服务 */
  toolType?: 'builtin' | 'mcp'
  /** 1=启用 0=禁用 */
  enabled?: number
  /** allow/ask/deny，空则用全局策略 */
  permissionOverride?: 'allow' | 'ask' | 'deny'
}

/** 智能体-技能绑定 — 对应 AgentSkillBindVO */
export interface AgentSkillBindVO {
  id?: string
  skillCode: string
  skillName?: string
  sortOrder?: number
}

/** 智能体-子智能体声明 — 对应 AgentSubagentBindVO */
export interface AgentSubagentBindVO {
  id?: string
  subAgentId?: string
  subAgentName?: string
  declaration?: Record<string, unknown>
  sortOrder?: number
}

/** 智能体-知识库绑定 — 对应 AgentKnowledgeBindVO */
export interface AgentKnowledgeBindVO {
  id?: string
  kbId: string
  kbName?: string
  retrievalConfig?: Record<string, unknown>
  sortOrder?: number
}

// ==================== 模型端点与模型清单（ag_model_provider / ag_model_config） ====================

/** 模型端点请求 — 对应 ModelProviderDTO */
export interface ModelProviderDTO {
  id?: string
  providerCode: string
  providerName: string
  baseUrl: string
  apiKey?: string
  headers?: Record<string, unknown>
  /** 1=默认 0=否 */
  isDefault?: number
}

/** 模型端点视图 — 对应 ModelProviderVO（apiKey 掩码） */
export interface ModelProviderVO {
  id: string
  providerCode: string
  providerName: string
  baseUrl: string
  apiKeyMasked?: string
  headers?: Record<string, unknown>
  isDefault?: number
  createTime?: string
  updateTime?: string
}

/** 模型端点查询条件 */
export interface ModelProviderQuery {
  current: number
  size: number
  providerCode?: string
  providerName?: string
  isDefault?: number
}

/** 模型配置请求 — 对应 ModelConfigDTO */
export interface ModelConfigDTO {
  id?: string
  providerId: string
  modelId: string
  /** chat/embedding */
  modelType?: string
  displayName?: string
  contextWindow?: number
  maxOutputTokens?: number
  supportStream?: number
  supportTools?: number
  supportVision?: number
  defaultParams?: Record<string, unknown>
  pricingInput?: number
  pricingOutput?: number
  /** 1=启用 0=禁用 */
  status?: number
  remark?: string
}

/** 模型配置视图 — 对应 ModelConfigVO */
export interface ModelConfigVO {
  id: string
  providerId: string
  providerCode?: string
  providerName?: string
  modelId: string
  fullModelId?: string
  modelType?: string
  displayName?: string
  contextWindow?: number
  maxOutputTokens?: number
  supportStream?: number
  supportTools?: number
  supportVision?: number
  defaultParams?: Record<string, unknown>
  pricingInput?: number
  pricingOutput?: number
  status: number
  remark?: string
  createTime?: string
}

/** 模型配置查询条件 */
export interface ModelConfigQuery {
  current: number
  size: number
  providerId?: string
  modelId?: string
  modelType?: string
  displayName?: string
  status?: number
}

/** 端点连通性测试结果 — 对应 ConnectTestVO */
export interface ConnectTestVO {
  success: boolean
  modelIds?: string[]
  latencyMs?: number
  errorMessage?: string
}

// ==================== 工具与 MCP（ag_tool / ag_mcp_server） ====================

/** 工具请求 — 对应 ToolDTO */
export interface ToolDTO {
  id?: string
  toolCode: string
  toolName: string
  toolType?: string
  groupCode?: string
  description?: string
  parametersSchema?: Record<string, unknown>
  isReadOnly?: number
  isDangerous?: number
  implRef?: string
  /** 1=启用 0=禁用 */
  status?: number
  remark?: string
}

/** 工具视图 — 对应 ToolVO */
export interface ToolVO {
  id: string
  toolCode: string
  toolName: string
  toolType?: string
  groupCode?: string
  description?: string
  parametersSchema?: Record<string, unknown>
  isReadOnly?: number
  isDangerous?: number
  implRef?: string
  status: number
  remark?: string
  createTime?: string
}

/** 工具查询条件 */
export interface ToolQuery {
  current: number
  size: number
  toolCode?: string
  toolName?: string
  toolType?: string
  groupCode?: string
  status?: number
}

/** MCP 服务请求 — 对应 McpServerDTO */
export interface McpServerDTO {
  id?: string
  serverCode: string
  serverName: string
  /** stdio/streamableHttp/sse */
  transport?: string
  command?: string
  args?: string[]
  endpoint?: string
  headers?: Record<string, unknown>
  env?: Record<string, unknown>
  timeoutMs?: number
  /** 1=启用 0=禁用 */
  status?: number
  remark?: string
}

/** MCP 服务视图 — 对应 McpServerVO */
export interface McpServerVO {
  id: string
  serverCode: string
  serverName: string
  transport?: string
  command?: string
  args?: string[]
  endpoint?: string
  headers?: Record<string, unknown>
  env?: Record<string, unknown>
  timeoutMs?: number
  status: number
  remark?: string
  createTime?: string
}

/** MCP 服务查询条件 */
export interface McpServerQuery {
  current: number
  size: number
  serverCode?: string
  serverName?: string
  transport?: string
  status?: number
}

// ==================== 技能（ag_skill） ====================

/** 技能请求 — 对应 SkillDTO */
export interface SkillDTO {
  id?: string
  skillCode: string
  skillName: string
  /** official/shared/private */
  scope?: string
  description?: string
  content?: string
  sourceRef?: string
  /** 1=启用 0=禁用 */
  status?: number
  remark?: string
}

/** 技能视图 — 对应 SkillVO */
export interface SkillVO {
  id: string
  skillCode: string
  skillName: string
  scope?: string
  description?: string
  content?: string
  sourceRef?: string
  status: number
  remark?: string
  createTime?: string
  updateTime?: string
}

/** 技能查询条件 */
export interface SkillQuery {
  current: number
  size: number
  skillCode?: string
  skillName?: string
  scope?: string
  status?: number
}

// ==================== 知识库（ag_knowledge_base / ag_knowledge_document） ====================

/** 知识库请求 — 对应 KnowledgeBaseDTO */
export interface KnowledgeBaseDTO {
  id?: string
  kbCode: string
  kbName: string
  description?: string
  embeddingModelConfigId?: string
  embeddingDim?: number
  chunkSize?: number
  chunkOverlap?: number
  retrievalConfig?: Record<string, unknown>
  /** 1=启用 0=禁用 */
  status?: number
  remark?: string
}

/** 知识库视图 — 对应 KnowledgeBaseVO */
export interface KnowledgeBaseVO {
  id: string
  kbCode: string
  kbName: string
  description?: string
  embeddingModelConfigId?: string
  embeddingModelName?: string
  embeddingDim?: number
  chunkSize?: number
  chunkOverlap?: number
  retrievalConfig?: Record<string, unknown>
  milvusCollection?: string
  docCount?: number
  chunkCount?: number
  status: number
  remark?: string
  createTime?: string
  updateTime?: string
}

/** 知识库查询条件 */
export interface KnowledgeBaseQuery {
  current: number
  size: number
  kbCode?: string
  kbName?: string
  status?: number
}

/** 知识库文档上传请求 — 对应 KnowledgeDocumentDTO */
export interface KnowledgeDocumentDTO {
  docName: string
  /** upload/url/text */
  sourceType?: string
  fileKey?: string
  sourceUrl?: string
  textContent?: string
  remark?: string
}

/** 知识库文档视图 — 对应 KnowledgeDocumentVO */
export interface KnowledgeDocumentVO {
  id: string
  kbId: string
  docName: string
  sourceType?: string
  fileKey?: string
  charCount?: number
  chunkCount?: number
  /** pending/parsing/success/failed */
  parseStatus?: string
  parseError?: string
  status: number
  createTime?: string
}

/** 知识库文档查询条件 */
export interface KnowledgeDocumentQuery {
  current: number
  size: number
  kbId?: string
  docName?: string
  sourceType?: string
  parseStatus?: string
  status?: number
}

/** 文档上传结果 — 对应 DocumentUploadVO */
export interface DocumentUploadVO {
  documentId: string
  docName: string
  parseStatus?: string
  message?: string
}

/** 重建索引请求 — 对应 ReindexRequest */
export interface ReindexRequest {
  documentIds?: string[]
}

// ==================== 工作流（ag_workflow / ag_workflow_execution） ====================

/** 工作流请求 — 对应 WorkflowDTO（graph 为 DAG JSON） */
export interface WorkflowDTO {
  id?: string
  workflowCode: string
  workflowName: string
  description?: string
  graph?: Record<string, unknown>
  /** api/manual/schedule 逗号分隔 */
  triggerTypes?: string
  remark?: string
}

/** 工作流视图 — 对应 WorkflowVO */
export interface WorkflowVO {
  id: string
  workflowCode: string
  workflowName: string
  description?: string
  graph?: Record<string, unknown>
  triggerTypes?: string
  /** draft/published/offline */
  workflowStatus?: string
  currentVersion?: number
  status: number
  remark?: string
  createTime?: string
  updateTime?: string
}

/** 工作流查询条件 */
export interface WorkflowQuery {
  current: number
  size: number
  workflowCode?: string
  workflowName?: string
  workflowStatus?: string
  triggerType?: string
  status?: number
}

/** 工作流触发请求 — 对应 WorkflowRunRequest */
export interface WorkflowRunRequest {
  inputData?: Record<string, unknown>
}

/** 工作流执行记录 — 对应 WorkflowExecutionVO */
export interface WorkflowExecutionVO {
  id: string
  workflowId: string
  workflowName?: string
  runCode: string
  triggerType?: string
  inputData?: Record<string, unknown>
  outputData?: Record<string, unknown>
  /** running/success/failed/canceled */
  execStatus?: string
  nodeStates?: Record<string, unknown>
  errorMessage?: string
  startTime?: string
  finishTime?: string
  durationMs?: number
  totalTokens?: number
  costAmount?: number
  createTime?: string
}

/** 工作流执行记录查询条件 */
export interface WorkflowExecutionQuery {
  current: number
  size: number
  workflowId?: string
  runCode?: string
  triggerType?: string
  execStatus?: string
  beginTime?: string
  endTime?: string
}

/** 工作流节点事件（SSE） — 对应 WorkflowNodeEventVO */
export interface WorkflowNodeEvent {
  runCode: string
  nodeId: string
  nodeType?: string
  nodeName?: string
  /** node_start/node_end/workflow_end 等 */
  eventType?: string
  output?: Record<string, unknown>
  errorMessage?: string
  latencyMs?: number
  eventTime?: string
}

// ==================== 会话与消息（ag_session / ag_message） ====================

/** 流式对话请求 — 对应 ChatRequest */
export interface ChatRequest {
  message: string
  attachments?: Record<string, unknown>[]
  metadata?: Record<string, unknown>
}

/** HITL 审批决策请求 — 对应 ConfirmRequest */
export interface ConfirmRequest {
  toolCallId: string
  /** approved=批准 denied=拒绝 */
  decision: 'approved' | 'denied'
  comment?: string
}

/** 消息反馈请求 — 对应 FeedbackRequest */
export interface FeedbackRequest {
  /** 1=踩 2=中立 3=赞 */
  rating: number
  comment?: string
}

/** 会话视图 — 对应 SessionVO */
export interface SessionVO {
  id: string
  sessionCode: string
  agentId: string
  agentCode?: string
  agentName?: string
  agentVersionNo?: number
  userId: string
  title?: string
  lastMessageTime?: string
  messageCount?: number
  summary?: string
  /** active=活跃 archived=已归档 */
  sessionStatus?: string
  createTime?: string
}

/** 会话查询条件 — 对应 SessionQuery（管理视角） */
export interface SessionQuery {
  current: number
  size: number
  agentId?: string
  userId?: string
  title?: string
  sessionStatus?: string
  beginTime?: string
  endTime?: string
}

/** 对话消息视图 — 对应 MessageVO */
export interface MessageVO {
  id: string
  sessionId: string
  msgId?: string
  /** user/assistant/system/tool */
  role: string
  contentBlocks?: Record<string, unknown>[]
  generateReason?: string
  toolCallId?: string
  tokenInput?: number
  tokenOutput?: number
  costAmount?: number
  durationMs?: number
  /** 1=踩 2=中立 3=赞，未反馈为空 */
  feedbackRating?: number
  createTime?: string
}

// ==================== 发布与 API Key（ag_publish / ag_api_key） ====================

/** 发布配置请求 — 对应 PublishDTO */
export interface PublishDTO {
  id?: string
  agentId: string
  agentVersionNo?: number
  /** webpage/iframe/api/webhook（后端 CHECK 枚举） */
  channel?: string
  /** token/public/auth（后端 CHECK 枚举） */
  accessMode?: string
  /** 明文令牌（仅创建/重置时提交，服务端仅存哈希） */
  token?: string
  expireTime?: string
  rateLimitQps?: number
  welcomeMessage?: string
  styleConfig?: Record<string, unknown>
  /** 1=启用 0=禁用 */
  status?: number
  remark?: string
}

/** 发布配置视图 — 对应 PublishVO */
export interface PublishVO {
  id: string
  agentId: string
  agentName?: string
  agentVersionNo?: number
  channel?: string
  accessMode?: string
  tokenPrefix?: string
  tokenConfigured?: boolean
  expireTime?: string
  rateLimitQps?: number
  welcomeMessage?: string
  styleConfig?: Record<string, unknown>
  shareUrl?: string
  status: number
  remark?: string
  createTime?: string
  updateTime?: string
}

/** 发布配置查询条件 */
export interface PublishQuery {
  current: number
  size: number
  agentId?: string
  channel?: string
  accessMode?: string
  status?: number
}

/** API Key 请求 — 对应 ApiKeyDTO */
export interface ApiKeyDTO {
  id?: string
  keyName: string
  agentIds?: string[]
  expiresAt?: string
  rateLimitQps?: number
  /** 1=启用 0=吊销 */
  status?: number
  remark?: string
}

/** API Key 视图 — 对应 ApiKeyVO（仅前缀可见） */
export interface ApiKeyVO {
  id: string
  keyName: string
  keyPrefix?: string
  agentIds?: string[]
  expiresAt?: string
  lastUsedTime?: string
  rateLimitQps?: number
  status: number
  remark?: string
  createTime?: string
  updateTime?: string
}

/** API Key 创建结果 — 对应 CreatedApiKeyVO（明文仅返回一次） */
export interface CreatedApiKeyVO extends ApiKeyVO {
  apiKey: string
}

/** API Key 查询条件 */
export interface ApiKeyQuery {
  current: number
  size: number
  keyName?: string
  keyPrefix?: string
  status?: number
}

// ==================== 用量统计（ag_usage_daily） ====================

/** 用量查询条件 — 对应 UsageStatQuery */
export interface UsageStatQuery {
  beginDate?: string
  endDate?: string
  agentId?: string
  modelConfigId?: string
}

/** 用量汇总 — 对应 UsageStatVO */
export interface UsageStatVO {
  requestCount: number
  tokenInput: number
  tokenOutput: number
  totalTokens: number
  totalCost: number
  avgLatencyMs?: number
  errorCount?: number
  errorRate?: number
}

/** 用量趋势（按日） — 对应 UsageTrendVO */
export interface UsageTrendVO {
  statDate: string
  requestCount: number
  tokenInput: number
  tokenOutput: number
  totalCost: number
  avgLatencyMs?: number
  errorCount?: number
}

/** Top 智能体用量 — 对应 GET /agent/usage/top-agents */
export interface UsageTopAgentVO {
  agentId: string
  agentName?: string
  requestCount: number
  totalTokens: number
  totalCost: number
}

// ==================== 执行轨迹（ag_agent_trace） ====================

/** 轨迹查询条件 — 对应 AgentTraceQuery */
export interface AgentTraceQuery {
  current: number
  size: number
  agentId?: string
  sessionId?: string
  traceId?: string
  /** llm/tool/retrieval/agent */
  nodeType?: string
  traceStatus?: string
  beginTime?: string
  endTime?: string
}

/** 执行轨迹 — 对应 AgentTraceVO */
export interface AgentTraceVO {
  id: string
  agentId?: string
  agentName?: string
  sessionId?: string
  traceId: string
  parentTraceId?: string
  nodeType?: string
  nodeName?: string
  inputSummary?: string
  outputSummary?: string
  /** success/failed/running */
  traceStatus?: string
  latencyMs?: number
  tokenInput?: number
  tokenOutput?: number
  costAmount?: number
  errorCode?: string
  errorMessage?: string
  requestJson?: Record<string, unknown>
  responseJson?: Record<string, unknown>
  createTime?: string
  /** 轨迹树子节点（GET /agent/traces/{traceId}/tree 嵌套返回） */
  children?: AgentTraceVO[]
}

// ==================== SSE 事件载荷 ====================

/** SSE 事件类型 */
export type AgentEventType =
  | 'session_created'
  | 'text_delta'
  | 'thinking'
  | 'tool_call_start'
  | 'tool_result'
  | 'require_user_confirm'
  | 'model_call'
  | 'done'
  | 'error'

/** 工具调用开始事件数据 */
export interface ToolCallStartData {
  toolCallId: string
  toolName?: string
  input?: Record<string, unknown>
}

/** 工具调用结果事件数据 */
export interface ToolResultData {
  toolCallId: string
  toolName?: string
  result?: unknown
  isError?: boolean
  durationMs?: number
}

/** HITL 审批请求事件数据 */
export interface RequireUserConfirmData {
  toolCallId: string
  toolName?: string
  input?: Record<string, unknown>
  reason?: string
  isDangerous?: boolean
}

/** 模型调用事件数据 */
export interface ModelCallData {
  modelId?: string
  tokenInput?: number
  tokenOutput?: number
  costAmount?: number
  durationMs?: number
}

/** 用量统计事件数据 */
export interface UsageData {
  tokenInput?: number
  tokenOutput?: number
  totalTokens?: number
  costAmount?: number
}

/** SSE 错误事件数据 */
export interface AgentErrorData {
  code?: string
  message?: string
}

/**
 * 流式对话 SSE 事件载荷。
 * 后端 SseEventBridge 以 `data:` 行下发 JSON，`event` 字段为事件类型判别器。
 */
export interface AgentEventPayload {
  event: AgentEventType
  /** 会话编码（运行时 sessionId） */
  sessionId?: string
  /** 消息ID */
  messageId?: string
  /** text_delta / thinking 的文本增量 */
  delta?: string
  /** tool_call_start 数据 */
  toolCall?: ToolCallStartData
  /** tool_result 数据 */
  toolResult?: ToolResultData
  /** require_user_confirm 数据 */
  confirm?: RequireUserConfirmData
  /** model_call 数据 */
  modelCall?: ModelCallData
  /** done 事件附带的用量统计 */
  usage?: UsageData
  /** error 事件数据 */
  error?: AgentErrorData
  /** 兜底：允许后端扩展字段 */
  [key: string]: unknown
}

/** useAgentChat 错误对象 */
export interface AgentChatError {
  /** network/http/parse/abort/business */
  type: 'network' | 'http' | 'parse' | 'abort' | 'business'
  message: string
  status?: number
  code?: string
}

/** 对外分享页配置（GET /v1/agent/share/{token}/config） */
export interface ShareConfigVO {
  agentName?: string
  avatar?: string
  welcomeMessage?: string
  styleConfig?: Record<string, unknown>
  suggestedQuestions?: string[]
  /** 访问模式（阶段 C 契约）：token=令牌访问 public=开放访问 auth=需登录 */
  accessMode?: 'token' | 'public' | 'auth'
  /** 是否需登录后使用（阶段 C 契约：auth 模式为 true，前端渲染前据此引导登录） */
  requireLogin?: boolean
}
