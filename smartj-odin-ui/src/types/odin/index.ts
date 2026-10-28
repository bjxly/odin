/** 数据源类型 */
export type DatasourceType = 'mysql' | 'postgresql' | 'dameng' | 'kingbase' | 'gaussdb' | 'oracle' | 'sqlserver' | 'sqlite' | 'duckdb_file'

/** 数据源 VO */
export interface DatasourceVO {
  id: string
  code: string
  name: string
  type: string
  host?: string
  port?: number
  dbName?: string
  schemaName?: string
  username?: string
  jdbcUrl?: string
  options?: string
  tags?: string
  poolMaxSize?: number
  poolMinIdle?: number
  poolMaxLifeS?: number
  connTimeoutS?: number
  queryTimeoutS?: number
  status: number
  remark?: string
  createTime: string
  updateTime?: string
}

/** 数据源 DTO（新增/编辑） */
export interface DatasourceDTO {
  id?: string
  code: string
  name: string
  type: string
  host?: string
  port?: number
  dbName?: string
  schemaName?: string
  username?: string
  password?: string
  jdbcUrl?: string
  options?: string
  tags?: string
  poolMaxSize?: number
  poolMinIdle?: number
  poolMaxLifeS?: number
  connTimeoutS?: number
  queryTimeoutS?: number
  status?: number
  remark?: string
}

/** 数据源查询条件 */
export interface DatasourceQuery {
  current: number
  size: number
  code?: string
  name?: string
  type?: string
  status?: number
}

/** 连接测试结果 */
export interface DatasourceTestResult {
  success: boolean
  latencyMs?: number
  dbVersion?: string
  errorMessage?: string
}

/** 表元数据 VO */
export interface SchemaTableVO {
  id: string
  schemaName?: string
  tableName: string
  tableType?: string
  comment?: string
  rowEstimate?: number
  lastSyncTime?: string
  fingerprint?: string
  status?: number
}

/** 列元数据 VO */
export interface SchemaColumnVO {
  id: string
  columnName: string
  dataType: string
  canonicalType?: string
  nullable?: number
  comment?: string
  ordinalPosition?: number
  isPrimaryKey?: number
  defaultValue?: string
  maxLength?: number
  numericPrecision?: number
  numericScale?: number
  isUnique?: number
  synonyms?: string
  status?: number
}

// Re-export sub-module types
export * from './ontology'
export * from './mapping'
export * from './query'
export * from './assistant'
export * from './insight'
export * from './connector'
export * from './layout'
export * from './trace'
export * from './workspace'
export * from './reason'
export * from './block'
