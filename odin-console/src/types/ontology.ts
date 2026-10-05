import type { ClassMap } from './mapping'

export interface OntoProperty {
  id: string
  label: string
  range: string
  isKey?: boolean
  sensitive?: boolean
  synonyms?: string[]
}

export interface OntoClass {
  id: string
  /** 后端数值主键（用于 API CRUD） */
  numericId?: number
  label: string
  description?: string
  domain?: string
  synonyms?: string[]
  subclassOf?: string
  abstract?: boolean
  properties: OntoProperty[]
  mappedSources: string[]
  rules?: string[]
  /** 后端 ont_class.position_x/position_y（设计期默认坐标，仅非零时携带） */
  position?: LayoutPos
}

export interface OntoRelation {
  id: string
  /** 后端数值主键（用于 API CRUD） */
  numericId?: number
  label: string
  domain: string
  range: string
  cardinality: string
  inverse?: string
  crossSource?: boolean
}

export interface OntoRule {
  id: string
  /** 后端数值主键（用于 API CRUD） */
  numericId?: number
  label: string
  classId: string
  description: string
  condition: string
  /** 推理动作（来自 action_json） */
  action?: string
  enabled?: boolean
}

export type SourceType =
  | 'postgres'
  | 'dameng'
  | 'kingbase'
  | 'opengauss'
  | 'gaussdb'
  | 'mysql'
  | 'oracle'
  | 'sqlite'
  | 'sqlserver'
  | 'clickhouse'

export const SOURCE_TYPES: SourceType[] = [
  'postgres',
  'dameng',
  'kingbase',
  'opengauss',
  'gaussdb',
  'mysql',
  'oracle',
  'sqlite',
  'sqlserver',
  'clickhouse',
]

export interface SourceInfo {
  id: string
  /** 业务编码（如 erp_pg），仅展示用；id 为后端数字主键的字符串形式 */
  code?: string
  label?: string
  type: SourceType
  status: 'ok' | 'warn' | 'down' | 'unknown'
  host?: string
  port?: number
  database?: string
  user?: string
  defaultSchema?: string
  readOnly?: boolean
  tags?: string[]
  description?: string
  tables?: number
  lastTestAt?: number
  lastIntrospectAt?: number
  latencyMs?: number
  testMessage?: string
}

/** 画布坐标（持久化到 ont_definition.layout_json） */
export interface LayoutPos {
  x: number
  y: number
}

/** 本体变更日志 */
export type ChangeKind =
  | 'add_class'
  | 'update_class'
  | 'delete_class'
  | 'add_property'
  | 'update_property'
  | 'delete_property'
  | 'add_relation'
  | 'update_relation'
  | 'delete_relation'
  | 'add_source'
  | 'update_source'
  | 'delete_source'
  | 'test_source'
  | 'introspect_source'
  | 'update_layout'

export interface ChangeLogItem {
  id: string
  kind: ChangeKind
  target: string
  summary: string
  at: number
}

/** 单个本体记录：包含本体内容 + 映射 + 元数据 */
export interface OntologyRecord {
  id: string
  name: string
  description?: string
  version: number
  createdAt: number
  updatedAt: number
  classes: OntoClass[]
  relations: OntoRelation[]
  rules: OntoRule[]
  layout: Record<string, LayoutPos>
  classMaps: ClassMap[]
  changeLog: ChangeLogItem[]
}

/** 旧单本体快照（仅用于迁移） */
export interface OntologySnapshot {
  version: number
  classes: OntoClass[]
  relations: OntoRelation[]
  rules: OntoRule[]
  sources: SourceInfo[]
  layout: Record<string, LayoutPos>
}
