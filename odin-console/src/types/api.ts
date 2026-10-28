/** 后端 API 响应类型定义 */

import type { ColumnMetaVO, JoinInfoVO, ReasonTraceVO } from './query'

/** 分页结果 */
export interface PageResult<T> {
  records: T[]
  total: number
  current: number
  size: number
}

/** 数据源 VO */
export interface DataSourceVO {
  id: number
  code?: string
  name: string
  type: string
  host?: string
  port?: number
  /** 后端 JSON tag：database_name */
  database_name?: string
  /** 兼容旧字段名 */
  database?: string
  username?: string
  config_json?: string
  schema_name?: string
  status: 'active' | 'inactive' | 'error' | 'testing'
  tags?: string
  description?: string
  /** 后端列表接口返回的已缓存表数量 */
  table_count?: number
  created_at?: string
  updated_at?: string
}

/**
 * 后端 `diff.added_columns` / `diff.removed_columns` 的元素（handler.diffColumn）。
 * 实测：列级差异是**对象数组**，不是字符串数组。
 */
export interface SchemaDiffColumnVO {
  table: string
  column: string
}

/**
 * 后端 `diff.changed_columns` 的元素（handler.diffColumnChange）。
 * old_type / new_type 为内省得到的 data_type（粒度较粗，如 varchar / text / decimal，不含长度精度）。
 */
export interface SchemaDiffColumnChangeVO {
  table: string
  column: string
  old_type: string
  new_type: string
}

/**
 * `POST /datasources/:id/schema/refresh` 响应中 `data.diff` 的**后端原始结构**
 * （对应 handler.schemaDiff）。后端用 newSchemaDiff() 初始化五个切片，
 * 故正常情况下恒为数组（不会是 null）；此处仍按可空声明以便前端容错。
 */
export interface SchemaRawDiffVO {
  /** 新增表名，string[] */
  added_tables?: string[] | null
  /** 删除表名，string[] */
  removed_tables?: string[] | null
  /** 新增列，{table, column}[]（新增整表时其列不计入此项） */
  added_columns?: SchemaDiffColumnVO[] | null
  /** 删除列，{table, column}[]（删除整表时其列不计入此项） */
  removed_columns?: SchemaDiffColumnVO[] | null
  /** 类型变更列，{table, column, old_type, new_type}[] */
  changed_columns?: SchemaDiffColumnChangeVO[] | null
}

/**
 * `POST /datasources/:id/schema/refresh` 的 `data` 结构（axios 拦截器已剥掉 code/message 外壳）：
 * `{ schema: 刷新后的完整 schema, diff: 差异摘要 }`。
 */
export interface SchemaRefreshResultVO {
  schema: DataSourceSchemaVO
  diff: SchemaRawDiffVO
}

/** diff 各类计数；`total === 0` 表示本次刷新无结构变化 */
export interface SchemaDiffCountsVO {
  addedTables: number
  removedTables: number
  addedColumns: number
  removedColumns: number
  changedColumns: number
  total: number
}

/**
 * 前端归一化后的 Schema 差异（由 `utils/schemaDiff.ts#normalizeSchemaDiff` 产出）。
 * 既保留后端的对象数组（便于按表分组 / 定位），也提供 `table.column` 展示串与计数。
 */
export interface SchemaDiffVO {
  /** 新增表名（后端 added_tables） */
  addedTables: string[]
  /** 删除表名（后端 removed_tables） */
  removedTables: string[]
  /** 新增列（后端 added_columns，保留 table/column 结构） */
  addedColumns: SchemaDiffColumnVO[]
  /** 删除列（后端 removed_columns） */
  removedColumns: SchemaDiffColumnVO[]
  /** 类型变更列（后端 changed_columns） */
  changedColumns: SchemaDiffColumnChangeVO[]
  /** 展示串：新增列 `table.column` */
  added: string[]
  /** 展示串：删除列 `table.column` */
  removed: string[]
  /** 展示串：变更列 `table.column: old_type → new_type` */
  changed: string[]
  /** 计数汇总 */
  counts: SchemaDiffCountsVO
}

/** 数据源 Schema（内省结果） */
export interface DataSourceSchemaVO {
  datasource_id: number
  tables: TableVO[]
  table_count?: number
  column_count?: number
  cached?: boolean
  duration_ms?: number
  /** 刷新接口归一化后的结构差异摘要（仅 refreshSchema 返回时存在） */
  diff?: SchemaDiffVO
  stats?: { tables?: number; columns?: number; primary_keys?: number }
}

export interface TableVO {
  /** 后端 schema_table 主键（schemaTableView.id） */
  id?: number
  name: string
  schema?: string
  type?: string
  comment?: string
  columns: ColumnVO[]
  foreign_keys?: unknown[]
  indexes?: unknown[]
}

export interface ColumnVO {
  /** 后端 schema_column 主键（schemaColumnView.id） */
  id?: number
  name: string
  /** 后端内省字段：数据类型 */
  data_type?: string
  is_nullable?: boolean
  comment?: string
  is_primary_key?: boolean
  is_foreign_key?: boolean
  default_value?: string
  ordinal_position?: number
  /** 兼容旧字段名 */
  type?: string
  nullable?: boolean
}

/** 本体定义 VO */
export interface OntologyVO {
  id: number
  name: string
  description?: string
  layout_json?: string
  /** 后端为字符串版本号（默认 "1.0.0"），由后端管理 */
  version?: string
  status?: string
  created_at?: string
  updated_at?: string
}

