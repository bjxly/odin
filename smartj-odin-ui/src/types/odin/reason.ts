/** 推理规则 VO */
export interface ReasonRuleVO {
  id: string
  ontologyId: string
  name: string
  code?: string
  description?: string
  ruleType: string
  expression?: string
  conditionJson?: string
  actionJson?: string
  priority?: number
  enabled: number
  status: number
  createTime: string
  updateTime?: string
}

/** 推理规则 DTO */
export interface ReasonRuleDTO {
  id?: string
  ontologyId: string
  name: string
  code?: string
  description?: string
  ruleType: string
  expression?: string
  conditionJson?: string
  actionJson?: string
  priority?: number
  enabled?: number
  status?: number
}

/** 推理规则查询条件 */
export interface ReasonRuleQuery {
  current: number
  size: number
  ontologyId?: string
  ruleType?: string
  name?: string
  enabled?: number
  status?: number
}

/** 推理路径节点 */
export interface PathNode {
  classCode: string
  className: string
  depth: number
  reason?: string
}

/** 推理路径 */
export interface ReasonPath {
  sourceClass: string
  targetClass: string
  nodes: PathNode[]
  hops: number
  relations: string[]
}

/** 推理试运行结果 */
export interface ReasonTraceResult {
  rulesFired: RuleFire[]
  expansions: ExpansionStep[]
  derivedFilters: DerivedFilter[]
  inferredRelations: InferredRelation[]
}

/** 规则触发记录 */
export interface RuleFire {
  ruleId: string
  ruleName: string
  ruleType: string
  conditionResult: boolean
  derived: boolean
  action: string
  detail: string
}

/** 虚拟类展开步骤 */
export interface ExpansionStep {
  from: string
  to: string
  reason: string
}

/** 派生过滤条件 */
export interface DerivedFilter {
  property: string
  op: string
  value?: unknown
  source: string
}

/** 推导关系 */
export interface InferredRelation {
  name: string
  label: string
  fromClass: string
  toClass: string
}

/** 推理解释请求参数 */
export interface ReasonExplainParam {
  ontologyId: string
  classCode: string
  filters?: Array<{ property: string; op: string; value?: unknown }>
}
