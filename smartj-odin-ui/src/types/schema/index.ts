// ==================== Schema Tool ====================

/** Schema工具视图 */
export interface SchemaToolVO {
  id: number | string
  toolId: string
  toolName: string
  description?: string
  category: string
  targetSystem?: string
  sourceType: string
  sourceConfig?: string
  currentVersion: string
  lifecycleStatus: string // draft/active/deprecated
  toolSynonyms?: string[]
  triggerPatterns?: string[]
  exampleMessages?: ExampleMessage[]
  tags?: string[]
  reviewStatus: string // pending/approved/rejected
  fields?: SchemaFieldVO[]
  fieldCount?: number
  tenantId?: number
  status?: number
  createBy?: string
  createTime?: string
  updateBy?: string
  updateTime?: string
  remark?: string
}

/** Schema工具摘要（列表用） */
export interface SchemaToolSummaryVO {
  id: number | string
  toolId: string
  toolName: string
  description?: string
  category: string
  sourceType: string
  currentVersion: string
  lifecycleStatus: string
  reviewStatus: string
  fieldCount?: number
  createBy?: string
  createTime?: string
  updateTime?: string
}

/** Schema工具创建/编辑DTO */
export interface SchemaToolDTO {
  id?: number | string
  toolId: string
  toolName: string
  description?: string
  category?: string
  targetSystem?: string
  sourceType?: string
  sourceConfig?: string
  toolSynonyms?: string[]
  triggerPatterns?: string[]
  exampleMessages?: ExampleMessage[]
  tags?: string[]
  fields?: SchemaFieldDTO[]
  changeDescription?: string
}

/** Schema工具分页查询 */
export interface SchemaToolQuery {
  current: number
  size: number
  toolName?: string
  toolId?: string
  category?: string
  lifecycleStatus?: string
  sourceType?: string
  reviewStatus?: string
  status?: number
}

// ==================== Schema Field ====================

/** Schema字段视图（三组属性完整覆盖） */
export interface SchemaFieldVO {
  id?: number | string
  toolId?: number | string
  parentId?: number | string
  fieldPath: string
  fieldLevel: number
  fieldKey: string
  fieldLabel: string
  // ========== 表单渲染属性 ==========
  widget: WidgetType
  fieldGroup?: string
  displayOrder: number
  placeholder?: string
  helpText?: string
  isReadonly?: boolean
  isHidden?: boolean
  conditionalShow?: ConditionalShowRule
  span?: number
  // ========== AI抽取属性 ==========
  aiDescription?: string
  synonyms?: string[]
  extractionHints?: string
  aliases?: FieldAlias[]
  patternHints?: string
  relatedFields?: string[]
  aiFillable?: boolean
  defaultReasoning?: string
  // ========== 数据校验属性 ==========
  dataType: DataType
  isRequired?: boolean
  enumValues?: any[]
  enumLabels?: Record<string, string>
  formatConstraint?: string
  fieldPattern?: string
  minLength?: number
  maxLength?: number
  minValue?: number
  maxValue?: number
  defaultValue?: string
  transformRule?: TransformRule
  minItems?: number
  maxItems?: number
  addButtonText?: string
  // ========== 子字段（嵌套） ==========
  children?: SchemaFieldVO[]
}

/** Schema字段DTO */
export interface SchemaFieldDTO {
  id?: number | string
  parentId?: number | string
  fieldPath: string
  fieldLevel: number
  fieldKey: string
  fieldLabel: string
  // 表单渲染
  widget?: WidgetType
  fieldGroup?: string
  displayOrder?: number
  placeholder?: string
  helpText?: string
  isReadonly?: boolean
  isHidden?: boolean
  conditionalShow?: ConditionalShowRule
  span?: number
  // AI抽取
  aiDescription?: string
  synonyms?: string[]
  extractionHints?: string
  aliases?: FieldAlias[]
  patternHints?: string
  relatedFields?: string[]
  aiFillable?: boolean
  defaultReasoning?: string
  // 数据校验
  dataType: DataType
  isRequired?: boolean
  enumValues?: any[]
  enumLabels?: Record<string, string>
  formatConstraint?: string
  fieldPattern?: string
  minLength?: number
  maxLength?: number
  minValue?: number
  maxValue?: number
  defaultValue?: string
  transformRule?: TransformRule
  minItems?: number
  maxItems?: number
  addButtonText?: string
}

// ==================== Schema Version ====================

export interface SchemaVersionVO {
  id: number | string
  toolId: number | string
  versionNumber: string
  versionType?: string // major/minor/patch
  changeDescription?: string
  snapshot?: string
  impactAnalysis?: SchemaImpactAnalysisVO
  isActive: boolean
  reviewStatus: string
  reviewedBy?: string
  reviewedTime?: string
  reviewComment?: string
  publishedBy?: string
  publishedTime?: string
  createBy?: string
  createTime?: string
}

// ==================== Form Render Config ====================

