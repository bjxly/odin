/** 物理 Schema（内省结果）与本体映射模型 */

export interface PhysColumn {
  name: string
  type: string
  nullable?: boolean
  comment?: string
  isPk?: boolean
}

export interface PhysTable {
  name: string
  schema?: string
  comment?: string
  columns: PhysColumn[]
}

export interface SourceSchema {
  sourceId: string
  tables: PhysTable[]
}

/** 单列映射：本体属性 → 物理列（可带枚举翻译） */
export interface PropertyMap {
  /** 本体属性 id，如 Customer.code */
  propertyId: string
  /** 物理列名 */
  column: string
  /** 物理值 → 本体枚举值 */
  valueMap?: Record<string, string>
  /** 后端存储的置信度（0~1）；缺省表示未记录 */
  confidence?: number | null
  /** 后端校验状态：validated / column_changed / column_deleted */
  validationStatus?: string
}

/** 一个类在某一源某张表上的映射 */
export interface ClassMap {
  id: string
  classId: string
  sourceId: string
  table: string
  schema?: string
  /** subject 业务键列 */
  keyColumns: string[]
  properties: PropertyMap[]
  filters?: string[]
  /** 是否为该类的主映射 */
  primary?: boolean
  /** 后端附带的数据源名称（跨源汇总展示用） */
  datasourceName?: string
  /** 后端附带的数据源类型（postgresql / mysql …） */
  datasourceType?: string
  /** 后端附带的本体类名 */
  className?: string
}

/** 映射编辑器中的一行：物理列 ↔ 本体属性（物理导向） */
export interface MappingEditorRow {
  /** 前端行唯一标识（仅用于列表渲染，不提交后端） */
  key: string
  /** 本体属性 name，如 code */
  propertyId: string
  /** 物理列名 */
  column: string
  /**
   * 置信度（0~1）：已保存行取后端存储值、建议行取建议值；
   * null / undefined 表示无存储值（手动行按 100% 展示）
   */
  confidence?: number | null
  /** 行来源：手动新增 / 接受建议 / 已保存映射 */
  origin?: 'manual' | 'suggest' | 'saved'
}

/** 映射编辑器「本体属性」下拉选项 */
export interface MappingPropertyOption {
  id: string
  label: string
  range?: string
  isKey?: boolean
  sensitive?: boolean
}

/** 映射编辑器「物理列」下拉选项 */
export interface MappingColumnOption {
  name: string
  type: string
  comment?: string
  isPk?: boolean
}

export interface MappingSnapshot {
  classMaps: ClassMap[]
}

export interface SuggestRow {
  column: string
  type: string
  comment?: string
  /** 建议的本体属性 id，空表示忽略 */
  propertyId: string | null
  confidence: number
  reason: string
}

export interface ValidateIssue {
  level: 'ok' | 'warn' | 'error'
  code: string
  message: string
}

export interface ValidateResult {
  ok: boolean
  issues: ValidateIssue[]
  sampleFillRate?: Record<string, number>
}
