/** 本体定义 VO */
export interface OntologyVO {
  id: string
  code: string
  name: string
  description?: string
  version?: string
  status: number
  remark?: string
  createTime: string
  updateTime?: string
  classCount?: number
  relationCount?: number
  /** 助手预置问题列表 */
  presetQuestions?: string[]
}

/** 本体 DTO */
export interface OntologyDTO {
  id?: string
  code: string
  name: string
  description?: string
  version?: string
  status?: number
  remark?: string
  /** 助手预置问题列表 */
  presetQuestions?: string[]
}

/** 本体查询 */
export interface OntologyQuery {
  current: number
  size: number
  code?: string
  name?: string
  status?: number
}

/** 本体类 VO */
export interface OntClassVO {
  id: string
  ontologyId: string
  code: string
  name: string
  description?: string
  parentClassId?: string
  icon?: string
  sortOrder?: number
  identityKeys?: string
  status: number
  remark?: string
  createTime: string
  children?: OntClassVO[]
  propertyCount?: number
}

/** 本体类 DTO */
export interface OntClassDTO {
  id?: string
  ontologyId: string
  code: string
  name: string
  description?: string
  parentClassId?: string
  icon?: string
  sortOrder?: number
  identityKeys?: string
  status?: number
  remark?: string
}

/** 属性 VO */
export interface OntPropertyVO {
  id: string
  classId: string
  code: string
  name: string
  dataType: string
  description?: string
  required?: number
  defaultValue?: string
  enumValues?: string
  unit?: string
  sortOrder?: number
  status: number
  createTime: string
}

/** 属性 DTO */
export interface OntPropertyDTO {
  id?: string
  classId: string
  code: string
  name: string
  dataType: string
  description?: string
  required?: number
  defaultValue?: string
  enumValues?: string
  unit?: string
  sortOrder?: number
  status?: number
}

/** 关系 VO */
export interface OntRelationVO {
  id: string
  ontologyId: string
  code: string
  name: string
  description?: string
  sourceClassId: string
  targetClassId: string
  sourceClassName?: string
  targetClassName?: string
  cardinality: string
  joinKeys?: string
  inverseName?: string
  status: number
  remark?: string
  createTime: string
}

/** 关系 DTO */
export interface OntRelationDTO {
  id?: string
  ontologyId: string
  code: string
  name: string
  description?: string
  sourceClassId: string
  targetClassId: string
  cardinality: string
  joinKeys?: string
  inverseName?: string
  status?: number
  remark?: string
}

/** 规则 VO */
export interface OntRuleVO {
  id: string
  ontologyId: string
  code: string
  name: string
  description?: string
  ruleType: string
  expression?: string
  priority?: number
  enabled?: number
  status: number
  createTime: string
}

/** 规则 DTO */
export interface OntRuleDTO {
  id?: string
  ontologyId: string
  code: string
  name: string
  description?: string
  ruleType: string
  expression?: string
  priority?: number
  enabled?: number
  status?: number
}

/** 动作 VO */
export interface OntActionVO {
  id: string
  ontologyId: string
  code: string
  name: string
  description?: string
  actionType: string
  targetClassId?: string
  targetClassName?: string
  parameters?: string
  preconditions?: string
  enabled?: number
  status: number
  createTime: string
}

/** 动作 DTO */
export interface OntActionDTO {
  id?: string
  ontologyId: string
  code: string
  name: string
  description?: string
  actionType: string
  targetClassId?: string
  parameters?: string
  preconditions?: string
  enabled?: number
  status?: number
}

/** 图谱 VO */
export interface OntologyGraphVO {
  nodes: GraphNode[]
  edges: GraphEdge[]
}
export interface GraphNode {
  id: string
  label: string
  type: string
  icon?: string
  properties?: string[]
  /** 属性明细（编码+名称+类型，供图谱节点展示） */
  propertyDetails?: GraphNodeProperty[]
}
export interface GraphNodeProperty {
  code: string
  name?: string
  dataType?: string
}
export interface GraphEdge {
  id: string
  source: string
  target: string
  label: string
  cardinality?: string
}

/** 词表 VO */
export interface VocabularyVO {
  ontologyId: string
  ontologyName: string
  version: string
  classes: ClassVocab[]
  relations: RelationVocab[]
}
export interface ClassVocab {
  code: string
  name: string
  description?: string
  properties: PropertyVocab[]
}
export interface PropertyVocab {
  code: string
  name: string
  dataType: string
  description?: string
}
export interface RelationVocab {
  code: string
  name: string
  sourceClass: string
  targetClass: string
  cardinality: string
  joinKeys?: string
}