/** 表单渲染配置（自包含，可直接传给SmSchemaForm组件） */
export interface FormRenderConfigVO {
  toolId: string
  toolName: string
  description?: string
  fields: SchemaFieldVO[]
}

// ==================== AI Prompt ====================

export interface AiPromptVO {
  toolId: string
  toolName: string
  prompt: string
  intentHint?: string
  extractionInstructions?: string[]
  fewShotExamples?: string[]
}

// ==================== Schema Validate ====================

export interface SchemaValidateDTO {
  toolId: string
  data: Record<string, unknown>
  applyTransform?: boolean
}

export interface SchemaValidationError {
  fieldKey: string
  fieldLabel: string
  errorMessage: string
}

export interface SchemaValidateResultVO {
  valid: boolean
  errors?: SchemaValidationError[]
  transformedData?: Record<string, unknown>
}

// ==================== Schema Match ====================

export interface SchemaMatchDTO {
  message: string
  contactId?: number
  topN?: number
}

export interface SchemaMatchResultVO {
  durationMs: number
  candidates: MatchCandidate[]
  matchStage: string
}

export interface MatchCandidate {
  toolId: number
  toolIdStr: string
  toolName: string
  score: number
  stage: string
  matchedKeywords?: string[]
}

// ==================== Schema Stats ====================

export interface SchemaStatsVO {
  totalTools: number
  activeTools: number
  draftTools: number
  deprecatedTools: number
  totalVersions: number
  totalMatches: number
  todayMatches: number
  matchSuccessRate: number
  categoryStats: CategoryStat[]
  matchTrend: DailyMatchStat[]
}

export interface CategoryStat {
  category: string
  count: number
}

export interface DailyMatchStat {
  date: string
  matchCount: number
  avgScore: number
}

// ==================== Schema Import ====================

export interface SchemaImportDTO {
  sourceType: string // mcp/openapi/database/json
  sourceConfig: string
  toolId?: string
  toolName?: string
  category?: string
  targetSystem?: string
}

export interface SchemaBatchImportDTO {
  sourceType?: string
  sourceConfig?: string
  imports: SchemaImportDTO[]
}

// ==================== Schema Impact Analysis ====================

export interface SchemaImpactAnalysisVO {
  versionType: string // major/minor/patch
  changedFields: string[]
  addedFields: string[]
  removedFields: string[]
  typeChangedFields: string[]
  affectedAreas: string[] // AI/Form/Validation/Approval
}

// ==================== Schema Preview ====================

export interface SchemaPreviewDTO {
  schemaId?: number
  fields?: SchemaFieldDTO[]
  mockData?: Record<string, unknown>
}

// ==================== Sub Types ====================

export interface ExampleMessage {
  message: string
  result: Record<string, unknown>
}

export interface FieldAlias {
  alias: string
  value: string
}

export interface ConditionalShowRule {
  logic: 'AND' | 'OR'
  rules: ConditionalRule[]
}

export interface ConditionalRule {
  field: string
  op: string // eq/ne/gt/lt/contains/in
  value: unknown
}

// ==================== Enums ====================

export type DataType = 'string' | 'integer' | 'number' | 'boolean' | 'array' | 'object'

export type WidgetType =
  | 'text'
  | 'number'
  | 'select'
  | 'date'
  | 'textarea'
  | 'phone'
  | 'email'
  | 'dynamic_table'
  | 'nested'
  | 'checkbox'

export type TransformRule =
  | 'trim'
  | 'uppercase'
  | 'lowercase'
  | 'phone_clean'
  | 'date_normalize'
  | 'amount_to_cents'

export type LifecycleStatus = 'draft' | 'active' | 'deprecated'

export type ReviewStatus = 'pending' | 'approved' | 'rejected'

// ==================== SmSchemaForm 独立组件类型 ====================

/** 自包含Schema JSON（可独立传输，不依赖后端接口） */
export interface FormSchema {
  toolId: string
  toolName: string
  description?: string
  fields: FormField[]
}

/** 字段定义（三组属性完整覆盖，供独立组件使用） */
export interface FormField {
  fieldKey: string
  fieldLabel: string
  parentId?: number | string
  fieldPath: string
  fieldLevel: number
  // 表单渲染
  widget: WidgetType
  fieldGroup?: string
  displayOrder: number
  placeholder?: string
  helpText?: string
  isReadonly?: boolean
  isHidden?: boolean
  conditionalShow?: ConditionalShowRule
  span?: number
  // 数据校验
  dataType: DataType
  isRequired?: boolean
  enumValues?: any[]
  enumLabels?: Record<string, string>
  formatConstraint?: string
  fieldPattern?: string
  minLength?: number
  maxLength?: number
  minValue?: number
  maxValue?: number
  defaultValue?: any
  transformRule?: TransformRule
  minItems?: number
  maxItems?: number
  addButtonText?: string
  // 子字段
  children?: FormField[]
}

// ==================== MCP ====================

export interface McpTool {
  name: string
  title?: string
  description: string
  inputSchema?: Record<string, unknown>
}
