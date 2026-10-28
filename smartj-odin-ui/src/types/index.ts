/**
 * 统一 API 响应
 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

/**
 * 分页响应
 */
export interface PageResult<T> {
  records: T[]
  total: number
  current: number
  size: number
}

/**
 * 分页查询参数
 */
export interface PageParams {
  current: number
  size: number
  orders?: string
}

/**
 * 任务信息 — 对应 task_info 表
 */
export interface Task {
  id: number
  taskName: string
  taskType: 'cron' | 'once'
  flowId: number
  cronExpression?: string
  timezone: string
  executeStrategy: 'single' | 'broadcast' | 'shard'
  shardTotal?: number
  shardParam?: string
  namespace: string
  timeout: number
  maxRetry: number
  retryInterval: number
  retryStrategy: 'fixed' | 'exponential'
  status: number // 0-禁用 1-启用
  createBy?: string
  updateBy?: string
  deleted: number
  createTime: string
  updateTime: string
}

/**
 * DAG 流程 — 对应 dag_flow 表
 */
export interface DAGFlow {
  id: number
  flowName: string
  description?: string
  version: number
  namespace: string
  nodes: DAGNode[]
  edges: DAGEdge[]
  variables?: FlowVar[]
  timeout: number
  failStrategy: 'fail_stop' | 'fail_continue'
  envGroupId?: number
  status: number // 0-禁用 1-启用
  createBy?: string
  updateBy?: string
  deleted: number
  createTime: string
  updateTime: string
}

/**
 * DAG 节点
 */
export interface DAGNode {
  nodeId: string
  nodeName: string
  nodeType: string
  config: Record<string, unknown>
  x?: number
  y?: number
  retries?: number
  timeout?: number
}

/**
 * DAG 边
 */
export interface DAGEdge {
  fromNodeId: string
  toNodeId: string
  condition?: string
}

/**
 * 流程变量
 */
export interface FlowVar {
  name: string
  type: string
  defaultValue?: string
  description?: string
}

/**
 * 流程执行实例 — 对应 flow_instance 表
 */
export interface FlowInstance {
  id: number
  instanceId: string
  flowId: number
  flowVersion: number
  triggerType: 'cron' | 'manual' | 'api' | 'subflow'
  status: number // 0-运行中 1-成功 2-失败 3-终止 4-暂停
  startTime?: string
  endTime?: string
  costTime?: number
  inputParams?: Record<string, unknown>
  outputResult?: Record<string, unknown>
  errorMsg?: string
  createTime: string
}

/**
 * 节点执行记录 — 对应 node_instance 表
 */
export interface NodeInstance {
  id: number
  instanceId: string
  nodeId: string
  nodeName: string
  nodeType: string
  workerId?: string
  status: number // 0-等待 1-运行中 2-成功 3-失败 4-跳过 5-终止
  input?: Record<string, unknown>
  output?: Record<string, unknown>
  logSummary?: string
  startTime?: string
  endTime?: string
  costTime?: number
  retryCount: number
  errorMsg?: string
  execId?: string
  createTime: string
}

/**
 * 执行日志 — 对应 execution_log 表
 */
export interface ExecutionLog {
  id: number
  execId: string
  nodeId: string
  level: 'INFO' | 'WARN' | 'ERROR' | 'DEBUG'
  message: string
  logTime: string
  createTime: string
}

/**
 * Worker 注册信息 — 对应 worker_registry 表
 * 后端返回 WorkerDetail 包含 online 字段
 */
export interface Worker {
  id: number
  workerId: string
  ip: string
  port: number
  hostname?: string
  tags?: string[]
  capacity: number
  status: number // 0-离线 1-在线 2-暂停
  online?: boolean
  lastHeartbeat?: string
  registerTime: string
  updateTime: string
}

/**
 * 数据源信息 — 对应 datasource_info 表
 */
export interface Datasource {
  id: number
  name: string
  type: 'mysql' | 'postgresql' | 'sqlserver' | 'dm' | 'kingbase' | 'opengauss'
  host: string
  port: number
  databaseName?: string
  username: string
  password?: string
  params?: string
  status: number // 0-禁用 1-启用
  createTime: string
  updateTime: string
}

/**
 * 环境变量组 — 对应 env_group 表
 */
export interface EnvGroup {
  id: number
  name: string
  description?: string
  status: number // 0-禁用 1-启用
  variableCount?: number // 列表接口返回
  variables?: EnvVariable[] // 详情接口返回
  createTime: string
  updateTime: string
}

/**
 * 环境变量 — 对应 env_variable 表
 */
export interface EnvVariable {
  id?: number
  groupId?: number
  varKey: string
  varValue: string
  isSecret: number // 0-明文 1-加密
  description?: string
}

/**
 * 环境变量组列表查询参数
 */
export interface EnvGroupListParams {
  page?: number
  pageSize?: number
  name?: string
}