/** 本体类 VO（对齐后端 classView = OntClass + property_count + parent_class_name） */
export interface OntClassVO {
  id: number
  ontology_id: number
  name: string
  label?: string
  description?: string
  /** normal / virtual，virtual 即虚拟（推理）类 */
  class_type?: string
  parent_class_id?: number
  /** 后端附带的父类名，避免前端二次解析 */
  parent_class_name?: string
  /** 后端附带的属性数量（不含属性明细） */
  property_count?: number
  icon?: string
  color?: string
  position_x?: number
  position_y?: number
  created_at?: string
  updated_at?: string
}

/** 本体属性 VO */
export interface OntPropertyVO {
  id: number
  ontology_id: number
  class_id: number
  name: string
  label?: string
  range: string
  is_key?: boolean
  sensitive?: boolean
  description?: string
  created_at?: string
  updated_at?: string
}

/** 本体关系 VO（后端 OntRelation 以 from_class_id/to_class_id 数值外键关联） */
export interface OntRelationVO {
  id: number
  ontology_id: number
  name: string
  label?: string
  from_class_id?: number
  to_class_id?: number
  relation_type?: string
  cardinality?: string
  description?: string
  join_condition_json?: string
  /** ListRelations 的 items 附带两端类名 */
  from_class_name?: string
  to_class_name?: string
  created_at?: string
  updated_at?: string
}

/** 推理规则 VO（后端 OntRule 为 ontology 级，无 class_id） */
export interface OntRuleVO {
  id: number
  ontology_id: number
  name: string
  description?: string
  rule_type?: string
  condition_json?: string
  action_json?: string
  priority?: number
  enabled?: boolean
  created_at?: string
  updated_at?: string
}

/** 单条属性映射（本体属性 → 物理列） */
export interface PropertyMappingVO {
  property_id?: number
  property_name?: string
  column_name?: string
  column_data_type?: string
  data_type?: string
  confidence?: number
  is_key?: boolean
  /** validated / column_changed / column_deleted 等 */
  validation_status?: string
}

/** 映射配置 VO */
export interface MappingVO {
  id: number
  ontology_id: number
  class_id: number
  datasource_id: number
  source_table: string
  source_schema?: string
  /** 后端返回的属性映射数组 */
  property_mappings?: PropertyMappingVO[]
  property_mappings_json?: string
  mapped_count?: number
  class_name?: string
  ontology_name?: string
  datasource_name?: string
  datasource_type?: string
  /** 兼容旧字段 */
  mappings_json?: string
  key_columns?: string
  filters?: string
  is_primary?: boolean
  status?: string
  created_at?: string
  updated_at?: string
}

/** 映射建议结果 */
export interface MappingSuggestionVO {
  column: string
  column_type: string
  column_comment?: string
  suggested_property_id?: number
  suggested_property_name?: string
  suggested_property_label?: string
  confidence: number
  reason?: string
}

/** 映射校验结果 */
export interface MappingValidationVO {
  valid: boolean
  issues: MappingIssueVO[]
}

export interface MappingIssueVO {
  level: 'error' | 'warning' | 'info'
  code: string
  message: string
  field?: string
}

/** 查询执行结果 VO（对齐后端 queryResponse） */
export interface QueryResultVO {
  /** 后端直接返回列名字符串数组 */
  columns: string[]
  /** 二维数组结果（与 columns 对应） */
  rows?: any[][]
  /** 对象数组结果（优先用于表格渲染） */
  records?: Record<string, any>[]
  row_count?: number
  duration_ms?: number
  sql?: string
  params?: any[]
  explanation?: string
  joins?: JoinInfoVO[] | null
  class_name?: string
  source_table?: string
  datasource_id?: number
  datasource_name?: string
  datasource_type?: string
  executed?: boolean
  truncated?: boolean
  /** 是否为聚合结果（存在 aggregates/group_by 时为 true） */
  aggregated?: boolean
  /** 列角色元信息，与 columns 顺序对齐（聚合查询下发） */
  column_meta?: ColumnMetaVO[] | null
  /** 后端真实推理轨迹（规则命中 / 虚拟展开 / 派生过滤） */
  reason_trace?: ReasonTraceVO | null
  /** 行级溯源：records 每行可携带 _src_datasource / _src_table */
  trace_id?: string
  request?: unknown
}

/** 查询 Dry-Run 结果（后端返回单条翻译结果，无 plans/warnings/errors） */
export interface QueryDryRunVO {
  executed?: boolean
  sql: string
  params?: any[]
  param_count?: number
  explanation?: string
  joins?: JoinInfoVO[] | null
  class_name?: string
  source_table?: string
  datasource_id?: number
  datasource_name?: string
  datasource_type?: string
  translate_ms?: number
  /** 是否为聚合查询 */
  aggregated?: boolean
  /** 列角色元信息（与 columns 顺序对齐） */
  column_meta?: ColumnMetaVO[] | null
  /** 后端真实推理轨迹 */
  reason_trace?: ReasonTraceVO | null
  trace_id?: string
  request?: unknown
}

/** 查询历史（对齐后端 GetQueryHistory item） */
export interface QueryHistoryVO {
  id: number
  ontology_id: number
  ontology_name?: string
  query_text?: string
  query_type?: string
  status: 'success' | 'error' | string
  result_count?: number
  execution_time_ms?: number
  error_message?: string
  created_at: string
}

/** 查询模板 */
export interface QueryTemplateVO {
  id: number
  ontology_id: number
  name: string
  description?: string
  query_json: string
  created_at?: string
  updated_at?: string
}
