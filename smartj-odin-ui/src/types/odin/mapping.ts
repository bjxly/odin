/** 数据映射配置 VO */
export interface MappingConfigVO {
  id: string
  ontologyId: string
  ontologyName?: string
  classId: string
  className?: string
  datasourceId: string
  datasourceName?: string
  tableName: string
  schemaName?: string
  columnMappings?: string
  filterCondition?: string
  priority?: number
  status: number
  remark?: string
  createTime: string
  updateTime?: string
}

/** 数据映射配置 DTO */
export interface MappingConfigDTO {
  id?: string
  ontologyId: string
  classId: string
  datasourceId: string
  tableName: string
  schemaName?: string
  columnMappings?: string
  filterCondition?: string
  priority?: number
  status?: number
  remark?: string
}

/** 映射查询条件 */
export interface MappingConfigQuery {
  current: number
  size: number
  ontologyId?: string
  classId?: string
  datasourceId?: string
  tableName?: string
  status?: number
}

/** AI 列映射建议 VO */
export interface MappingSuggestVO {
  columnName: string
  propertyName: string
  propertyCode: string
  confidence: number
  matchReason?: string
}

/**
 * 列映射规范行（column_mappings JSON 数组元素）。
 *
 * 以 propertyCode 为稳定连接键，propertyName 仅作展示快照。
 * 读取兼容历史键（property_name/column_name/validation_status 等），写入统一产规范键。
 */
export interface ColumnMappingRow {
  /** 本体属性编码（稳定连接键） */
  propertyCode: string
  /** 属性展示名快照（可选） */
  propertyName?: string
  /** 物理列名 */
  columnName: string
  /** 值转换：none/uppercase/lowercase/trim/split/json_extract */
  transform?: string
  /** 值映射表，如 {"M":"Male"}；可为对象或其 JSON 字符串 */
  valueMap?: Record<string, string> | string | null
  /** 置信度 (0~1) */
  confidence?: number
  /** 校验状态：validated/pending 等 */
  validationStatus?: string
}
